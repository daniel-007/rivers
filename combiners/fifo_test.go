package combiners_test

import (
	"testing"
	"github.com/drborges/riversv2"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/drborges/riversv2/rx"
	"github.com/drborges/riversv2/combiners"
)

func TestFifo(t *testing.T) {
	Convey("Given I have a context", t, func() {
		context := rivers.NewContext()

		Convey("And a stream of data", func() {
			in1, out1 := rx.NewStream(2)
			out1 <- 1
			out1 <- 2
			close(out1)

			in2, out2 := rx.NewStream(2)
			out2 <- 3
			out2 <- 4
			close(out2)

			Convey("When I apply the combiner to the streams", func() {
				combined := combiners.New(context).FIFO().Combine(in1, in2)

				Convey("Then a transformed stream is returned", func() {
					So(combined.Read(), ShouldResemble, []rx.T{1, 2, 3, 4})
				})
			})

			Convey("When I close the context", func() {
				context.Close()

				Convey("And I apply the transformer to the stream", func() {
					combined := combiners.New(context).Zip().Combine(in1, in2)

					Convey("Then no item is sent to the next stage", func() {
						So(combined.Read(), ShouldBeEmpty)
					})
				})
			})
		})
	})
}