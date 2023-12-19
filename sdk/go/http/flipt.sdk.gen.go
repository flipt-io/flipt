// Code generated by protoc-gen-go-flipt-sdk. DO NOT EDIT.

package http

import (
	bytes "bytes"
	context "context"
	fmt "fmt"
	flipt "go.flipt.io/flipt/rpc/flipt"
	grpc "google.golang.org/grpc"
	protojson "google.golang.org/protobuf/encoding/protojson"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	io "io"
	http "net/http"
	url "net/url"
)

type FliptClient struct {
	client *http.Client
	addr   string
}

func (x *FliptClient) Evaluate(ctx context.Context, v *flipt.EvaluationRequest, _ ...grpc.CallOption) (*flipt.EvaluationResponse, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/evaluate", v.NamespaceKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.EvaluationResponse
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) BatchEvaluate(ctx context.Context, v *flipt.BatchEvaluationRequest, _ ...grpc.CallOption) (*flipt.BatchEvaluationResponse, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/batch-evaluate", v.NamespaceKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.BatchEvaluationResponse
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) GetNamespace(ctx context.Context, v *flipt.GetNamespaceRequest, _ ...grpc.CallOption) (*flipt.Namespace, error) {
	var body io.Reader
	values := url.Values{}
	values.Set("reference", v.Reference)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, x.addr+fmt.Sprintf("/api/v1/namespaces/%v", v.Key), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Namespace
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) ListNamespaces(ctx context.Context, v *flipt.ListNamespaceRequest, _ ...grpc.CallOption) (*flipt.NamespaceList, error) {
	var body io.Reader
	values := url.Values{}
	values.Set("limit", fmt.Sprintf("%v", v.Limit))
	values.Set("offset", fmt.Sprintf("%v", v.Offset))
	values.Set("pageToken", v.PageToken)
	values.Set("reference", v.Reference)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, x.addr+"/api/v1/namespaces", body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.NamespaceList
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) CreateNamespace(ctx context.Context, v *flipt.CreateNamespaceRequest, _ ...grpc.CallOption) (*flipt.Namespace, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, x.addr+"/api/v1/namespaces", body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Namespace
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) UpdateNamespace(ctx context.Context, v *flipt.UpdateNamespaceRequest, _ ...grpc.CallOption) (*flipt.Namespace, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, x.addr+fmt.Sprintf("/api/v1/namespaces/%v", v.Key), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Namespace
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) DeleteNamespace(ctx context.Context, v *flipt.DeleteNamespaceRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	var body io.Reader
	var values url.Values
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, x.addr+fmt.Sprintf("/api/v1/namespaces/%v", v.Key), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output emptypb.Empty
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) GetFlag(ctx context.Context, v *flipt.GetFlagRequest, _ ...grpc.CallOption) (*flipt.Flag, error) {
	var body io.Reader
	values := url.Values{}
	values.Set("reference", v.Reference)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v", v.NamespaceKey, v.Key), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Flag
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) ListFlags(ctx context.Context, v *flipt.ListFlagRequest, _ ...grpc.CallOption) (*flipt.FlagList, error) {
	var body io.Reader
	values := url.Values{}
	values.Set("limit", fmt.Sprintf("%v", v.Limit))
	values.Set("offset", fmt.Sprintf("%v", v.Offset))
	values.Set("pageToken", v.PageToken)
	values.Set("reference", v.Reference)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags", v.NamespaceKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.FlagList
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) CreateFlag(ctx context.Context, v *flipt.CreateFlagRequest, _ ...grpc.CallOption) (*flipt.Flag, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags", v.NamespaceKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Flag
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) UpdateFlag(ctx context.Context, v *flipt.UpdateFlagRequest, _ ...grpc.CallOption) (*flipt.Flag, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v", v.NamespaceKey, v.Key), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Flag
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) DeleteFlag(ctx context.Context, v *flipt.DeleteFlagRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	var body io.Reader
	var values url.Values
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v", v.NamespaceKey, v.Key), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output emptypb.Empty
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) CreateVariant(ctx context.Context, v *flipt.CreateVariantRequest, _ ...grpc.CallOption) (*flipt.Variant, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/variants", v.NamespaceKey, v.FlagKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Variant
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) UpdateVariant(ctx context.Context, v *flipt.UpdateVariantRequest, _ ...grpc.CallOption) (*flipt.Variant, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/variants/%v", v.NamespaceKey, v.FlagKey, v.Id), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Variant
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) DeleteVariant(ctx context.Context, v *flipt.DeleteVariantRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	var body io.Reader
	var values url.Values
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/variants/%v", v.NamespaceKey, v.FlagKey, v.Id), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output emptypb.Empty
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) GetRule(ctx context.Context, v *flipt.GetRuleRequest, _ ...grpc.CallOption) (*flipt.Rule, error) {
	var body io.Reader
	values := url.Values{}
	values.Set("reference", v.Reference)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rules/%v", v.NamespaceKey, v.FlagKey, v.Id), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Rule
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) ListRules(ctx context.Context, v *flipt.ListRuleRequest, _ ...grpc.CallOption) (*flipt.RuleList, error) {
	var body io.Reader
	values := url.Values{}
	values.Set("limit", fmt.Sprintf("%v", v.Limit))
	values.Set("offset", fmt.Sprintf("%v", v.Offset))
	values.Set("pageToken", v.PageToken)
	values.Set("reference", v.Reference)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rules", v.NamespaceKey, v.FlagKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.RuleList
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) CreateRule(ctx context.Context, v *flipt.CreateRuleRequest, _ ...grpc.CallOption) (*flipt.Rule, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rules", v.NamespaceKey, v.FlagKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Rule
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) UpdateRule(ctx context.Context, v *flipt.UpdateRuleRequest, _ ...grpc.CallOption) (*flipt.Rule, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rules/%v", v.NamespaceKey, v.FlagKey, v.Id), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Rule
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) OrderRules(ctx context.Context, v *flipt.OrderRulesRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rules/order", v.NamespaceKey, v.FlagKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output emptypb.Empty
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) DeleteRule(ctx context.Context, v *flipt.DeleteRuleRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	var body io.Reader
	var values url.Values
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rules/%v", v.NamespaceKey, v.FlagKey, v.Id), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output emptypb.Empty
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) GetRollout(ctx context.Context, v *flipt.GetRolloutRequest, _ ...grpc.CallOption) (*flipt.Rollout, error) {
	var body io.Reader
	values := url.Values{}
	values.Set("reference", v.Reference)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rollouts/%v", v.NamespaceKey, v.FlagKey, v.Id), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Rollout
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) ListRollouts(ctx context.Context, v *flipt.ListRolloutRequest, _ ...grpc.CallOption) (*flipt.RolloutList, error) {
	var body io.Reader
	values := url.Values{}
	values.Set("limit", fmt.Sprintf("%v", v.Limit))
	values.Set("pageToken", v.PageToken)
	values.Set("reference", v.Reference)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rollouts", v.NamespaceKey, v.FlagKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.RolloutList
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) CreateRollout(ctx context.Context, v *flipt.CreateRolloutRequest, _ ...grpc.CallOption) (*flipt.Rollout, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rollouts", v.NamespaceKey, v.FlagKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Rollout
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) UpdateRollout(ctx context.Context, v *flipt.UpdateRolloutRequest, _ ...grpc.CallOption) (*flipt.Rollout, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rollouts/%v", v.NamespaceKey, v.FlagKey, v.Id), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Rollout
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) DeleteRollout(ctx context.Context, v *flipt.DeleteRolloutRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	var body io.Reader
	var values url.Values
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rollouts/%v", v.NamespaceKey, v.FlagKey, v.Id), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output emptypb.Empty
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) OrderRollouts(ctx context.Context, v *flipt.OrderRolloutsRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rollouts/order", v.NamespaceKey, v.FlagKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output emptypb.Empty
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) CreateDistribution(ctx context.Context, v *flipt.CreateDistributionRequest, _ ...grpc.CallOption) (*flipt.Distribution, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rules/%v/distributions", v.NamespaceKey, v.FlagKey, v.RuleId), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Distribution
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) UpdateDistribution(ctx context.Context, v *flipt.UpdateDistributionRequest, _ ...grpc.CallOption) (*flipt.Distribution, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rules/%v/distributions/%v", v.NamespaceKey, v.FlagKey, v.RuleId, v.Id), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Distribution
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) DeleteDistribution(ctx context.Context, v *flipt.DeleteDistributionRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	var body io.Reader
	values := url.Values{}
	values.Set("variantId", v.VariantId)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/flags/%v/rules/%v/distributions/%v", v.NamespaceKey, v.FlagKey, v.RuleId, v.Id), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output emptypb.Empty
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) GetSegment(ctx context.Context, v *flipt.GetSegmentRequest, _ ...grpc.CallOption) (*flipt.Segment, error) {
	var body io.Reader
	values := url.Values{}
	values.Set("reference", v.Reference)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/segments/%v", v.NamespaceKey, v.Key), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Segment
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) ListSegments(ctx context.Context, v *flipt.ListSegmentRequest, _ ...grpc.CallOption) (*flipt.SegmentList, error) {
	var body io.Reader
	values := url.Values{}
	values.Set("limit", fmt.Sprintf("%v", v.Limit))
	values.Set("offset", fmt.Sprintf("%v", v.Offset))
	values.Set("pageToken", v.PageToken)
	values.Set("reference", v.Reference)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/segments", v.NamespaceKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.SegmentList
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) CreateSegment(ctx context.Context, v *flipt.CreateSegmentRequest, _ ...grpc.CallOption) (*flipt.Segment, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/segments", v.NamespaceKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Segment
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) UpdateSegment(ctx context.Context, v *flipt.UpdateSegmentRequest, _ ...grpc.CallOption) (*flipt.Segment, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/segments/%v", v.NamespaceKey, v.Key), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Segment
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) DeleteSegment(ctx context.Context, v *flipt.DeleteSegmentRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	var body io.Reader
	var values url.Values
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/segments/%v", v.NamespaceKey, v.Key), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output emptypb.Empty
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) CreateConstraint(ctx context.Context, v *flipt.CreateConstraintRequest, _ ...grpc.CallOption) (*flipt.Constraint, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/segments/%v/constraints", v.NamespaceKey, v.SegmentKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Constraint
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) UpdateConstraint(ctx context.Context, v *flipt.UpdateConstraintRequest, _ ...grpc.CallOption) (*flipt.Constraint, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/segments/%v/constraints/%v", v.NamespaceKey, v.SegmentKey, v.Id), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output flipt.Constraint
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (x *FliptClient) DeleteConstraint(ctx context.Context, v *flipt.DeleteConstraintRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	var body io.Reader
	var values url.Values
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, x.addr+fmt.Sprintf("/api/v1/namespaces/%v/segments/%v/constraints/%v", v.NamespaceKey, v.SegmentKey, v.Id), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output emptypb.Empty
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (t Transport) FliptClient() flipt.FliptClient {
	return &FliptClient{client: t.client, addr: t.addr}
}
