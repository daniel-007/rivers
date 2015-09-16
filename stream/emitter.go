package stream

type emitter struct {
	context  Context
	writable Writable
}

func NewEmitter(context Context, w Writable) Emitter {
	return &emitter{context, w}
}

func (emitter *emitter) Emit(data T) {
	select {
	case <-emitter.context.Failure():
		panic(Done)
	default:
		emitter.writable <- data
	}
}
