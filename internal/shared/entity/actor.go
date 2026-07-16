package entity

import "github.com/golang-jwt/jwt/v5"

// ActorType represents who is making the request.
type ActorType string

const (
	ActorTypeAdmin    ActorType = "admin"
	ActorTypeEmployee ActorType = "employee"
)

// ActorRole represents the role of the actor within a store.
type ActorRole string

const (
	ActorRoleOwner    ActorRole = "owner"
	ActorRoleEmployee ActorRole = "employee"
)

// ActorClaims is the JWT payload structure shared across all authenticated endpoints.
// The auth middleware extracts and validates this from the Bearer token,
// then stores it in the Gin context for downstream handlers and usecases.
//
// All data queries must be scoped to StoreID — never query without it.
type ActorClaims struct {
	ActorID   string    `json:"actor_id"`
	ActorType ActorType `json:"actor_type"`
	StoreID   string    `json:"store_id"`
	Role      ActorRole `json:"role"`
	jwt.RegisteredClaims
}

// ContextKey is the type used for storing values in context to avoid collisions.
type ContextKey string

const (
	// ContextKeyActor is the key used to store *ActorClaims in Gin context.
	ContextKeyActor ContextKey = "actor"
)
