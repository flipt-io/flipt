package flipt

import (
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

// Now returns a timestamppb.Time pointer which is rounded to
// microsecond precision.
// This is enough for Flipts purposes and the lowest common
// denominator for the various backends.
func Now() *timestamppb.Timestamp {
	now := timestamppb.Now()
	now.Nanos = int32(now.AsTime().Round(time.Microsecond).Nanosecond())
	return now
}
