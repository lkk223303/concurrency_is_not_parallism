package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// Define a struct to represent the load balancer.
type LoadBalancer struct {
	serverPool []string // The pool of servers to balance requests among.
}

// Method to add a server to the pool.
func (l *LoadBalancer) AddServer(server string) {
	l.serverPool = append(l.serverPool, server)
}

// Method to balance a request among the servers in the pool.
func (l LoadBalancer) Balance(w http.ResponseWriter, r *http.Request) {
	// Use the default http.ServeMux to dispatch the request to the selected server.
	http.DefaultServeMux.ServeHTTP(w, r)
}

// Method to select a server from the pool at random.
func (l *LoadBalancer) selectServer() string {
	// Use the current time as a seed for the random number generator.
	rand.Seed(time.Now().Unix())
	// Select a server at random from the pool.
	server := l.serverPool[rand.Intn(len(l.serverPool))]
	return server
}

// Method to handle requests to the load balancer.
func (l LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Select a server from the pool at random.
	server := l.selectServer()
	// Dispatch the request to the selected server.
	http.DefaultServeMux.Handle(server, http.HandlerFunc(l.Balance))
}

func AnotherLoadBalancer() {
	// Create a new load balancer.
	lb := LoadBalancer{}

	// Add some servers to the pool.
	lb.AddServer("http://localhost:8080")
	lb.AddServer("http://localhost:8081")
	lb.AddServer("http://localhost:8082")

	// Start the load balancer.
	fmt.Println("Load balancer listening on port 8083")
	http.ListenAndServe(":8083", lb)
}
