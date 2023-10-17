package sdk

import (
	context "context"

	flipt "go.flipt.io/flipt/rpc/flipt"
	evaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
)

func ExampleNew() {
	// see the following subpackages for transport implementations:
	// - grpc
	// - http
	var transport Transport

	client := New(transport)

	client.Flipt().GetFlag(context.Background(), &flipt.GetFlagRequest{
		NamespaceKey: "my_namespace",
		Key:          "my_flag",
	})
}

func ExampleEvaluation_Boolean() {
	// see the following subpackages for transport implementations:
	// - grpc
	// - http
	var transport Transport

	client := New(transport)

	client.Evaluation().Boolean(context.Background(), &evaluation.EvaluationRequest{
		NamespaceKey: "my_namespace",
		FlagKey:      "my_flag",
		EntityId:     "my_entity",
		Context:      map[string]string{"some": "context"},
	})
}

func ExampleEvaluation_Variant() {
	// see the following subpackages for transport implementations:
	// - grpc
	// - http
	var transport Transport

	client := New(transport)

	client.Evaluation().Variant(context.Background(), &evaluation.EvaluationRequest{
		NamespaceKey: "my_namespace",
		FlagKey:      "my_flag",
		EntityId:     "my_entity",
		Context:      map[string]string{"some": "context"},
	})
}
