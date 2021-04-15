package receiver

import "encoding/json"

type SourceError struct {
	Message string `json:"message"`
}

func (e *SourceError) Error() string {
	err, _ := json.Marshal(e)
	return string(err)
}

func NewSourceError(message string) *SourceError {
	return &SourceError{Message: message}
}
