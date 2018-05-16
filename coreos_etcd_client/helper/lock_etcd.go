package helper

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/coreos/etcd/client"
)

// EtcdLocker is an ETCD-backed Locker
type EtcdLocker struct {
	etcdEndpoints []string
	lockPath      string
}

// NewEtcdLocker returns a new EtcdLocker
func NewEtcdLocker(etcdEndpoints []string, lockPath string) *EtcdLocker {
	return &EtcdLocker{etcdEndpoints: etcdEndpoints, lockPath: lockPath}
}

// Lock tries to create an inter-services lock in a etcd directory, it will
// return true if it was able to grab the lock or false otherwise.
func (el *EtcdLocker) Lock(ctx context.Context, lockName string, ttl time.Duration) (bool, error) {
	var err error
	// create a new etcd client
	var c client.Client
	c, err = client.New(client.Config{Endpoints: el.etcdEndpoints})
	if err != nil {
		return false, err
	}
	// create a new key API
	keyAPI := client.NewKeysAPI(c)
	// always try to create directory
	if _, err = keyAPI.Get(ctx, el.lockPath, nil); client.IsKeyNotFound(err) {
		if _, err = keyAPI.Set(ctx, el.lockPath, "", &client.SetOptions{
			Dir: true,
		}); err != nil {
			return false, err
		}
	}
	// set the key
	_, err = keyAPI.Set(ctx, path.Join(el.lockPath, lockName), el.getOwner(), &client.SetOptions{
		PrevExist: client.PrevNoExist,
		TTL:       ttl,
	})
	if err != nil {
		// we do not return an error if the error is a node exist!
		if err, ok := err.(client.Error); ok && err.Code == client.ErrorCodeNodeExist {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// Unlock removes the lock from etcd
func (el *EtcdLocker) Unlock(ctx context.Context, lockName string) error {
	var err error
	// create a new etcd client
	var c client.Client
	c, err = client.New(client.Config{Endpoints: el.etcdEndpoints})
	if err != nil {
		return err
	}
	// create a new key API
	keyAPI := client.NewKeysAPI(c)
	_, err = keyAPI.Delete(ctx, path.Join(el.lockPath, lockName), &client.DeleteOptions{PrevValue: el.getOwner()})
	return err
}

// Touch tries to keep the inter-services lock in a etcd directory, it will
// return true if it was able to grab the lock or false otherwise.
func (el *EtcdLocker) Touch(ctx context.Context, lockName string, ttl time.Duration) (bool, error) {
	var err error
	// create a new etcd client
	var c client.Client
	c, err = client.New(client.Config{Endpoints: el.etcdEndpoints})
	if err != nil {
		return false, err
	}
	// create a new key API
	keyAPI := client.NewKeysAPI(c)
	// initialize a set options to touch the key
	setOptions := &client.SetOptions{
		PrevValue: el.getOwner(),
		PrevExist: client.PrevExist,
		TTL:       ttl,
	}
	// set the key
	_, err = keyAPI.Set(ctx, path.Join(el.lockPath, lockName), el.getOwner(), setOptions)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (el *EtcdLocker) getOwner() string {
	host, err := os.Hostname()
	if err != nil {
		panic(fmt.Errorf("error getting the hostname of the server: %s", err))
	}
	return fmt.Sprintf("%s-%d", host, os.Getpid())
}
