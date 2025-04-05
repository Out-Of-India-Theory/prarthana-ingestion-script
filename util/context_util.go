package util

import "context"

const accessTokenKey = "access-token"

func GetZohoAccessTokenFromContext(ctx context.Context) string {
	lang, ok := ctx.Value(accessTokenKey).(string)
	if ok {
		return lang
	}
	return ""
}

func SetZohoAccessTokenInContext(ctx context.Context, accessToken string) context.Context {
	ctx = context.WithValue(ctx, accessTokenKey, accessToken)
	return ctx
}
