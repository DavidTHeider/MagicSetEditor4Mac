// **Eine Übernahme, die wirklich schreibt — an jeder Fläche ohne Fenster** —
// Werkstück A14 von Phase 5.3
// (#1453).
//
// # Warum es diesen Lauf bis A14 nicht gab
//
// **Weil keiner eine Laufzeit hatte.** Die Läufe über die KI-Flächen prüfen
// bisher **Abweisungen** — kein Schalter, kein Modell, keine Freigabe —, und
// ihr Kopf sagt es ausdrücklich: „was er nicht deckt: eine Übernahme, die
// wirklich schreibt. Sie brauchte eine laufende KI-Laufzeit."
//
// **Die gibt es seit `kilaufzeitstoff_test.go`.** Damit lässt sich der Satz
// aus dem Issue endlich messen statt umschreiben:
//
// > Lesen, Vorschläge, Assets, Apply und Providerstatus auf Wails/CLI/REST/MCP
// > fachlich gleich.
//
// # Was dieser Lauf deckt und was daneben steht
//
//   - **Apply** — hier, an drei Flächen, mit demselben geschriebenen Wert und
//     derselben Herkunft.
//   - **Lesen und Vorschläge** — die Abweisungen stehen in `kiauswahl_test.go`
//     und `kimodelle_test.go`, der ganze Weg in `agentenlauf_test.go`.
//   - **Providerstatus** — `kixai_test.go` und `kimodelle_test.go`
//     unterscheiden „nicht eingeschaltet" von „gibt es nicht".
//   - **Assets** stehen **nicht** hier, und das ist begründet und nicht
//     vergessen: ein Bildvorschlag lebt im Prozess und wird angesehen, bevor
//     ihn jemand übernimmt — Kommandozeile, MCP und Dienst bauen die Anwendung
//     je Aufruf ([0102](../../docs/entscheidungen/0102-wo-ein-erzeugtes-bild-liegt.md),
//     `docs/FAEHIGKEITEN.md`).
//
// # Die Gegenprobe
//
// **Der Dienst übernimmt nicht mehr** (`bezug.Uebernehmen` durch `false`
// ersetzt in `cmd/mse-rest/sammlung.go`): der Lauf wird rot, und zwar
// zweimal — die Karte trägt ihren alten Wert, und sie hat null KI-Schritte.
//
// **Der erste Anlauf traf die falsche Stelle**: die Marke `json:"uebernehmen"`
// umbenannt, und der Dienst antwortete **400** — er weist unbekannte Felder
// ab. Das ist ein guter Riegel und war eine schlechte Gegenprobe: rot war der
// Lauf an der Anfrage und nicht an der Übernahme (§10).
//
// # Warum jede Fläche eine eigene Karte bekommt
//
// **Weil sonst die zweite gegen einen Stand liefe, den die erste geändert
// hat.** Der Auftrag ist an die **Zielrevision** gebunden (A6); drei
// Übernahmen auf dieselbe Karte prüften ab der zweiten den Konflikt und nicht
// die Gleichheit.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

