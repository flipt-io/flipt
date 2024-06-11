package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	vegeta "github.com/tsenart/vegeta/lib"
)

const (
	postMethod = "POST"
	getMethod  = "GET"
)

type Evaluation struct {
	EntityId string            `json:"entityId"`
	FlagKey  string            `json:"flagKey"`
	Context  map[string]string `json:"context"`
}

var (
	dur               time.Duration
	rate              int
	fliptAddr         string
	fliptAuthToken    string
	fliptCacheEnabled bool
)

func init() {
	flag.DurationVar(&dur, "duration", 60*time.Second, "duration of load test for evaluations")
	flag.IntVar(&rate, "rate", 100, "rate of requests per second")
	flag.StringVar(&fliptAddr, "flipt-addr", "http://flipt:8080", "address of the flipt instance")
	flag.StringVar(&fliptAuthToken, "flipt-auth-token", "", "token to use for authentication with flipt")
	flag.BoolVar(&fliptCacheEnabled, "flipt-cache-enabled", false, "whether the flipt cache is enabled")
}

func main() {
	flag.Parse()

	variantTargeter := vegeta.Targeter(func(t *vegeta.Target) error {
		t.Header = http.Header{
			"Authorization": []string{fmt.Sprintf("Bearer %s", fliptAuthToken)},
		}
		t.Method = postMethod
		t.URL = fmt.Sprintf("%s/evaluate/v1/variant", fliptAddr)

		variantEvaluation := Evaluation{
			EntityId: uuid.Must(uuid.NewV4()).String(),
			FlagKey:  "flag_010",
			Context: map[string]string{
				"in_segment": "baz",
			},
		}

		ve, err := json.Marshal(variantEvaluation)
		if err != nil {
			return err
		}
		t.Body = ve
		return nil
	})

	booleanTargeter := func(t *vegeta.Target) error {
		t.Header = http.Header{
			"Authorization": []string{fmt.Sprintf("Bearer %s", fliptAuthToken)},
		}
		t.Method = postMethod
		t.URL = fmt.Sprintf("%s/evaluate/v1/boolean", fliptAddr)

		booleanEvaluation := Evaluation{
			EntityId: uuid.Must(uuid.NewV4()).String(),
			FlagKey:  "flag_boolean",
			Context: map[string]string{
				"in_segment": "baz",
			},
		}

		be, err := json.Marshal(booleanEvaluation)
		if err != nil {
			return err
		}

		t.Body = be
		return nil
	}

	flagTargeter := func(t *vegeta.Target) error {
		t.Header = http.Header{
			"Authorization": []string{fmt.Sprintf("Bearer %s", fliptAuthToken)},
		}
		t.Method = getMethod
		t.URL = fmt.Sprintf("%s/api/v1/namespaces/default/flags/flag_001", fliptAddr)

		return nil
	}

	rate := vegeta.Rate{Freq: rate, Per: time.Second}
	attack := func(name string, t vegeta.Targeter) {
		name = strings.ToUpper(name)
		var metrics vegeta.Metrics

		fmt.Printf("\n[%s]: Vegeta attack for %f seconds...\n", name, dur.Seconds())
		attacker := vegeta.NewAttacker()
		for res := range attacker.Attack(t, rate, dur, name) {
			metrics.Add(res)
		}
		metrics.Close()

		fmt.Printf("[%s]: Vegeta attack complete\n\n", name)
		fmt.Printf("Authentication Enabled: %v\n", fliptAuthToken != "")
		fmt.Printf("Cache Enabled: %v\n", fliptCacheEnabled)

		fmt.Printf("--------------------\n")
		fmt.Printf("Mean: %s\n", metrics.Latencies.Mean)
		fmt.Printf("95th percentile: %s\n", metrics.Latencies.P95)
		fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
		fmt.Printf("Max: %s\n", metrics.Latencies.Max)
		fmt.Printf("Requests: %d\n", metrics.Requests)
		fmt.Printf("Throughput: %f\n", metrics.Throughput)
		fmt.Printf("Errors: %v\n", metrics.Errors)
		fmt.Printf("Status Codes: %v\n", metrics.StatusCodes)
		fmt.Printf("Success: %f %%\n\n", metrics.Success*100.0)
	}

	attack("flag", flagTargeter)
	attack("variant", variantTargeter)
	attack("boolean", booleanTargeter)
}
