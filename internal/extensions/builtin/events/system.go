// Package events provides event publishing and subscription capabilities for extensions.
package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/llamagate/llamagate/internal/extensions/builtin/core"
)

// System provides event publishing and subscription
type System struct {
	name        string
	version     string
	mu          sync.RWMutex
	subscribers map[string][]*subscription
	registry    *core.Registry
}

type subscription struct {
	id      string
	filter  *core.EventFilter
	handler core.EventHandler
	active  bool
}

// NewSystem creates a new event system
func NewSystem(name, version string) *System {
	return &System{
		name:        name,
		version:     version,
		subscribers: make(map[string][]*subscription),
		registry:    core.GetRegistry(),
	}
}

// Name returns the name of the extension
func (s *System) Name() string {
	return s.name
}

// Version returns the version of the extension
func (s *System) Version() string {
	return s.version
}

// Initialize initializes the system
func (s *System) Initialize(_ context.Context, _ map[string]interface{}) error {
	return nil
}

// Shutdown shuts down the system
func (s *System) Shutdown(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Deactivate all subscriptions
	for _, subs := range s.subscribers {
		for _, sub := range subs {
			sub.active = false
		}
	}

	return nil
}

// Publish publishes an event
func (s *System) Publish(ctx context.Context, event *core.Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	if event.ID == "" {
		event.ID = fmt.Sprintf("event-%d", time.Now().UnixNano())
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	s.mu.RLock()
	subs := s.getMatchingSubscriptions(event)
	s.mu.RUnlock()

	// Notify subscribers
	for _, sub := range subs {
		if sub.active {
			go func(sub *subscription) {
				_ = sub.handler(ctx, event)
			}(sub)
		}
	}

	return nil
}

// Subscribe subscribes to events matching the filter
func (s *System) Subscribe(_ context.Context, filter *core.EventFilter, handler core.EventHandler) (core.Subscription, error) {
	if handler == nil {
		return nil, fmt.Errorf("handler cannot be nil")
	}

	sub := &subscription{
		id:      fmt.Sprintf("sub-%d", time.Now().UnixNano()),
		filter:  filter,
		handler: handler,
		active:  true,
	}

	s.mu.Lock()
	// Group by event type for efficient lookup
	key := "all"
	if filter != nil && len(filter.Types) > 0 {
		key = filter.Types[0] // Use first type as key
	}
	s.subscribers[key] = append(s.subscribers[key], sub)
	s.mu.Unlock()

	return &SubscriptionImpl{system: s, id: sub.id}, nil
}

// getMatchingSubscriptions returns subscriptions that match the event
func (s *System) getMatchingSubscriptions(event *core.Event) []*subscription {
	var matches []*subscription

	// Check all subscriptions
	for _, subs := range s.subscribers {
		for _, sub := range subs {
			if !sub.active {
				continue
			}
			if s.matchesFilter(event, sub.filter) {
				matches = append(matches, sub)
			}
		}
	}

	return matches
}

// matchesFilter checks if an event matches a filter
func (s *System) matchesFilter(event *core.Event, filter *core.EventFilter) bool {
	if filter == nil {
		return true
	}

	// Check types
	if len(filter.Types) > 0 {
		matched := false
		for _, typ := range filter.Types {
			if typ == event.Type {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check sources
	if len(filter.Sources) > 0 {
		matched := false
		for _, source := range filter.Sources {
			if source == event.Source {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check match criteria
	if len(filter.Match) > 0 {
		for k, v := range filter.Match {
			if eventVal, ok := event.Data[k]; !ok || eventVal != v {
				return false
			}
		}
	}

	return true
}

// SubscriptionImpl implements the Subscription interface
type SubscriptionImpl struct {
	system *System
	id     string
}

// Unsubscribe unsubscribes from events
func (si *SubscriptionImpl) Unsubscribe(_ context.Context) error {
	si.system.mu.Lock()
	defer si.system.mu.Unlock()

	// Find and deactivate subscription
	for key, subs := range si.system.subscribers {
		for i, sub := range subs {
			if sub.id == si.id {
				sub.active = false
				// Remove from slice
				si.system.subscribers[key] = append(subs[:i], subs[i+1:]...)
				return nil
			}
		}
	}

	return fmt.Errorf("subscription %s not found", si.id)
}
