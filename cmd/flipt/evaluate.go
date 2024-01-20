package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	sdk "go.flipt.io/flipt/sdk/go"
	"go.uber.org/zap"
)

type evaluateCommand struct {
	address       string
	requestID     string
	entityID      string
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
		&c.namespace,
		"namespace", "n",
		"default",
		"flag namespace.",
	)
	cmd.Flags().StringVarP(
		&c.entityID,
		"entity-id", "e",
		uuid.Must(uuid.NewV4()).String(),
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
	flag, err := sdk.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{NamespaceKey: c.namespace, Key: flagKey})
	if err != nil {
		return err
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
		FlagKey:      flagKey,
		NamespaceKey: c.namespace,
		EntityId:     c.entityID,
		RequestId:    c.requestID,
		Context:      values,
	}

	if !c.watch {
		return c.evaluate(ctx, flag.Type, sdk, request)
	}

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := c.evaluate(ctx, flag.Type, sdk, request)
			if err != nil {
				defaultLogger.Error("failed to evaluate", zap.Error(err))
			}
		}
	}
}

func (c *evaluateCommand) evaluate(ctx context.Context, flagType flipt.FlagType, sdk *sdk.SDK, req *evaluation.EvaluationRequest) error {
	client := sdk.Evaluation()
	switch flagType {
	case flipt.FlagType_BOOLEAN_FLAG_TYPE:
		response, err := client.Boolean(ctx, req)
		if err != nil {
			return err
		}
		defaultLogger.Info("boolean",
			zap.String("flag", response.FlagKey),
			zap.Bool("enabled", response.Enabled),
			zap.String("reason", response.Reason.String()),
			zap.String("requestId", response.RequestId),
			zap.Float64("requestDurationMillis", response.RequestDurationMillis),
			zap.Time("timestamp", response.Timestamp.AsTime()),
		)
		return nil
	case flipt.FlagType_VARIANT_FLAG_TYPE:
		response, err := client.Variant(ctx, req)
		if err != nil {
			return err
		}
		defaultLogger.Info("variant",
			zap.String("flag", response.FlagKey),
			zap.Bool("match", response.Match),
			zap.String("reason", response.Reason.String()),
			zap.String("variantKey", response.VariantKey),
			zap.String("variantAttachment", response.VariantAttachment),
			zap.Strings("segmentKeys", response.SegmentKeys),
			zap.String("requestId", response.RequestId),
			zap.Float64("requestDurationMillis", response.RequestDurationMillis),
			zap.Time("timestamp", response.Timestamp.AsTime()),
		)
		return nil
	default:
		return fmt.Errorf("unsupported flag type: %v", flagType)
	}
}
