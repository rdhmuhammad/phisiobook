package localerror

type InvalidDataError struct {
	Msg string
}

func (e InvalidDataError) Error() string {
	return e.Msg
}

type AccessControlError struct {
	Msg string
}

func (e AccessControlError) Error() string {
	return e.Msg
}
