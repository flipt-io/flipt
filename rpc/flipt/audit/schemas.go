package audit

import (
	_ "embed"
)

//go:embed event.avsc
var AvroSchema string

//go:embed event.proto
var ProtoSchema string
