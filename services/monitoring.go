package services

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	serviceContext "github.com/cloakd/common/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

const (
	MONITORING_SVC          = "monitoring_svc"
	SERVICE_NAME            = "ven_backend"
	DEFAULT_PROMETHEUS_PORT = 2112
)

// HTTP Metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"endpoint", "method", "status"},
	)

	httpRequestsSuccessfulTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_successful_total",
			Help: "Total successful HTTP requests (2xx status codes)",
		},
		[]string{"endpoint", "method"},
	)

	httpRequestsFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_failed_total",
			Help: "Total failed HTTP requests (4xx, 5xx status codes)",
		},
		[]string{"endpoint", "method"},
	)

	httpRequestsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_active",
			Help: "Number of active concurrent HTTP requests",
		},
		[]string{"endpoint", "method"},
	)

	httpRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"endpoint", "method", "status"},
	)

	httpResponseSizeBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response payload size in bytes",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
		},
		[]string{"endpoint", "method"},
	)
)

// System Metrics
var (
	heapAllocBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "heap_alloc_bytes",
			Help: "Heap memory allocated in bytes",
		},
	)

	heapSysBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "heap_sys_bytes",
			Help: "Heap memory obtained from system in bytes",
		},
	)

	gcTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gc_total",
			Help: "Total number of garbage collections",
		},
	)

	memoryUsageBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_usage_bytes",
			Help: "Current memory usage in bytes",
		},
	)

	memoryUsagePercent = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_usage_percent",
			Help: "Memory usage percentage",
		},
	)
)

// Trace Metrics
var (
	traceSpanDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "trace_span_duration_seconds",
			Help:    "Span duration in seconds for distributed tracing",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"service", "operation", "span_kind"},
	)
)

type MonitoringService struct {
	serviceContext.DefaultService

	port     int
	register *prometheus.Registry

	closed      chan struct{}
	server      *fiber.App
	lastGCCount uint32
}

func (svc *MonitoringService) Id() string {
	return MONITORING_SVC
}

func (svc *MonitoringService) Start() error {
	svc.closed = make(chan struct{}, 1)

	portStr := os.Getenv("PROMETHEUS_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = DEFAULT_PROMETHEUS_PORT
	}
	svc.port = port

	// Create new registry
	reg := prometheus.NewRegistry()

	// Register default collectors (includes Go runtime metrics like memory)
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	// Register custom metrics
	reg.MustRegister(
		httpRequestsTotal,
		httpRequestsSuccessfulTotal,
		httpRequestsFailedTotal,
		httpRequestsActive,
		httpRequestDurationSeconds,
		httpResponseSizeBytes,
		heapAllocBytes,
		heapSysBytes,
		gcTotal,
		memoryUsageBytes,
		memoryUsagePercent,
		traceSpanDurationSeconds,
	)

	svc.register = reg

	svc.initializeMetrics()

	// Start memory metrics updater
	go svc.updateMemoryMetrics()

	config := fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		},
	}

	svc.server = fiber.New(config)
	svc.server.Use(recover.New())

	svc.server.Get("/metrics", svc.metricsHandler)
	svc.server.Get("/health", svc.healthHandler)

	log.Info().Int("port", svc.port).Msg("Prometheus metrics server started")
	return svc.server.Listen(fmt.Sprintf(":%v", svc.port))
}

func (svc *MonitoringService) Shutdown() {
	svc.closed <- struct{}{}
	if svc.server != nil {
		_ = svc.server.Shutdown()
	}
}

func (svc *MonitoringService) metricsHandler(c *fiber.Ctx) error {
	handler := promhttp.HandlerFor(svc.register, promhttp.HandlerOpts{})
	return adaptor.HTTPHandler(handler)(c)
}

func (svc *MonitoringService) healthHandler(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":    "healthy",
		"service":   SERVICE_NAME,
		"timestamp": time.Now().Unix(),
	})
}

