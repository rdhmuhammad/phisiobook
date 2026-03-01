package middleware

type SessionDataUser struct {
	UserReference string `json:"userReference"`
	RoleName      string `json:"roleName"`
	Lang          string `json:"lang"`
	TimeZone      string `json:"timeZone"`
}
