package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
	vegeta "github.com/tsenart/vegeta/lib"
)

const (
	postMethod = "POST"
)

type Evaluation struct {
	EntityId string            `json:"entityId"`
	FlagKey  string            `json:"flagKey"`
	Context  map[string]string `json:"context"`
}

type BooleanRollout struct {
	EntityId string            `json:"entityId"`
	FlagKey  string            `json:"flagKey"`
	Context  map[string]string `json:"context"`
}

var (
	dur       time.Duration
	fliptAddr string
)

func init() {
	flag.DurationVar(&dur, "duration", 120*time.Second, "duration of load test for evaluations")
	flag.StringVar(&fliptAddr, "flipt-addr", "http://127.0.0.1:8080", "address of the flipt instance")
}

func run() error {
	flag.Parse()

	evaluationTargets := make([]vegeta.Target, 0, 2)

	variantEvaluation := Evaluation{
		EntityId: uuid.Must(uuid.NewV4()).String(),
		FlagKey:  "flag_010",
	}

	ve, _ := json.Marshal(variantEvaluation)

	evaluationTargets = append(evaluationTargets, vegeta.Target{
		Method: postMethod,
		URL:    fmt.Sprintf("%s/api/v1/evaluate", fliptAddr),
		Body:   ve,
	})

	booleanEvaluation := BooleanRollout{
		EntityId: uuid.Must(uuid.NewV4()).String(),
		FlagKey:  "flag_boolean",
	}

	be, _ := json.Marshal(booleanEvaluation)

	evaluationTargets = append(evaluationTargets, vegeta.Target{
		Method: postMethod,
		URL:    fmt.Sprintf("%s/evaluate/v1/boolean", fliptAddr),
		Body:   be,
	})

	evaluationTargeter := vegeta.NewStaticTargeter(evaluationTargets...)

	// TODO: make configurable
	rate := vegeta.Rate{Freq: 100, Per: time.Second}

	var metrics vegeta.Metrics

	fmt.Printf("About to start vegeta attack on evalutions for %f seconds...\n", dur.Seconds())
	attacker := vegeta.NewAttacker()
	for res := range attacker.Attack(evaluationTargeter, rate, dur, "Doing evaluation request...") {
		metrics.Add(res)
	}
	metrics.Close()

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
