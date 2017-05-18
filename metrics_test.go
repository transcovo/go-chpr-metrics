package metrics

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/transcovo/go-chpr-logger"
	"github.com/transcovo/go-chpr-metrics/utils/tests"
)

var port = 12345
var localhost = "127.0.0.1"

/*
Util function to restore the metrics handler
*/
func resetSender() {
	sender = nil
}

/*
Ensure uniqueness in port number (tests are ran in parallel)
*/
func getUniquePort() string {
	port++
	return strconv.Itoa(port)
}

/*
Start an UDP server that does nothing. Waits for the test to finish to close the server
Notify the channel when cleaning is done
*/
func simpleUDPServer(host, port string, end chan string) {
	pc, err := net.ListenPacket("udp", host+":"+port)
	if err != nil {
		fmt.Println("Error creating the udp server", err)
		panic(err)
	}
	end <- "ready" // Notify that ready
	<-end          // Wait for the test to notify that it's finishing
	pc.Close()     // We do not use defer here because we need to close cleanly the connection
	end <- "closed"
}

/*
Notify and wait for a answer
*/
func notifyToClose(end chan string) {
	end <- "test ended"
	<-end
}

/*
Test the defaultClientErrorHandler
*/
func TestDefaultClientErrorHandler(t *testing.T) {
	err := errors.New("This is an error")

	stdout := tests.CaptureStdout(func() { defaultClientErrorHandler(err) })
	assert.Contains(t, stdout, `msg="[METRICS] Error caught in metrics" error="This is an error"`)
}

/**
This tests expect buildClient to fail if the fields (like the port) in the config object are ill formatted
*/
func TestBuildClient_Failure(t *testing.T) {
	wrongConfig := &clientConfig{
		Host:   "an ill formatted host",
		Prefix: "prefix.",
	}

	assert.Panics(t, func() {
		defaultClientBuilder.buildClient(wrongConfig)
	})
}

/**
Tests that with a started udp server, the statsd client is correctly started
*/
func TestBuildClient_Success(t *testing.T) {
	port := getUniquePort()
	host := localhost + ":" + port

	end := make(chan string)
	go simpleUDPServer(localhost, port, end)
	<-end
	correctConfig := &clientConfig{
		Host:   host,
		Prefix: "prefix.",
	}

	mockClientBuilder := &clientBuilder{
		errorLogger: func(err error) {
			logger.WithField("error", err).Error("this is some error")
			panic("It should not happen")
		},
	}

	client := mockClientBuilder.buildClient(correctConfig)
	assert.NotNil(t, client)
	notifyToClose(end)
}

/*
Tests that the configuration is built correctly from a metricsDestination object
*/
func TestClientConfigFromDestination_Success(t *testing.T) {
	destination := &metricsDestination{
		Host:   "some.host",
		Port:   "1234",
		Prefix: "prefix.",
	}

	expectedConfig := &clientConfig{
		Host:   "some.host:1234",
		Prefix: "prefix.",
	}

	config := clientConfigFromDestination(destination)
	assert.Equal(t, config, expectedConfig)
}

/*
Tests that if there is only METRICS_HOST (but not METICS_PORT or METRICS_PREFIX), it panics
*/
func TestInitStandardClient_Failure(t *testing.T) {
	resetSender()

	env := map[string]string{
		"METRICS_HOST": "127.0.0.1",
		"METRICS_PORT": "",
	}

	tests.WithEnvVars(env, func() {
		assert.Panics(t, func() { GetMetricsSender() })
	})
	assert.Len(t, sender.Clients, 0)
}

/*
Tests that if there is only METRICS_HOST (as well as PORT and PREFIX), there is only one client
*/
func TestInitStandardClient_Success(t *testing.T) {
	resetSender()

	port := getUniquePort()
	end := make(chan string)
	go simpleUDPServer(localhost, port, end)
	<-end

	env := map[string]string{
		"METRICS_HOST":   localhost,
		"METRICS_PORT":   port,
		"METRICS_PREFIX": "prefix1.",
	}
	tests.WithEnvVars(env, func() {
		GetMetricsSender()
	})
	assert.Len(t, sender.Clients, 1)
	notifyToClose(end)
}

