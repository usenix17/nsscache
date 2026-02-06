package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"nsscache-http/cache"
	"nsscache-http/config"
	"nsscache-http/handlers"
	"nsscache-http/ldap"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Create LDAP client
	ldapClient := ldap.NewClient(&cfg.LDAP)

	// Create and start cache
	dataCache := cache.New(ldapClient, cfg.Cache.TTL)
	if err := dataCache.Start(); err != nil {
		log.Fatalf("failed to start cache: %v", err)
	}
	defer dataCache.Stop()

	// Create handlers
	h := handlers.New(dataCache)

	// Register routes
	http.HandleFunc("/passwd.json", h.PasswdJSON)
	http.HandleFunc("/passwd", h.PasswdFlat)
	http.HandleFunc("/group.json", h.GroupJSON)
	http.HandleFunc("/group", h.GroupFlat)
	http.HandleFunc("/shadow.json", h.ShadowJSON)
	http.HandleFunc("/shadow", h.ShadowFlat)
	http.HandleFunc("/health", h.Health)

	// Handle shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("starting server on %s", cfg.Server.Listen)
		if err := http.ListenAndServe(cfg.Server.Listen, nil); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-done
	log.Println("shutting down")
}
