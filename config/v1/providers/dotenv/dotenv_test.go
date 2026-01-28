package dotenv_test

import (
	"errors"
	"os"
	"testing"

	"github.com/raf555/salome/config/v1/providers/dotenv"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Test with a valid .env file
	testFile := "test.env"
	os.WriteFile(testFile, []byte("KEY=value\n"), 0644)
	defer os.Remove(testFile) // Clean up after test

	provider, err := dotenv.New(testFile)
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	// Test with a non-existent .env file
	_, err = dotenv.New("non_existent.env")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, dotenv.ErrDotenvFileNotFound))
}

func TestFetchConfig(t *testing.T) {
	// Create a temporary .env file
	testFile := "test.env"
	os.WriteFile(testFile, []byte("KEY=value\n"), 0644)
	defer os.Remove(testFile) // Clean up after test

	provider, _ := dotenv.New(testFile)

	configs, err := provider.FetchConfig(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, "value", configs["KEY"])

	os.WriteFile(testFile, []byte("KEY=value2\n"), 0644)

	configs, err = provider.FetchConfig(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, "value2", configs["KEY"])
}

func TestReadConfig(t *testing.T) {
	// Create a temporary .env file
	testFile := "test.env"
	os.WriteFile(testFile, []byte("KEY=value\n"), 0644)
	defer os.Remove(testFile) // Clean up after test

	provider, _ := dotenv.New(testFile)

	configs, err := provider.Config(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, "value", configs["KEY"])
}
