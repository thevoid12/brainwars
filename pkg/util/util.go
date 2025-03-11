// contains all misc utilities
package util

import (
	constants "brainwars/constant"
	"context"

	"github.com/google/uuid"
)

// get User from ctx
func GetUserIDFromctx(ctx context.Context) uuid.UUID {
	userID := ctx.Value(constants.CONTEXT_KEY_USER_ID)
	return uuid.MustParse(userID.(string)) // type assersion

}
