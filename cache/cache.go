package cache

import (
	"log"
	"sync"
	"time"

	"nsscache-http/ldap"
	"nsscache-http/models"
)

// Cache holds the in-memory cache of LDAP data
type Cache struct {
	mu      sync.RWMutex
	users   []models.User
	groups  []models.Group
	shadows []models.Shadow

	client    *ldap.Client
	ttl       time.Duration
	lastFetch time.Time
	stopChan  chan struct{}
}

// New creates a new cache instance
func New(client *ldap.Client, ttlSeconds int) *Cache {
	return &Cache{
		client:   client,
		ttl:      time.Duration(ttlSeconds) * time.Second,
		stopChan: make(chan struct{}),
	}
}

// Start begins the background refresh goroutine
func (c *Cache) Start() error {
	// Initial load
	if err := c.Refresh(); err != nil {
		return err
	}

	// Start background refresh
	go c.refreshLoop()

	return nil
}

// Stop stops the background refresh goroutine
func (c *Cache) Stop() {
	close(c.stopChan)
}

// refreshLoop periodically refreshes the cache
func (c *Cache) refreshLoop() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.Refresh(); err != nil {
				log.Printf("cache refresh failed: %v", err)
			}
		case <-c.stopChan:
			return
		}
	}
}

// Refresh fetches fresh data from LDAP
func (c *Cache) Refresh() error {
	// Connect and bind for each refresh to handle connection drops
	if err := c.client.Connect(); err != nil {
		return err
	}
	defer c.client.Close()

	if err := c.client.Bind(); err != nil {
		return err
	}

	users, err := c.client.FetchUsers()
	if err != nil {
		return err
	}

	groups, err := c.client.FetchGroups()
	if err != nil {
		return err
	}

	shadows, err := c.client.FetchShadow()
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.users = users
	c.groups = groups
	c.shadows = shadows
	c.lastFetch = time.Now()
	c.mu.Unlock()

	log.Printf("cache refreshed: %d users, %d groups, %d shadow entries",
		len(users), len(groups), len(shadows))

	return nil
}

// GetUsers returns the cached users
func (c *Cache) GetUsers() []models.User {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.users
}

// GetGroups returns the cached groups
func (c *Cache) GetGroups() []models.Group {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.groups
}

// GetShadows returns the cached shadow entries
func (c *Cache) GetShadows() []models.Shadow {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.shadows
}

// LastFetch returns the time of the last successful fetch
func (c *Cache) LastFetch() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastFetch
}

// Stats returns cache statistics
func (c *Cache) Stats() (users, groups, shadows int, lastFetch time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.users), len(c.groups), len(c.shadows), c.lastFetch
}
