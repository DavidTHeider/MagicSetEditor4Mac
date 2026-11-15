package app_test

// **Die zwei Wege zum Katalog ohne Schalterprobe** — Werkstück A1 von
// Phase 5.3 (#1440): der kopflose Weg über einen Bestand und die
// Fensterbindung. Aus `dienstkatalog_test.go` herausgelöst, als die Datei
// über die Zeilengrenze aus §2 ging.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der kopflose Weg liest den Bestand** — dieselbe Wahl wie die Suche, mit
// Riegel; ohne gesetzten Bund liest er den Schlüsselbund nur für Dienste, die
// an sind (§3: kein Griff in den echten Schlüsselbund im Test).
func TestDerKopfloseKatalogLiestDieEinstellungenDesBestands(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischeAppdatenbank(t)

	katalog, err := app.DienstkatalognahtUeber(db).Katalog(probekontext(t))
	if err != nil {
		t.Fatalf("der Katalog: %v", err)
	}

	if suche := eintrag(t, katalog, "scryfall", "suche"); suche.Status != domain.StatusBereit {
		t.Errorf("scryfall/suche: %+v — erwartet bereit, Scryfall ist die Vorgabe", suche)
	}

	if scrydex := eintrag(t, katalog, "scrydex", "suche"); scrydex.Status != domain.StatusAus {
		t.Errorf("scrydex/suche: %+v — erwartet aus, die drei neuen fangen aus an", scrydex)
	}

	mitBund := app.DienstkatalognahtUeber(db).MitBund(bundprobe{})
	if _, err := mitBund.Katalog(probekontext(t)); err != nil {
		t.Errorf("mit eigenem Bund: %v", err)
	}
}

// **Die Fensterbindung reicht mit dem Kontext der Anwendung durch.**
func TestDieFensterbindungLiefertDenKatalog(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselbund(neuerBundspeicher())

	antwort := app.NeueBindungen(anwendung).Dienstkatalog(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Eintraege) != len(domain.Dienstvertraege()) {
		t.Errorf("%d Einträge, erwartet %d", len(antwort.Eintraege), len(domain.Dienstvertraege()))
	}
}
