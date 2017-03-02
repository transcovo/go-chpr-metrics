package metrics

import (
	"time"

	"github.com/transcovo/go-chpr-logger"
)

var (
	defaultClientBuilder *clientBuilder
	sender               = &Sender{}
	isSenderInitialized  = false
)

/*
Init function
- init a default client builder
*/
func init() {
	defaultClientBuilder = &clientBuilder{
		errorLogger: defaultClientErrorHandler,
	}
}

/*
Initialize the clients from the two possible configurations
*/
func (sender *Sender) initSender() {
	sender.initStandardClient(defaultClientBuilder)
	sender.initDestinationClients(defaultClientBuilder)
	if len(sender.Clients) == 0 {
		logger.Error("[METRICS] No metrics client initialized")
		panic("[METRICS] No metrics client initialized")
	}
}

/*
GetMetricsSender returns a handler (Singleton) on the metrics sender
*/
func GetMetricsSender() *Sender {
	if !isSenderInitialized {
		sender.initSender()
		isSenderInitialized = true
	}
	return sender
}

/*
Count sends a count metric
*/
func (sender *Sender) Count(bucket string, n interface{}) {
	for _, client := range sender.Clients {
		client.Count(bucket, n)
	}
}

/*
Increment sends an increment metric (a Count with 1 as the quantifier)
*/
func (sender *Sender) Increment(bucket string) {
	for _, client := range sender.Clients {
		client.Increment(bucket)
	}
}

/*
A Timing is an helper object that eases sending timing values.
*/
type Timing struct {
	start  time.Time
	sender *Sender
}

/*
NewTiming generates a timing object
Call it where you need to start timing, then call send on the returned object
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
	duration := int(timing.Duration() / time.Millisecond)
	for _, client := range timing.sender.Clients {
		client.Timing(bucket, duration)
	}
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
