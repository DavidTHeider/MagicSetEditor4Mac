package flaechen_test

// **Der Tauschabgleich über jede Fläche** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Was hier gemessen wird
//
// **Die Abnahmepunkte des Issues an der tatsächlich serialisierten Antwort:**
//
//  1. Dieselbe Eingabe ergibt dieselben Kandidaten — an beiden Flächen und
//     bei beiden Läufen.
//  2. **Es wird nichts erfunden**: das Quellbeispiel kann keine zwei Solitude
//     herbeiführen, solange keine dastehen.
//  3. Ein leeres Angebot wird **nicht** als bester Tausch empfohlen.
//  4. Der Bestand ändert sich dabei nicht.
//
// # Warum über beide Flächen und nicht über eine
//
// **Weil §16 dieselbe Antwort an jeder verlangt.** Ein Abgleich, der an der
// Kommandozeile etwas anderes sagt als im Dienst, ist zweimal gebaut und
// einmal geprüft.

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// derFremdeStand ist, was ein Gegenüber freigegeben hat.
//
// **Derselbe Aufbau, den `kontakt-sicht --json` hergibt** — zwei Listen mit
// je sechs Feldern. Was hier nicht steht, geht auch nicht hinein.
const derFremdeStand = `{
  "angebot":  [{"printing":"b2c3d4e5-0000-4000-8000-00000000000b","menge":4,
                "zustand":"played","sprache":"de","finish":"etched",
                "bezeichnung":"Force of Will"}],
  "wuensche": [{"printing":"a1b2c3d4-0000-4000-8000-00000000000a","menge":2,
                "zustand":"played","sprache":"de","finish":"etched",
                "bezeichnung":"Lightning Bolt"}]
}`

func TestDerTauschabgleichUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	const (
		bolt  = "a1b2c3d4-0000-4000-8000-00000000000a"
		force = "b2c3d4e5-0000-4000-8000-00000000000b"
	)

	// **Drei Bolts im Bestand, einer davon bleibt** — zwei sind überzählig.
	mussKommandozeile(t, umgebung, "exemplar", bolt, "3", "played", "de",
		"etched", "2026-08-01", "4.17", "EUR", "aus dem Tausch",
		"Lightning Bolt", "lea", "161", "Kiste 4")

	// **Und ein Wunsch nach Force of Will** — zwei Stück, dringend.
	//
	// **Rang und Art stehen an den letzten zwei Stellen** (T4): sie gelten
	// nur an einem Wunsch, und die Kommandozeile kann beides sagen — sonst
	// führte sie nicht dieselbe Sammlung wie das Fenster (§16).
	mussKommandozeile(t, umgebung, "exemplar", force, "2", "played", "de",
		"etched", "", "", "", "", "Force of Will", "all", "28", "",
		"wunsch", "hoch", "genau")

	pfad := filepath.Join(t.TempDir(), "stand.json")
	if fehler := os.WriteFile(pfad, []byte(derFremdeStand), 0o600); fehler != nil {
		t.Fatalf("den Stand ablegen: %v", fehler)
	}

	ausgabe, _, code := kommandozeile(t, umgebung, "--json",
		"tausch-abgleich", "bgegenueber", pfad, "1")
	if code != 0 {
		t.Fatalf("die Kommandozeile gab %d: %s", code, ausgabe)
	}

	vonHand := abgleichAus(t, ausgabe)

	// **Zwei Bolts gehen hin, zwei Force of Will kommen her.**
	if len(vonHand.Gebe) != 1 || vonHand.Gebe[0].Menge != 2 {
		t.Errorf("gebe: %+v", vonHand.Gebe)
	}

	if len(vonHand.Bekomme) != 1 || vonHand.Bekomme[0].Menge != 2 {
		t.Errorf("bekomme: %+v", vonHand.Bekomme)
	}

	if !vonHand.HatVorschlag {
		t.Error("es kam kein Vorschlag, obwohl beide Seiten etwas haben")
	}

	// **Nichts wird erfunden**: was im Stand nicht steht, steht auch nicht im
	// Ergebnis — das ist der vierte Abnahmepunkt des Issues.
	for _, zeile := range vonHand.Bekomme {
		if zeile.Angebot != "Force of Will" {
			t.Errorf("im Ergebnis steht %q", zeile.Angebot)
		}
	}

	// **Derselbe Abgleich im Dienst** — und dieselbe Antwort.
	status, rumpf := an.ruf(t, http.MethodPost,
		"/peer/kontakte/bgegenueber/abgleich?behalten=1", derFremdeStand)
	if status != http.StatusOK {
		t.Fatalf("der Dienst gab %d: %s", status, rumpf)
	}

	ueberDenDienst := abgleichAus(t, rumpf)

	if len(ueberDenDienst.Gebe) != len(vonHand.Gebe) ||
		len(ueberDenDienst.Bekomme) != len(vonHand.Bekomme) ||
		ueberDenDienst.HatVorschlag != vonHand.HatVorschlag {
		t.Errorf("die Flächen sagen Verschiedenes: %+v gegen %+v",
			ueberDenDienst, vonHand)
	}

	// **Der Bestand hat sich dabei nicht geändert** — ein Abgleich rechnet.
	danach, _, code := kommandozeile(t, umgebung, "--json", "sammlung")
	if code != 0 {
		t.Fatalf("die Sammlung lesen gab %d: %s", code, danach)
	}

	if anzahl := sammlungszeilen(t, danach); anzahl != 2 {
		t.Errorf("im Bestand stehen %d Zeilen", anzahl)
	}
}

