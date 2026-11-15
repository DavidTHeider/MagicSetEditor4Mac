package storage_test

// **Die Einstellungen, die dem Menschen gehören** — Werkstück 8 von
// Phase 4.12
// (#1351).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Paar kommt so zurück, wie es hineinging — und ein zweites ersetzt es.**
func TestEineNutzereinstellungKommtZurueck(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := probekontext(t)

	if fehler := ablage.NutzereinstellungSetzen(
		ctx, 1, domain.NutzerBewegung, domain.BewegungReduziert); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	gelesen, lesefehler := ablage.Nutzereinstellungen(ctx, 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gelesen[domain.NutzerBewegung] != domain.BewegungReduziert {
		t.Errorf("gelesen wurde %v", gelesen)
	}

	// **Dasselbe Paar noch einmal ersetzt** — und legt keine zweite Zeile an.
	if fehler := ablage.NutzereinstellungSetzen(
		ctx, 1, domain.NutzerBewegung, domain.NachSystem); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	danach := mussEinstellungen(t, ablage, 1)
	if len(danach) != 1 || danach[domain.NutzerBewegung] != domain.NachSystem {
		t.Errorf("gelesen wurde %v", danach)
	}
}

// **Ein Nutzer ohne Einstellungen bekommt eine leere Abbildung, kein `nil`**
// (§3).
func TestEinNutzerOhneEinstellungenBekommtEineAbbildung(t *testing.T) {
	gelesen, fehler := storage.NeueNutzerablage(frischeDatenbank(t)).
		Nutzereinstellungen(probekontext(t), 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if gelesen == nil {
		t.Error("die Abbildung ist nil")
	}
}

// **Ein unbekannter Schlüssel kommt gar nicht erst in den Bestand.**
//
// **Die Prüfung steht in `domain` und wird hier gerufen** (§2): der
// MCP-Server schreibt auf dieselbe Datei, und was er hineinschreibt, muss
// dieselben Regeln bestehen.
func TestEineUnbekannteNutzereinstellungKommtNichtInDenBestand(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := probekontext(t)

	if fehler := ablage.NutzereinstellungSetzen(
		ctx, 1, "gibtsnicht", "egal"); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("angenommen wurde: %v", fehler)
	}

	if fehler := ablage.NutzereinstellungSetzen(
		ctx, 1, domain.NutzerBewegung, "viel"); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("angenommen wurde: %v", fehler)
	}

	gelesen := mussEinstellungen(t, ablage, 1)
	if len(gelesen) != 0 {
		t.Errorf("es steht etwas da: %v", gelesen)
	}
}

// **Wer geht, nimmt seine Einstellungen mit.**
func TestWerGehtNimmtSeineEinstellungenMit(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := probekontext(t)

	zweiter, fehler := ablage.NutzerAnlegen(ctx, domain.Nutzer{Name: "Zweite"})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if err := ablage.NutzereinstellungSetzen(
		ctx, zweiter.ID, domain.NutzerKontrast, domain.KontrastHoch); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if err := ablage.NutzerLoeschen(ctx, zweiter.ID, "Zweite"); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	gelesen := mussEinstellungen(t, ablage, zweiter.ID)
	if len(gelesen) != 0 {
		t.Errorf("die Einstellungen blieben: %v", gelesen)
	}
}

// mussEinstellungen liest sie und bricht ab, wenn es nicht glückt.
func mussEinstellungen(
	t *testing.T, ablage *storage.Nutzerablage, nutzerID int64,
) map[string]string {
	t.Helper()

	gelesen, fehler := ablage.Nutzereinstellungen(probekontext(t), nutzerID)
	if fehler != nil {
		t.Fatalf("die Einstellungen lesen: %v", fehler)
	}

	return gelesen
}
