package transformers_test

import (
	"github.com/drborges/rivers"
	"github.com/drborges/rivers/stream"
	"github.com/drborges/rivers/transformers"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type batch struct {
	size  int
	items []stream.T
}

func (batch *batch) Full() bool {
	return len(batch.items) == batch.size
}

func (batch *batch) Empty() bool {
	return len(batch.items) == 0
}

func (batch *batch) Commit(emitter stream.Emitter) {
	emitter.Emit(batch.items)
	batch.items = []stream.T{}
}

func (batch *batch) Add(data stream.T) {
	batch.items = append(batch.items, data)
}

func TestBatcher(t *testing.T) {
	Convey("Given I have a context", t, func() {
		context := rivers.NewContext()

		Convey("And a stream of data", func() {
			in, out := stream.New(3)
			out <- 1
			out <- 2
			out <- 3
			close(out)

			Convey("When I apply the batch transformer to the stream", func() {
				transformer := transformers.Batch(2)
				transformer.Attach(context)
				next := transformer.Transform(in)

				Convey("Then a transformed stream is returned", func() {
					So(next.ReadAll(), ShouldResemble, []stream.T{[]stream.T{1, 2}, []stream.T{3}})
				})
			})

			Convey("When I apply the batch by transformer to the stream", func() {
				transformer := transformers.BatchBy(&batch{size: 1})
				transformer.Attach(context)
				next := transformer.Transform(in)

				Convey("Then a transformed stream is returned", func() {
					So(next.ReadAll(), ShouldResemble, []stream.T{[]stream.T{1}, []stream.T{2}, []stream.T{3}})
				})
			})

			Convey("When I close the context", func() {
				context.Close(stream.Done)

				Convey("And I apply the transformer to the stream", func() {
					transformer := transformers.Flatten()
					transformer.Attach(context)
					next := transformer.Transform(in)

					Convey("Then no item is sent to the next stage", func() {
						So(next.ReadAll(), ShouldBeEmpty)
					})
				})
			})
		})
	})
}
