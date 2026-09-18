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

type PollRepository struct {
	col *mongo.Collection
}

func NewPollRepository(db *mongo.Database) *PollRepository {
	repo := &PollRepository{col: db.Collection("polls")}
	_, _ = repo.col.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "publicId", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "creatorId", Value: 1}}},
	})
	return repo
}

func (r *PollRepository) Create(ctx context.Context, p *models.Poll) error {
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	res, err := r.col.InsertOne(ctx, p)
	if err != nil {
		return err
	}
	p.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *PollRepository) FindByPublicID(ctx context.Context, publicID string) (*models.Poll, error) {
	var p models.Poll
	if err := r.col.FindOne(ctx, bson.M{"publicId": publicID}).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PollRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Poll, error) {
	var p models.Poll
	if err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PollRepository) ListByCreator(ctx context.Context, creatorID primitive.ObjectID) ([]models.Poll, error) {
	cur, err := r.col.Find(ctx, bson.M{"creatorId": creatorID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var polls []models.Poll
	if err := cur.All(ctx, &polls); err != nil {
		return nil, err
	}
	return polls, nil
}

func (r *PollRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.PollStatus) error {
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": status, "updatedAt": time.Now()}})
	return err
}

func (r *PollRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updatedAt"] = time.Now()
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (r *PollRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
