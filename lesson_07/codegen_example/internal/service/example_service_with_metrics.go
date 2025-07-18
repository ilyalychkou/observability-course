// Code generated by gowrap. DO NOT EDIT.
// template: https://raw.githubusercontent.com/hexdigest/gowrap/629c2e966eaf72a2446886fd2dd4885ac1b3fbda/templates/prometheus
// gowrap: http://github.com/hexdigest/gowrap

package service

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ExampleServiceWithPrometheus implements ExampleService interface with all methods wrapped
// with Prometheus metrics
type ExampleServiceWithPrometheus struct {
	base         ExampleService
	instanceName string
}

var exampleserviceDurationSummaryVec = promauto.NewSummaryVec(
	prometheus.SummaryOpts{
		Name:       "exampleservice_duration_seconds",
		Help:       "exampleservice runtime duration and result",
		MaxAge:     time.Minute,
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	},
	[]string{"instance_name", "method", "result"})

// NewExampleServiceWithPrometheus returns an instance of the ExampleService decorated with prometheus summary metric
func NewExampleServiceWithPrometheus(base ExampleService, instanceName string) ExampleServiceWithPrometheus {
	return ExampleServiceWithPrometheus{
		base:         base,
		instanceName: instanceName,
	}
}

// ExampleMethod implements ExampleService
func (_d ExampleServiceWithPrometheus) ExampleMethod(ctx context.Context, param string) (err error) {
	_since := time.Now()
	defer func() {
		result := "ok"
		if err != nil {
			result = "error"
		}

		exampleserviceDurationSummaryVec.WithLabelValues(_d.instanceName, "ExampleMethod", result).Observe(time.Since(_since).Seconds())
	}()
	return _d.base.ExampleMethod(ctx, param)
}
