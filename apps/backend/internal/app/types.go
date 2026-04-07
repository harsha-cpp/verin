package app

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type contextKey string

const AuthContextKey contextKey = "auth-context"

var ErrUnauthorized = errors.New("unauthorized")

type AuthContext struct {
	UserID       string
	OrgID        string
	Email        string
	FullName     string
	AvatarURL    string
	Roles        []string
	RoleIDs      []pgtype.UUID
	Permissions  []string
	IsAdmin      bool
	CSRFToken    string
	Authenticated bool
}

func ToPGUUID(value string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return pgtype.UUID{}, err
	}

	var result pgtype.UUID
	if err := result.Scan(parsed.String()); err != nil {
		return pgtype.UUID{}, err
	}
	return result, nil
}

func MustPGUUID(value string) pgtype.UUID {
	result, err := ToPGUUID(value)
	if err != nil {
		panic(err)
	}
	return result
}

func UUIDString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	parsed, err := uuid.FromBytes(value.Bytes[:])
	if err != nil {
		return ""
	}
	return parsed.String()
}

func JSONBytes(value any) []byte {
	encoded, _ := json.Marshal(value)
	return encoded
}

func AuthFromContext(ctx context.Context) (AuthContext, bool) {
	value, ok := ctx.Value(AuthContextKey).(AuthContext)
	return value, ok
}
