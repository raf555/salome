package config

import (
	"testing"

	"go.uber.org/mock/gomock"
)

func TestLoadConfigTo(t *testing.T) {
	type Config struct {
		Test string `env:"TEST,required" validate:"len=3"`
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		providerMock := NewMockProvider(ctrl)

		providerMock.EXPECT().Config(gomock.Any()).Return(map[string]string{
			"TEST": "123",
		}, nil)

		conf, err := LoadConfigTo[Config](providerMock)
		if err != nil {
			t.Errorf("expecting nil error, got %v", err)
		}

		if len(conf.Test) != 3 {
			t.Errorf("expecting Config.Test to have length 3, got %d", len(conf.Test))
		}
	})
}

func TestLoadDynamicConfigToWithNotify(t *testing.T) {
	type Config struct {
		Test string `env:"TEST,required" validate:"len=3"`
	}

	t.Run("success - Get returns current config", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		dynamicMock := NewMockDynamicConfigManager(ctrl)

		dynamicMock.EXPECT().
			RegisterConfigWithNotify(gomock.AssignableToTypeOf(&Config{}), gomock.AssignableToTypeOf(func() any { return nil })).
			Return(callbackAdderFunc(func(func(any)) {}), nil)

		dynamicMock.EXPECT().GetConfig(gomock.AssignableToTypeOf(&Config{})).Return(&Config{Test: "123"})
		dynamicMock.EXPECT().GetConfig(gomock.AssignableToTypeOf(&Config{})).Return(&Config{Test: "456"})

		conf, err := LoadDynamicConfigToWithNotify[Config](dynamicMock)
		if err != nil {
			t.Errorf("expecting nil error, got %v", err)
		}

		first := conf.Get()
		if expected := (Config{Test: "123"}); first != expected {
			t.Errorf("expecting %v, got %v", expected, first)
		}

		second := conf.Get()
		if expected := (Config{Test: "456"}); second != expected {
			t.Errorf("expecting %v, got %v", expected, second)
		}
	})

	t.Run("success - RegisterCallback wraps and forwards to adder", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		dynamicMock := NewMockDynamicConfigManager(ctrl)

		var registeredRawCb func(any)
		dynamicMock.EXPECT().
			RegisterConfigWithNotify(gomock.AssignableToTypeOf(&Config{}), gomock.AssignableToTypeOf(func() any { return nil })).
			Return(callbackAdderFunc(func(cb func(any)) { registeredRawCb = cb }), nil)

		conf, err := LoadDynamicConfigToWithNotify[Config](dynamicMock)
		if err != nil {
			t.Errorf("expecting nil error, got %v", err)
		}

		var notified Config
		conf.RegisterCallback(func(c Config) {
			notified = c
		})

		if registeredRawCb == nil {
			t.Fatal("expecting adder to be called when RegisterCallback is invoked")
		}

		registeredRawCb(&Config{Test: "789"})
		if expected := (Config{Test: "789"}); notified != expected {
			t.Errorf("expecting notified value %v, got %v", expected, notified)
		}
	})
}

func TestLoadDynamicConfigTo(t *testing.T) {
	type Config struct {
		Test string `env:"TEST,required" validate:"len=3"`
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		dynamicMock := NewMockDynamicConfigManager(ctrl)

		dynamicMock.EXPECT().RegisterConfig(gomock.AssignableToTypeOf(&Config{}), gomock.AssignableToTypeOf(func() any { return nil })).Return(nil)

		dynamicMock.EXPECT().GetConfig(gomock.AssignableToTypeOf(&Config{})).Return(&Config{Test: "123"})
		dynamicMock.EXPECT().GetConfig(gomock.AssignableToTypeOf(&Config{})).Return(&Config{Test: "456"})

		conf, err := LoadDynamicConfigTo[Config](dynamicMock)
		if err != nil {
			t.Errorf("expecting nil error, got %v", err)
		}

		first := conf.Get()
		if expected := (Config{Test: "123"}); first != expected {
			t.Errorf("expecting %v, got %v", expected, first)
		}

		second := conf.Get()
		if expected := (Config{Test: "456"}); second != expected {
			t.Errorf("expecting %v, got %v", expected, second)
		}
	})
}
