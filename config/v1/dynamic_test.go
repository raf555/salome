package config

import (
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"go.uber.org/mock/gomock"
)

func TestDynamic(t *testing.T) {
	type Config1 struct {
		Test string `env:"TEST1"`
	}

	type Config2 struct {
		Test string `env:"TEST2"`
	}

	t.Run("E2E with notify - callbacks fired on change", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			providerMock := NewMockProvider(ctrl)

			providerMock.EXPECT().Config(gomock.Any()).Return(map[string]string{
				"TEST1": "abc",
				"TEST2": "def",
			}, nil)
			providerMock.EXPECT().FetchConfig(gomock.Any()).Return(map[string]string{
				"TEST1": "abc1",
				"TEST2": "def1",
			}, nil)

			dynamic, err := NewDynamic(providerMock)
			if err != nil {
				t.Fatalf("expecting nil error when initializing dynamic, got %v", err)
			}

			dynamic.Start()
			defer dynamic.Close()

			cfg1 := Config1{}
			adder, err := dynamic.RegisterConfigWithNotify(&cfg1, func() any {
				return &Config1{}
			})
			if err != nil {
				t.Fatalf("expecting nil error when registering Config1 with notify, got %v", err)
			}

			var cb1Count atomic.Int32
			var cb1LastValue Config1
			adder.Add(func(v any) {
				cb1Count.Add(1)
				cb1LastValue = *v.(*Config1)
			})

			var cb2Count atomic.Int32
			adder.Add(func(v any) {
				cb2Count.Add(1)
			})

			if cb1Count.Load() != 0 {
				t.Errorf("expecting 0 callbacks before update, got %d", cb1Count.Load())
			}

			time.Sleep(15 * time.Second)

			if cb1Count.Load() != 1 {
				t.Errorf("expecting 1 callback after update, got %d", cb1Count.Load())
			}
			if expected := (Config1{Test: "abc1"}); cb1LastValue != expected {
				t.Errorf("expecting callback value %v, got %v", expected, cb1LastValue)
			}
			if cb2Count.Load() != 1 {
				t.Errorf("expecting 1 callback for second subscriber after update, got %d", cb2Count.Load())
			}
		})
	})

	t.Run("E2E with notify - no callback when config unchanged", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			providerMock := NewMockProvider(ctrl)

			providerMock.EXPECT().Config(gomock.Any()).Return(map[string]string{
				"TEST1": "abc",
			}, nil)
			providerMock.EXPECT().FetchConfig(gomock.Any()).Return(map[string]string{
				"TEST1": "abc", // same as initial
			}, nil)

			dynamic, err := NewDynamic(providerMock)
			if err != nil {
				t.Fatalf("expecting nil error when initializing dynamic, got %v", err)
			}

			dynamic.Start()
			defer dynamic.Close()

			cfg1 := Config1{}
			adder, err := dynamic.RegisterConfigWithNotify(&cfg1, func() any {
				return &Config1{}
			})
			if err != nil {
				t.Fatalf("expecting nil error when registering Config1 with notify, got %v", err)
			}

			var cbCount atomic.Int32
			adder.Add(func(v any) {
				cbCount.Add(1)
			})

			time.Sleep(15 * time.Second)

			if cbCount.Load() != 0 {
				t.Errorf("expecting 0 callbacks when config unchanged, got %d", cbCount.Load())
			}
		})
	})

	t.Run("E2E success", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			errCbCounter := atomic.Int32{}
			ctrl := gomock.NewController(t)
			providerMock := NewMockProvider(ctrl)

			providerMock.EXPECT().Config(gomock.Any()).Return(map[string]string{
				"TEST1": "abc",
				"TEST2": "def",
			}, nil)
			providerMock.EXPECT().FetchConfig(gomock.Any()).Return(map[string]string{
				"TEST1": "abc1",
				"TEST2": "def1",
			}, nil)

			dynamic, err := NewDynamic(providerMock,
				WithErrCallback(func(_ error) {
					errCbCounter.Add(1)
				}),
			)
			if err != nil {
				t.Fatalf("expecting nil error when initializing dynamic, got %v", err)
			}

			dynamic.Start()
			defer dynamic.Close()

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

			time.Sleep(15 * time.Second)

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
