package os_test

import (
	"os"
	"testing"

	osprov "github.com/raf555/salome/config/v1/providers/os"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Set a test environment variable
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	provider := osprov.New()
	assert.NotNil(t, provider)

	// Verify that initial config was loaded
	config, err := provider.Config(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, "test_value", config["TEST_KEY"])
}

func TestConfig(t *testing.T) {
	os.Setenv("TEST_CONFIG_KEY", "test_config_value")
	defer os.Unsetenv("TEST_CONFIG_KEY")

	provider := osprov.New()

	config, err := provider.Config(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test_config_value", config["TEST_CONFIG_KEY"])

	// Verify that Config returns a clone, not the original
	config["TEST_CONFIG_KEY"] = "modified"
	config2, _ := provider.Config(t.Context())
	assert.Equal(t, "test_config_value", config2["TEST_CONFIG_KEY"])
}

func TestFetchConfig(t *testing.T) {
	// Set multiple test environment variables
	os.Setenv("FETCH_KEY_1", "fetch_value_1")
	os.Setenv("FETCH_KEY_2", "fetch_value_2")
	defer os.Unsetenv("FETCH_KEY_1")
	defer os.Unsetenv("FETCH_KEY_2")

	provider := osprov.New()

	configs, err := provider.FetchConfig(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, configs)
	assert.Equal(t, "fetch_value_1", configs["FETCH_KEY_1"])
	assert.Equal(t, "fetch_value_2", configs["FETCH_KEY_2"])

	os.Setenv("FETCH_KEY_2", "fetch_value_3")
	configs, err = provider.FetchConfig(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, configs)
	assert.Equal(t, "fetch_value_1", configs["FETCH_KEY_1"])
	assert.Equal(t, "fetch_value_3", configs["FETCH_KEY_2"])
}

func TestFetchConfigWithEmptyValue(t *testing.T) {
	// Set an environment variable with an empty value
	os.Setenv("EMPTY_KEY", "")
	defer os.Unsetenv("EMPTY_KEY")

	provider := osprov.New()

	configs, err := provider.FetchConfig(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, configs)
	assert.Equal(t, "", configs["EMPTY_KEY"])
}

func TestFetchConfigReturnsAllEnvironmentVariables(t *testing.T) {
	os.Setenv("TEST_ENV_VAR", "test_env_value")
	defer os.Unsetenv("TEST_ENV_VAR")

	provider := osprov.New()

	configs, err := provider.FetchConfig(t.Context())
	assert.NoError(t, err)
	assert.Greater(t, len(configs), 0)
	assert.Equal(t, "test_env_value", configs["TEST_ENV_VAR"])
}
