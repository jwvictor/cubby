package server

import (
	"errors"
	lru "github.com/hashicorp/golang-lru"
	"io/ioutil"
)

type StaticCache struct {
	cacheProvider *lru.Cache
}

func NewStaticCache() *StaticCache {
	lruCache, _ := lru.New(64)
	return &StaticCache{cacheProvider: lruCache}
}

func (c *StaticCache) Get(filename string) ([]byte, error) {

	// Check the cache
	res, ok := c.cacheProvider.Get(filename)
	if ok {
		if bs, ok2 := res.([]byte); ok2 {
			return bs, nil
		} else {
			return nil, errors.New("InternalError")
		}
	}

	// Read the file
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Update the cache and return
	c.cacheProvider.Add(filename, bs)
	return bs, nil
}
