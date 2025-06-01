package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})

	httpRequestsLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_requests_latency_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: []float64{0.1, 0.25, 0.5, 0.75, 1},
	}, []string{"method", "path", "status"})
)

func main() {
	rand.Seed(time.Now().UnixNano())

	collector := newCollector()
	prometheus.MustRegister(collector)

	r := mux.NewRouter()
	r.Use(middleware)

	r.HandleFunc("/send-message/{userID}", sendMessageHandler).Methods("POST")
	r.Handle("/metrics", promhttp.Handler())

	fmt.Println("Service listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Message sent to %s\n", userID)))
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)

		duration := time.Since(start)
		statusCode := strconv.Itoa(ww.statusCode)

		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate() // Get the route pattern
		httpRequestsTotal.With(
			prometheus.Labels{"method": r.Method, "path": path, "status": statusCode},
		).Inc()

		httpRequestsLatency.With(
			prometheus.Labels{"method": r.Method, "path": path, "status": statusCode}).
			Observe(duration.Seconds())
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (ww *responseWriter) WriteHeader(statusCode int) {
	ww.statusCode = statusCode
	ww.ResponseWriter.WriteHeader(statusCode)
}

type Collector struct {
	heavyMetric *prometheus.Desc
	cache       float64
	mu          sync.RWMutex
}

func newCollector() *Collector {
	c := &Collector{
		heavyMetric: prometheus.NewDesc("my_heavy_metric",
			"Heavy metric calculated periodically in background",
			nil, nil,
		),
	}
	go c.backgroundUpdater()
	return c
}

func (c *Collector) backgroundUpdater() {
	for {
		value := calculateHeavyMetric()
		c.mu.Lock()
		c.cache = value
		c.mu.Unlock()
		time.Sleep(5 * time.Second)
	}
}

func (collector *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.heavyMetric
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	log.Println("Running my_heavy_metric ollection")
	c.mu.RLock()
	value := c.cache
	c.mu.RUnlock()

	ch <- prometheus.MustNewConstMetric(c.heavyMetric, prometheus.GaugeValue, value)
}

func calculateHeavyMetric() float64 {
	log.Println("Recalculating my_heavy_metric in background...")
	time.Sleep(time.Second * time.Duration(rand.Intn(5)+1)) // simulate slow logic
	return rand.Float64() * 100
}
