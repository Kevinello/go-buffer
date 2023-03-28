package container

// Container data Container in the Buffer
//
//	@author kevineluo
//	@update 2023-03-15 09:35:51
type Container[T any] interface {
	// Put int an element. Will NEVER check if is full so be caution.
	Put(data T) error
	// Flush will apply some action on this container. SHOULD RESET CONTAINER when called
	// when it's a sync Flush, container.put will be blocked until Flush, so Container can empty it's data properly
	// when it's a async Flush, please set the chanBufSize of the buffer to 0 to block container.put,
	// or container.put will still being called when doing Flush, so Container should split a batch from it's data to be flushed and reset itself
	Flush() error
	// IsFull return true if this container is full
	IsFull() bool
	// will call Reset when flush return error
	Reset()
}
