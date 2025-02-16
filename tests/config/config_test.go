package config_test

import (
	"Shop/config"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadEnv(t *testing.T) {
	err := os.WriteFile(".env", []byte("TEST_VAR=12345"), 0644)
	assert.NoError(t, err)

	config.LoadEnv()

	testVar := os.Getenv("TEST_VAR")
	assert.Equal(t, "12345", testVar)

	err = os.Remove(".env")
	assert.NoError(t, err)
}

