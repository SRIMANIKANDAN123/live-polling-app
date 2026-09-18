package repositories

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"live-polling-app/backend/internal/models"
)

type VoteRepository struct {
	col *mongo.Collection
}

func NewVoteRepository(db *mongo.Database) *VoteRepository {
	repo := &VoteRepository{col: db.Collection("votes")}
	_, _ = repo.col.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "pollId", Value: 1}}},
		// Compound unique index: this is the durable, database-level
		// guarantee against duplicate votes (Redis SETNX is the fast
		// pre-check, this is the fallback of record).
		{
			Keys:    bson.D{{Key: "pollId", Value: 1}, {Key: "voterIdentifier", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	return repo
}

func (r *VoteRepository) Create(ctx context.Context, v *models.Vote) error {
	v.CreatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, v)
	return err
}

// CountsByPoll aggregates persisted votes per option — used to seed the
// Redis cache when it's cold (e.g. after a restart).
func (r *VoteRepository) CountsByPoll(ctx context.Context, pollID primitive.ObjectID) (map[string]int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"pollId": pollID}}},
		{{Key: "$group", Value: bson.M{"_id": "$optionId", "count": bson.M{"$sum": 1}}}},
	}
	cur, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make(map[string]int64)
	var results []struct {
		ID    string `bson:"_id"`
		Count int64  `bson:"count"`
	}
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}
	for _, r := range results {
		out[r.ID] = r.Count
	}
	return out, nil
}

func (r *VoteRepository) HasVoted(ctx context.Context, pollID primitive.ObjectID, voterIdentifier string) (bool, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{"pollId": pollID, "voterIdentifier": voterIdentifier})
	return count > 0, err
}

func (r *VoteRepository) TotalVotes(ctx context.Context, pollID primitive.ObjectID) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{"pollId": pollID})
}
