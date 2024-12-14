package main

const dbTemplate = `package db

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Import the PostgreSQL driver
)

// DB defines the interface for database operations
type DB interface {
	Close() error
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// PostgresDB is a concrete implementation of the DB interface
type PostgresDB struct {
	conn *sql.DB
}

// NewPostgresDB creates a new PostgresDB connection
func NewPostgresDB(dataSourceName string) (*PostgresDB, error) {
	conn, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &PostgresDB{conn: conn}, nil
}

// Close closes the database connection
func (db *PostgresDB) Close() error {
	return db.conn.Close()
}

// Query executes a query on the database
func (db *PostgresDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.QueryContext(ctx, query, args...)
}
`

const cacheTemplate = `package cache

import (
	"context"
	"sync"
)

// Cache defines the interface for cache operations
type Cache interface {
	Set(ctx context.Context, key string, value interface{})
	Get(ctx context.Context, key string) (interface{}, bool)
	Delete(ctx context.Context, key string)
}

// InMemoryCache is a concrete implementation of the Cache interface
type InMemoryCache struct {
	mu    sync.RWMutex
	store map[string]interface{}
}

// NewInMemoryCache creates a new InMemoryCache instance
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		store: make(map[string]interface{}),
	}
}

// Set adds a value to the cache
func (c *InMemoryCache) Set(ctx context.Context, key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(ctx context.Context, key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, found := c.store[key]
	return value, found
}

// Delete removes a value from the cache
func (c *InMemoryCache) Delete(ctx context.Context, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}
`

const queueTemplate = `package queue

import (
	"context"
	"sync"
)

// Queue defines the interface for queue operations
type Queue interface {
	Enqueue(ctx context.Context, item interface{})
	Dequeue(ctx context.Context) (interface{}, bool)
	Size(ctx context.Context) int
}

// FIFOQueue is a concrete implementation of the Queue interface
type FIFOQueue struct {
	mu    sync.Mutex
	items []interface{}
}

// NewFIFOQueue creates a new FIFOQueue instance
func NewFIFOQueue() *FIFOQueue {
	return &FIFOQueue{
		items: make([]interface{}, 0),
	}
}

// Enqueue adds an item to the queue
func (q *FIFOQueue) Enqueue(ctx context.Context, item interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, item)
}

// Dequeue removes an item from the queue
func (q *FIFOQueue) Dequeue(ctx context.Context) (interface{}, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return nil, false
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, true
}

// Size returns the number of items in the queue
func (q *FIFOQueue) Size(ctx context.Context) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}
`

const redisQueueTemplate = `package queue

import (
	"context"
	"github.com/go-redis/redis/v8"
)

// RedisQueue is a concrete implementation of the Queue interface using Redis
type RedisQueue struct {
	client *redis.Client
	key    string
}

// NewRedisQueue creates a new RedisQueue instance
func NewRedisQueue(addr, password, key string, db int) *RedisQueue {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisQueue{
		client: client,
		key:    key,
	}
}

// Enqueue adds an item to the Redis queue
func (q *RedisQueue) Enqueue(ctx context.Context, item interface{}) {
	q.client.RPush(ctx, q.key, item)
}

// Dequeue removes an item from the Redis queue
func (q *RedisQueue) Dequeue(ctx context.Context) (interface{}, bool) {
	result, err := q.client.LPop(ctx, q.key).Result()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		return nil, false
	}
	return result, true
}

// Size returns the number of items in the Redis queue
func (q *RedisQueue) Size(ctx context.Context) int {
	size, err := q.client.LLen(ctx, q.key).Result()
	if err != nil {
		return 0
	}
	return int(size)
}
`

const loggingTemplate = `package logging

import (
	"context"
	"log"
	"os"
)

// Logger represents a simple logger
type Logger struct {
	logger *log.Logger
}

// NewLogger creates a new Logger instance
func NewLogger(prefix string) *Logger {
	return &Logger{
		logger: log.New(os.Stdout, prefix, log.LstdFlags),
	}
}

// Info logs an informational message
func (l *Logger) Info(ctx context.Context, msg string) {
	l.logger.Println("INFO:", msg)
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, msg string) {
	l.logger.Println("ERROR:", msg)
}
`

