package bouteitech

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSigner(t *testing.T) {
	res, err := NewSigner(os.Getenv(envCACert), os.Getenv(envCAKey))
	assert.NoError(t, err)
	log.Println(res)
}
