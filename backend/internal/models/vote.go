package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Vote is a single durable vote record. VoterIdentifier is a hash of the
// signal used to prevent duplicate voting (authenticated user id, or a
// hashed IP+UserAgent fingerprint for anonymous visitors) — never raw PII.
type Vote struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID          primitive.ObjectID `bson:"pollId" json:"pollId"`
	OptionID        string             `bson:"optionId" json:"optionId"`
	VoterIdentifier string             `bson:"voterIdentifier" json:"-"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
}

// PollResult is the live, aggregated view of a poll's votes, served from
// the Redis cache and broadcast over WebSocket.
type PollResult struct {
	OptionID string `json:"optionId"`
	Count    int64  `json:"count"`
}

type PollResultsPayload struct {
	Type       string       `json:"type"`
	PollID     string       `json:"pollId"`
	Results    []PollResult `json:"results"`
	TotalVotes int64        `json:"totalVotes"`
}
