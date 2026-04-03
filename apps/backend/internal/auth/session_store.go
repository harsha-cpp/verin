package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type SessionStore struct {
	client *redis.Client
}

type Session struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	OrgID            string    `json:"org_id"`
	Authenticated    bool      `json:"authenticated"`
	PendingMFA       bool      `json:"pending_mfa"`
	CSRFTok          string    `json:"csrf_token"`
	PendingMFASecret string    `json:"pending_mfa_secret"`
	PendingMFASetup  bool      `json:"pending_mfa_setup"`
	ExpiresAt        time.Time `json:"expires_at"`
}

func NewSessionStore(client *redis.Client) *SessionStore {
	return &SessionStore{client: client}
}

func (s *SessionStore) Create(ctx context.Context, session Session, ttl time.Duration) (Session, error) {
	if session.ID == "" {
		session.ID = uuid.NewString()
	}
	session.ExpiresAt = time.Now().Add(ttl)
	if err := s.Save(ctx, session, ttl); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (s *SessionStore) Save(ctx context.Context, session Session, ttl time.Duration) error {
	payload, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	if err := s.client.Set(ctx, key(session.ID), payload, ttl).Err(); err != nil {
		return fmt.Errorf("store session: %w", err)
	}

	return nil
}

func (s *SessionStore) Get(ctx context.Context, sessionID string) (Session, error) {
	value, err := s.client.Get(ctx, key(sessionID)).Result()
	if err != nil {
		return Session{}, err
	}

	var session Session
	if err := json.Unmarshal([]byte(value), &session); err != nil {
		return Session{}, fmt.Errorf("unmarshal session: %w", err)
	}

	return session, nil
}

func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	return s.client.Del(ctx, key(sessionID)).Err()
}

func key(sessionID string) string {
	return "verin:session:" + sessionID
}
