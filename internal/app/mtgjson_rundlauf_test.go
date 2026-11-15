package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// **Der Rundlauf über den echten Bestand: MTG JSON → SQLite → MTG JSON.**
//
// # Warum dieser Test hier steht
//
// Der Zweitblick über Phase 9 hat die Lücke benannt: die Rundläufe der
// Fachbibliothek gehen **an SQLite vorbei**. Sie prüfen `Schreibe(Lies(x))` im
// Arbeitsspeicher — und genau dazwischen lag der Fehler. Der Import legte den
// Setcode als Setangabe ab, die Ausfuhr las die Angaben nicht, und der
// Schreiber griff auf den Setnamen zurück: aus `PRB` wurde `Probeset`.
//
// Das ist dieselbe Klasse, vor der AGENTS.md §3 warnt: *eine Vorrichtung, die
// großzügiger ist als der Betrieb*. Hier geht der Weg durch beide Schichten.
func TestDerRundlaufUeberDenBestand(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	quelle, err := filepath.Abs(filepath.Join("..", "..", "testdaten", "mtgjson", "probe.json"))
	if err != nil {
		t.Fatalf("Pfad: %v", err)
	}

	geoeffnet := anwendung.MTGJSONOeffnen(t.Context(), quelle, "magic")
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "wieder.json")

	geschrieben := anwendung.MTGJSONSchreiben(t.Context(), geoeffnet.SetID, "magic", ziel)
	if geschrieben.Fehler != nil {
		t.Fatalf("schreiben: %+v", geschrieben.Fehler)
	}

	inhalt, err := os.ReadFile(ziel) // #nosec G304 -- der Pfad stammt aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// **Der Setcode überlebt** — das war der Befund.
	if !strings.Contains(string(inhalt), `"code": "PRB"`) {
		t.Errorf("der Setcode ist unterwegs verlorengegangen:\n%s", inhalt)
	}

	// **Und die Ausgabe ist gültig**, geprüft mit demselben Prüfer, den auch
	// eine bezogene Datei besteht.
	if befunde := mtgjson.Pruefe(inhalt, ziel); len(befunde) > 0 {
		t.Errorf("die Ausgabe ist nicht gültig: %v", befunde)
	}

	// Und die Kartenwerte sind dieselben wie in der Quelle.
	vorher, err := mtgjson.Lies(quelle, mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("Quelle lesen: %v", err)
	}

	nachher, err := mtgjson.LiesInhalt(inhalt, ziel, mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("Ausgabe lesen: %v", err)
	}

	vergleicheKarten(t, vorher.Sets[0].Karten, nachher.Sets[0].Karten)
}

// vergleicheKarten hält die Werte zweier Kartenlisten gegeneinander.
func vergleicheKarten(t *testing.T, vorher, nachher []*mtgjson.Karte) {
	t.Helper()

	if len(nachher) != len(vorher) {
		t.Fatalf("%d Karten, erwartet %d", len(nachher), len(vorher))
	}

	for nummer, karte := range vorher {
		alt := werteVon(karte)
		neu := werteVon(nachher[nummer])

		for feld, wert := range alt {
			if neu[feld] != wert {
				t.Errorf("Karte %d, Feld %q: %q, erwartet %q", nummer+1, feld, neu[feld], wert)
			}
		}
	}
}

// werteVon legt die Werte einer Karte als Abbildung aus.
func werteVon(karte *mtgjson.Karte) map[string]string {
	werte := map[string]string{}
	for _, wert := range append(append([]domain.Value{}, karte.Werte...), karte.Rest...) {
		werte[wert.Field] = wert.Text
	}

	return werte
}
