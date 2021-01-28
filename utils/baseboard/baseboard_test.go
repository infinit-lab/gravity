package baseboard

import (
	"log"
	"testing"
)

func TestGetBaseBoardUUID(t *testing.T) {
	uuid, err := GetUUID()
	if err != nil {
		t.Fatal("Failed to GetBaseBoardUUID. error: ", err)
	}
	log.Println(uuid)
}
