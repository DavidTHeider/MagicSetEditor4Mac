package domain_test

import (
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func BenchmarkGesehen(b *testing.B) {
	exemplare := make([]*domain.Exemplar, 10000)
	for i := range exemplare {
		exemplare[i] = &domain.Exemplar{
			ID:       int64(i),
			Printing: fmt.Sprintf("druck-%05d", i),
			Menge:    1,
			Name:     fmt.Sprintf("Ätherkarte %05d", (i*7919)%10000),
			Setcode:  "m11",
		}
	}

	sicht := domain.Sammlungssicht{}

	b.ResetTimer()
	for b.Loop() {
		_ = sicht.Gesehen(exemplare)
	}
}
