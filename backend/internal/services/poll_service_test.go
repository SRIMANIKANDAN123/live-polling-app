package services

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// These tests exercise validation logic in PollService.Create that does
// not require a live MongoDB connection: they hit the (nil-repo-safe)
// validation branches before any database interaction. Full integration
// tests against a real MongoDB instance are recommended in CI (see
// README "Testing" section) using a test container or mongodb-memory-server
// equivalent for Go (e.g. testcontainers-go).

func TestCreatePollRejectsTooFewOptions(t *testing.T) {
	s := NewPollService(nil)
	_, err := s.Create(context.Background(), primitive.NewObjectID(), CreatePollInput{
		Question: "Pick one",
		Options:  []string{"only one"},
	})
	if err != ErrTooFewOptions {
		t.Fatalf("expected ErrTooFewOptions, got %v", err)
	}
}

func TestCreatePollRejectsTooManyOptions(t *testing.T) {
	s := NewPollService(nil)
	opts := make([]string, 11)
	for i := range opts {
		opts[i] = "option"
	}
	// Give them unique text so duplicate-detection isn't what trips first.
	for i := range opts {
		opts[i] = opts[i] + string(rune('a'+i))
	}
	_, err := s.Create(context.Background(), primitive.NewObjectID(), CreatePollInput{
		Question: "Pick one",
		Options:  opts,
	})
	if err != ErrTooManyOptions {
		t.Fatalf("expected ErrTooManyOptions, got %v", err)
	}
}

func TestCreatePollRejectsEmptyQuestion(t *testing.T) {
	s := NewPollService(nil)
	_, err := s.Create(context.Background(), primitive.NewObjectID(), CreatePollInput{
		Question: "   ",
		Options:  []string{"a", "b"},
	})
	if err != ErrInvalidQuestion {
		t.Fatalf("expected ErrInvalidQuestion, got %v", err)
	}
}

func TestCreatePollRejectsDuplicateOptions(t *testing.T) {
	s := NewPollService(nil)
	_, err := s.Create(context.Background(), primitive.NewObjectID(), CreatePollInput{
		Question: "Pick one",
		Options:  []string{"Go", "go"}, // case-insensitive duplicate
	})
	if err != ErrDuplicateOption {
		t.Fatalf("expected ErrDuplicateOption, got %v", err)
	}
}
