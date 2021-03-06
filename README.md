# Go-chpr-metrics

[![CircleCI](https://circleci.com/gh/transcovo/go-chpr-metrics.svg?style=shield)](https://circleci.com/gh/transcovo/go-chpr-metrics)
[![codecov](https://codecov.io/gh/transcovo/go-chpr-metrics/branch/master/graph/badge.svg)](https://codecov.io/gh/transcovo/go-chpr-metrics)
[![GoDoc](https://godoc.org/github.com/transcovo/go-chpr-metrics?status.svg)](https://godoc.org/github.com/transcovo/go-chpr-metrics)

----------------- 

This utility library implements our standard statsd configuration. 
The base for this library is our [nodejs statsd client](https://github.com/transcovo/chpr-metrics) 

See the [godoc](https://godoc.org/github.com/transcovo/go-chpr-metrics)

## Requirements 
 
Minimum Go version: 1.7 
 
## Installation 
 
if using govendor 
```bash 
govendor fetch github.com/transcovo/go-chpr-metrics 
``` 
 
standard way (not recommended) 
```bash 
go get -u github.com/transcovo/go-chpr-metrics 
``` 
 
## Configuration 
 
This simple configuration allows you to send metrics to a single statds server 
 
* METRICS_HOST 
* METRICS_PORT 
* METRICS_PREFIX (prepended to all metrics name - we don't had a 'dot' by default)  
 
## Advanced configuration 
 
If you need to send metrics to several destinations, you can use the METRICS_DESTINATIONS 
variable, which allows you to specify a list of destinations as a JSON array: 
 
Complete example: 
 
    [{ 
      "METRICS_HOST": "host1.yourstats.com", 
      "METRICS_PORT": "1234", 
      "METRICS_PREFIX": "prefix1." 
    }, { 
      "METRICS_HOST": "your-other-host.com", 
      "METRICS_PORT": "44444", 
      "METRICS_PREFIX": "anotherPrefix." 
    }] 
 
If you use this in combination with the individual configuration variables listed 
in the previous sections, it will send metrics to all directions. 
 
**NB**: This variable should be stringified to be set as a Environment Var 
 
## How to use 
 
If you import this library, you need to have at least the standard or the advanced configuration filled. 
If you don't, when you try to require the handler with `GetMetricsSender()`, the program will panic. 
 
```golang 
import "github.com/transcovo/go-chpr-metrics"
 
// Count: Increments a stat by a value (default is 1) 
metrics.Count('my_counter', 3) 
 
// Increment: Increments a stat by a value (default is 1) 
// Special case of count with value set to 1 
metrics.Increment('my_counter') 

// Duration: send a timing based on a duration.
metrics.Duration("my_timing", endTime.Sub(startTime))

// Timing: first instantiate a timer object, then call the send function of this object 
timer := metrics.NewTiming() 
someTask() 
timer.Send('someTask.timing') 
``` 
 
The exported object is a handler on a multi statsd client: see https://github.com/alexcesaro/statsd 
 
## Misc 
 
The policy for this lib regarding vendoring is not to include any dependency.  
The main reason for this is to avoid any conflict between your project and go-chpr-metrics. 
