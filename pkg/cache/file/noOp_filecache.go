package cache

type NoOpFilecache struct {
}

func (nc *NoOpFilecache) Increment(path string) {
	return
}
