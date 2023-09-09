package privacy

type apiServiceError struct {
	msg string `json:"message"`
}

const defaultPrivacyErrorCode = -32800

func (e *apiServiceError) Error() string  { return e.msg }
func (e *apiServiceError) ErrorCode() int { return -32801 }

func NewApiServiceError(msg string) *apiServiceError {
	return &apiServiceError{msg: msg}
}
