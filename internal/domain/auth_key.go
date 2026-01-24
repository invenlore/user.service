package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthKeyStatus string

const (
	AuthKeyStatusActive   AuthKeyStatus = "active"
	AuthKeyStatusRetiring AuthKeyStatus = "retiring"
	AuthKeyStatusRevoked  AuthKeyStatus = "revoked"
)

type AuthKey struct {
	Id            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Kid           string             `bson:"kid" json:"kid"`
	Alg           string             `bson:"alg" json:"alg"`
	Use           string             `bson:"use" json:"use"`
	PrivateKeyPEM string             `bson:"private_key_pem" json:"-"`
	PublicKeyPEM  string             `bson:"public_key_pem" json:"-"`
	Status        AuthKeyStatus      `bson:"status" json:"status"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	RotatedAt     *time.Time         `bson:"rotated_at,omitempty" json:"rotated_at,omitempty"`
}