const dbTestTemplate = `package db_test

import (
	"context"
	"database/sql"
	"testing"
)

func TestNewPostgresDB(t *testing.T) {
	// Setup test database connection string
	dataSourceName := "user=test dbname=test sslmode=disable"

	// Test creating a new PostgresDB
	db, err := NewPostgresDB(dataSourceName)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()
}

func TestPostgresDB_Query(t *testing.T) {
	// Setup test database connection string
	dataSourceName := "user=test dbname=test sslmode=disable"

	// Create a new PostgresDB
	db, err := NewPostgresDB(dataSourceName)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	// Test query execution
	ctx := context.Background()
	_, err = db.Query(ctx, "SELECT 1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
`

const cacheTestTemplate = `package cache_test

import (
	"context"
	"testing"
)

func TestInMemoryCache_SetAndGet(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	// Test setting and getting a value
	cache.Set(ctx, "key", "value")
	value, found := cache.Get(ctx, "key")
	if !found || value != "value" {
		t.Fatalf("expected value 'value', got %v", value)
	}
}

func TestInMemoryCache_Delete(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	// Test deleting a value
	cache.Set(ctx, "key", "value")
	cache.Delete(ctx, "key")
	_, found := cache.Get(ctx, "key")
	if found {
		t.Fatalf("expected key to be deleted")
	}
}
`

const queueTestTemplate = `package queue_test

import (
	"context"
	"testing"
)

func TestFIFOQueue_EnqueueAndDequeue(t *testing.T) {
	queue := NewFIFOQueue()
	ctx := context.Background()

	// Test enqueue and dequeue
	queue.Enqueue(ctx, "item1")
	queue.Enqueue(ctx, "item2")

	item, ok := queue.Dequeue(ctx)
	if !ok || item != "item1" {
		t.Fatalf("expected 'item1', got %v", item)
	}

	item, ok = queue.Dequeue(ctx)
	if !ok || item != "item2" {
		t.Fatalf("expected 'item2', got %v", item)
	}
}

func TestFIFOQueue_Size(t *testing.T) {
	queue := NewFIFOQueue()
	ctx := context.Background()

	// Test size
	queue.Enqueue(ctx, "item1")
	if size := queue.Size(ctx); size != 1 {
		t.Fatalf("expected size 1, got %d", size)
	}
}
`

const loggingTestTemplate = `package logging_test

import (
	"bytes"
	"context"
	"log"
	"testing"
)

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", log.LstdFlags)
	l := &Logger{logger: logger}

	ctx := context.Background()
	l.Info(ctx, "info message")

	if !bytes.Contains(buf.Bytes(), []byte("INFO: info message")) {
		t.Fatalf("expected log to contain 'INFO: info message'")
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", log.LstdFlags)
	l := &Logger{logger: logger}

	ctx := context.Background()
	l.Error(ctx, "error message")

	if !bytes.Contains(buf.Bytes(), []byte("ERROR: error message")) {
		t.Fatalf("expected log to contain 'ERROR: error message'")
	}
}
`

const rabbitMQQueueTemplate = `package queue

import (
	"context"
	"github.com/streadway/amqp"
	"log"
)

// RabbitMQQueue is a concrete implementation of the Queue interface using RabbitMQ
type RabbitMQQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

// NewRabbitMQQueue creates a new RabbitMQQueue instance
func NewRabbitMQQueue(amqpURL, queueName string) (*RabbitMQQueue, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	return &RabbitMQQueue{
		conn:    conn,
		channel: channel,
		queue:   queue,
	}, nil
}

// Enqueue adds an item to the RabbitMQ queue
func (q *RabbitMQQueue) Enqueue(ctx context.Context, item string) error {
	return q.channel.Publish(
		"",         // exchange
		q.queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(item),
		})
}

// Dequeue removes an item from the RabbitMQ queue
func (q *RabbitMQQueue) Dequeue(ctx context.Context) (string, bool) {
	msgs, err := q.channel.Consume(
		q.queue.Name,
		"",
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Println("Failed to register a consumer:", err)
		return "", false
	}

	for msg := range msgs {
		return string(msg.Body), true
	}

	return "", false
}

// Close closes the RabbitMQ connection
func (q *RabbitMQQueue) Close() {
	q.channel.Close()
	q.conn.Close()
}
`

