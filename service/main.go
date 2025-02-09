package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/middleware"
	"github.com/nibroos/s-erp-api/service/internal/routes"
	"github.com/nibroos/s-erp-api/service/internal/validators"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Use middleware to track metrics
// app.Use(func(c *fiber.Ctx) error {
// func PromDurationMiddleware(next fiber.Handler) fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		start := time.Now()
// 		err := c.Next()
// 		respStatus := c.Response().StatusCode()
// 		duration := time.Since(start)
// 		httpRequestDuration.With(map[string]string{
// 			"response_status": strconv.Itoa(respStatus),
// 		}).Observe(duration.Seconds())
// 		httpRequestsTotal.With(map[string]string{
// 			"response_status": strconv.Itoa(respStatus),
// 		}).Inc()

//			return err
//		}
//	}

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "response_status"},
	)
	// httpRequestDuration = promauto.NewHistogramVec(
	// 	prometheus.HistogramOpts{
	// 		Name:    "request_duration_seconds",
	// 		Buckets: []float64{.00005, .0005, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
	// 	},
	// 	[]string{"method", "endpoint", "response_status"},
	// )
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "response_status"},
	)
)

func PromDurationMiddleware(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next() // Call the next handler
	respStatus := c.Response().StatusCode()
	duration := time.Since(start)

	// Record metrics
	httpRequestDuration.WithLabelValues(c.Method(), c.Path(), strconv.Itoa(respStatus)).Observe(duration.Seconds())
	httpRequestsTotal.WithLabelValues(c.Method(), c.Path(), strconv.Itoa(respStatus)).Inc()

	return err
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	// Create a Prometheus registry
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry

	// Expose Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Println("Starting Prometheus metrics server on :9090")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Fatalf("Failed to start Prometheus metrics server: %v", err)
		}
	}()

	// Determine the environment (production or test)
	env := os.Getenv("APP_ENV")
	var dbURL string
	if env == "test" {
		dbURL = config.GetTestDatabaseURL()
	} else {
		dbURL = config.GetDatabaseURL()
	}

	// Initialize the SQLx database connection
	sqlDB, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to the SQL database: %v", err)
	}

	// Configure SQLx connection pool
	sqlDB.SetMaxOpenConns(100)          // Maximum number of open connections
	sqlDB.SetMaxIdleConns(10)           // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(time.Hour) // Maximum lifetime of a connection

	// Initialize the Gorm database connection
	gormDB, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the Gorm database: %v", err)
	}

	// Configure Gorm connection pool
	sqlDBGorm, err := gormDB.DB()
	if err != nil {
		log.Fatalf("Failed to get Gorm DB instance: %v", err)
	}
	sqlDBGorm.SetMaxOpenConns(100)          // Maximum number of open connections
	sqlDBGorm.SetMaxIdleConns(10)           // Maximum number of idle connections
	sqlDBGorm.SetConnMaxLifetime(time.Hour) // Maximum lifetime of a connection

	// Initialize the Redis client
	// if env == "test" {
	// 	config.InitRedisClientTest()
	// } else {
	// 	config.InitRedisClient()
	// }

	// Fetch needed data from the database and cache it in Redis
	config.FetchCachedData(context.Background(), sqlDB)

	// Initialize the validator with the database connection
	validators.InitValidator(sqlDB)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	app.Use(PromDurationMiddleware)

	// Attach middleware
	app.Use(middleware.ConvertEmptyStringsToNull())
	app.Use(middleware.ConvertRequestToFilters())

	// Setup REST routes
	routes.SetupRoutes(app, gormDB, sqlDB)

	// Protect routes with JWT middleware
	// app.Use(middleware.JWTMiddleware())

	var wg sync.WaitGroup

	// Check if the service type is "scheduler"
	if os.Getenv("SERVICE_TYPE") == "scheduler" {
		// Start the scheduler in a separate goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Initialize the cron scheduler
			cron := cron.New()
			schedulerController := rest.NewSchedulerController(cron, gormDB, sqlDB)

			// Reload schedules from the database
			if err := schedulerController.ReloadSchedules(); err != nil {
				log.Printf("Failed to reload schedules: %v", err)
				return
			}

			// Start the cron scheduler
			cron.Start()
			log.Println("Scheduler started successfully")
		}()
	} else {
		// Start REST server
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := app.Listen(":4001"); err != nil {
				log.Fatalf("Failed to start REST server: %v", err)
			}

			println("Server started on :4001")
		}()

		// Start gRPC server
		// wg.Add(1)
		// go func() {
		// 	defer wg.Done()
		// 	if err := runGRPCServer(grpcUserController); err != nil {
		// 		log.Fatalf("Failed to run gRPC server: %v", err)
		// 	}
		// }()
	}

	// Wait for all servers to exit
	wg.Wait()
}

// func runGRPCServer(grpcUserController grpcController.GRPCUserController) error {
// 	lis, err := net.Listen("tcp", ":50051")
// 	if err != nil {
// 		return err
// 	}

// 	server := grpc.NewServer(
// 		grpc.UnaryInterceptor(interceptor.UnaryServerInterceptor()),
// 	)

// 	grpcController.RegisterUserServiceServer(server, grpcUserController)

// 	log.Printf("gRPC server listening on %v", lis.Addr())
// 	return server.Serve(lis)
// }
