package flipt

import (
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const DefaultNamespace = "default"

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
	if x.RequestId == "" {
		x.RequestId = id
	}

	return x.RequestId
}

// SetRequestIDIfNotBlank attempts to set the provided ID on the instance
// If the ID was blank, it returns the ID provided to this call.
// If the ID was not blank, it returns the ID found on the instance.
func (x *BatchEvaluationRequest) SetRequestIDIfNotBlank(id string) string {
	if x.RequestId == "" {
		x.RequestId = id
	}

	return x.RequestId
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

// SetTimestamps records the start and end times on the target instance.
func (x *EvaluationResponse) SetTimestamps(start, end time.Time) {
	x.Timestamp = timestamppb.New(end)
	x.RequestDurationMillis = float64(end.Sub(start)) / float64(time.Millisecond)
}

// SetTimestamps records the start and end times on the target instance.
func (x *BatchEvaluationResponse) SetTimestamps(start, end time.Time) {
	x.RequestDurationMillis = float64(end.Sub(start)) / float64(time.Millisecond)
	for _, r := range x.Responses {
		r.Timestamp = timestamppb.New(end)
	}
}
