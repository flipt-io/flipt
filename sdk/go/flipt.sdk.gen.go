// Code generated by protoc-gen-go-flipt-sdk. DO NOT EDIT.

package sdk

import (
	context "context"
	flipt "go.flipt.io/flipt/rpc/flipt"
)

type Flipt struct {
	transport              flipt.FliptClient
	authenticationProvider ClientAuthenticationProvider
}

func (x *Flipt) GetNamespace(ctx context.Context, v *flipt.GetNamespaceRequest) (*flipt.Namespace, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.GetNamespace(ctx, v)
}

func (x *Flipt) ListNamespaces(ctx context.Context, v *flipt.ListNamespaceRequest) (*flipt.NamespaceList, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.ListNamespaces(ctx, v)
}

func (x *Flipt) CreateNamespace(ctx context.Context, v *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.CreateNamespace(ctx, v)
}

func (x *Flipt) UpdateNamespace(ctx context.Context, v *flipt.UpdateNamespaceRequest) (*flipt.Namespace, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.UpdateNamespace(ctx, v)
}

func (x *Flipt) DeleteNamespace(ctx context.Context, v *flipt.DeleteNamespaceRequest) error {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return err
	}
	_, err = x.transport.DeleteNamespace(ctx, v)
	return err
}

func (x *Flipt) GetFlag(ctx context.Context, v *flipt.GetFlagRequest) (*flipt.Flag, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.GetFlag(ctx, v)
}

func (x *Flipt) ListFlags(ctx context.Context, v *flipt.ListFlagRequest) (*flipt.FlagList, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.ListFlags(ctx, v)
}

func (x *Flipt) CreateFlag(ctx context.Context, v *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.CreateFlag(ctx, v)
}

func (x *Flipt) UpdateFlag(ctx context.Context, v *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.UpdateFlag(ctx, v)
}

func (x *Flipt) DeleteFlag(ctx context.Context, v *flipt.DeleteFlagRequest) error {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return err
	}
	_, err = x.transport.DeleteFlag(ctx, v)
	return err
}

func (x *Flipt) CreateVariant(ctx context.Context, v *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.CreateVariant(ctx, v)
}

func (x *Flipt) UpdateVariant(ctx context.Context, v *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.UpdateVariant(ctx, v)
}

func (x *Flipt) DeleteVariant(ctx context.Context, v *flipt.DeleteVariantRequest) error {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return err
	}
	_, err = x.transport.DeleteVariant(ctx, v)
	return err
}

func (x *Flipt) GetRule(ctx context.Context, v *flipt.GetRuleRequest) (*flipt.Rule, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.GetRule(ctx, v)
}

func (x *Flipt) ListRules(ctx context.Context, v *flipt.ListRuleRequest) (*flipt.RuleList, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.ListRules(ctx, v)
}

func (x *Flipt) CreateRule(ctx context.Context, v *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.CreateRule(ctx, v)
}

func (x *Flipt) UpdateRule(ctx context.Context, v *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.UpdateRule(ctx, v)
}

func (x *Flipt) OrderRules(ctx context.Context, v *flipt.OrderRulesRequest) error {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return err
	}
	_, err = x.transport.OrderRules(ctx, v)
	return err
}

func (x *Flipt) DeleteRule(ctx context.Context, v *flipt.DeleteRuleRequest) error {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return err
	}
	_, err = x.transport.DeleteRule(ctx, v)
	return err
}

func (x *Flipt) GetRollout(ctx context.Context, v *flipt.GetRolloutRequest) (*flipt.Rollout, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.GetRollout(ctx, v)
}

func (x *Flipt) ListRollouts(ctx context.Context, v *flipt.ListRolloutRequest) (*flipt.RolloutList, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.ListRollouts(ctx, v)
}

func (x *Flipt) CreateRollout(ctx context.Context, v *flipt.CreateRolloutRequest) (*flipt.Rollout, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.CreateRollout(ctx, v)
}

func (x *Flipt) UpdateRollout(ctx context.Context, v *flipt.UpdateRolloutRequest) (*flipt.Rollout, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.UpdateRollout(ctx, v)
}

func (x *Flipt) DeleteRollout(ctx context.Context, v *flipt.DeleteRolloutRequest) error {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return err
	}
	_, err = x.transport.DeleteRollout(ctx, v)
	return err
}

func (x *Flipt) OrderRollouts(ctx context.Context, v *flipt.OrderRolloutsRequest) error {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return err
	}
	_, err = x.transport.OrderRollouts(ctx, v)
	return err
}

func (x *Flipt) CreateDistribution(ctx context.Context, v *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.CreateDistribution(ctx, v)
}

func (x *Flipt) UpdateDistribution(ctx context.Context, v *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.UpdateDistribution(ctx, v)
}

func (x *Flipt) DeleteDistribution(ctx context.Context, v *flipt.DeleteDistributionRequest) error {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return err
	}
	_, err = x.transport.DeleteDistribution(ctx, v)
	return err
}

func (x *Flipt) GetSegment(ctx context.Context, v *flipt.GetSegmentRequest) (*flipt.Segment, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.GetSegment(ctx, v)
}

func (x *Flipt) ListSegments(ctx context.Context, v *flipt.ListSegmentRequest) (*flipt.SegmentList, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.ListSegments(ctx, v)
}

func (x *Flipt) CreateSegment(ctx context.Context, v *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.CreateSegment(ctx, v)
}

func (x *Flipt) UpdateSegment(ctx context.Context, v *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.UpdateSegment(ctx, v)
}

func (x *Flipt) DeleteSegment(ctx context.Context, v *flipt.DeleteSegmentRequest) error {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return err
	}
	_, err = x.transport.DeleteSegment(ctx, v)
	return err
}

func (x *Flipt) CreateConstraint(ctx context.Context, v *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.CreateConstraint(ctx, v)
}

func (x *Flipt) UpdateConstraint(ctx context.Context, v *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.UpdateConstraint(ctx, v)
}

func (x *Flipt) DeleteConstraint(ctx context.Context, v *flipt.DeleteConstraintRequest) error {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return err
	}
	_, err = x.transport.DeleteConstraint(ctx, v)
	return err
}
