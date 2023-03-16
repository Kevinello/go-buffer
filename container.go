package buffer

// Container data Container in the Buffer
//
//	@author kevineluo
//	@update 2023-03-15 09:35:51
type Container[T any] interface {
	// put int an element. Will NEVER check if is full so be caution.
	put(element T) error
	// flush will apply some action on this container. SHOULD RESET CONTAINER when called
	// when it's a sync flush, container.put won't be called until flush, so Container can empty it's data
	// when it's a async flush, container.put is still being called when doing flush, so Container should split a batch from it's data to be flushed and reset itself
	flush() error
	// isFull return true if this container is full, then should call `execute` to reset the container
	isFull() bool
	// will call reset when flush return error
	reset()
}

// putAndCheck put an element into container and execute user function when full
//
//	@param buffer *Buffer[T]
//	@param data T
//	@return error
//	@author kevineluo
//	@update 2023-03-15 09:46:37
func putAndCheck[T any](buffer *Buffer[T], data T) error {
	if err := buffer.container.put(data); err != nil {
		buffer.logger.Error(err, "buffer cannot write message to container")
		return err
	}

	if buffer.container.isFull() {
		buffer.logger.Info("buffer if full, will call container.flush")
		if err := buffer.container.flush(); err != nil {
			buffer.logger.Error(err, "error when call Container.execute")
			return err
		}
	}

	return nil
}
