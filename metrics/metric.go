//Package metrics bootstrap metrics reporter, and supply 2 metrics registry
//native prometheus registry and rcrowley/go-metrics registry
//system registry is the place where go-chassis feed metrics data to
//you can get system registry and report them to varies monitoring system
package metrics

import (
	"errors"
	"sync"

	"github.com/ServiceComb/go-chassis/core/lager"

	"github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rcrowley/go-metrics"
)

var errMonitoringFail = errors.New("can not report metrics to CSE monitoring service")

// constants for header parameters
const (
	defaultName = "default_metric_registry"
	// Metrics is the constant string
	Metrics = "PrometheusMetrics"
)

var metricRegistries = make(map[string]metrics.Registry)
var prometheusRegistry = prometheus.NewRegistry()
var l sync.RWMutex

//GetSystemRegistry return metrics registry which go chassis use
func GetSystemRegistry() metrics.Registry {
	return GetOrCreateRegistry(defaultName)
}

//GetSystemPrometheusRegistry return prometheus registry which go chassis use
func GetSystemPrometheusRegistry() *prometheus.Registry {
	return prometheusRegistry
}

//GetOrCreateRegistry return a go-metrics registry which go chassis framework use to report metrics
func GetOrCreateRegistry(name string) metrics.Registry {
	r, ok := metricRegistries[name]
	if !ok {
		l.Lock()
		r = metrics.NewRegistry()
		metricRegistries[name] = r
		l.Unlock()
	}
	return r
}

// HTTPHandleFunc is a go-restful handler which can expose metrics in http server
func HTTPHandleFunc(req *restful.Request, rep *restful.Response) {
	promhttp.HandlerFor(GetSystemPrometheusRegistry(), promhttp.HandlerOpts{}).ServeHTTP(rep.ResponseWriter, req.Request)
}

//Init prepare the metrics registry and report metrics to other systems
func Init() error {
	metricRegistries[defaultName] = metrics.DefaultRegistry
	for k, report := range reporterPlugins {
		lager.Logger.Info("report metrics to " + k)
		if err := report(GetSystemRegistry()); err != nil {
			lager.Logger.Warnf(err.Error(), err)
			return err
		}
	}
	return nil
}
