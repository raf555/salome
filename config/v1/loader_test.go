package config_test

import (
	"testing"

	"github.com/raf555/salome/config/v1"
	"github.com/raf555/salome/config/v1/configtest"
	"go.uber.org/mock/gomock"
)

func TestLoadConfigTo(t *testing.T) {
	type Config struct {
		Test string `env:"TEST,required" validate:"len=3"`
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		providerMock := configtest.NewMockProvider(ctrl)

		providerMock.EXPECT().Config(gomock.Any()).Return(map[string]string{
			"TEST": "123",
		}, nil)

		conf, err := config.LoadConfigTo[Config](providerMock)
		if err != nil {
			t.Errorf("expecting nil error, got %v", err)
		}

		if len(conf.Test) != 3 {
			t.Errorf("expecting Config.Test to have length 3, got %d", len(conf.Test))
		}
	})
}

func TestLoadDynamicConfigTo(t *testing.T) {
	type Config struct {
		Test string `env:"TEST,required" validate:"len=3"`
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		dynamicMock := configtest.NewMockDynamicConfigManager(ctrl)

		dynamicMock.EXPECT().RegisterConfig(Config{}, gomock.AssignableToTypeOf(func() any { return nil })).Return(nil)

		dynamicMock.EXPECT().GetConfig(Config{}).Return(&Config{Test: "123"}) // 1st call
		dynamicMock.EXPECT().GetConfig(Config{}).Return(&Config{Test: "456"}) // 2nd call

		conf, err := config.LoadDynamicConfigTo[Config](dynamicMock)
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
