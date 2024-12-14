package main

const dbIntegrationTestTemplate = `package db_integration_test

import (
	"context"
	"testing"
	"your_project/db"
)

func TestPostgresDB_Integration(t *testing.T) {
	// Setup test database connection string
	dataSourceName := "user=test dbname=test sslmode=disable"

	// Create a new PostgresDB
	db, err := db.NewPostgresDB(dataSourceName)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer db.Close()

	// Test a real query execution
	ctx := context.Background()
	_, err = db.Query(ctx, "SELECT * FROM your_table")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
`

const cacheIntegrationTestTemplate = `package cache_integration_test

import (
	"context"
	"testing"
	"your_project/cache"
)

func TestRedisCache_Integration(t *testing.T) {
	// Setup Redis connection
	addr := "localhost:6379"
	password := ""
	dbIndex := 0

	// Create a new RedisCache
	cache := cache.NewRedisCache(addr, password, dbIndex)

	ctx := context.Background()

	// Test setting and getting a value
	cache.Set(ctx, "key", "value")
	value, found := cache.Get(ctx, "key")
	if !found || value != "value" {
		t.Fatalf("expected value 'value', got %v", value)
	}

	// Test deleting a value
	cache.Delete(ctx, "key")
	_, found = cache.Get(ctx, "key")
	if found {
		t.Fatalf("expected key to be deleted")
	}
}
`

const queueIntegrationTestTemplate = `package queue_integration_test

import (
	"context"
	"testing"
	"your_project/queue"
)

func TestRedisQueue_Integration(t *testing.T) {
	// Setup Redis connection
	addr := "localhost:6379"
	password := ""
	dbIndex := 0
	key := "test_queue"

	// Create a new RedisQueue
	q := queue.NewRedisQueue(addr, password, key, dbIndex)

	ctx := context.Background()

	// Test enqueue and dequeue
	q.Enqueue(ctx, "item1")
	item, ok := q.Dequeue(ctx)
	if !ok || item != "item1" {
		t.Fatalf("expected 'item1', got %v", item)
	}
}
`

const httpClientIntegrationTestTemplate = `package httpclient_integration_test

import (
	"context"
	"testing"
	"your_project/httpclient"
	"net/http"
	"net/http/httptest"
)

func TestHTTPClient_Integration(t *testing.T) {
	// Setup a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(` + "`" + `{"message": "success"}` + "`" + `))
	}))
	defer server.Close()

	// Create a new HTTPClient
	client := httpclient.NewHTTPClient(5*time.Second, 3, 1*time.Second)

	ctx := context.Background()

	// Test GET request
	resp, err := client.Get(ctx, server.URL, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}
`
