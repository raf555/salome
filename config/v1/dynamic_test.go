package config_test

import (
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/raf555/salome/config/v1"
	"github.com/raf555/salome/config/v1/configtest"
	"go.uber.org/mock/gomock"
)

func TestDynamic(t *testing.T) {
	type Config1 struct {
		Test string `env:"TEST1"`
	}

	type Config2 struct {
		Test string `env:"TEST2"`
	}

	t.Run("E2E success", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			errCbCounter := atomic.Int32{}
			ctrl := gomock.NewController(t)
			providerMock := configtest.NewMockProvider(ctrl)

			// init call
			providerMock.EXPECT().Config(gomock.Any()).Return(map[string]string{
				"TEST1": "abc",
				"TEST2": "def",
			}, nil)

			dynamic, err := config.NewDynamic(providerMock,
				config.WithErrCallback(func(_ error) {
					errCbCounter.Add(1)
				}),
			)
			if err != nil {
				t.Fatalf("expecting nil error when initializing dynamic, got %v", err)
			}

			dynamic.Start()
			defer dynamic.Close()

			// simulate separate goroutines

			cfg1 := Config1{}
			cfg2 := Config2{}

			err = dynamic.RegisterConfig(&cfg1, func() any {
				return &Config1{}
			})
			if err != nil {
				t.Fatalf("expecting nil error when registering Config1, got %v", err)
			}

			err = dynamic.RegisterConfig(&cfg2, func() any {
				return &Config2{}
			})
			if err != nil {
				t.Fatalf("expecting nil error when registering Config2, got %v", err)
			}

			conf1 := dynamic.GetConfig(&cfg1).(*Config1)
			if expected := (Config1{Test: "abc"}); expected != *conf1 {
				t.Errorf("expecting first conf1 to be %v, got %v", expected, conf1)
			}
			conf2 := dynamic.GetConfig(&cfg2).(*Config2)
			if expected := (Config2{Test: "def"}); expected != *conf2 {
				t.Errorf("expecting first conf2 to be %v, got %v", expected, conf2)
			}

			providerMock.EXPECT().FetchConfig(gomock.Any()).Return(map[string]string{
				"TEST1": "abc1",
				"TEST2": "def1",
			}, nil)
			time.Sleep(15 * time.Second) // wait until it gets refreshed

			conf1 = dynamic.GetConfig(&cfg1).(*Config1)
			if expected := (Config1{Test: "abc1"}); expected != *conf1 {
				t.Errorf("expecting second conf1 to be %v, got %v", expected, conf1)
			}
			conf2 = dynamic.GetConfig(&cfg2).(*Config2)
			if expected := (Config2{Test: "def1"}); expected != *conf2 {
				t.Errorf("expecting second conf2 to be %v, got %v", expected, conf2)
			}
		})
	})
}
