// Rückgängig, durch die ganze Anwendung.
//
// **Der Weg, den ein Mensch nimmt**: Set öffnen, ein Feld ändern, ⌘Z. Die
// Fachtests der Schichten laufen gegen Doubles und sehen diese Naht nicht — und
// genau dort lag der Befund, der dieses Werkstück aufgehalten hat: `FeldSetzen`
// ging am Verlauf vorbei, und ⌘Z hätte nichts zurückgenommen, was jemand
// getippt hat. Der MCP-Server schrieb ihn, das Fenster nicht.
package app_test

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// geoeffneteAnwendung startet die Anwendung und öffnet `probe.mse-set`.
func geoeffneteAnwendung(t *testing.T) (anwendung *app.App, setID, karteID string) {
	t.Helper()

	anwendung = app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	quelle, err := filepath.Abs(probeset)
	if err != nil {
		t.Fatalf("Pfad: %v", err)
	}

	geoeffnet := anwendung.SetOeffnen(t.Context(), quelle)
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	return anwendung, geoeffnet.SetID, karten.Karten[0].ID
}

// namenswert liest das Feld `name` der ersten Karte aus dem Bestand.
//
// **Ohne Feldparameter**, weil es nur einen Aufrufwert gäbe — `unparam` meldet
// das zu Recht: ein Parameter, der immer dasselbe trägt, sieht nach Allgemeinheit
// aus und ist keine.
func namenswert(t *testing.T, anwendung *app.App, setID string) string {
	t.Helper()

	karten := anwendung.Karten(t.Context(), setID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	for _, wert := range karten.Karten[0].Werte {
		if wert.Feld == "name" {
			return wert.Wert
		}
	}

	return ""
}

// **Was das Fenster ändert, nimmt ⌘Z zurück.**
func TestEineAenderungAusDemFensterLaesstSichZuruecknehmen(t *testing.T) {
	anwendung, setID, karteID := geoeffneteAnwendung(t)

	if vorher := namenswert(t, anwendung, setID); vorher != "Feuerball" {
		t.Fatalf("Ausgangswert %q", vorher)
	}

	if antwort := anwendung.FeldSetzen(
		probekontext(t), setID, karteID, "name", "Eisball"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	if jetzt := namenswert(t, anwendung, setID); jetzt != "Eisball" {
		t.Fatalf("nach dem Setzen %q", jetzt)
	}

	genommen := anwendung.Zuruecknehmen(t.Context(), setID, "")
	if genommen.Fehler != nil {
		t.Fatalf("zurücknehmen: %+v", genommen.Fehler)
	}

	if !genommen.Bewegt {
		t.Fatal("nichts zurückgenommen — das Fenster schreibt seinen Verlauf nicht")
	}

	// **Die betroffene Karte wird genannt**, damit die Oberfläche nur sie
	// nachlädt (§7).
	if genommen.KarteID != karteID || genommen.Feld != "name" {
		t.Errorf("%+v, erwartet war Karte %s / Feld name", genommen, karteID)
	}

	// **Und der Umfang** — eine Einzeländerung ist ein Schritt über einer Karte.
	if genommen.Umfang != 1 {
		t.Errorf("Umfang %d, erwartet eins", genommen.Umfang)
	}

	// **Und der Urheber**, denn zwei Prozesse schreiben (#52).
	if genommen.Urheber != string(domain.UrheberFenster) {
		t.Errorf("Urheber %q, erwartet %q", genommen.Urheber, domain.UrheberFenster)
	}

	if wieder := namenswert(t, anwendung, setID); wieder != "Feuerball" {
		t.Errorf("nach der Rücknahme %q, erwartet Feuerball", wieder)
	}

	// … und ⇧⌘Z bringt sie zurück.
	gebracht := anwendung.Wiederherstellen(t.Context(), setID, "")
	if gebracht.Fehler != nil || !gebracht.Bewegt {
		t.Fatalf("wiederherstellen: %+v / %v", gebracht.Fehler, gebracht.Bewegt)
	}

	if jetzt := namenswert(t, anwendung, setID); jetzt != "Eisball" {
		t.Errorf("nach dem Wiederherstellen %q, erwartet Eisball", jetzt)
	}
}

// **Ein leerer Verlauf ist kein Fehler** — ⌘Z am Anfang ist eine gewöhnliche
// Lage.
func TestZuruecknehmenOhneVerlaufMeldetNichts(t *testing.T) {
	anwendung, setID, _ := geoeffneteAnwendung(t)

	genommen := anwendung.Zuruecknehmen(t.Context(), setID, "")
	if genommen.Fehler != nil {
		t.Fatalf("zurücknehmen: %+v", genommen.Fehler)
	}

	if genommen.Bewegt {
		t.Error("etwas zurückgenommen, obwohl nichts geändert wurde")
	}
}

// **Eine Massenänderung kommt in **einem** ⌘Z zurück.**
//
// Der Weg geht durch die ganze Anwendung: drei Karten in einem Zug ändern, ⌘Z,
// nachlesen. Ohne den Verlauf der Massenänderung nähme ⌘Z hier die **vorige**
// Änderung zurück, während die drei Karten stehen blieben — und das sähe aus,
// als hätte es gewirkt (#52).
func TestEineMassenaenderungKommtInEinemZugZurueck(t *testing.T) {
	anwendung, setID, _ := geoeffneteAnwendung(t)

	karten := anwendung.Karten(t.Context(), setID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	kennungen := make([]string, 0, len(karten.Karten))
	for _, karte := range karten.Karten {
		kennungen = append(kennungen, karte.ID)
	}

	// **Erst die Vorschau, dann das Anwenden mit ihrer Erwartung** — der
	// einzige Weg des Fensters seit U6 (#1417).
	vorschau := anwendung.FelderVorschau(probekontext(t), setID, kennungen, "typ", "Artefakt")
	if vorschau.Fehler != nil {
		t.Fatalf("Vorschau: %+v", vorschau.Fehler)
	}

	erwartet := make([]app.FeldvorwertDTO, 0, len(vorschau.Zeilen))
	for _, zeile := range vorschau.Zeilen {
		erwartet = append(erwartet, app.FeldvorwertDTO{
			KarteID: zeile.KarteID, Wert: zeile.Alt, GabEs: zeile.GabAlt,
		})
	}

	if antwort := anwendung.FelderSetzenMitErwartung(
		probekontext(t), setID, erwartet, "typ", "Artefakt"); antwort.Fehler != nil {
		t.Fatalf("Massenänderung: %+v", antwort.Fehler)
	}

	genommen := anwendung.Zuruecknehmen(t.Context(), setID, "")
	if genommen.Fehler != nil || !genommen.Bewegt {
		t.Fatalf("zurücknehmen: %+v / %v", genommen.Fehler, genommen.Bewegt)
	}

	if genommen.Umfang != len(kennungen) {
		t.Errorf("Umfang %d, erwartet %d", genommen.Umfang, len(kennungen))
	}

	// **Alle Karten sind zurück** — und das Feld ist entfernt, nicht geleert:
	// es gab es vorher nicht.
	nachher := anwendung.Karten(t.Context(), setID)
	if nachher.Fehler != nil {
		t.Fatalf("Karten: %+v", nachher.Fehler)
	}

	for _, karte := range nachher.Karten {
		for _, wert := range karte.Werte {
			if wert.Feld == "typ" {
				t.Errorf("Karte %s trägt noch typ=%q", karte.ID, wert.Wert)
			}
		}
	}
}
