package chat_io

type Transporter struct {
	ToID    string `json:"to_id"`
	Message string `json:"message"`
	RoomID  string `json:"room_id"`
	Online  bool   `json:"online"`
	MsgType string `json:"MsgType"` // type = message, status
}
