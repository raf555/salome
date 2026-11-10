package layered_test

import (
	"context"
	"errors"
	"testing"

	"github.com/raf555/salome/config/v1/providers/layered"
	"github.com/stretchr/testify/assert"
)

type fakeProvider struct {
	initial map[string]string
	fetched map[string]string
	err     error
}

func (f *fakeProvider) Config(_ context.Context) (map[string]string, error) {
	return f.initial, f.err
}

func (f *fakeProvider) FetchConfig(_ context.Context) (map[string]string, error) {
	return f.fetched, f.err
}

func TestNewWithoutProviders(t *testing.T) {
	_, err := layered.New()
	assert.ErrorIs(t, err, layered.ErrNoProviders)
}

func TestNewProviderError(t *testing.T) {
	errBoom := errors.New("boom")

	_, err := layered.New(&fakeProvider{}, &fakeProvider{err: errBoom})
	assert.ErrorIs(t, err, errBoom)
}

func TestConfigLaterProviderWins(t *testing.T) {
	base := &fakeProvider{initial: map[string]string{"SHARED": "base", "BASE_ONLY": "base"}}
	overlay := &fakeProvider{initial: map[string]string{"SHARED": "overlay", "OVERLAY_ONLY": "overlay"}}

	provider, err := layered.New(base, overlay)
	assert.NoError(t, err)

	cfg, err := provider.Config(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"SHARED":       "overlay",
		"BASE_ONLY":    "base",
		"OVERLAY_ONLY": "overlay",
	}, cfg)

	// Config returns a copy.
	cfg["SHARED"] = "modified"
	cfg, _ = provider.Config(t.Context())
	assert.Equal(t, "overlay", cfg["SHARED"])

	// Merging leaves the providers' own maps alone.
	assert.Equal(t, "base", base.initial["SHARED"])
}

func TestFetchConfigLaterProviderWins(t *testing.T) {
	base := &fakeProvider{fetched: map[string]string{"SHARED": "base", "BASE_ONLY": "base"}}
	overlay := &fakeProvider{fetched: map[string]string{"SHARED": "overlay"}}

	provider, err := layered.New(base, overlay)
	assert.NoError(t, err)

	cfg, err := provider.FetchConfig(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"SHARED": "overlay", "BASE_ONLY": "base"}, cfg)
}

func TestFetchConfigProviderError(t *testing.T) {
	errBoom := errors.New("boom")
	overlay := &fakeProvider{}

	provider, err := layered.New(&fakeProvider{}, overlay)
	assert.NoError(t, err)

	overlay.err = errBoom
	_, err = provider.FetchConfig(t.Context())
	assert.ErrorIs(t, err, errBoom)
}
