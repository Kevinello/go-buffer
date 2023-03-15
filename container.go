package buffer

// Container data Container in the Buffer
//
//	@author kevineluo
//	@update 2023-03-15 09:35:51
type Container[T any] interface {
	// put int an element. Will NEVER check if is full so be caution.
	put(element T) error
	// execute will apply some action on this container. SHOULD RESET CONTAINER when called
	execute() error
	// isFull return true if this container is full, then should call `execute` to reset the container
	isFull() bool
	// recover will be called when error arisen
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
		buffer.logger.Info("buffer if full, will execute batchFunc")
		if err := buffer.container.execute(); err != nil {
			buffer.logger.Error(err, "error when call Container.execute")
			return err
		}
	}

	return nil
}
