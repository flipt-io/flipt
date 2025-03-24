package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	sdk "go.flipt.io/flipt/sdk/go"
)

type evaluateCommand struct {
	root          *rootCommand
	address       string
	requestID     string
	entityID      string
	namespace     string
	watch         bool
	token         string
	interval      time.Duration
	contextValues []string
}

func newEvaluateCommand(root *rootCommand) *cobra.Command {
	c := &evaluateCommand{
		root: root,
	}

	cmd := &cobra.Command{
		Use:   "evaluate [flagKey]",
		Short: "Evaluate a flag",
		Args:  cobra.ExactArgs(1),
		RunE:  c.run,
	}

	cmd.Flags().StringVarP(
		&c.namespace,
		"namespace", "n",
		"default",
		"flag namespace.",
	)
	cmd.Flags().StringVarP(
		&c.entityID,
		"entity-id", "e",
		uuid.NewString(),
		"evaluation request entity id.",
	)
	cmd.Flags().StringVarP(
		&c.requestID,
		"request-id", "r",
		"",
		"evaluation request id.",
	)

	cmd.Flags().StringArrayVarP(
		&c.contextValues,
		"context", "c",
		[]string{},
		"evaluation request context as key=value.",
	)

	cmd.Flags().StringVarP(
		&c.address,
		"address", "a",
		"http://localhost:8080",
		"address of Flipt instance.",
	)

	cmd.Flags().StringVarP(
		&c.token,
		"token", "t",
		"",
		"client token used to authenticate access to Flipt instance.",
	)

	cmd.Flags().BoolVarP(
		&c.watch,
		"watch", "w",
		false,
		"watch for changes to evaluations",
	)

	cmd.Flags().DurationVarP(
		&c.interval,
		"interval", "i",
		5*time.Second,
		"interval at which to poll for changes",
	)

	cmd.Flags().StringVar(&c.root.configFile, "config", c.root.configFile, "path to config file")

	return cmd
}

func (c *evaluateCommand) run(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	if c.requestID == "" {
		c.requestID = uuid.NewString()
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalCh
		cancel()
	}()

	context := map[string]string{}
	for _, c := range c.contextValues {
		i := strings.Index(c, "=")
		if i == -1 {
			return fmt.Errorf("invalid context value %q, must be of form key=value", c)
		}

		k, v := c[0:i], c[i+1:]
		context[k] = v
	}

	req := &evaluation.EvaluationRequest{
		FlagKey:      args[0],
		EntityId:     c.entityID,
		Context:      context,
		RequestId:    c.requestID,
		NamespaceKey: c.namespace,
	}

	sdkClient, err := fliptSDK(c.address, c.token)
	if err != nil {
		return err
	}

	client := sdkClient.Flipt()
	resp, err := client.GetFlag(ctx, &flipt.GetFlagRequest{
		Key:          args[0],
		NamespaceKey: c.namespace,
	})
	if err != nil {
		return err
	}

	if c.watch {
		return c.evaluate(ctx, resp.Type, sdkClient, req)
	}

	return c.evaluate(ctx, resp.Type, sdkClient, req)
}

type booleanEvaluateResponse struct {
	FlagKey               string    `json:"flag_key,omitempty"`
	Enabled               bool      `json:"enabled"`
	Reason                string    `json:"reason,omitempty"`
	RequestID             string    `json:"request_id,omitempty"`
	RequestDurationMillis float64   `json:"request_duration_millis,omitempty"`
	Timestamp             time.Time `json:"timestamp,omitempty"`
}

type variantEvaluationResponse struct {
	FlagKey               string    `json:"flag_key,omitempty"`
	Match                 bool      `json:"match"`
	Reason                string    `json:"reason,omitempty"`
	VariantKey            string    `json:"variant_key,omitempty"`
	VariantAttachment     string    `json:"variant_attachment,omitempty"`
	SegmentKeys           []string  `json:"segment_keys,omitempty"`
	RequestID             string    `json:"request_id,omitempty"`
	RequestDurationMillis float64   `json:"request_duration_millis,omitempty"`
	Timestamp             time.Time `json:"timestamp,omitempty"`
}

func (c *evaluateCommand) evaluate(ctx context.Context, flagType flipt.FlagType, sdkClient *sdk.SDK, req *evaluation.EvaluationRequest) error {
	colorBool := color.New(color.FgCyan, color.Bold).SprintFunc()
	colorFloat := color.New(color.FgGreen).SprintFunc()

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	evaluationClient := sdkClient.Evaluation()

	for {
		switch flagType {
		case flipt.FlagType_BOOLEAN_FLAG_TYPE:
			resp, err := evaluationClient.Boolean(ctx, req)
			if err != nil {
				return err
			}

			res := &booleanEvaluateResponse{
				FlagKey:               resp.FlagKey,
				Enabled:               resp.Enabled,
				Reason:                resp.Reason.String(),
				RequestID:             resp.RequestId,
				Timestamp:             time.Now().UTC(),
				RequestDurationMillis: resp.RequestDurationMillis,
			}

			b, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				return err
			}

			if c.watch {
				fmt.Printf("Evaluating flag %s at %s with request ID: %s\n", colorBool(resp.FlagKey), colorFloat(res.Timestamp.Format(time.RFC3339)), resp.RequestId)
			}
			fmt.Println(string(b))
		case flipt.FlagType_VARIANT_FLAG_TYPE:
			resp, err := evaluationClient.Variant(ctx, req)
			if err != nil {
				return err
			}

			segmentKeys := make([]string, 0, len(resp.SegmentKeys))
			for _, s := range resp.SegmentKeys {
				segmentKeys = append(segmentKeys, s)
			}

			sort.Strings(segmentKeys)

			res := &variantEvaluationResponse{
				FlagKey:               resp.FlagKey,
				Match:                 resp.Match,
				Reason:                resp.Reason.String(),
				VariantKey:            resp.VariantKey,
				VariantAttachment:     resp.VariantAttachment,
				SegmentKeys:           segmentKeys,
				RequestID:             resp.RequestId,
				Timestamp:             time.Now().UTC(),
				RequestDurationMillis: resp.RequestDurationMillis,
			}

			b, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				return err
			}

			if c.watch {
				fmt.Printf("Evaluating flag %s at %s with request ID: %s\n", colorBool(resp.FlagKey), colorFloat(res.Timestamp.Format(time.RFC3339)), resp.RequestId)
			}
			fmt.Println(string(b))
		default:
			return fmt.Errorf("unknown flag type: %s", flagType)
		}

		if !c.watch {
			return nil
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