// **Ein leeres Angebot ist kein bester Tausch** — der leere Tausch ist keiner.
func TestEinLeeresAngebotIstKeinVorschlag(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	const leer = `{"angebot":[],"wuensche":[]}`

	status, rumpf := an.ruf(t, http.MethodPost,
		"/peer/kontakte/bgegenueber/abgleich", leer)
	if status != http.StatusOK {
		t.Fatalf("der Dienst gab %d: %s", status, rumpf)
	}

	if abgleichAus(t, rumpf).HatVorschlag {
		t.Error("der leere Tausch wurde empfohlen")
	}
}

// abgleichsicht ist, was von einem Abgleich zurückkommt.
//
// **Eine eigene Gestalt im Lauf** — sie hält die Antwort gegen das, was das
// Issue verlangt, und nicht gegen den Verbund, der sie erzeugt hat.
type abgleichsicht struct {
	Gebe []struct {
		Angebot string `json:"angebot"`
		Menge   int    `json:"menge"`
	} `json:"gebe"`

	Bekomme []struct {
		Angebot string `json:"angebot"`
		Menge   int    `json:"menge"`
	} `json:"bekomme"`

	HatVorschlag bool `json:"hatVorschlag"`
}

// abgleichAus liest die Antwort.
func abgleichAus(t *testing.T, roh string) abgleichsicht {
	t.Helper()

	var sicht abgleichsicht
	if fehler := json.Unmarshal([]byte(roh), &sicht); fehler != nil {
		t.Fatalf("die Antwort deuten: %v — %s", fehler, roh)
	}

	return sicht
}

// sammlungszeilen zählt die Exemplare einer Sammlungsantwort.
func sammlungszeilen(t *testing.T, roh string) int {
	t.Helper()

	var antwort struct {
		Exemplare []json.RawMessage `json:"exemplare"`
	}

	if fehler := json.Unmarshal([]byte(roh), &antwort); fehler != nil {
		t.Fatalf("die Sammlung deuten: %v — %s", fehler, roh)
	}

	return len(antwort.Exemplare)
}

// **Rang und Art stehen an jeder Fläche** — Werkstück `T4` von Phase 5.6.
//
// # Warum das hier gemessen wird und nicht nur an der Naht
//
// **Weil jede Fläche ihre eigene Gestalt hat.** `cmd/mse`, `internal/rest` und
// `internal/mcp` führen je einen eigenen Typ für ein Exemplar — genau
// deshalb, weil die Wails-DTO innen ist und sie außen sind (0033, Regel 5).
// Ein Feld, das nur in der DTO steht, erreicht dann **keine** von ihnen, und
// die Lücke sieht an der Naht wie Erfolg aus.
func TestRangUndArtStehenAnJederFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	const force = "b2c3d4e5-0000-4000-8000-00000000000b"

	mussKommandozeile(t, umgebung, "exemplar", force, "2", "played", "de",
		"etched", "", "", "", "", "Force of Will", "all", "28", "",
		"wunsch", "hoch", "jederdruck")

	ausgabe, _, code := kommandozeile(t, umgebung, "--json", "sammlung")
	if code != 0 {
		t.Fatalf("die Kommandozeile gab %d: %s", code, ausgabe)
	}

	pruefeRangUndArt(t, "Kommandozeile", ausgabe)

	status, rumpf := an.ruf(t, http.MethodGet, "/sammlung", "")
	if status != http.StatusOK {
		t.Fatalf("der Dienst gab %d: %s", status, rumpf)
	}

	pruefeRangUndArt(t, "Dienst", rumpf)

	text, istFehler := mcpRuf(t, umgebung, "lesen.sammlung", map[string]any{})
	if istFehler {
		t.Fatalf("das Werkzeug gab einen Fehler: %s", text)
	}

	pruefeRangUndArt(t, "MCP", text)
}

// pruefeRangUndArt sucht die Wunschzeile und hält sie gegen die Eingabe.
func pruefeRangUndArt(t *testing.T, flaeche, roh string) {
	t.Helper()

	var antwort struct {
		Exemplare []struct {
			Wunsch     bool   `json:"wunsch"`
			Wunschrang string `json:"wunschrang"`
			Wunschart  string `json:"wunschart"`
		} `json:"exemplare"`
	}

	if fehler := json.Unmarshal([]byte(roh), &antwort); fehler != nil {
		t.Fatalf("%s: die Antwort deuten: %v — %s", flaeche, fehler, roh)
	}

	for _, zeile := range antwort.Exemplare {
		if !zeile.Wunsch {
			continue
		}

		if zeile.Wunschrang != "hoch" || zeile.Wunschart != "jederdruck" {
			t.Errorf("%s nennt %q und %q", flaeche, zeile.Wunschrang, zeile.Wunschart)
		}

		return
	}

	t.Errorf("%s nennt gar keinen Wunsch: %s", flaeche, roh)
}
