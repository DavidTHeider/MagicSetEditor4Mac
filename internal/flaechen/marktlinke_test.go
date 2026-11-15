package flaechen_test

// **Die Wege nach draußen über jede Fläche** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Was hier gemessen wird
//
//   - **Dass alle drei kopflosen Flächen dieselben Adressen nennen** (§16) —
//     nicht dieselbe Gestalt, sondern denselben Satz aus Anbieter, Art und
//     Adresse. Eine Fläche, die einen Anbieter ausließe oder eine andere Art
//     behauptete, fiele hier auf.
//   - **Dass ein Druck, den es nicht gibt, kein Fehler ist.** Er hat keine
//     Kennung, und deshalb bleibt es beim Namen — genau der Fall einer
//     eigenen Karte.
//   - **Dass die Kommandozeile eine fremde Adresse nicht öffnet.** Das ist
//     die Ablehnung, und ohne sie wäre nur der Erfolgsweg geprüft (§10).
//
// # Was er **nicht** misst, und warum nicht
//
// **Das geglückte Öffnen.** Es macht einen Browser auf diesem Rechner auf;
// ein Lauf, der das täte, hinterließe ein Fenster und führe nach draußen
// (§3). Gemessen ist es über `internal/service` mit einem eingesetzten
// Öffner — hier zählt, **dass** eine fremde Adresse vorher scheitert.
//
// **Und das Fehlen an zwei Flächen ist selbst die Messung**: weder der Dienst
// noch MCP kennen einen Weg zum Öffnen, und das ist eine Entscheidung
// ([0106](../../docs/entscheidungen/0106-drei-vertrauensgrenzen-und-nur-eine-wird-gebaut.md)),
// keine Lücke. Der Lauf hält sie fest.

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"testing"
)

// derLinkweg ist der Endpunkt, wie er im Katalog steht.
const derLinkweg = "GET /drucke/{printing}/links"

// derFremdeDruck ist eine Kennung, die im Bestand nicht steht.
//
// **Und das ist hier der Normalfall** — eine eigene Karte hat keinen
// kanonischen Druck.
const derFremdeDruck = "gibtesnicht"

// dieProbekarte trägt die Suche.
const dieProbekarte = "Waldläufer der Probe"

// marktweg ist ein Weg, wie ihn jede Fläche schreibt.
type marktweg struct {
	Anbieter string `json:"anbieter"`
	Art      string `json:"art"`
	Herkunft string `json:"herkunft"`
	Adresse  string `json:"adresse"`
}

// marktwegeAus liest die Antwort einer Fläche und ordnet sie.
//
// **Geordnet, weil verglichen wird und nicht aufgezählt**: welche Reihenfolge
// eine Fläche wählt, ist ihre Sache — welche Wege sie nennt, nicht.
func marktwegeAus(t *testing.T, roh string) []marktweg {
	t.Helper()

	var wege []marktweg
	if fehler := json.Unmarshal([]byte(roh), &wege); fehler != nil {
		t.Fatalf("die Antwort deuten: %v — %s", fehler, roh)
	}

	sort.Slice(wege, func(eins, zwei int) bool {
		return wege[eins].Anbieter < wege[zwei].Anbieter
	})

	return wege
}

func TestDieMarktwegeStehenUeberallGleich(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	ausCLI := marktwegeAus(t, mitEingabe(t, umgebung, "",
		"--json", "markt-links", derFremdeDruck, dieProbekarte))

	text, istFehler := mcpRuf(t, umgebung, "lesen.marktlinks",
		map[string]any{"printing": derFremdeDruck, "name": dieProbekarte})
	if istFehler {
		t.Fatalf("MCP klagte: %s", text)
	}

	ausMCP := marktwegeAus(t, text)

	status, rumpf := an.ruf(t, http.MethodGet,
		strings.ReplaceAll(pfadAus(derLinkweg), "{printing}", derFremdeDruck)+
			"?name="+strings.ReplaceAll(dieProbekarte, " ", "%20"), "")
	if status != http.StatusOK {
		t.Fatalf("der Dienst gab %d: %s", status, rumpf)
	}

	ausREST := marktwegeAus(t, rumpf)

	// **Sechs Anbieter, und zwar überall dieselben sechs.**
	if len(ausCLI) != 6 {
		t.Fatalf("die Kommandozeile nennt %d Wege: %+v", len(ausCLI), ausCLI)
	}

	for name, andere := range map[string][]marktweg{"MCP": ausMCP, "REST": ausREST} {
		if len(andere) != len(ausCLI) {
			t.Fatalf("%s nennt %d Wege, die Kommandozeile %d",
				name, len(andere), len(ausCLI))
		}

		for stelle, weg := range andere {
			if weg != ausCLI[stelle] {
				t.Errorf("%s nennt %+v, die Kommandozeile %+v",
					name, weg, ausCLI[stelle])
			}
		}
	}

	// **Und keiner davon ist ein Produktlink.** Es gibt keine Kennung; wer
	// hier einen behauptete, führte auf eine fremde Karte.
	for _, weg := range ausCLI {
		if weg.Art == "produkt" {
			t.Errorf("%s verspricht ein Produkt ohne Kennung: %s",
				weg.Anbieter, weg.Adresse)
		}

		if !strings.HasPrefix(weg.Adresse, "https://") {
			t.Errorf("%s geht nicht über https: %s", weg.Anbieter, weg.Adresse)
		}
	}
}

// **Eine fremde Adresse wird nicht geöffnet.**
//
// **Der Wirt entscheidet und nicht das Schema**: `https://` allein ist keine
// Erlaubnis, und ein Anbietername vor einer fremden Adresse erst recht nicht.
func TestDieKommandozeileOeffnetKeineFremdeAdresse(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	for name, adresse := range map[string]string{
		"fremder Wirt":        "https://beispiel.invalid/karte",
		"Wirt als Vorsatz":    "https://www.cardmarket.com.beispiel.invalid/x",
		"kein https":          "http://www.cardmarket.com/en/Magic",
		"mit Benutzerangabe":  "https://wer@www.cardmarket.com/en/Magic",
		"mit eigenem Port":    "https://www.cardmarket.com:8443/en/Magic",
		"Geheimnis im Anhang": "https://www.cardmarket.com/en/Magic?token=abc",
	} {
		t.Run(name, func(t *testing.T) {
			ausgabe, klage, code := kommandozeile(t, umgebung,
				"--json", "markt-oeffnen", "cardmarket", adresse)
			if code == 0 {
				t.Fatalf("die Adresse kam durch: %s", ausgabe)
			}

			if strings.Contains(ausgabe, "Geöffnet") {
				t.Errorf("es wurde geöffnet: %s", ausgabe)
			}

			if strings.TrimSpace(klage+ausgabe) == "" {
				t.Error("die Ablehnung sagt nichts")
			}
		})
	}
}

// **Und ein unbekannter Anbieter ebenso.**
//
// **Die Positivliste ist eine Liste von Namen und keine Gewohnheit**: wer
// einen siebten Anbieter nennt, bekommt keinen Wirt, gegen den geprüft würde.
func TestEinUnbekannterAnbieterHatKeinenWirt(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	ausgabe, _, code := kommandozeile(t, umgebung, "--json", "markt-oeffnen",
		"kartenkiosk", "https://www.cardmarket.com/en/Magic")
	if code == 0 {
		t.Fatalf("der fremde Anbieter kam durch: %s", ausgabe)
	}
}
