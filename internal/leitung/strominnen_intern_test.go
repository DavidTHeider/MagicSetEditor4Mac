package leitung

// **Die drei Wege, auf denen ein fertiger Strom nicht zur Leitung wird** —
// Werkstück N3 von Phase 5.4
// (#1456).
//
// **Innen**, weil zwei davon nur über die Zufallsnaht zu erreichen sind: ein
// Zeugnis entsteht mit einer gezogenen Nummer, und ohne sie gibt es keines.
// Von außen wäre der Fall nur mit einem kaputten Betriebssystem herzustellen.

import (
	"errors"
	"net"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ohne erwartete Kennung wird nicht gewählt** — auch nicht über ein
// fremdes Netz. Wer wählt, weiß, wen er erreichen will; sonst bekäme ein
// Fremder an derselben Leitung dieselbe Sitzung.
func TestUeberStromOhneErwartungIstEinBefund(t *testing.T) {
	hier := eineIdentitaet(t)

	hiesig, dortig := net.Pipe()
	t.Cleanup(func() { _ = hiesig.Close(); _ = dortig.Close() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	_, fehler := UeberStrom(t.Context(), hiesig, hier, "")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("erwartet war ein Befund, kam: %v", fehler)
	}
}

// **Und ohne Zeugnis geht keiner der beiden Wege.**
func TestOhneZeugnisGehtKeinFertigerStrom(t *testing.T) {
	hier := eineIdentitaet(t)
	dort := eineIdentitaet(t)

	ohneZufall(t)

	hiesig, dortig := net.Pipe()
	t.Cleanup(func() { _ = hiesig.Close(); _ = dortig.Close() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	if _, fehler := UeberStrom(
		t.Context(), hiesig, hier, dort.Kennung()); fehler == nil {
		t.Error("UeberStrom kam ohne Zeugnis durch")
	}

	if _, fehler := NimmStrom(t.Context(), dortig, dort); fehler == nil {
		t.Error("NimmStrom kam ohne Zeugnis durch")
	}
}
