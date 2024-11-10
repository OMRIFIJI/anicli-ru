package api

type NothingFoundError struct {
	Message string
}

func (e *NothingFoundError) Error() string {
	return e.Message
}

type NoConnectionError struct {
	Message string
}

func (e *NoConnectionError) Error() string {
	return e.Message
}