/*
Tests that if there is only METRICS_HOST (as well as PORT and PREFIX), there is only one client
*/
func TestInitDestinationClients_FailureErrorJSON(t *testing.T) {
	resetSender()

	env := map[string]string{
		"METRICS_DESTINATIONS": "[{",
	}

	tests.WithEnvVars(env, func() {
		assert.Panics(t, func() {
			GetMetricsSender()
		})
	})
	assert.Len(t, sender.Clients, 0)
}

/*
Tests that if there is only METRICS_DESTINATIONS, with one element in the array, it builds one client
*/
func TestInitDestinationClients_SuccessOneDestination(t *testing.T) {
	resetSender()

	end := make(chan string)
	port := getUniquePort()
	go simpleUDPServer(localhost, port, end)
	<-end

	env := map[string]string{
		"METRICS_DESTINATIONS": fmt.Sprintf(`[{"METRICS_HOST": "127.0.0.1","METRICS_PORT":"%s","METRICS_PREFIX": "testPrefix."}]`, port),
	}

	tests.WithEnvVars(env, func() {
		GetMetricsSender()
	})
	assert.Len(t, sender.Clients, 1)
	notifyToClose(end)
}

/*
Tests that if there is only METRICS_DESTINATIONS, with twos elements in the array, it builds two clients
*/
func TestInitDestinationClients_SuccessMultipleDestinations(t *testing.T) {
	resetSender()
	end1 := make(chan string)
	end2 := make(chan string)
	port1 := getUniquePort()
	port2 := getUniquePort()
	go simpleUDPServer(localhost, port1, end1)
	go simpleUDPServer(localhost, port2, end2)
	<-end1
	<-end2

	env := map[string]string{
		"METRICS_DESTINATIONS": fmt.Sprintf(`[{"METRICS_HOST": "127.0.0.1","METRICS_PORT":"%s","METRICS_PREFIX": "testPrefix."}, {"METRICS_HOST": "127.0.0.1","METRICS_PORT":"%s","METRICS_PREFIX": "testPrefix2."}]`, port1, port2),
	}

	tests.WithEnvVars(env, func() {
		GetMetricsSender()
	})
	assert.Len(t, sender.Clients, 2)

	notifyToClose(end1)
	notifyToClose(end2)
}

/*
Tests that the program panic if there is no client initialized
*/
func TestInitSender_EmptyConfig(t *testing.T) {
	resetSender()
	env := map[string]string{
		"METRICS_DESTINATIONS": "",
		"METRICS_HOST":         "",
	}
	tests.WithEnvVars(env, func() {
		GetMetricsSender()
	})
	assert.Len(t, sender.Clients, 0)
}

/*
Test that the singleton pattern is correctly implemented and it returns the metrics handler
*/
func TestGetMetricsSender_SingletonSuccess(t *testing.T) {
	resetSender()
	end := make(chan string)
	port := getUniquePort()
	go simpleUDPServer(localhost, port, end)
	<-end

	env := map[string]string{
		"METRICS_HOST":   localhost,
		"METRICS_PORT":   port,
		"METRICS_PREFIX": "prefix1.",
	}
	var metricsSender1 *Sender
	var metricsSender2 *Sender
	tests.WithEnvVars(env, func() {
		metricsSender1 = GetMetricsSender()
		metricsSender2 = GetMetricsSender()
	})
	assert.NotNil(t, metricsSender1)
	assert.Equal(t, metricsSender1, metricsSender2)
	assert.Equal(t, sender, metricsSender1)

	notifyToClose(end)
}

