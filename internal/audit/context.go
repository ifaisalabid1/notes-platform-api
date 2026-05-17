package audit

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const actorIDContextKey contextKey = "actor_id"

func ContextWithActorID(ctx context.Context, actorID uuid.UUID) context.Context {
	return context.WithValue(ctx, actorIDContextKey, actorID)
}

func ActorIDFromContext(ctx context.Context) *uuid.UUID {
	actorID, ok := ctx.Value(actorIDContextKey).(uuid.UUID)
	if !ok {
		return nil
	}

	return &actorID
}
