package metrics_test

import (
	"fmt"
	"time"

	metrics "github.com/transcovo/go-chpr-metrics"
)

func Example() {
	// Count: Increments a stat by a value (default is 1)
	metrics.Count("my_counter", 3)

	// Increment: Increments a stat by a value (default is 1)
	// Special case of count with value set to 1
	metrics.Increment("my_counter")

	// Timing: first instantiate a timer object, then call the send function of this object
	timer := metrics.NewTiming()
	//
	func() {
		fmt.Println("I am doing some task")
		time.Sleep(time.Second * 1)
		fmt.Println("I am done doing some task")
	}()
	timer.Send("someTask.timing")

	// You can also use defer to time a function run time
	defer metrics.NewTiming().Send("someTask.timing")
}

/*
This example presents a standard way to use the metrics lib
*/
func ExampleGetMetricsSender() {
	metrics := metrics.GetMetricsSender()

	// Count: Increments a stat by a value (default is 1)
	metrics.Count("my_counter", 3)

	// Increment: Increments a stat by a value (default is 1)
	// Special case of count with value set to 1
	metrics.Increment("my_counter")

	// Timing: first instantiate a timer object, then call the send function of this object
	timer := metrics.NewTiming()
	//
	func() {
		fmt.Println("I am doing some task")
		time.Sleep(time.Second * 1)
		fmt.Println("I am done doing some task")
	}()
	timer.Send("someTask.timing")
}
