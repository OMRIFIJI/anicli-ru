package models

type NotFoundError struct {
	Msg string
}

type NotAvailableError struct {
	Msg string
}

func (e *NotFoundError) Error() string     { return e.Msg }
func (e *NotAvailableError) Error() string { return e.Msg }
