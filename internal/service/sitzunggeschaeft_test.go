package service_test

// **Ein Übergang der Sitzung ist ein Geschäft** — Nachlauf über ID-02,
// Befund 2 (#1701).
//
// Beim Anmelden standen die Zeiger, bevor der Zustand festgehalten war; bei
// der gesperrten Wiederherstellung Kennwort und Code. Scheiterte das
// Festhalten, blieb der erste Schritt stehen.

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// festesZiel zeigt jede Anmeldung auf Profil 1 und zieht nichts nach.
var festesZiel = service.Anmeldeziel{
	Profil: func(context.Context, int64) (int64, error) { return 1, nil },
}

// zaehlendesZiel zählt, wie oft der Prozess nachgezogen wurde.
func zaehlendesZiel(gezaehlt *int) service.Anmeldeziel {
	return service.Anmeldeziel{
		Profil:     festesZiel.Profil,
		Nachziehen: func(context.Context, int64) { *gezaehlt++ },
	}
}

// AnmeldungFesthalten stellt Zeiger und Zustand zusammen — oder keines.
func (d *sitzungsdoppel) AnmeldungFesthalten(_ context.Context, nutzerID, profilID int64) error {
	d.gerufen = append(d.gerufen, "AnmeldungFesthalten")

	if d.festhaltefehler != nil {
		return d.festhaltefehler
	}

	d.festgehalten = domain.SitzungAngemeldet
	d.zeiger = [2]int64{nutzerID, profilID}

	return nil
}

func TestEinGescheitertesFesthaltenLaesstDieAnmeldungAus(t *testing.T) {
	t.Parallel()

	sitzung, doppel := neueSitzung(t)
	if fehler := sitzung.Abmelden(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	nachgezogen := 0
	doppel.festhaltefehler = errors.New("die Platte ist voll")

	if fehler := sitzung.Anmelden(t.Context(), 1, "richtig-und-lang", "", zaehlendesZiel(&nachgezogen)); !errors.Is(fehler, doppel.festhaltefehler) {
		t.Fatalf("das Festhalten scheiterte still: %v", fehler)
	}

	if zustand, _ := sitzung.Stand(); zustand != domain.SitzungAbgemeldet || nachgezogen != 0 || doppel.zeiger != [2]int64{} {
		t.Errorf("halber Stand: %s, nachgezogen %d, Zeiger %v", zustand, nachgezogen, doppel.zeiger)
	}

	doppel.festhaltefehler = nil

	if fehler := sitzung.Anmelden(t.Context(), 1, "richtig-und-lang", "", zaehlendesZiel(&nachgezogen)); fehler != nil {
		t.Fatal(fehler)
	}

	if zustand, wer := sitzung.Stand(); zustand != domain.SitzungAngemeldet || wer != 1 || nachgezogen != 1 || doppel.zeiger != [2]int64{1, 1} {
		t.Errorf("%s/%d, nachgezogen %d, Zeiger %v", zustand, wer, nachgezogen, doppel.zeiger)
	}
}

func TestDieGesperrteWiederherstellungHaeltDenZustandImGeschaeft(t *testing.T) {
	t.Parallel()

	sitzung, doppel := neueSitzung(t)
	if fehler := sitzung.Sperren(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	doppel.codesetzfehler = errors.New("das Geschäft scheitert")

	if fehler := sitzung.Wiederherstellen(t.Context(), 1, "zettel-1", "neues-und-lang"); !errors.Is(fehler, doppel.codesetzfehler) {
		t.Fatalf("still: %v", fehler)
	}

	if zustand, _ := sitzung.Stand(); zustand != domain.SitzungGesperrt || doppel.festgehalten != domain.SitzungGesperrt {
		t.Errorf("nach dem gescheiterten Geschäft: %s, festgehalten %s", zustand, doppel.festgehalten)
	}

	doppel.codesetzfehler = nil
	doppel.gerufen = nil

	if fehler := sitzung.Wiederherstellen(t.Context(), 1, "zettel-1", "neues-und-lang"); fehler != nil {
		t.Fatal(fehler)
	}

	// **Kein zweiter Schritt danach** — der Zustand ging mit dem Geschäft.
	if zustand, _ := sitzung.Stand(); zustand != domain.SitzungAngemeldet || doppel.festgehalten != domain.SitzungAngemeldet ||
		!slices.Equal(doppel.gerufen, []string{"CodeFinden", "KennwortMitCodeSetzen"}) {
		t.Errorf("%s, festgehalten %s, gerufen %v", zustand, doppel.festgehalten, doppel.gerufen)
	}
}
