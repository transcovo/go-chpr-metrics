package metrics

import (
	"encoding/json"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/alexcesaro/statsd"
	"github.com/transcovo/go-chpr-logger"
)

func defaultClientErrorHandler(err error) {
	// Do we choose to make the metrics depend on the logger ?
	// That sounds like a bad idea to me but this would be useful to have logs about the connection to the statsd server
	// fmt.Println("Error caught in metrics", error)
	logger.WithField("error", err).Error("[METRICS] Error caught in metrics")
}

/*
Sender is a struct responsible for keeping all the statsd clients
*/
type Sender struct {
	Clients []*statsd.Client
}

/*
Array of metricDestination (retrieve each metric destination from METRIC_DESTINATIONS)
*/
type metricsDestinations []*metricsDestination

/*
Struct target from METRICS_DESTINATION env var
*/
type metricsDestination struct {
	Host   string `json:"METRICS_HOST"`
	Port   string `json:"METRICS_PORT"`
	Prefix string `json:"METRICS_PREFIX"`
}

/*
The config to pass to a client builder
*/
type clientConfig struct {
	Host, Prefix string
}

/*
Return a client for a metric destination struct
*/
func clientConfigFromDestination(destination *metricsDestination) *clientConfig {
	return &clientConfig{
		Host:   destination.Host + ":" + destination.Port,
		Prefix: destination.Prefix,
	}
}

/*
Struct responsible for building clients
*/
type clientBuilder struct {
	errorLogger func(error)
}

/*
Return a statsd client from our custom clientConfig
statsd.new will resolve the given host and panics if it fails
The statsd lib we use sends a message upon connection
if there is nothing listening it may crash
TODO : check that this code is removed https://github.com/alexcesaro/statsd/blob/master/conn.go#L50
*/
func (builder *clientBuilder) buildClient(config *clientConfig) *statsd.Client {
	client, err := statsd.New(
		statsd.Address(config.Host),
		statsd.Prefix(config.Prefix),
		statsd.ErrorHandler(builder.errorLogger),
	)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"config": config,
			"error":  err,
		}).Error("Error creating the statsd client")
		panic("Error creating statsd client")
	}

	return client
}

/*
Initialize the client from the standard configuration of metrics (METRICS_HOST, METRICS_PORT, METRICS_PREFIX)
*/
func (sender *Sender) initStandardClient(clientBuilder *clientBuilder) {
	host := os.Getenv("METRICS_HOST")

	if host == "" {
		logger.Info("[METRICS] METRICS_HOST empty, not initializing a client for the standard configuration")
		return
	}
	port := os.Getenv("METRICS_PORT")
	prefix := os.Getenv("METRICS_PREFIX")

	if port == "" || prefix == "" {
		panic("[METRICS] Basic configuration can not have any empty port or prefix")
	}

	basicConfig := &clientConfig{
		Host:   host + ":" + port,
		Prefix: prefix,
	}
	sender.Clients = append(sender.Clients, clientBuilder.buildClient(basicConfig))
}

/*
Initialize one or several clients from the METRICS_DESTINATIONS env var
*/
func (sender *Sender) initDestinationClients(clientBuilder *clientBuilder) {
	metricsDestinationVar := os.Getenv("METRICS_DESTINATIONS")

	if metricsDestinationVar == "" {
		logger.Info("[METRICS] METRICS_DESTINATIONS empty, not initializing a client for the advanced configuration")
		return
	}
	destinations := &metricsDestinations{}
	err := json.Unmarshal([]byte(metricsDestinationVar), destinations)

	if err != nil {
		logger.WithFields(logrus.Fields(map[string]interface{}{
			"metrics_destinations": metricsDestinationVar,
			"error":                err,
		})).Error("Error parsing env METRICS_DESTINATION")
		panic("[METRICS] Error creating statsd client - JSON unmarshalling failed")
	}

	for _, destination := range *destinations {
		config := clientConfigFromDestination(destination)
		sender.Clients = append(sender.Clients, clientBuilder.buildClient(config))
	}
}
