package constant

import "strings"

const (
	ContextDashboard = "DASHBOARD"
	ContextMobile    = "MOBILE"
	ContextTherapist = "TERAPIS"

	CacheKeyOTP   = "KEY_OTP_"
	CacheKeyLogin = "USER_LOGIN_"

	RolesIsMobile = "USER"

	RoleIsAdmin = "ADMIN"
	RoleIsUser  = "USER"
)

func LoginCacheKey(userReference string) string {
	if strings.HasPrefix(userReference, CacheKeyLogin) {
		return userReference
	}

	return CacheKeyLogin + userReference
}