const redisCacheTemplate = `package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
)

// RedisCache is a concrete implementation of the Cache interface using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new RedisCache instance
func NewRedisCache(addr, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisCache{
		client: client,
	}
}

// Set adds a value to the Redis cache
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}) {
	c.client.Set(ctx, key, value, 0)
}

// Get retrieves a value from the Redis cache
func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, bool) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		return nil, false
	}
	return val, true
}

// Delete removes a value from the Redis cache
func (c *RedisCache) Delete(ctx context.Context, key string) {
	c.client.Del(ctx, key)
}
`

const sqlxDBTemplate = `package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Import the PostgreSQL driver
)

// SQLxDB is a concrete implementation of the DB interface using SQLx
type SQLxDB struct {
	db *sqlx.DB
}

// NewSQLxDB creates a new SQLxDB instance
func NewSQLxDB(dataSourceName string) (*SQLxDB, error) {
	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &SQLxDB{db: db}, nil
}

// Close closes the SQLx database connection
func (s *SQLxDB) Close() error {
	return s.db.Close()
}

// Query executes a query on the SQLx database
func (s *SQLxDB) Query(query string, args ...interface{}) (*sqlx.Rows, error) {
	return s.db.Queryx(query, args...)
}
`

const robustHttpClientTemplate = `package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// MiddlewareFunc defines a function to process middleware
type MiddlewareFunc func(req *http.Request) error

// HTTPClient is an advanced HTTP client with middleware, retry, and logging capabilities
type HTTPClient struct {
	client      *http.Client
	retryCount  int
	retryDelay  time.Duration
	middlewares []MiddlewareFunc
}

// NewHTTPClient creates a new HTTPClient instance with a default timeout, retry count, and delay
func NewHTTPClient(timeout time.Duration, retryCount int, retryDelay time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		retryCount: retryCount,
		retryDelay: retryDelay,
	}
}

// Use adds a middleware to the HTTP client
func (hc *HTTPClient) Use(middleware MiddlewareFunc) {
	hc.middlewares = append(hc.middlewares, middleware)
}

// Response represents a structured HTTP response
type Response struct {
	StatusCode int
	Body       string
	Headers    http.Header
}

// Get performs a GET request with optional headers and query parameters
func (hc *HTTPClient) Get(ctx context.Context, baseURL string, queryParams map[string]string, headers map[string]string) (*Response, error) {
	return hc.doRequest(ctx, http.MethodGet, baseURL, queryParams, nil, headers)
}

// PostJSON performs a POST request with a JSON body and optional headers
func (hc *HTTPClient) PostJSON(ctx context.Context, url string, jsonBody interface{}, headers map[string]string) (*Response, error) {
	body, err := json.Marshal(jsonBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	headers["Content-Type"] = "application/json"
	return hc.doRequest(ctx, http.MethodPost, url, nil, body, headers)
}

// doRequest performs the HTTP request with retry logic and middleware
func (hc *HTTPClient) doRequest(ctx context.Context, method, baseURL string, queryParams map[string]string, body []byte, headers map[string]string) (*Response, error) {
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Add query parameters
	if queryParams != nil {
		q := reqURL.Query()
		for key, value := range queryParams {
			q.Add(key, value)
		}
		reqURL.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// Apply middlewares
	for _, middleware := range hc.middlewares {
		if err := middleware(req); err != nil {
			return nil, fmt.Errorf("middleware error: %w", err)
		}
	}

	var lastErr error
	for i := 0; i <= hc.retryCount; i++ {
		resp, err := hc.client.Do(req)
		if err != nil {
			lastErr = err
			log.Printf("Request failed: %v. Retrying... (%d/%d)", err, i+1, hc.retryCount)
			time.Sleep(hc.retryDelay)
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		log.Printf("Request to %s completed with status %d", req.URL, resp.StatusCode)
		return &Response{
			StatusCode: resp.StatusCode,
			Body:       string(body),
			Headers:    resp.Header,
		}, nil
	}
	return nil, fmt.Errorf("request failed after %d attempts: %w", hc.retryCount, lastErr)
}
`
