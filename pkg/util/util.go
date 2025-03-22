// contains all misc utilities
package util

import (
	constants "brainwars/constant"
	usermodel "brainwars/pkg/users/model"
	"context"
)

// get User from ctx
func GetUserInfoFromctx(ctx context.Context) *usermodel.UserInfo {
	userInfo := ctx.Value(constants.CONTEXT_KEY_USER)
	return userInfo.(*usermodel.UserInfo) // type assersion
}

func SetUserInfoInctx(ctx context.Context, userInfo *usermodel.UserInfo) context.Context {

	ctx = context.WithValue(ctx, constants.CONTEXT_KEY_USER, userInfo)
	return ctx
}
