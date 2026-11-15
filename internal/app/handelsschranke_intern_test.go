package app

// **Die Handelsrichtlinie aus den Einstellungen** — #2435: was die
// Einstellung sagt, gilt; ohne Bestand und bei einem Lesefehler hält es an.

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieEinstellungsschrankeLiestWasGilt(t *testing.T) {
	t.Parallel()

	db, fehler := storage.Oeffnen(t.Context(), filepath.Join(t.TempDir(), "bestand.db"))
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Aufräumen

	if migriert := storage.Migriere(t.Context(), db, nil); migriert != nil {
		t.Fatal(migriert)
	}

	gezeigt, fehler := storage.MitGezeigtemProfil(t.Context(), db)
	if fehler != nil {
		t.Fatal(fehler)
	}

	anwendung := NeueApp("")
	anwendung.einstellungen = einstellungsverwalterUeber(db)
	schranke := einstellungsschranke{a: anwendung}

	if richtlinie, fehler := schranke.Handelsrichtlinie(gezeigt); fehler != nil || richtlinie != domain.Standardhandelsrichtlinie() {
		t.Errorf("ohne Wahl: %+v, %v", richtlinie, fehler)
	}

	if gesetzt := anwendung.einstellungen.Setzen(gezeigt, domain.RichtlinieHandel, domain.HandelSchrankeAus); gesetzt != nil {
		t.Fatal(gesetzt)
	}

	if richtlinie, fehler := schranke.Handelsrichtlinie(gezeigt); fehler != nil || richtlinie.Handeln {
		t.Errorf("nach „aus“: %+v, %v", richtlinie, fehler)
	}

	// **Ein Lesefehler hält an** und lässt nicht die Vorgabe gelten.
	if fehler := db.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := schranke.Handelsrichtlinie(gezeigt); fehler == nil {
		t.Error("ein geschlossener Bestand gab eine Richtlinie")
	}
}

func TestOhneBestandGibtEsKeineHandelsrichtlinie(t *testing.T) {
	t.Parallel()

	if _, fehler := (einstellungsschranke{a: NeueApp("")}).Handelsrichtlinie(t.Context()); fehler == nil {
		t.Error("ohne Bestand kam eine Richtlinie")
	}
}
