package main

import (
	"math/rand"
	"strconv"
	"time"
)

// Client represents a provider's client or SDK
type Client interface {
	GetContent(userIP string, count int) ([]*ContentItem, error)
}

// ContentItem represent one piece of content fetched from a provider
type ContentItem struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	Source  string    `json:"source"`
	Summary string    `json:"summary"`
	Link    string    `json:"link"`
	Expiry  time.Time `json:"expiry"`
}

// Provider represent the 3rd party from which we are getting content
type Provider string

var (

	// Sample Providers, put here as an example

	Provider1 = Provider("1")
	Provider2 = Provider("2")
	Provider3 = Provider("3")
)

// SampleContentProvider is an example for a Provider's client
type SampleContentProvider struct {
	Source Provider
}

// GetContent returns content items given a user IP, and the number of content items desired.
func (cp SampleContentProvider) GetContent(userIP string, count int) ([]*ContentItem, error) {
	resp := make([]*ContentItem, count)
	for i, _ := range resp {
		resp[i] = &ContentItem{
			ID:     strconv.Itoa(rand.Int()),
			Title:  "title",
			Source: string(cp.Source),
			Expiry: time.Now(),
		}

	}
	return resp, nil
}
