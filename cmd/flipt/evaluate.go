package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	sdk "go.flipt.io/flipt/sdk/go"
	sdkgrpc "go.flipt.io/flipt/sdk/go/grpc"
	sdkhttp "go.flipt.io/flipt/sdk/go/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type evaluateCommand struct {
	address       string
	requestID     string
	entityID      string
	environment   string
	namespace     string
	watch         bool
	token         string
	interval      time.Duration
	contextValues []string
}

func newEvaluateCommand() *cobra.Command {
	c := &evaluateCommand{}
	cmd := &cobra.Command{
		Use:   "evaluate [flagKey]",
		Short: "Evaluate a flag",
		Args:  cobra.ExactArgs(1),
		RunE:  c.run,
	}

	cmd.Flags().StringVarP(
		&c.environment,
		"environment", "",
		"default",
		"flag environment",
	)
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
		"enable watch mode.",
	)

	cmd.Flags().DurationVarP(
		&c.interval,
		"interval", "i",
		time.Second,
		"interval between requests in watch mode.",
	)

	return cmd
}

func (c *evaluateCommand) run(cmd *cobra.Command, args []string) error {
	sdk, err := fliptSDK(c.address, c.token)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	flagKey := strings.TrimSpace(args[0])
	flags, err := sdk.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
		NamespaceKey:   c.namespace,
		EnvironmentKey: c.environment,
	})
	if err != nil {
		return err
	}

	flagType := flipt.FlagType_BOOLEAN_FLAG_TYPE
	for _, f := range flags.Flags {
		if f.Key == flagKey {
			flagType = f.Type
			break
		}
	}

	values := make(map[string]string, len(c.contextValues))
	for _, v := range c.contextValues {
		tokens := strings.SplitN(v, "=", 2)
		if len(tokens) != 2 {
			return fmt.Errorf("invalid context pair: %v", v)
		}
		values[strings.TrimSpace(tokens[0])] = tokens[1]
	}

	request := &evaluation.EvaluationRequest{
		FlagKey:        flagKey,
		NamespaceKey:   c.namespace,
		EnvironmentKey: c.environment,
		EntityId:       c.entityID,
		RequestId:      c.requestID,
		Context:        values,
	}

	if !c.watch {
		return c.evaluate(ctx, flagType, sdk, request)
	}

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := c.evaluate(ctx, flagType, sdk, request)
			if err != nil {
				fmt.Printf("failed to evaluate: %s\n", err)
			}
		}
	}
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

func (c *evaluateCommand) evaluate(ctx context.Context, flagType flipt.FlagType, sdk *sdk.SDK, req *evaluation.EvaluationRequest) error {
	client := sdk.Evaluation()
	switch flagType {
	case flipt.FlagType_BOOLEAN_FLAG_TYPE:
		response, err := client.Boolean(ctx, req)
		if err != nil {
			return err
		}

		boolResponse := &booleanEvaluateResponse{
			FlagKey:               response.FlagKey,
			Enabled:               response.Enabled,
			Reason:                response.Reason.String(),
			RequestID:             response.RequestId,
			RequestDurationMillis: response.RequestDurationMillis,
			Timestamp:             response.Timestamp.AsTime(),
		}

		out, err := json.Marshal(boolResponse)
		if err != nil {
			return err
		}

		fmt.Println(string(out))

		return nil
	case flipt.FlagType_VARIANT_FLAG_TYPE:
		response, err := client.Variant(ctx, req)
		if err != nil {
			return err
		}

		variantResponse := &variantEvaluationResponse{
			FlagKey:               response.FlagKey,
			Match:                 response.Match,
			Reason:                response.Reason.String(),
			VariantKey:            response.VariantKey,
			VariantAttachment:     response.VariantAttachment,
			SegmentKeys:           response.SegmentKeys,
			RequestID:             response.RequestId,
			RequestDurationMillis: response.RequestDurationMillis,
			Timestamp:             response.Timestamp.AsTime(),
		}

		out, err := json.Marshal(variantResponse)
		if err != nil {
			return err
		}

		fmt.Println(string(out))

		return nil
	default:
		return fmt.Errorf("unsupported flag type: %v", flagType)
	}
}

func fliptSDK(address, token string) (*sdk.SDK, error) {
	addr, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("invalid Flipt URL: %w", err)
	}

	var transport sdk.Transport
	switch addr.Scheme {
	case "http", "https":
		transport = sdkhttp.NewTransport(address)
	case "grpc":
		conn, err := grpc.NewClient(addr.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("failed to dial Flipt %w", err)
		}

		transport = sdkgrpc.NewTransport(conn)
	default:
		return nil, fmt.Errorf("unexpected protocol %s", addr.Scheme)
	}

	var opts []sdk.Option
	if token != "" {
		opts = append(opts, sdk.WithAuthenticationProvider(
			sdk.StaticTokenAuthenticationProvider(token),
		))
	}
	s := sdk.New(transport, opts...)
	return &s, nil
}
