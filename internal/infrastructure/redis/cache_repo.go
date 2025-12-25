package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CacheRepository provides caching functionality
type CacheRepository struct {
	client *Client
}

// NewCacheRepository creates a new cache repository
func NewCacheRepository(client *Client) *CacheRepository {
	return &CacheRepository{client: client}
}

// Get retrieves a value from cache
func (r *CacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Set stores a value in cache with expiration
func (r *CacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.Set(ctx, key, data, expiration)
}

// Delete removes a key from cache
func (r *CacheRepository) Delete(ctx context.Context, keys ...string) error {
	return r.client.Delete(ctx, keys...)
}

// Exists checks if a key exists in cache
func (r *CacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key)
	return count > 0, err
}

// Cache keys
const (
	CacheKeyUser        = "user:%s"
	CacheKeyTODO        = "todo:%s"
	CacheKeyTeam        = "team:%s"
	CacheKeyTeamMembers = "team:%s:members"
	CacheKeyUserSession = "session:%s"
	CacheKeyTODOList    = "todos:user:%s:filter:%s"
)

// GenerateUserCacheKey generates a cache key for a user
func GenerateUserCacheKey(userID string) string {
	return fmt.Sprintf(CacheKeyUser, userID)
}

// GenerateTODOCacheKey generates a cache key for a TODO
func GenerateTODOCacheKey(todoID string) string {
	return fmt.Sprintf(CacheKeyTODO, todoID)
}

// GenerateTeamCacheKey generates a cache key for a team
func GenerateTeamCacheKey(teamID string) string {
	return fmt.Sprintf(CacheKeyTeam, teamID)
}

// GenerateSessionCacheKey generates a cache key for a session
func GenerateSessionCacheKey(sessionID string) string {
	return fmt.Sprintf(CacheKeyUserSession, sessionID)
}
