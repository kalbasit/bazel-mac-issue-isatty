package helper

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const lockBasePath = "/locks/cron"

func lockPath() string {
	f, err := ioutil.TempFile("", "testing-cron")
	if err != nil {
		panic(fmt.Errorf("error creating a temporary file to use as filename for the lock: %s", err))
	}
	p := path.Join(lockBasePath, path.Base(f.Name()))
	if err = f.Close(); err != nil {
		panic(fmt.Errorf("error closing the temporary file: %s", err))
	}
	if err = os.Remove(f.Name()); err != nil {
		panic(fmt.Errorf("error removing the temporary file: %s", err))
	}

	return p
}

func TestLock(t *testing.T) {
	ctx := context.Background()
	etcdEndpoints := []string{"http://127.0.0.1:2379"}

	// create a new etcd client
	c, err := client.New(client.Config{Endpoints: etcdEndpoints})
	require.NoError(t, err)

	// make sure we are actually connected to etcd
	_, err = c.GetVersion(context.Background())
	if err != nil {
		t.Logf("error connecting to etcd: %s", err)
		t.SkipNow()
	}

	// define the el
	el := NewEtcdLocker(etcdEndpoints, lockPath())

	// create new key api
	keyAPI := client.NewKeysAPI(c)
	_, err = keyAPI.Set(ctx, el.lockPath, "", &client.SetOptions{Dir: true})
	require.NoError(t, err)

	var ok bool
	t.Run("grabbing the lock from the same owner", func(t *testing.T) {
		t.Run("try to grab the lock, it should work", func(t *testing.T) {
			ok, err = el.Lock(ctx, "lock-name", 1*time.Millisecond)
			if assert.NoError(t, err) {
				assert.True(t, ok)
			}
		})

		t.Run("extend the lock for 20 minutes", func(t *testing.T) {
			ok, err = el.Touch(ctx, "lock-name", 20*time.Minute)
			if assert.NoError(t, err) {
				assert.True(t, ok)
			}
		})

		t.Run("try to grab the lock, it should not work", func(t *testing.T) {
			ok, err = el.Lock(ctx, "lock-name", 1*time.Millisecond)
			if assert.NoError(t, err) {
				assert.False(t, ok)
			}
		})
	})

	t.Run("grabbing the lock from a different owner", func(t *testing.T) {
		_, err = keyAPI.Set(ctx, path.Join(el.lockPath, "new-lock-name"), "hostname", nil)
		assert.NoError(t, err)

		t.Run("try to grab the lock, it should not work", func(t *testing.T) {
			ok, err = el.Lock(ctx, "new-lock-name", 1*time.Millisecond)
			if assert.NoError(t, err) {
				assert.False(t, ok)
			}
		})

		t.Run("extend the lock for 20 minutes, it should not work", func(t *testing.T) {
			ok, err = el.Touch(ctx, "new-lock-name", 20*time.Minute)
			assert.Error(t, err)
		})
	})
}

func TestUnlock(t *testing.T) {
	ctx := context.Background()
	etcdEndpoints := []string{"http://127.0.0.1:2379"}

	// create a new etcd client
	c, err := client.New(client.Config{Endpoints: etcdEndpoints})
	require.NoError(t, err)

	// make sure we are actually connected to etcd
	_, err = c.GetVersion(context.Background())
	if err != nil {
		t.Logf("error connecting to etcd: %s", err)
		t.SkipNow()
	}

	// define the el
	el := NewEtcdLocker(etcdEndpoints, lockPath())

	// create new key api
	keyAPI := client.NewKeysAPI(c)
	_, err = keyAPI.Set(ctx, el.lockPath, "", &client.SetOptions{Dir: true})
	require.NoError(t, err)

	var ok bool
	t.Run("try to grab the lock, it should work", func(t *testing.T) {
		ok, err = el.Lock(ctx, "test", 1*time.Millisecond)
		if assert.NoError(t, err) {
			assert.True(t, ok)
		}
	})

	t.Run("try to remove the lock, it should work", func(t *testing.T) {
		err = el.Unlock(ctx, "test")
		assert.NoError(t, err)
	})

	t.Run("make sure the lock was actually removed", func(t *testing.T) {
		_, err = keyAPI.Get(ctx, path.Join(el.lockPath, "test"), nil)
		assert.True(t, client.IsKeyNotFound(err))
	})
}
