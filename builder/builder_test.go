package builder

import (
	"log"
	"testing"
)

func TestBuilder(t *testing.T) {
	Build("gosh://0:b00a7a5a24740e4a7d6487d31969732f1febcaea412df5cc307400818055ad58/at-test/telepresence-build-gosh")

	log.Print("test")
}
