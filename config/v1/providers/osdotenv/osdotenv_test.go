package osdotenv_test

import (
	"os"
	"testing"

	"github.com/raf555/salome/config/v1/providers/osdotenv"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Test with a valid .env file
	testFile := "test_osdotenv.env"
	os.WriteFile(testFile, []byte("DOTENV_KEY=dotenv_value\n"), 0644)
	defer os.Remove(testFile)

	provider, err := osdotenv.New(testFile)
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	// Test with a non-existent .env file (should not error, just use OS env)
	provider, err = osdotenv.New("non_existent_osdotenv.env")
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestConfig(t *testing.T) {
	// Set OS environment variable
	os.Setenv("OS_TEST_KEY", "os_value")
	defer os.Unsetenv("OS_TEST_KEY")

	testFile := "test_config_osdotenv.env"
	os.WriteFile(testFile, []byte("DOTENV_TEST_KEY=dotenv_value\n"), 0644)
	defer os.Remove(testFile)

	provider, _ := osdotenv.New(testFile)

	config, err := provider.Config(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Config should contain both OS and .env values
	assert.Equal(t, "os_value", config["OS_TEST_KEY"])
	assert.Equal(t, "dotenv_value", config["DOTENV_TEST_KEY"])

	// Verify that Config returns a clone
	config["OS_TEST_KEY"] = "modified"
	config2, _ := provider.Config(t.Context())
	assert.Equal(t, "os_value", config2["OS_TEST_KEY"])
}

func TestFetchConfig(t *testing.T) {
	// Set OS environment variables
	os.Setenv("OS_FETCH_KEY_1", "os_fetch_value_1")
	os.Setenv("OS_FETCH_KEY_2", "os_fetch_value_2")
	defer os.Unsetenv("OS_FETCH_KEY_1")
	defer os.Unsetenv("OS_FETCH_KEY_2")

	testFile := "test_fetch_osdotenv.env"
	os.WriteFile(testFile, []byte("DOTENV_FETCH_KEY=dotenv_fetch_value\n"), 0644)
	defer os.Remove(testFile)

	provider, _ := osdotenv.New(testFile)

	configs, err := provider.FetchConfig(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, configs)

	// Verify OS environment variables
	assert.Equal(t, "os_fetch_value_1", configs["OS_FETCH_KEY_1"])
	assert.Equal(t, "os_fetch_value_2", configs["OS_FETCH_KEY_2"])

	// Verify .env file variables
	assert.Equal(t, "dotenv_fetch_value", configs["DOTENV_FETCH_KEY"])
}

func TestDotenvPrecedence(t *testing.T) {
	// Test that .env values take precedence over OS environment variables
	os.Setenv("PRECEDENCE_KEY", "os_value")
	defer os.Unsetenv("PRECEDENCE_KEY")

	testFile := "test_precedence_osdotenv.env"
	os.WriteFile(testFile, []byte("PRECEDENCE_KEY=dotenv_value\n"), 0644)
	defer os.Remove(testFile)

	provider, _ := osdotenv.New(testFile)

	config, _ := provider.Config(t.Context())
	// .env file should take precedence
	assert.Equal(t, "dotenv_value", config["PRECEDENCE_KEY"])
}

func TestFetchConfigWithoutDotenvFile(t *testing.T) {
	// Test FetchConfig when .env file doesn't exist
	os.Setenv("OS_ONLY_KEY", "os_only_value")
	defer os.Unsetenv("OS_ONLY_KEY")

	provider, _ := osdotenv.New("non_existent_file_osdotenv.env")

	configs, err := provider.FetchConfig(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, configs)
	assert.Equal(t, "os_only_value", configs["OS_ONLY_KEY"])
}

func TestMultipleValuesFromBoth(t *testing.T) {
	// Test with multiple values from both OS and .env
	os.Setenv("OS_VAR_1", "os_value_1")
	os.Setenv("OS_VAR_2", "os_value_2")
	os.Setenv("SHARED_KEY", "os_shared")
	defer os.Unsetenv("OS_VAR_1")
	defer os.Unsetenv("OS_VAR_2")
	defer os.Unsetenv("SHARED_KEY")

	testFile := "test_multiple_osdotenv.env"
	dotenvContent := "DOTENV_VAR_1=dotenv_value_1\nDOTENV_VAR_2=dotenv_value_2\nSHARED_KEY=dotenv_shared\n"
	os.WriteFile(testFile, []byte(dotenvContent), 0644)
	defer os.Remove(testFile)

	provider, _ := osdotenv.New(testFile)

	config, _ := provider.Config(t.Context())

	// Verify OS variables
	assert.Equal(t, "os_value_1", config["OS_VAR_1"])
	assert.Equal(t, "os_value_2", config["OS_VAR_2"])

	// Verify .env variables
	assert.Equal(t, "dotenv_value_1", config["DOTENV_VAR_1"])
	assert.Equal(t, "dotenv_value_2", config["DOTENV_VAR_2"])

	// Verify .env takes precedence
	assert.Equal(t, "dotenv_shared", config["SHARED_KEY"])
}

func TestEmptyValues(t *testing.T) {
	// Test handling of empty values
	os.Setenv("EMPTY_OS_KEY", "")
	defer os.Unsetenv("EMPTY_OS_KEY")

	testFile := "test_empty_osdotenv.env"
	os.WriteFile(testFile, []byte("EMPTY_DOTENV_KEY=\n"), 0644)
	defer os.Remove(testFile)

	provider, _ := osdotenv.New(testFile)

	config, _ := provider.Config(t.Context())

	assert.Equal(t, "", config["EMPTY_OS_KEY"])
	assert.Equal(t, "", config["EMPTY_DOTENV_KEY"])
}