/*
Tests that an increment sends a metric
*/
func TestIncrement_Success(t *testing.T) {
	resetSender()

	env, pc, tearDown := setupTestUDPServer()
	defer tearDown()

	assert.Nil(t, sender)

	var payload string
	tests.WithEnvVars(env, func() {
		GetMetricsSender()

		Increment("test.increment")
		buffer := make([]byte, 1024)
		var bytesReadCount int
		for {
			bytesReadCount, _, _ = pc.ReadFrom(buffer)
			if bytesReadCount != 0 {
				break
			}
		}
		payload = string(buffer)[:bytesReadCount]
	})
	assert.Equal(t, "prefix1.test.increment:1|c", payload)
}

/*
Tests that a count sends a metric
*/
func TestCount_Success(t *testing.T) {
	resetSender()

	env, pc, tearDown := setupTestUDPServer()
	defer tearDown()

	assert.Nil(t, sender)

	var payload string
	tests.WithEnvVars(env, func() {
		GetMetricsSender()

		Count("test.count", 3)

		buffer := make([]byte, 1024)
		var bytesReadCount int
		for {
			bytesReadCount, _, _ = pc.ReadFrom(buffer)
			if bytesReadCount != 0 {
				break
			}
		}
		payload = string(buffer)[:bytesReadCount]
		assert.Equal(t, "prefix1.test.count:3|c", payload)
	})
}

/*
Tests that a count sends a metric
*/
func TestGauge_Success(t *testing.T) {
	resetSender()

	env, pc, tearDown := setupTestUDPServer()
	defer tearDown()

	assert.Nil(t, sender)

	var payload string
	tests.WithEnvVars(env, func() {
		GetMetricsSender()

		Gauge("test.gauge", 123)

		buffer := make([]byte, 1024)
		var bytesReadCount int
		for {
			bytesReadCount, _, _ = pc.ReadFrom(buffer)
			if bytesReadCount != 0 {
				break
			}
		}
		payload = string(buffer)[:bytesReadCount]
	})
	assert.Equal(t, "prefix1.test.gauge:123|g", payload)
}

/*
TestDuration_Success tests that Duration sends a timing metric
*/
func TestDuration_Success(t *testing.T) {
	resetSender()

	env, pc, tearDown := setupTestUDPServer()
	defer tearDown()

	assert.Nil(t, sender)

	var payload string
	tests.WithEnvVars(env, func() {
		GetMetricsSender()

		Duration(time.Microsecond*123456, "test.duration")

		buffer := make([]byte, 1024)
		var bytesReadCount int
		for {
			bytesReadCount, _, _ = pc.ReadFrom(buffer)
			if bytesReadCount != 0 {
				break
			}
		}
		payload = string(buffer)[:bytesReadCount]
	})
	assert.Equal(t, "prefix1.test.duration:123|ms", payload)
}

/*
Tests that a Timing sends a metric
*/
func TestTiming_Success(t *testing.T) {
	resetSender()

	env, pc, tearDown := setupTestUDPServer()
	defer tearDown()

	assert.Nil(t, sender)

	var payload string
	tests.WithEnvVars(env, func() {
		GetMetricsSender()

		// stub now function
		start := time.Date(2017, 2, 27, 0, 0, 0, 0, time.UTC)
		now = func() time.Time { return start }
		timing := NewTiming()
		func() {
			// Simulate some other task
		}()
		now = func() time.Time { return start.Add(time.Second) }
		defer func() { now = time.Now }()
		timing.Send("test.timing")

		buffer := make([]byte, 1024)
		var bytesReadCount int
		for {
			bytesReadCount, _, _ = pc.ReadFrom(buffer)
			if bytesReadCount != 0 {
				break
			}
		}
		payload = string(buffer)[:bytesReadCount]
	})
	assert.Equal(t, "prefix1.test.timing:1000|ms", payload)
}

func setupTestUDPServer() (map[string]string, net.PacketConn, func() error) {
	port := getUniquePort()

	pc, err := net.ListenPacket("udp", localhost+":"+port)
	if err != nil {
		fmt.Println("Error creating the udp server", err)
		panic(err)
	}
	env := map[string]string{
		"METRICS_HOST":   localhost,
		"METRICS_PORT":   port,
		"METRICS_PREFIX": "prefix1.",
	}
	return env, pc, pc.Close
}
