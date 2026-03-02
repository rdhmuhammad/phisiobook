package security

import (
	"github.com/golang-jwt/jwt/v4"
)

var (
	CtxKeySession = "auth.middleware"
	KeyBranchID   = "branchId"

	// Session Check Middleware
	RequestParams   = "request-params"
	RequestQuery    = "request-query"
	RequestBodyJSON = "request-body-json"

	// Role Name
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type SingleTokenClaim struct {
	UserData
	jwt.RegisteredClaims
}

type UserData struct {
	UserId   string `json:"userId"`
	Lang     string `json:"lang"`
	Timezone string `json:"timezone"`
	Email    string `json:"email"`
	RoleName string `json:"roleName"`
}
