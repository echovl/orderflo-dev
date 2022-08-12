package redis

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/echovl/orderflo-dev/testhelpers/docker"
)

func TestRedis_Get(t *testing.T) {
	testcases := []struct {
		name          string
		currentValues map[string]string
		key           string
		expectedValue []byte
		wantErr       bool
	}{
		{
			name: "not found",
			currentValues: map[string]string{
				"key1": "obj1",
				"key2": "obj2",
			},
			expectedValue: nil,
			key:           "key3",
			wantErr:       true,
		},
		{
			name: "found",
			currentValues: map[string]string{
				"key1": "obj1",
				"key2": "obj2",
			},
			expectedValue: []byte("obj1"),
			key:           "key1",
		},
	}

	cleanup, addr := prepareTestContainer(t)
	defer cleanup()

	db, err := New(&Config{
		Addr: addr,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close(context.TODO())

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := db.Del(context.TODO(), "key1", "key2", "key3", "key4")
			if err != nil {
				t.Fatal(err)
			}

			for k, v := range tc.currentValues {
				err := db.Set(context.TODO(), k, v, time.Minute)
				if err != nil {
					t.Fatal(err)
				}
			}

			val, err := db.Get(context.TODO(), tc.key)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error:\ngot: %v", val)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(val, tc.expectedValue) {
				t.Fatalf("mismatched get value:\ngot: %v\nwant: %v", val, tc.expectedValue)
			}
		})
	}
}

func TestRedis_Set(t *testing.T) {
	testcases := []struct {
		name       string
		key        string
		value      []byte
		expiration time.Duration
	}{
		{
			name:       "with expiration",
			key:        "key1",
			value:      []byte("obj1"),
			expiration: 100 * time.Millisecond,
		},
		{
			name:  "no expiration",
			key:   "key1",
			value: []byte("obj2"),
		},
	}

	cleanup, addr := prepareTestContainer(t)
	defer cleanup()

	db, err := New(&Config{
		Addr: addr,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close(context.TODO())

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := db.Del(context.TODO(), "key1", "key2", "key3", "key4")
			if err != nil {
				t.Fatal(err)
			}

			err = db.Set(context.TODO(), tc.key, tc.value, tc.expiration)
			if err != nil {
				t.Fatal(err)
			}

			val, err := db.Get(context.TODO(), tc.key)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(val, tc.value) {
				t.Fatalf("mismatched get value:\ngot: %v\nwant: %v", val, tc.value)
			}

			if tc.expiration > 0 {
				time.Sleep(2 * tc.expiration)
				_, err := db.Get(context.TODO(), tc.key)
				if err == nil {
					t.Fatalf("key %s should have expired", tc.key)
				}
			}
		})
	}
}

func TestRedis_Del(t *testing.T) {
	testcases := []struct {
		name          string
		currentValues map[string]string
		key           string
	}{
		{
			name: "not found",
			currentValues: map[string]string{
				"key1": "obj1",
				"key2": "obj2",
			},
			key: "key3",
		},
		{
			name: "found",
			currentValues: map[string]string{
				"key1": "obj1",
				"key2": "obj2",
			},
			key: "key1",
		},
	}

	cleanup, addr := prepareTestContainer(t)
	defer cleanup()

	db, err := New(&Config{
		Addr: addr,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close(context.TODO())

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := db.Del(context.TODO(), "key1", "key2", "key3", "key4")
			if err != nil {
				t.Fatal(err)
			}

			for k, v := range tc.currentValues {
				err := db.Set(context.TODO(), k, v, time.Minute)
				if err != nil {
					t.Fatal(err)
				}
			}

			err = db.Del(context.TODO(), tc.key)
			if err != nil {
				t.Fatal(err)
			}

			val, err := db.Get(context.TODO(), tc.key)
			if err == nil {
				t.Fatalf("expected error:\ngot: %v", val)
			}
		})
	}
}

type cfg struct {
	docker.ServiceHostPort
	Addr string
}

var _ docker.ServiceConfig = &cfg{}

func prepareTestContainer(t *testing.T) (func(), string) {
	if url := os.Getenv("TEST_REDIS_ADDR"); url != "" {
		return func() {}, url
	}

	runner, err := docker.NewServiceRunner(docker.RunOptions{
		ImageRepo:     "redis",
		ImageTag:      "latest",
		ContainerName: "redis-test",
		Ports:         []string{"6379/tcp"},
		Env:           []string{},
	})
	if err != nil {
		t.Fatalf("could not start local redis: %s", err)
	}

	svc, err := runner.StartService(context.Background(), connect)
	if err != nil {
		t.Fatalf("could not start local redis: %s", err)
	}

	return svc.Cleanup, svc.Config.(*cfg).Addr
}

func connect(ctx context.Context, host string, port int) (docker.ServiceConfig, error) {
	hostIP := docker.NewServiceHostPort(host, port)
	client := redis.NewClient(&redis.Options{
		Addr: hostIP.Address(),
	})
	defer client.Close()

	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	return &cfg{
		ServiceHostPort: *hostIP,
		Addr:            hostIP.Address(),
	}, nil
}
