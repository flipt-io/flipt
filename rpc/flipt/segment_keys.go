package flipt

// removeDuplicates is a utility function that will deduplicate a slice of strings.
func removeDuplicates(src []string) []string {
	allKeys := make(map[string]bool)

	dest := []string{}

	for _, item := range src {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			dest = append(dest, item)
		}
	}

	return dest
}

// SanitizeSegmentKeys takes in a rule request and combines the two `SegmentKey“
// and `SegmentKeys` field and deduplicates the values.
func (r *CreateRuleRequest) SanitizeSegmentKeys() {
	segmentKeys := make([]string, 0)

	if len(r.SegmentKeys) > 0 {
		segmentKeys = append(segmentKeys, r.SegmentKeys...)
	} else if r.SegmentKey != "" {
		segmentKeys = append(segmentKeys, r.SegmentKey)
	}

	r.SegmentKeys = removeDuplicates(segmentKeys)
}

func (r *UpdateRuleRequest) SanitizeSegmentKeys() {
	segmentKeys := make([]string, 0)

	if len(r.SegmentKeys) > 0 {
		segmentKeys = append(segmentKeys, r.SegmentKeys...)
	} else if r.SegmentKey != "" {
		segmentKeys = append(segmentKeys, r.SegmentKey)
	}

	r.SegmentKeys = removeDuplicates(segmentKeys)
}

func (r *CreateRolloutRequest) SanitizeSegmentKeys() {
	_, ok := r.GetRule().(*CreateRolloutRequest_Segment)
	if !ok {
		return
	}

	segmentKeys := make([]string, 0)

	if len(r.GetSegment().GetSegmentKeys()) > 0 {
		segmentKeys = append(segmentKeys, r.GetSegment().GetSegmentKeys()...)
	} else if r.GetSegment().GetSegmentKey() != "" {
		segmentKeys = append(segmentKeys, r.GetSegment().GetSegmentKey())
	}

	r.GetSegment().SegmentKeys = removeDuplicates(segmentKeys)
}

func (r *UpdateRolloutRequest) SanitizeSegmentKeys() {
	_, ok := r.GetRule().(*UpdateRolloutRequest_Segment)
	if !ok {
		return
	}

	segmentKeys := make([]string, 0)

	if len(r.GetSegment().GetSegmentKeys()) > 0 {
		segmentKeys = append(segmentKeys, r.GetSegment().GetSegmentKeys()...)
	} else if r.GetSegment().GetSegmentKey() != "" {
		segmentKeys = append(segmentKeys, r.GetSegment().GetSegmentKey())
	}

	r.GetSegment().SegmentKeys = removeDuplicates(segmentKeys)
}
