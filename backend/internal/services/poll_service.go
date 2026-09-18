package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"live-polling-app/backend/internal/models"
	"live-polling-app/backend/internal/repositories"
)

var (
	ErrInvalidQuestion = errors.New("question is required (max 300 characters)")
	ErrTooFewOptions   = errors.New("a poll needs at least 2 options")
	ErrTooManyOptions  = errors.New("a poll can have at most 10 options")
	ErrDuplicateOption = errors.New("options must be unique")
	ErrOptionTooLong   = errors.New("each option must be 1-120 characters")
	ErrPollNotFound    = errors.New("poll not found")
	ErrNotPollOwner    = errors.New("you do not own this poll")
)

type CreatePollInput struct {
	Question        string
	Description     string
	Options         []string
	ExpiresAt       *time.Time
	AnonymousVoting bool
}

type PollService struct {
	polls *repositories.PollRepository
}

func NewPollService(polls *repositories.PollRepository) *PollService {
	return &PollService{polls: polls}
}

func generatePublicID() (string, error) {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *PollService) Create(ctx context.Context, creatorID primitive.ObjectID, in CreatePollInput) (*models.Poll, error) {
	q := strings.TrimSpace(in.Question)
	if q == "" || len(q) > 300 {
		return nil, ErrInvalidQuestion
	}
	if len(in.Options) < 2 {
		return nil, ErrTooFewOptions
	}
	if len(in.Options) > 10 {
		return nil, ErrTooManyOptions
	}

	seen := make(map[string]bool)
	options := make([]models.PollOption, 0, len(in.Options))
	for i, raw := range in.Options {
		opt := strings.TrimSpace(raw)
		if opt == "" || len(opt) > 120 {
			return nil, ErrOptionTooLong
		}
		key := strings.ToLower(opt)
		if seen[key] {
			return nil, ErrDuplicateOption
		}
		seen[key] = true
		options = append(options, models.PollOption{ID: "opt-" + itoa(i+1), Text: opt})
	}

	publicID, err := generatePublicID()
	if err != nil {
		return nil, err
	}

	poll := &models.Poll{
		PublicID:        publicID,
		CreatorID:       creatorID,
		Question:        q,
		Description:     strings.TrimSpace(in.Description),
		Options:         options,
		Status:          models.PollStatusActive,
		AnonymousVoting: in.AnonymousVoting,
		ExpiresAt:       in.ExpiresAt,
	}
	if err := s.polls.Create(ctx, poll); err != nil {
		return nil, err
	}
	return poll, nil
}

func (s *PollService) GetByPublicID(ctx context.Context, publicID string) (*models.Poll, error) {
	poll, err := s.polls.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, ErrPollNotFound
	}
	// Lazily expire polls whose expiry has passed.
	if poll.Status == models.PollStatusActive && poll.ExpiresAt != nil && time.Now().After(*poll.ExpiresAt) {
		_ = s.polls.UpdateStatus(ctx, poll.ID, models.PollStatusClosed)
		poll.Status = models.PollStatusClosed
	}
	return poll, nil
}

func (s *PollService) ListMine(ctx context.Context, creatorID primitive.ObjectID) ([]models.Poll, error) {
	return s.polls.ListByCreator(ctx, creatorID)
}

func (s *PollService) Close(ctx context.Context, publicID string, ownerID primitive.ObjectID) error {
	poll, err := s.polls.FindByPublicID(ctx, publicID)
	if err != nil {
		return ErrPollNotFound
	}
	if poll.CreatorID != ownerID {
		return ErrNotPollOwner
	}
	return s.polls.UpdateStatus(ctx, poll.ID, models.PollStatusClosed)
}

func (s *PollService) Delete(ctx context.Context, publicID string, ownerID primitive.ObjectID) error {
	poll, err := s.polls.FindByPublicID(ctx, publicID)
	if err != nil {
		return ErrPollNotFound
	}
	if poll.CreatorID != ownerID {
		return ErrNotPollOwner
	}
	return s.polls.Delete(ctx, poll.ID)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
