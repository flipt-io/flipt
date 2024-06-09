package audit

import (
	"errors"
	"fmt"
	"strings"
)

// EventPairChecker is the contract for checking if an event pair exists and if it should be emitted to configured sinks.
type EventPairChecker interface {
	Check(eventPair string) bool
	Events() []string
}

// Checker holds a map that maps event pairs to a dummy struct. It is basically
// used as a set to check for existence.
type Checker struct {
	eventActions map[string]struct{}
}

// NewChecker is the constructor for a Checker.
func NewChecker(eventPairs []string) (*Checker, error) {
	nouns := map[string][]string{
		"constraint":   {"constraint"},
		"distribution": {"distribution"},
		"flag":         {"flag"},
		"namespace":    {"namespace"},
		"rollout":      {"rollout"},
		"rule":         {"rule"},
		"segment":      {"segment"},
		"token":        {"token"},
		"variant":      {"variant"},
		"*":            {"constraint", "distribution", "flag", "namespace", "rollout", "rule", "segment", "token", "variant"},
	}

	verbs := map[string][]string{
		"created": {"created"},
		"deleted": {"deleted"},
		"updated": {"updated"},
		"*":       {"created", "deleted", "updated"},
	}

	eventActions := make(map[string]struct{})
	for _, ep := range eventPairs {
		epSplit := strings.Split(ep, ":")
		if len(epSplit) < 2 {
			return nil, fmt.Errorf("invalid event pair: %s", ep)
		}

		eventNouns, ok := nouns[epSplit[0]]
		if !ok {
			return nil, fmt.Errorf("invalid noun: %s", epSplit[0])
		}

		eventVerbs, ok := verbs[epSplit[1]]
		if !ok {
			return nil, fmt.Errorf("invalid verb: %s", epSplit[1])
		}

		for _, en := range eventNouns {
			for _, ev := range eventVerbs {
				eventPair := fmt.Sprintf("%s:%s", en, ev)

				_, ok := eventActions[eventPair]
				if ok {
					return nil, fmt.Errorf("repeated event pair: %s", eventPair)
				}

				eventActions[eventPair] = struct{}{}
			}
		}
	}

	if len(eventActions) == 0 {
		return nil, errors.New("no event pairs exist")
	}

	return &Checker{
		eventActions: eventActions,
	}, nil
}

// Check checks if an event pair exists in the Checker data structure for event emission.
func (c *Checker) Check(eventPair string) bool {
	if c == nil || c.eventActions == nil {
		return false
	}

	_, ok := c.eventActions[eventPair]
	return ok
}

// Events returns the type of events we would like to emit to configured sinks.
func (c *Checker) Events() []string {
	var events = make([]string, 0, len(c.eventActions))
	for k := range c.eventActions {
		events = append(events, k)
	}

	return events
}

type NoOpChecker struct{}

func (n *NoOpChecker) Check(eventPair string) bool {
	return false
}

func (n *NoOpChecker) Events() []string {
	return []string{}
}
