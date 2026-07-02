package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	requestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests by method, path and status.",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: []float64{0.05, 0.1, 0.2, 0.3, 0.5, 1, 2, 5},
		},
		[]string{"method", "path"},
	)

	ordersCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of orders created.",
		},
	)

	ordersFailed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_failed_total",
			Help: "Total number of failed order requests.",
		},
	)

	inventoryLookup = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "inventory_lookup_total",
			Help: "Total number of inventory lookups.",
		},
	)
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

func normalizePath(path string) string {
	if strings.HasPrefix(path, "/orders/") {
		return "/orders/{id}"
	}
	return path
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rr, r)
		path := normalizePath(r.URL.Path)
		status := strconv.Itoa(rr.statusCode)
		requestCounter.WithLabelValues(r.Method, path, status).Inc()
		requestDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
	})
}

func writeJSON(w http.ResponseWriter, status int, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "orders-service"})
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		ordersCreated.Inc()
		writeJSON(w, http.StatusCreated, map[string]any{"message": "order created", "orderId": rand.Intn(100000)})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"orders": []string{"ORD-1001", "ORD-1002", "ORD-1003"}})
}

func orderByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/orders/")
	writeJSON(w, http.StatusOK, map[string]any{"orderId": id, "status": "confirmed"})
}

func inventoryHandler(w http.ResponseWriter, r *http.Request) {
	inventoryLookup.Inc()
	writeJSON(w, http.StatusOK, map[string]any{"sku": "SKU-1001", "available": true, "quantity": rand.Intn(100)})
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	ordersFailed.Inc()
	writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "simulated internal server error"})
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	delay := time.Duration(500+rand.Intn(2500)) * time.Millisecond
	time.Sleep(delay)
	writeJSON(w, http.StatusOK, map[string]any{"message": "slow response", "delay": delay.String()})
}

func initTracer(ctx context.Context) func(context.Context) error {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		log.Println("OTEL_EXPORTER_OTLP_ENDPOINT not set; traces will use no-op provider")
		return func(context.Context) error { return nil }
	}

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		log.Printf("failed to create OTLP exporter: %v", err)
		return func(context.Context) error { return nil }
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("orders-service"),
			semconv.ServiceVersion("v1"),
		),
	)
	if err != nil {
		log.Printf("failed to create OTEL resource: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx := context.Background()
	shutdown := initTracer(ctx)
	defer func() { _ = shutdown(ctx) }()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/orders", ordersHandler)
	mux.HandleFunc("/orders/", orderByIDHandler)
	mux.HandleFunc("/inventory", inventoryHandler)
	mux.HandleFunc("/error", errorHandler)
	mux.HandleFunc("/slow", slowHandler)
	mux.Handle("/metrics", promhttp.Handler())

	wrapped := metricsMiddleware(otelhttp.NewHandler(mux, "orders-service-http"))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("orders-service listening on :%s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), wrapped))
}
