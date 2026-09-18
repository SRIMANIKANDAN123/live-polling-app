package redisclient

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// Client wraps a redis.Client with the specific operations this app needs:
// atomic live vote counters and pub/sub for realtime fan-out.
type Client struct {
	rdb *redis.Client
}

func New(redisURL string) (*Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(opt)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func countsKey(pollID string) string {
	return fmt.Sprintf("poll:%s:counts", pollID)
}

func channelName(pollID string) string {
	return fmt.Sprintf("poll:%s:updates", pollID)
}

// IncrVote atomically increments the live count for an option and returns
// the full updated count map for the poll. This is the fast path read by
// GET /results and by the WebSocket broadcast — MongoDB remains the
// durable source of truth, Redis is just the hot cache.
func (c *Client) IncrVote(ctx context.Context, pollID, optionID string) (map[string]int64, error) {
	if err := c.rdb.HIncrBy(ctx, countsKey(pollID), optionID, 1).Err(); err != nil {
		return nil, err
	}
	return c.GetCounts(ctx, pollID)
}

// GetCounts reads the current live counts for a poll from the Redis hash.
func (c *Client) GetCounts(ctx context.Context, pollID string) (map[string]int64, error) {
	raw, err := c.rdb.HGetAll(ctx, countsKey(pollID)).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(raw))
	for k, v := range raw {
		var n int64
		fmt.Sscanf(v, "%d", &n)
		out[k] = n
	}
	return out, nil
}

// SeedCounts primes the Redis hash from MongoDB (e.g. on server restart
// with an empty cache, or first read of a poll with existing votes).
func (c *Client) SeedCounts(ctx context.Context, pollID string, counts map[string]int64) error {
	if len(counts) == 0 {
		return nil
	}
	fields := make(map[string]interface{}, len(counts))
	for k, v := range counts {
		fields[k] = v
	}
	return c.rdb.HSet(ctx, countsKey(pollID), fields).Err()
}

// Publish pushes a JSON-encoded update onto the poll's pub/sub channel.
func (c *Client) Publish(ctx context.Context, pollID string, payload []byte) error {
	return c.rdb.Publish(ctx, channelName(pollID), payload).Err()
}

// Subscribe returns a pub/sub subscription for a single poll's channel.
func (c *Client) Subscribe(ctx context.Context, pollID string) *redis.PubSub {
	return c.rdb.Subscribe(ctx, channelName(pollID))
}

// CheckAndSetVoted uses SETNX with no expiry to atomically record that a
// voter has already voted on a poll, preventing duplicate votes even under
// concurrent requests. Returns true if this call newly recorded the vote
// (i.e. the voter had not voted before).
func (c *Client) CheckAndSetVoted(ctx context.Context, pollID, voterHash string) (bool, error) {
	key := fmt.Sprintf("poll:%s:voted:%s", pollID, voterHash)
	ok, err := c.rdb.SetNX(ctx, key, "1", 0).Result()
	if err != nil {
		log.Printf("redis SETNX error (falling back to Mongo-only duplicate check): %v", err)
		return true, err
	}
	return ok, nil
}

// IncrViewers/DecrViewers track how many clients currently have a poll's
// results page open, for the "N people viewing" live counter.
func (c *Client) IncrViewers(ctx context.Context, pollID string) (int64, error) {
	return c.rdb.Incr(ctx, fmt.Sprintf("poll:%s:viewers", pollID)).Result()
}

func (c *Client) DecrViewers(ctx context.Context, pollID string) (int64, error) {
	n, err := c.rdb.Decr(ctx, fmt.Sprintf("poll:%s:viewers", pollID)).Result()
	if n < 0 {
		c.rdb.Set(ctx, fmt.Sprintf("poll:%s:viewers", pollID), 0, 0)
		return 0, err
	}
	return n, err
}
