package storage_test

// **Ein Kennwort wechselt, und sonst nichts** — ID-02
// (#1701).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinKennwortWechseltUndSonstNichts(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	angelegt, fehler := ablage.NutzerAnlegen(t.Context(), domain.Nutzer{Name: "Ada", Email: "ada@example.org", Kennwort: "alt"})
	if fehler != nil {
		t.Fatal(fehler)
	}

	if gesetzt := ablage.KennwortSetzen(t.Context(), angelegt.ID, "neu"); gesetzt != nil {
		t.Fatal(gesetzt)
	}

	gelesen, fehler := ablage.NutzerZuKennung(t.Context(), "ada@example.org")
	if fehler != nil || gelesen.Kennwort != "neu" || gelesen.Name != "Ada" {
		t.Errorf("%+v %v", gelesen, fehler)
	}

	if fehler := ablage.KennwortSetzen(t.Context(), 4711, "x"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("unbekannter Nutzer: %v", fehler)
	}
}

// **Code und Kennwort in einem Geschäft** — Zweitblick über ID-02, Befund 5:
// ein verbrauchter Code setzt nichts, und was nicht gesetzt wird, verbraucht
// keinen Code.
func TestCodeUndKennwortGehenZusammen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	ada, fehler := ablage.NutzerAnlegen(t.Context(), domain.Nutzer{Name: "Ada", Email: "ada@example.org", Kennwort: "alt"})
	if fehler != nil {
		t.Fatal(fehler)
	}

	if setzfehler := ablage.CodesSetzen(t.Context(), ada.ID, []string{"gehuetet-1"}); setzfehler != nil {
		t.Fatal(setzfehler)
	}

	frei, fehler := ablage.FreieCodes(t.Context(), ada.ID)
	if fehler != nil || len(frei) != 1 {
		t.Fatalf("%v %v", frei, fehler)
	}

	// **Ein unbekannter Nutzer setzt nichts** — und der Code bleibt frei.
	if gesetzt := ablage.KennwortMitCodeSetzen(t.Context(), 4711, frei[0].ID, "neu", domain.SitzungAngemeldet); !errors.Is(gesetzt, domain.ErrNotFound) {
		t.Errorf("unbekannter Nutzer: %v", gesetzt)
	}

	if nochFrei, lesefehler := ablage.FreieCodes(t.Context(), ada.ID); lesefehler != nil || len(nochFrei) != 1 {
		t.Errorf("ein gescheitertes Geschäft hat den Code verbraucht: %v", lesefehler)
	}

	if gesetzt := ablage.KennwortMitCodeSetzen(t.Context(), ada.ID, frei[0].ID, "neu", domain.SitzungAngemeldet); gesetzt != nil {
		t.Fatal(gesetzt)
	}

	gelesen, lesefehler := ablage.NutzerZuKennung(t.Context(), "ada@example.org")
	if lesefehler != nil || gelesen.Kennwort != "neu" {
		t.Errorf("Kennwort: %q %v", gelesen.Kennwort, lesefehler)
	}

	// **Ein verbrauchter Code setzt nichts.**
	if gesetzt := ablage.KennwortMitCodeSetzen(t.Context(), ada.ID, frei[0].ID, "noch-neuer", domain.SitzungAngemeldet); !errors.Is(gesetzt, domain.ErrNotFound) {
		t.Errorf("zweimal: %v", gesetzt)
	}

	if danach, lesefehler := ablage.NutzerZuKennung(t.Context(), "ada@example.org"); lesefehler != nil || danach.Kennwort != "neu" {
		t.Errorf("nach dem zweiten Versuch: %q %v", danach.Kennwort, lesefehler)
	}
}
