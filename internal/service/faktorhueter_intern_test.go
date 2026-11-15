package service

// **Der Weg, den ein tauglicher Code nie nimmt** — Werkstück 7 von Phase 4.12
// (#1351).
//
// **Ein interner Lauf**, weil die Naht unexportiert ist — und sie ist es aus
// gutem Grund: einen Hüter von außen setzen zu können wäre eine Fläche, die
// niemand braucht.
//
// # Warum es diesen Lauf gibt
//
// **Weil der Zweig sonst eine Behauptung bliebe.** `kennwort.Huete` scheitert
// nur an der Länge, und ein gruppierter Code hat vierzehn Zeichen — er liegt
// sicher zwischen `Mindestlaenge` und `Hoechstlaenge`. Der Fehlerweg ist
// **richtig** und **unerreichbar**; ohne diese Naht stünde er da, ohne dass
// jemand wüsste, ob er trägt (§3).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Scheitert das Hüten, entsteht kein halbes Konto.**
func TestScheitertDasHuetenEntstehtKeinKonto(t *testing.T) {
	t.Parallel()

	eigener := errors.New("der Hüter hat aufgegeben")
	ablage := &hueterdoppel{}

	dienst := NeuerZweiterfaktor(ablage, ablage, ablage)
	dienst.huete = func(string) (string, error) { return "", eigener }

	_, fehler := dienst.TOTPEinrichten(t.Context(), 1)
	if !errors.Is(fehler, eigener) {
		t.Fatalf("gemeldet wurde: %v", fehler)
	}

	if ablage.gesetzt {
		t.Error("es wurden trotzdem Codes abgelegt")
	}
}

// hueterdoppel ist die kleinste Ablage, die diesen Lauf trägt.
//
// **Ein eigenes und nicht das aus `faktordoppel_test.go`**: das steht im
// äußeren Testpaket und ist von hier nicht erreichbar. Ein paar Zeilen Doppel
// sind billiger als ein Paketwechsel für die ganze Datei.
type hueterdoppel struct{ gesetzt bool }

func (*hueterdoppel) Faktoren(context.Context, int64) ([]domain.Faktor, error) {
	return nil, nil
}

func (*hueterdoppel) FaktorAnlegen(_ context.Context, neu domain.Faktor) (
	domain.Faktor, error,
) {
	return neu, nil
}

func (*hueterdoppel) FaktorLoeschen(context.Context, int64, domain.Faktorart) error {
	return nil
}

func (h *hueterdoppel) CodesSetzen(context.Context, int64, []string) error {
	h.gesetzt = true

	return nil
}

func (*hueterdoppel) FreieCodes(context.Context, int64) (
	[]domain.Wiederherstellungscode, error,
) {
	return nil, nil
}

func (*hueterdoppel) CodeVerbrauchen(context.Context, int64) error { return nil }

func (*hueterdoppel) SchrittVerbrauchen(context.Context, int64, int64) (bool, error) {
	return true, nil
}

func (*hueterdoppel) FaktorversuchBeginnen(context.Context, int64, time.Time) (time.Time, error) {
	return time.Time{}, nil
}

func (*hueterdoppel) FaktorversuchGelungen(context.Context, int64) error { return nil }

func (*hueterdoppel) Setze(context.Context, int64, string) error { return nil }

func (*hueterdoppel) Hole(context.Context, int64) (string, error) { return "", nil }

func (*hueterdoppel) Nimm(context.Context, int64) error { return nil }

func (*hueterdoppel) Vorhanden(context.Context) bool { return false }

func (*hueterdoppel) Bestaetige(context.Context, string) error { return nil }
