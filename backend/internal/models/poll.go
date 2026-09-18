package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PollOption is one selectable choice on a poll.
type PollOption struct {
	ID   string `bson:"id" json:"id"`
	Text string `bson:"text" json:"text"`
}

// PollStatus enumerates the lifecycle of a poll.
type PollStatus string

const (
	PollStatusActive PollStatus = "active"
	PollStatusClosed PollStatus = "closed"
)

// Poll is the persistent, source-of-truth record for a poll. Live vote
// counts are cached in Redis for speed, but MongoDB always holds the
// durable question/options/status data.
type Poll struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PublicID        string             `bson:"publicId" json:"publicId"`
	CreatorID       primitive.ObjectID `bson:"creatorId" json:"creatorId"`
	Question        string             `bson:"question" json:"question"`
	Description     string             `bson:"description" json:"description"`
	Options         []PollOption       `bson:"options" json:"options"`
	Status          PollStatus         `bson:"status" json:"status"`
	AnonymousVoting bool               `bson:"anonymousVoting" json:"anonymousVoting"`
	ExpiresAt       *time.Time         `bson:"expiresAt" json:"expiresAt"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
}
