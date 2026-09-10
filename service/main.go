package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nibroos/s-erp-api/service/internal/chat"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/consumer"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/middleware"
	"github.com/nibroos/s-erp-api/service/internal/routes"
	"github.com/nibroos/s-erp-api/service/internal/validators"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	apmfiber "go.elastic.co/apm/module/apmfiber/v2"
	// APM-instrumented gorm dialector — DB queries carrying a request context
	// become spans. sqlx here shares gorm's *sql.DB, so it is covered too.
	postgres "go.elastic.co/apm/module/apmgormv2/v2/driver/postgres"
	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"
	"gorm.io/gorm"
)

// apmResponseBodyMax caps how much of the response body is attached to the APM
// transaction (labels are keyword-truncated anyway); keeps memory/PII bounded.
const apmResponseBodyMax = 2048

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "response_status"},
	)
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
	// Load environment variables. `.env.local` (host-run overrides, gitignored)
	// is loaded first so its values win; godotenv never overrides keys that are
	// already set — so in the container the compose-injected env always wins and
	// these Load calls are harmless no-ops. `.env` is a symlink to the canonical
	// docker/.env (the single source of truth).
	_ = godotenv.Load(".env.local")
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Create a Prometheus registry
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry

	// Register the metrics with Prometheus
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)

	// Expose Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Println("Starting Prometheus metrics server on :9090")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Fatalf("Failed to start Prometheus metrics server: %v", err)
		}
	}()

	// Instrument ALL outbound HTTP once: any http.Client using the default
	// transport now creates an APM span and injects the traceparent header, so
	// downstream services continue the same distributed trace.
	http.DefaultTransport = apmhttp.WrapRoundTripper(http.DefaultTransport)

	// Determine the environment (production or test)
	env := os.Getenv("APP_ENV")
	var dbURL string
	if env == "test" {
		dbURL = config.GetTestDatabaseURL()
	} else {
		dbURL = config.GetDatabaseURL()
	}

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

	// // Initialize the SQLx database connection
	// sqlDB, err := sqlx.Connect("postgres", dbURL)

	// Convert *sql.Tx to *sqlx.Tx using sqlx.NewTx
	sqlDB := sqlx.NewDb(sqlDBGorm, "postgres")

	if err != nil {
		log.Fatalf("Failed to connect to the SQL database: %v", err)
	}

	// Configure SQLx connection pool
	sqlDB.SetMaxOpenConns(100)          // Maximum number of open connections
	sqlDB.SetMaxIdleConns(10)           // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(time.Hour) // Maximum lifetime of a connection

	// Initialize the Redis client (shared s-erp-redis). Best-effort: if Redis is
	// unreachable the API keeps running with caching disabled.
	config.InitRedisClientSafe()

	// When AI replies are dispatched to the consumer-service (CHAT_AI_QUEUE), the
	// worker delivers them back through a Redis fan-out channel. This socket-
	// serving process must subscribe so those replies reach the user's WebSocket.
	// (The worker itself uses a publish-only hub — see the AI-reply consumer.)
	if os.Getenv("CHAT_AI_QUEUE") == "true" && os.Getenv("SERVICE_TYPE") != "consumer" {
		if config.RedisClient != nil {
			chat.GlobalHub.EnableRedis(config.RedisClient, chat.FanoutChannel, true)
		} else {
			log.Println("WARNING: CHAT_AI_QUEUE=true but Redis is unavailable — queued AI replies will not reach sockets")
		}
	}

	// Fetch needed data from the database and cache it in Redis
	config.FetchCachedData(&fiber.Ctx{}, sqlDB)

	// Initialize the validator with the database connection
	validators.InitValidator(sqlDB)

	// Tracing is handled by Elastic APM (see apmfiber middleware below); the
	// Jaeger tracer has been retired. The route/consumer layers still take an
	// opentracing.Tracer, so pass a no-op — their spans compile but go nowhere.
	var tracer opentracing.Tracer = opentracing.NoopTracer{}
	opentracing.SetGlobalTracer(tracer)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
		// Increase read/write timeouts
		ReadTimeout:  time.Second * 60,
		WriteTimeout: time.Second * 60,
	})

	app.Use(middleware.RecoverMiddleware())

	// Elastic APM: opens a transaction per request (endpoint timing + errors).
	app.Use(apmfiber.Middleware())
	// Attach the response body to the APM transaction (APM has no native
	// response-body capture). Opt-in via APM_CAPTURE_RESPONSE_BODY=true; runs
	// after apmfiber so the transaction is active.
	if os.Getenv("APM_CAPTURE_RESPONSE_BODY") == "true" {
		app.Use(func(c *fiber.Ctx) error {
			err := c.Next()
			if tx := apm.TransactionFromContext(c.Context()); tx != nil {
				b := c.Response().Body()
				if len(b) > apmResponseBodyMax {
					b = b[:apmResponseBodyMax]
				}
				tx.Context.SetLabel("response_body", string(b))
			}
			return err
		})
	}

	// Add recover middleware with more detailed logging
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			// Log the full stack trace
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			log.Printf("Panic recovered: %v\nStack trace: %s\n", e, buf[:n])
		},
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	app.Use(PromDurationMiddleware)

	// Attach middleware
	// app.Use(middleware.JaegerTracingMiddleware(tracer))
	app.Use(middleware.ConvertEmptyStringsToNull())
	app.Use(middleware.ConvertRequestToFilters())

	// static folder on /public/uploads
	app.Static("/public", "./public")

	// Initialize RabbitMQ for publishers
	rabbitmq, err := config.NewRabbitMQ()
	if err != nil {
		log.Printf("Warning: Failed to initialize RabbitMQ: %v", err)
		// Don't fatal here, allow service to run without notifications
	}
	defer rabbitmq.Close()

	// Setup REST routes
	routes.SetupRoutes(app, gormDB, sqlDB, rabbitmq, tracer)

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
	} else if os.Getenv("SERVICE_TYPE") == "consumer" {
		rabbitmq, err := config.NewRabbitMQ()
		if err != nil {
			log.Fatalf("Failed to initialize RabbitMQ: %v", err)
		}
		defer rabbitmq.Close()

		// Start the consumer service
		wg.Add(1)
		go func() {
			defer wg.Done()
			consumerService := consumer.NewConsumerRouter(rabbitmq, gormDB, sqlDB, tracer)
			if err := consumerService.SetupConsumers(); err != nil {
				log.Printf("Consumer service error: %v", err)
			}
			log.Println("Consumer service started successfully")

			// Keep the goroutine running
			select {}
		}()

		go func() {
			healthApp := fiber.New()
			healthHandler := &consumer.ConsumerHealth{
				DB:       gormDB,
				RabbitMQ: rabbitmq.Connection,
			}
			healthApp.Get("/health", healthHandler.Handler)
			log.Println("Starting consumer health endpoint on :4010/health")
			if err := healthApp.Listen(":4010"); err != nil {
				log.Printf("Health endpoint error: %v", err)
			}
		}()
	} else {
		// Start REST server
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				log.Println("Starting REST server on :4001...")
				if err := app.Listen(":4001"); err != nil {
					log.Printf("Server error: %v", err)
					log.Println("Restarting server in 5 seconds...")
					time.Sleep(5 * time.Second)
					continue
				}
				break
			}
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
