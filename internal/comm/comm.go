package comm

import (
	"log"

	"github.com/google/uuid"
)

func GenUniqueId() string {
	var (
		id  uuid.UUID
		err error
	)

	for {
		id, err = uuid.NewRandom()
		if err != nil {
			log.Printf("uuid.NewRandom() returned an error: %v" + err.Error())
			continue
		}
		break
	}

	return id.String()
}
