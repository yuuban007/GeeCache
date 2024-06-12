package singleflight

import "sync"

// wg is used to wait for the goroutine to complete.
// val holds the value returned by the function call.
// err holds any error that occurred during the function call.
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do executes and returns the result of the function `fn` if the given `key` is not already being processed.
// If the `key` is being processed by another goroutine, `Do` waits for that goroutine to complete and returns its result.
// The result of the function `fn` is stored in the cache for future use.
// The `Do` method is safe for concurrent use.
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Lock()

	c.val, c.err = fn()
	c.wg.Done()
	g.mu.Unlock()
	return c.val, c.err
}
