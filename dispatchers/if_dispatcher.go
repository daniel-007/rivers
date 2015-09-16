package dispatchers

import (
	"github.com/drborges/rivers/stream"
)

type ifDispatcher struct {
	context stream.Context
	fn      stream.PredicateFn
}

func (dispatcher *ifDispatcher) Dispatch(in stream.Readable, writables ...stream.Writable) stream.Readable {
	notDispatchedReadable, notDispatchedWritable := stream.New(in.Capacity())

	dispatchedCount := 0
	done := make(chan bool, len(writables))

	closeWritables := func() {
		defer func() {
			for _, writable := range writables {
				close(writable)
			}
		}()

		expectedDoneMessages := dispatchedCount * len(writables)
		for i := 0; i < expectedDoneMessages; i++ {
			select {
			case <-dispatcher.context.Failure():
				return
			case <-done:
				continue
			}
		}
	}

	go func() {
		defer dispatcher.context.Recover()
		defer close(notDispatchedWritable)
		defer closeWritables()

		for data := range in {
			select {
			case <-dispatcher.context.Failure():
				return
			default:
				if dispatcher.fn(data) {
					dispatchedCount++
					for _, writable := range writables {
						// dispatch data asynchronously so that
						// slow receivers don't block the dispatch
						// process
						go func(w stream.Writable, d stream.T) {
							w <- d
							done <- true
						}(writable, data)
					}
				} else {
					notDispatchedWritable <- data
				}
			}
		}
	}()

	return notDispatchedReadable
}
