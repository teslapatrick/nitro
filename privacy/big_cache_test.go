package privacy

import (
	"bytes"
	"context"
	"errors"
	"github.com/allegro/bigcache"
	"github.com/offchainlabs/nitro/das/dastree"
	"testing"
)

func TestBigCache(t *testing.T) {
	notFound := bigcache.ErrEntryNotFound
	ctx := context.Background()
	bc, err := bigcache.NewBigCache(bigcache.DefaultConfig(BigCacheConfigDefault.Expiration))
	if err != nil {
		t.Error(err)
	}

	bcService := &BigCacheStorageService{
		config:   BigCacheConfigDefault,
		bigCache: bc,
	}

	val := "this is a big cache test value"
	correctKey := dastree.Hash([]byte(val))
	incorrectKey := dastree.Hash(append([]byte(val), []byte{1}...))

	_, err = bcService.Get(ctx, correctKey.String())
	if !errors.Is(err, notFound) {
		t.Fatal(err)
	}

	err = bcService.Set(ctx, correctKey.String(), []byte(val), uint64(BigCacheConfigDefault.Expiration.Seconds()))
	if err != nil {
		t.Fatal(err)
	}

	v, err := bcService.Get(ctx, correctKey.String())
	if !bytes.Equal(v, []byte(val)) {
		t.Fatal(err)
	}

	_, err = bcService.Get(ctx, incorrectKey.String())
	if !errors.Is(err, notFound) {
		t.Error(err)
	}
}
