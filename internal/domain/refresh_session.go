package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefreshSession struct {
	Id               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SessionID        string             `bson:"session_id" json:"session_id"`
	UserID           primitive.ObjectID `bson:"user_id" json:"user_id"`
	RefreshTokenHash string             `bson:"refresh_token_hash" json:"-"`
	UserAgent        string             `bson:"user_agent" json:"user_agent"`
	IPAddress        string             `bson:"ip_address" json:"ip_address"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	ExpiresAt        time.Time          `bson:"expires_at" json:"expires_at"`
	RevokedAt        *time.Time         `bson:"revoked_at,omitempty" json:"revoked_at,omitempty"`
}
