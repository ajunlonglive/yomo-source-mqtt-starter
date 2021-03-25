package comm

import (
	"log"
	"time"

	"github.com/google/uuid"
)

const (
	// ACCEPT_MIN_SLEEP is the minimum acceptable sleep times on temporary errors.
	ACCEPT_MIN_SLEEP = 100 * time.Millisecond
	// ACCEPT_MAX_SLEEP is the maximum acceptable sleep times on temporary errors
	ACCEPT_MAX_SLEEP = 10 * time.Second
)

func GenUniqueId() string {
	id, err := uuid.NewRandom()
	if err != nil {
		log.Printf("uuid.NewRandom() returned an error: %v" + err.Error())
	}
	return id.String()
}
