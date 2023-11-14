package evaluation

import (
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

// SetRequestIDIfNotBlank attempts to set the provided ID on the instance
// If the ID was blank, it returns the ID provided to this call.
// If the ID was not blank, it returns the ID found on the instance.
func (x *EvaluationRequest) SetRequestIDIfNotBlank(id string) string {
	if x.RequestId == "" {
		x.RequestId = id
	}

	return x.RequestId
}

// SetRequestIDIfNotBlank attempts to set the provided ID on the instance
// If the ID was blank, it returns the ID provided to this call.
// If the ID was not blank, it returns the ID found on the instance.
func (x *EvaluationResponse) SetRequestIDIfNotBlank(id string) string {
	switch r := x.Response.(type) {
	case *EvaluationResponse_VariantResponse:
		if r.VariantResponse.RequestId == "" {
			r.VariantResponse.RequestId = id
		}
		return r.VariantResponse.RequestId
	case *EvaluationResponse_BooleanResponse:
		if r.BooleanResponse.RequestId == "" {
			r.BooleanResponse.RequestId = id
		}
		return r.BooleanResponse.RequestId
	}

	return ""
}

func (x *BatchEvaluationRequest) SetRequestIDIfNotBlank(id string) string {
	if x.RequestId == "" {
		x.RequestId = id
	}

	return x.RequestId
}

func (x *VariantEvaluationResponse) SetRequestIDIfNotBlank(id string) string {
	if x.RequestId == "" {
		x.RequestId = id
	}

	return x.RequestId
}

func (x *BooleanEvaluationResponse) SetRequestIDIfNotBlank(id string) string {
	if x.RequestId == "" {
		x.RequestId = id
	}

	return x.RequestId
}

func (x *BatchEvaluationRequest) GetNamespaceKeys() (keys []string) {
	for _, r := range x.Requests {
		keys = append(keys, r.NamespaceKey)
	}
	return
}

// SetRequestIDIfNotBlank attempts to set the provided ID on the instance
// If the ID was blank, it returns the ID provided to this call.
// If the ID was not blank, it returns the ID found on the instance.
func (x *BatchEvaluationResponse) SetRequestIDIfNotBlank(id string) string {
	if x.RequestId == "" {
		x.RequestId = id
	}

	return x.RequestId
}

func (x *EvaluationResponse) GetTimestamp() *timestamppb.Timestamp {
	switch r := x.Response.(type) {
	case *EvaluationResponse_VariantResponse:
		return r.VariantResponse.GetTimestamp()
	case *EvaluationResponse_BooleanResponse:
		return r.BooleanResponse.GetTimestamp()
	}

	return nil
}

func (x *EvaluationResponse) GetRequestId() string {
	switch r := x.Response.(type) {
	case *EvaluationResponse_VariantResponse:
		return r.VariantResponse.GetRequestId()
	case *EvaluationResponse_BooleanResponse:
		return r.BooleanResponse.GetRequestId()
	}

	return ""
}

func (x *EvaluationResponse) GetRequestDurationMillis() float64 {
	switch r := x.Response.(type) {
	case *EvaluationResponse_VariantResponse:
		return r.VariantResponse.GetRequestDurationMillis()
	case *EvaluationResponse_BooleanResponse:
		return r.BooleanResponse.GetRequestDurationMillis()
	}

	return 0
}

// SetTimestamps records the start and end times on the target instance.
func (x *EvaluationResponse) SetTimestamps(start, end time.Time) {
	switch r := x.Response.(type) {
	case *EvaluationResponse_VariantResponse:
		r.VariantResponse.Timestamp = timestamppb.New(end)
		r.VariantResponse.RequestDurationMillis = float64(end.Sub(start)) / float64(time.Millisecond)
	case *EvaluationResponse_BooleanResponse:
		r.BooleanResponse.Timestamp = timestamppb.New(end)
		r.BooleanResponse.RequestDurationMillis = float64(end.Sub(start)) / float64(time.Millisecond)
	}
}

func (x *VariantEvaluationResponse) SetTimestamps(start, end time.Time) {
	x.Timestamp = timestamppb.New(end)
	x.RequestDurationMillis = float64(end.Sub(start)) / float64(time.Millisecond)
}

func (x *BooleanEvaluationResponse) SetTimestamps(start, end time.Time) {
	x.Timestamp = timestamppb.New(end)
	x.RequestDurationMillis = float64(end.Sub(start)) / float64(time.Millisecond)
}

// SetTimestamps records the start and end times on the target instance.
func (x *BatchEvaluationResponse) SetTimestamps(start, end time.Time) {
	x.RequestDurationMillis = float64(end.Sub(start)) / float64(time.Millisecond)
	for _, r := range x.Responses {
		r.SetTimestamps(start, end)
	}
}
