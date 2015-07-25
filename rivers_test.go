package rivers_test

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/drborges/riversv2"
	"github.com/drborges/riversv2/rx"
	"net"
	"github.com/drborges/riversv2/scanners"
	"strings"
)

func TestRiversAPI(t *testing.T) {
	toString := func(data rx.T) rx.T { return string(data.([]byte)) }
	nonEmptyLines := func(data rx.T) bool { return data.(string) != "" }
	splitWord := func(data rx.T) rx.T { return strings.Split(data.(string), " ") }
	evens := func(data rx.T) bool { return data.(int) %2 == 0 }
	sum := func(a, b rx.T) rx.T { return a.(int) + b.(int) }
	add := func(n int) rx.MapFn {
		return func(data rx.T) rx.T { return data.(int) + n }
	}

	append := func(c string) rx.MapFn {
		return func(data rx.T) rx.T { return data.(string) + c }
	}

	addOrAppend := func(n int, c string) rx.MapFn {
		return func(data rx.T) rx.T {
			if num, ok := data.(int); ok {
				return num + n
			}
			if letter, ok := data.(string); ok {
				return letter + "_"
			}
			return data
		}
	}

	alphabeticOrder := func(a, b rx.T) bool {
		return a.(string) < b.(string)
	}

	listen := func() (net.Listener, string) {
		port := ":8080"
		ln, err := net.Listen("tcp", port)
		if err != nil {
			port = ":8081"
			ln, _ = net.Listen("tcp", port)
		}
		return ln, port
	}

	Convey("rivers API", t, func() {

		Convey("From Range -> Filter -> Map -> Reduce -> Each -> Sink", func() {
			stream := rivers.New().FromRange(1, 5).
				Filter(evens).
				Map(add(1)).
				Reduce(0, sum).
				Sink()

			So(stream.Read(), ShouldResemble, []rx.T{8})
		})

		Convey("From Data -> Flatten -> Map -> Sort By -> Batch -> Sink", func() {
			stream := rivers.New().FromData([]rx.T{"a", "c"}, "b", []rx.T{"d", "e"}).
				Flatten().
				Map(append("_")).
				SortBy(alphabeticOrder).
				Batch(2).
				Sink()

			So(stream.Read(), ShouldResemble, []rx.T{
				[]rx.T{"a_", "b_"},
				[]rx.T{"c_", "d_"},
				[]rx.T{"e_"},
			})
		})

		Convey("From Slice -> Dispatch If -> Map -> Sink", func() {
			in, out := rx.NewStream(2)

			notDispatched := rivers.New().FromSlice([]rx.T{1, 2, 3, 4, 5}).
				DispatchIf(evens, out).
				Map(add(2)).
				Sink()

			So(in.Read(), ShouldResemble, []rx.T{2, 4})
			So(notDispatched.Read(), ShouldResemble, []rx.T{3, 5, 7})
		})

		Convey("Combine Zipping -> Map -> Sink", func() {
			streams := rivers.New()
			numbers := streams.FromData(1, 2, 3, 4)
			letters := streams.FromData("a", "b", "c")

			combined := streams.CombineZipping(numbers.Sink(), letters.Sink()).Map(addOrAppend(1, "_")).Sink()

			So(combined.Read(), ShouldResemble, []rx.T{2, "a_", 3, "b_", 4, "c_", 5})
		})

		Convey("Combine Zipping By -> Map -> Sink", func() {
			streams := rivers.New()
			numbers := streams.FromData(1, 2, 3, 4)
			moreNumbers := streams.FromData(4, 4, 1)

			combined := streams.CombineZippingBy(sum, numbers.Sink(), moreNumbers.Sink()).Filter(evens).Sink()

			So(combined.Read(), ShouldResemble, []rx.T{6, 4, 4})
		})

		Convey("From Data -> Drain", func() {
			numbers := rivers.New().FromData(1, 2, 3, 4)
			numbers.Drain()

			data, opened := <-numbers.Sink()
			So(data, ShouldBeNil)
			So(opened, ShouldBeFalse)
		})

		Convey("From Socket -> Map -> Filter -> Map -> Flatten -> Sink", func() {
			ln, port := listen()

			go func() {
				conn, _ := ln.Accept()
				defer conn.Close()
				conn.Write([]byte("Hello there\n"))
				conn.Write([]byte("\n"))
				conn.Write([]byte("rivers!\n"))
				conn.Write([]byte("super cool!\n"))
			}()

			words := rivers.New().FromSocketWithScanner("tcp", port, scanners.NewLineScanner()).
				Map(toString).
				Filter(nonEmptyLines).
				Map(splitWord).
				Flatten().
				Sink()

			So(words.Read(), ShouldResemble, []rx.T{"Hello", "there", "rivers!", "super", "cool!"})
		})
	})
}