func (svc *MonitoringService) initializeMetrics() {
	// Initialize HTTP metrics with zero values
	httpRequestsTotal.WithLabelValues("/health", "GET", "200").Add(0)
	httpRequestsSuccessfulTotal.WithLabelValues("/health", "GET").Add(0)
	httpRequestsActive.WithLabelValues("/health", "GET").Set(0)
	httpRequestDurationSeconds.WithLabelValues("/health", "GET", "200").Observe(0)
	httpResponseSizeBytes.WithLabelValues("/health", "GET").Observe(0)

	// Initialize system metrics
	heapAllocBytes.Set(0)
	heapSysBytes.Set(0)
	memoryUsageBytes.Set(0)
	memoryUsagePercent.Set(0)

	log.Info().Msg("Metrics initialized successfully")
}

// updateMemoryMetrics updates memory-related metrics every 15 seconds
func (svc *MonitoringService) updateMemoryMetrics() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	log.Info().Msg("Memory metrics updater started")

	for {
		select {
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			// Update heap metrics
			heapAllocBytes.Set(float64(m.Alloc))
			heapSysBytes.Set(float64(m.Sys))

			// Update GC count (only increment difference)
			if m.NumGC > svc.lastGCCount {
				gcTotal.Add(float64(m.NumGC - svc.lastGCCount))
				svc.lastGCCount = m.NumGC
			}

			// Update memory usage
			memoryUsageBytes.Set(float64(m.Alloc))

			// Calculate memory usage percentage (approximate)
			memPercent := float64(m.Alloc) / float64(m.Sys) * 100
			if memPercent > 100 {
				memPercent = 100
			}
			memoryUsagePercent.Set(memPercent)

		case <-svc.closed:
			log.Info().Msg("Memory metrics updater stopped")
			return
		}
	}
}

// RecordRequest records HTTP request metrics
func (svc *MonitoringService) RecordRequest(method, endpoint, status string, duration time.Duration, responseSize int) {
	httpRequestsTotal.WithLabelValues(endpoint, method, status).Inc()
	httpRequestDurationSeconds.WithLabelValues(endpoint, method, status).Observe(duration.Seconds())
	httpResponseSizeBytes.WithLabelValues(endpoint, method).Observe(float64(responseSize))

	statusCode, _ := strconv.Atoi(status)
	if statusCode >= 200 && statusCode < 400 {
		httpRequestsSuccessfulTotal.WithLabelValues(endpoint, method).Inc()
	} else if statusCode >= 400 {
		httpRequestsFailedTotal.WithLabelValues(endpoint, method).Inc()
	}
}

// IncrementActiveRequests increments the active requests gauge
func (svc *MonitoringService) IncrementActiveRequests(endpoint, method string) {
	httpRequestsActive.WithLabelValues(endpoint, method).Inc()
}

// DecrementActiveRequests decrements the active requests gauge
func (svc *MonitoringService) DecrementActiveRequests(endpoint, method string) {
	httpRequestsActive.WithLabelValues(endpoint, method).Dec()
}

// RecordTraceSpan records a trace span duration
func (svc *MonitoringService) RecordTraceSpan(service, operation, spanKind string, duration time.Duration) {
	traceSpanDurationSeconds.WithLabelValues(service, operation, spanKind).Observe(duration.Seconds())
}

// MonitoringMiddleware creates a Fiber middleware for monitoring HTTP requests
func MonitoringMiddleware(monitoringSvc *MonitoringService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip metrics endpoint
		if c.Path() == "/metrics" {
			return c.Next()
		}

		start := time.Now()
		endpoint := c.Route().Path // Get route pattern, not actual path
		method := c.Method()

		// Increment active requests
		monitoringSvc.IncrementActiveRequests(endpoint, method)
		defer monitoringSvc.DecrementActiveRequests(endpoint, method)

		// Process request
		err := c.Next()

		// Calculate duration and response size
		duration := time.Since(start)
		status := strconv.Itoa(c.Response().StatusCode())
		responseSize := len(c.Response().Body())

		// Record metrics
		monitoringSvc.RecordRequest(method, endpoint, status, duration, responseSize)

		// Record trace span
		monitoringSvc.RecordTraceSpan(SERVICE_NAME, endpoint, "server", duration)

		return err
	}
}
