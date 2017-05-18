package metrics

import (
	"time"

	"github.com/transcovo/go-chpr-logger"
)

var (
	defaultClientBuilder *clientBuilder
	sender               *Sender
)

/*
Init function
- init a default client builder
*/
func init() {
	defaultClientBuilder = &clientBuilder{
		errorLogger: defaultClientErrorHandler,
	}
	sender = GetMetricsSender()
}

/*
Initialize the clients from the two possible configurations
*/
func (sender *Sender) initSender() {
	sender.initStandardClient(defaultClientBuilder)
	sender.initDestinationClients(defaultClientBuilder)
	if len(sender.Clients) == 0 {
		logger.Warning("[METRICS] No metrics client initialized")
	}
}

/*
GetMetricsSender returns a handler (Singleton) on the metrics sender
*/
func GetMetricsSender() *Sender {
	if sender == nil {
		sender = &Sender{}
		sender.initSender()
	}
	return sender
}

// Count sends a count metric
func Count(bucket string, n interface{}) { sender.Count(bucket, n) }

// Count sends a count metric
func (sender *Sender) Count(bucket string, n interface{}) {
	for _, client := range sender.Clients {
		client.Count(bucket, n)
	}
}

// Increment sends an increment metric (a Count with 1 as the quantifier)
func Increment(bucket string) { sender.Increment(bucket) }

// Increment sends an increment metric (a Count with 1 as the quantifier)
func (sender *Sender) Increment(bucket string) {
	for _, client := range sender.Clients {
		client.Increment(bucket)
	}
}

// Gauge sends a gauge metric
func Gauge(bucket string, value interface{}) { sender.Gauge(bucket, value) }

// Gauge sends a gauge metric
func (sender *Sender) Gauge(bucket string, value interface{}) {
	for _, client := range sender.Clients {
		client.Gauge(bucket, value)
	}
}

/*
A Timing is an helper object that eases sending timing values.
*/
type Timing struct {
	start  time.Time
	sender *Sender
}

// Duration can emit a timing metric from a duration.
func Duration(bucket string, duration time.Duration) { sender.Duration(bucket, duration) }

/*
Duration can emit a timing metric from a duration.
Typical use: metrics.SendDuration(toTime.Sub(startTime))
It is useful when the start time is not time.Now(), in which case NewTiming cannot apply.
*/
func (sender *Sender) Duration(bucket string, duration time.Duration) {
	durationMs := int(duration / time.Millisecond)
	for _, client := range sender.Clients {
		client.Timing(bucket, durationMs)
	}
}

/*
NewTiming generates a timing object.
Call it where you need to start timing, then call send on the returned object.
*/
func NewTiming() *Timing { return sender.NewTiming() }

/*
NewTiming generates a timing object
*/
func (sender *Sender) NewTiming() *Timing {
	return &Timing{
		start:  now(),
		sender: sender,
	}
}

/*
Send the timing metric from a previously generated Timing object
*/
func (timing *Timing) Send(bucket string) {
	timing.sender.Duration(bucket, timing.Duration())
}

/*
Duration returns the duration since Timing was generated
*/
func (timing *Timing) Duration() time.Duration {
	return now().Sub(timing.start)
}

// TODO : you want some more metrics ? be my PR-guest !

// Stubbed out for testing
var now = time.Now
