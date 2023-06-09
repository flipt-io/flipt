package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"go.flipt.io/flipt/internal/ext"
	"gopkg.in/yaml.v2"
)

var (
	namespace              = flag.String("namespace", "", "Namespace of resulting document")
	flagCount              = flag.Int("flags", 50, "Number of flags to generate")
	flagVariantCount       = flag.Int("flag-variants", 2, "Number of variants per flag")
	flagRuleCount          = flag.Int("flag-rules", 50, "Numer of rules per flag")
	flagRuleDistCount      = flag.Int("flag-rule-distributions", 2, "Numer of rules per flag")
	segmentCount           = flag.Int("segments", 50, "Number of segments to generate")
	segmentConstraintCount = flag.Int("segment-contraints", 2, "Number of contraints per segment")
)

func main() {
	flag.Parse()

	doc := ext.Document{
		Version:   "1.0",
		Namespace: *namespace,
	}

	for i := 0; i < *segmentCount; i++ {
		key := fmt.Sprintf("segment_%03d", i+1)
		segment := &ext.Segment{
			Key:         key,
			Name:        strings.ToUpper(key),
			Description: "Some Segment Description",
			MatchType:   "ALL_MATCH_TYPE",
		}

		for j := 0; j < *segmentConstraintCount; j++ {
			constraint := &ext.Constraint{
				Type:     "STRING_COMPARISON_TYPE",
				Property: "in_segment",
				Operator: "eq",
				Value:    key,
			}

			segment.Constraints = append(segment.Constraints, constraint)
		}

		doc.Segments = append(doc.Segments, segment)
	}

	for i := 0; i < *flagCount; i++ {
		key := fmt.Sprintf("flag_%03d", i+1)
		flag := &ext.Flag{
			Key:         key,
			Name:        strings.ToUpper(key),
			Enabled:     true,
			Description: "Some Description",
		}

		for j := 0; j < *flagVariantCount; j++ {
			key := fmt.Sprintf("variant_%03d", j+1)
			variant := &ext.Variant{
				Key:  key,
				Name: strings.ToUpper(key),
			}

			flag.Variants = append(flag.Variants, variant)
		}

		for k := 0; k < *flagRuleCount; k++ {
			rule := &ext.Rule{
				Rank:       uint(k + 1),
				SegmentKey: doc.Segments[k%len(doc.Segments)].Key,
			}

			for l := 0; l < *flagRuleDistCount; l++ {
				rule.Distributions = append(rule.Distributions, &ext.Distribution{
					Rollout:    100.0 / float32(*flagRuleDistCount),
					VariantKey: flag.Variants[l%len(flag.Variants)].Key,
				})
			}

			flag.Rules = append(flag.Rules, rule)
		}

		doc.Flags = append(doc.Flags, flag)
	}

	if err := yaml.NewEncoder(os.Stdout).Encode(&doc); err != nil {
		log.Fatal(err)
	}
}
