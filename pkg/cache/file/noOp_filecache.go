package cache

type NoOpFilecache struct {
}

func (nc *NoOpFilecache) Increment(path string) {
	return
}

func (nc *NoOpFilecache) Stop() {
	return
}

func (nc *NoOpFilecache) Lookup(path string) ([]byte, bool) {
	return nil, false
}
