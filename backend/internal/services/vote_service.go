package services

import (
	"context"
	"encoding/json"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"

	"live-polling-app/backend/internal/models"
	"live-polling-app/backend/internal/redisclient"
	"live-polling-app/backend/internal/repositories"
)

var (
	ErrPollClosed    = errors.New("this poll is closed")
	ErrInvalidOption = errors.New("selected option does not exist on this poll")
	ErrAlreadyVoted  = errors.New("you have already voted on this poll")
)

type VoteService struct {
	polls *repositories.PollRepository
	votes *repositories.VoteRepository
	redis *redisclient.Client
}

func NewVoteService(polls *repositories.PollRepository, votes *repositories.VoteRepository, redis *redisclient.Client) *VoteService {
	return &VoteService{polls: polls, votes: votes, redis: redis}
}

// CastVote runs the full pipeline described in the spec:
//
//	vote request -> validate poll/option -> dedupe check -> persist to
//	MongoDB -> increment Redis counter -> publish to Redis channel ->
//	WebSocket hub picks it up and broadcasts to every connected client.
func (s *VoteService) CastVote(ctx context.Context, publicID, optionID, voterIdentifier string) (*models.PollResultsPayload, error) {
	poll, err := s.polls.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, ErrPollNotFound
	}
	if poll.Status != models.PollStatusActive {
		return nil, ErrPollClosed
	}

	validOption := false
	for _, o := range poll.Options {
		if o.ID == optionID {
			validOption = true
			break
		}
	}
	if !validOption {
		return nil, ErrInvalidOption
	}

	// Fast duplicate-vote pre-check via Redis SETNX (cheap, catches the
	// common case immediately). The MongoDB unique compound index on
	// (pollId, voterIdentifier) below is the durable guarantee in case
	// Redis is unavailable or two requests race.
	firstTime, _ := s.redis.CheckAndSetVoted(ctx, publicID, voterIdentifier)
	if !firstTime {
		return nil, ErrAlreadyVoted
	}

	vote := &models.Vote{
		PollID:          poll.ID,
		OptionID:        optionID,
		VoterIdentifier: voterIdentifier,
	}
	if err := s.votes.Create(ctx, vote); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrAlreadyVoted
		}
		return nil, err
	}

	counts, err := s.redis.IncrVote(ctx, publicID, optionID)
	if err != nil {
		return nil, err
	}

	payload := buildResultsPayload(publicID, poll, counts)

	raw, err := json.Marshal(payload)
	if err == nil {
		_ = s.redis.Publish(ctx, publicID, raw) // best-effort; polling clients can still re-fetch /results
	}

	return payload, nil
}

// GetResults returns the current live tally for a poll, seeding the Redis
// cache from MongoDB on a cold cache (e.g. right after a server restart).
func (s *VoteService) GetResults(ctx context.Context, publicID string) (*models.PollResultsPayload, error) {
	poll, err := s.polls.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, ErrPollNotFound
	}

	counts, err := s.redis.GetCounts(ctx, publicID)
	if err != nil {
		return nil, err
	}
	if len(counts) == 0 {
		if dbCounts, err := s.votes.CountsByPoll(ctx, poll.ID); err == nil && len(dbCounts) > 0 {
			_ = s.redis.SeedCounts(ctx, publicID, dbCounts)
			counts = dbCounts
		}
	}

	return buildResultsPayload(publicID, poll, counts), nil
}

func buildResultsPayload(publicID string, poll *models.Poll, counts map[string]int64) *models.PollResultsPayload {
	results := make([]models.PollResult, 0, len(poll.Options))
	var total int64
	for _, o := range poll.Options {
		c := counts[o.ID]
		total += c
		results = append(results, models.PollResult{OptionID: o.ID, Count: c})
	}
	return &models.PollResultsPayload{
		Type:       "vote_update",
		PollID:     publicID,
		Results:    results,
		TotalVotes: total,
	}
}
