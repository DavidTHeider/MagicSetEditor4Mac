package domain_test

import (
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// TestZahleinstellungenNennenIhreVorgabe hält Zweitblick über Phase 6.4,
// Befund 12: die Zahlen des Betriebs standen in `EINSTELLUNGEN.md` und in der
// Geltung als „—", obwohl ohne Wahl eine feste Zahl gilt.
//
// **Und die genannte Vorgabe ist die, die gilt** — sonst stünde neben der
// Regel eine zweite Zahl, die niemand prüft (§10).
func TestZahleinstellungenNennenIhreVorgabe(t *testing.T) {
	t.Parallel()

	protokoll := func(werte map[string]string) any { return domain.ProtokollregelAus(werte) }
	messung := func(werte map[string]string) any { return domain.MessregelAus(werte) }
	speicher := func(werte map[string]string) any {
		return domain.ZwischenspeichergrenzeAus(werte[domain.ZwischenspeicherGrenze])
	}

	for _, fall := range []struct {
		schluessel, erwartet string
		regel                func(map[string]string) any
	}{
		{domain.ProtokollGrenze, "4", protokoll},
		{domain.ProtokollAufbewahrung, "1", protokoll},
		{domain.MessungAbtastung, "100", messung},
		{domain.MessungSpurgrenze, "1000", messung},
		{domain.MessungAufbewahrung, "60", messung},
		{domain.MessungProfile, "3", messung},
		{domain.ZwischenspeicherGrenze, "512", speicher},
	} {
		t.Run(fall.schluessel, func(t *testing.T) {
			t.Parallel()

			vorgabe := domain.EinstellungsVorgabe(fall.schluessel)
			if vorgabe != fall.erwartet {
				t.Fatalf("Vorgabe = %q, erwartet %q", vorgabe, fall.erwartet)
			}

			if !reflect.DeepEqual(fall.regel(map[string]string{fall.schluessel: vorgabe}), fall.regel(nil)) {
				t.Errorf("die genannte Vorgabe %q ist nicht die, die ohne Wahl gilt", vorgabe)
			}
		})
	}
}
