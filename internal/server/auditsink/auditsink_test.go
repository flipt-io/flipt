package auditsink

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"gotest.tools/assert"
)

const (
	retries = 10
)

type sampleSink struct {
	mtx  sync.Mutex
	hits int
	fmt.Stringer
}

func (s *sampleSink) SendAudits([]AuditEvent) error {
	s.mtx.Lock()
	s.hits++
	s.mtx.Unlock()

	return nil
}

func (s *sampleSink) GetHits() int {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.hits
}

func TestPublisher(t *testing.T) {
	ss := &sampleSink{}
	publisher := NewSinkPublisher(zap.NewNop(), 2, []AuditSink{ss}, 10*time.Second)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			publisher.Publish(&AuditEvent{
				ResourceName: "sample-flag",
			})
		}()
	}

	wg.Wait()

	<-time.After(10 * time.Second)

	for i := 0; i < retries; i++ {
		if ss.GetHits() == 5 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	assert.Equal(t, ss.GetHits(), 5)
}
