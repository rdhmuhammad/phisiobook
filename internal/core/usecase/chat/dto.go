package chat

type CreateSessionResponse struct {
	SessionId string `json:"sessionId"`
	RoomId    string `json:"roomId"`
}
