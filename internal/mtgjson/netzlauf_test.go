//go:build netzlauf || dienstdrift

// **Der Lauf gegen die echte Datei** — Werkstück A14 von Phase 5.3
// (#1453).
//
// # Was dieses Paket liest, und warum das ein Dienst ist
//
// **`internal/mtgjson` liest ein Dateiformat und ruft niemanden.** Genau
// deshalb stand hier bisher kein Lauf: es gibt keinen Klienten, der driften
// könnte. **Aber das Format hat einen Herausgeber**, und der schreibt es
// weiter: `https://mtgjson.com/api/v5/` führt jede Setdatei, und die Zuordnung
// in diesem Paket ist gegen **selbst geschriebene** Beispiele geprüft.
//
// **Was daran fehlt**, ist die Frage, ob eine Datei von dort heute noch so
// aussieht, wie dieses Programm sie liest — und die beantwortet nur ein Ruf.
//
// # Warum eine Setdatei und nicht `AllPrintings`
//
// **`AllPrintings.json` führt jedes Set, das es gibt.** Eine einzelne Datei
// sagt über die **Gestalt** dasselbe und kostet ein Tausendstel; wer mehr
// Sicherheit über die Breite will, holt mehrere Sets, nicht eine große Datei
// (§3, §16 Regel 4).
//
// **`LEA` ist das erste Set von 1993**: die Karten darin ändern sich nicht
// mehr, und wenn es verschwände, wäre das eine Nachricht über den Dienst.
//
// # Warum er nicht in der Pipeline läuft
//
// Er lädt von einem fremden Rechner.
//
//	go test -tags netzlauf ./internal/mtgjson/

package mtgjson_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// diePruefdatei ist die Setdatei, gegen die gelaufen wird.
const diePruefdatei = "https://mtgjson.com/api/v5/LEA.json"

// **Eine echte Setdatei geht durch die Zuordnung — mit Karten und ohne
// Schemabefund.**
func TestDieEchteSetdateiWirdGelesen(t *testing.T) {
	zuordnung, da := mtgjson.ZuordnungFuer("magic")
	if !da {
		t.Fatal("die Zuordnung für `magic` fehlt — der Lauf greift daneben")
	}

	inhalt := holeDieSetdatei(t)

	// **Die Schemaprüfung zuerst** — sie ist die Frage nach der Drift.
	if befunde := mtgjson.Pruefe(inhalt, "LEA.json"); len(befunde) > 0 {
		for _, befund := range befunde {
			t.Errorf("Schemabefund: %s — %s", befund.Ort, befund.Meldung)
		}
	}

	// **Und die Fassung** — MTGJSON zählt seine Hauptversion mit.
	angabe, befunde := mtgjson.Fassung(inhalt, "LEA.json")
	for _, befund := range befunde {
		t.Errorf("Fassungsbefund: %s — %s", befund.Ort, befund.Meldung)
	}

	t.Logf("Fassung %s vom %s", angabe.Version, angabe.Datum)

	quelle, fehler := mtgjson.LiesInhalt(inhalt, "LEA.json", zuordnung)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(quelle.Sets) != 1 {
		t.Fatalf("%d Sets statt einem", len(quelle.Sets))
	}

	karten := quelle.Sets[0].Karten
	if len(karten) < 100 {
		t.Fatalf("%d Karten — LEA hat über 290", len(karten))
	}

	// **Jede Karte trägt Felder** — das ist der Vertrag der Zuordnung: sie
	// bildet MTGJSONs Schlüssel auf Kartenfelder ab, und eine Karte ohne
	// einen einzigen bekannten Wert hieße, dass keiner mehr so heißt.
	ohneWerte := 0

	for _, karte := range karten {
		if len(karte.Werte) == 0 {
			ohneWerte++
		}
	}

	if ohneWerte > 0 {
		t.Errorf("%d von %d Karten trugen kein bekanntes Feld — hat MTGJSON "+
			"seine Schlüssel geändert?", ohneWerte, len(karten))
	}

	// **Und die Befunde stehen da** — sie sind kein Fehler, aber sie sagen,
	// was dieses Programm an der echten Datei **nicht** kennt.
	for _, befund := range quelle.Befunde {
		t.Logf("Befund: %s — %s", befund.Ort, befund.Meldung)
	}

	t.Logf("gelesen: %d Karten aus %q, %d Befunde",
		len(karten), quelle.Sets[0].Name, len(quelle.Befunde))
}

// holeDieSetdatei lädt sie.
func holeDieSetdatei(t *testing.T) []byte {
	t.Helper()

	anfrage, fehler := http.NewRequestWithContext(t.Context(),
		http.MethodGet, diePruefdatei, http.NoBody)
	if fehler != nil {
		t.Fatalf("die Anfrage bauen: %v", fehler)
	}

	// **Die Anwendung nennt sich** (§16, Satz 2).
	anfrage.Header.Set("User-Agent", "MagicSetEditor4Mac/netzlauf")

	antwort, fehler := http.DefaultClient.Do(anfrage) //nolint:bodyclose // unten
	if fehler != nil {
		t.Fatalf("die Setdatei holen: %v", fehler)
	}

	defer func() { _ = antwort.Body.Close() }() //nolint:errcheck // gelesen ist gelesen

	if antwort.StatusCode != http.StatusOK {
		t.Fatalf("die Setdatei gab %d", antwort.StatusCode)
	}

	inhalt, fehler := io.ReadAll(antwort.Body)
	if fehler != nil {
		t.Fatalf("die Setdatei lesen: %v", fehler)
	}

	return inhalt
}
