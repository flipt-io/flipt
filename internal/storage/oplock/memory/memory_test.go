package memory

import (
	"testing"

	oplocktesting "go.flipt.io/flipt/internal/storage/oplock/testing"
)

func Test_Harness(t *testing.T) {
	oplocktesting.Harness(t, New())
}