// **Dieselbe Übernahme über Kommandozeile, MCP und Dienst.**
func TestEineUebernahmeUeberJedeFlaeche(t *testing.T) {
	t.Parallel()

	umgebung, laufzeit := einAgentenbestand(t)
	setID, karten := dreiAbnahmekarten(t, umgebung)
	an := starteDienst(t, umgebung)

	mussKommandozeile(t, umgebung, "ki-vorschlag", karten[0],
		"ollama", probemodell, derAuftrag, "--uebernehmen", "--felder="+dasFeld)

	if _, istFehler := mcpRuf(t, umgebung, "schreiben.ki.vorschlag",
		map[string]any{
			"karteId": zahl(t, karten[1]), "anbieter": "ollama",
			"alias": probemodell, "auftrag": derAuftrag,
			"uebernehmen": true, "felder": []string{dasFeld},
		}); istFehler {
		t.Fatalf("die Übernahme über MCP scheiterte")
	}

	if status, antwort := an.ruf(t, http.MethodPost,
		"/karten/"+karten[2]+"/ki-vorschlag",
		`{"anbieter":"ollama","alias":"`+probemodell+`","auftrag":"`+derAuftrag+
			`","uebernehmen":true,"felder":["`+dasFeld+`"]}`); status != http.StatusOK {
		t.Fatalf("die Übernahme über den Dienst: %d — %s", status, antwort)
	}

	// **Drei Aufträge und nicht mehr** — eine Fläche, die zweimal fragte,
	// bezahlte bei einem Cloudanbieter doppelt.
	if gefragt := laufzeit.auftraege(); gefragt != 3 {
		t.Errorf("die Laufzeit sah %d Aufträge statt dreien", gefragt)
	}

	// **Und alle drei Karten tragen dasselbe.**
	stand := mussKommandozeile(t, umgebung, "karten", setID, "--json")
	for _, karteID := range karten {
		if wert := feldwertDerKarte(t, stand, karteID); wert != "Aus dem Modell" {
			t.Errorf("die Karte %s trägt %q", karteID, wert)
		}
	}

	// **Mit derselben Herkunft, und je Karte genau eine.**
	//
	// **Das ist die eigentliche Zusage.** Drei gleiche Werte könnten drei
	// verschiedene Wege genommen haben; die Herkunft `ki` sagt, dass es
	// dieselbe Fachlogik war (A11) — und dass keine Fläche nebenher noch ein
	// zweites Mal geschrieben hat.
	geschrieben := kischritteJeKarte(t, mussMCP(t, umgebung, "lesen.verlauf",
		map[string]any{"setID": zahl(t, setID)}))

	for _, karteID := range karten {
		if geschrieben[karteID] != 1 {
			t.Errorf("die Karte %s hat %d KI-Schritte statt einem",
				karteID, geschrieben[karteID])
		}
	}
}

// kischritteJeKarte zählt je Karte die Verlaufsschritte mit KI-Herkunft.
func kischritteJeKarte(t *testing.T, roh string) map[string]int {
	t.Helper()

	var antwort struct {
		Schritte []struct {
			KarteID  string `json:"karteID"`
			Feld     string `json:"feld"`
			Herkunft string `json:"herkunft"`
		} `json:"schritte"`
	}

	if fehler := json.Unmarshal([]byte(roh), &antwort); fehler != nil {
		t.Fatalf("den Verlauf lesen: %v — %s", fehler, roh)
	}

	gezaehlt := map[string]int{}

	for _, schritt := range antwort.Schritte {
		if schritt.Feld == dasFeld && schritt.Herkunft == "ki" {
			gezaehlt[schritt.KarteID]++
		}
	}

	return gezaehlt
}

// dreiAbnahmekarten fährt das Probeset ein und gibt Set und drei Karten.
func dreiAbnahmekarten(t *testing.T, umgebung []string) (setID string, karten []string) {
	t.Helper()

	setID = setkennungAus(t, mussKommandozeile(t, umgebung,
		"oeffnen", probeset(t), "--json"))

	zeilen := kartenzeilenAus(t, mussKommandozeile(t, umgebung,
		"karten", setID, "--json"))

	if len(zeilen) < 3 {
		t.Fatalf("das Probeset hat nur %d Karten", len(zeilen))
	}

	for _, zeile := range zeilen[:3] {
		karten = append(karten, zeile.ID)
	}

	return setID, karten
}

// feldwertDerKarte liest [dasFeld] einer bestimmten Karte aus `mse karten`.
func feldwertDerKarte(t *testing.T, roh, karteID string) string {
	t.Helper()

	var antwort struct {
		Karten []struct {
			ID    string            `json:"id"`
			Werte map[string]string `json:"werte"`
		} `json:"karten"`
	}

	if fehler := json.Unmarshal([]byte(roh), &antwort); fehler != nil {
		t.Fatalf("die Kartenliste lesen: %v — %s", fehler, roh)
	}

	for _, karte := range antwort.Karten {
		if karte.ID == karteID {
			return karte.Werte[dasFeld]
		}
	}

	t.Fatalf("die Karte %s steht nicht in der Liste", karteID)

	return ""
}
