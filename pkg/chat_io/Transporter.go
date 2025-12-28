package chat_io

type Transporter struct {
	FromID  string `json:"from_id"`
	ToID    string `json:"to_id"`
	Message string `json:"message"`
}
