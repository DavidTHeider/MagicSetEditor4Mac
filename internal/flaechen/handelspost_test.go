package flaechen_test

// **Der Poststand über jede Fläche** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// # Was hier gemessen wird
//
// **Dass beide Flächen dieselbe Zahl sagen** (§16) — und dass sie **nur**
// Zahlen sagen. Das Issue verlangt „technische Zustände und Zähler, ohne
// Rohsammlungen/Decklisten oder Geheimnisse"; ein Stand, der eine Nachricht
// durchscheinen ließe, wäre eine zweite Sicht auf den Handel.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestDerPoststandUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	const wer = "bgegenueber"

	ausgabe, _, code := kommandozeile(t, umgebung, "--json", "handelspost", wer)
	if code != 0 {
		t.Fatalf("die Kommandozeile gab %d: %s", code, ausgabe)
	}

	vonHand := poststandAus(t, ausgabe)

	// **Ohne Handel steht nichts aus** — und das ist eine Auskunft und kein
	// Fehler.
	if vonHand.Offen != 0 || vonHand.HoechsteVersuche != 0 {
		t.Errorf("die Kommandozeile sagt %+v", vonHand)
	}

	if vonHand.Kennung != wer {
		t.Errorf("der Stand nennt %q", vonHand.Kennung)
	}

	status, rumpf := an.ruf(t, http.MethodGet, "/peer/kontakte/"+wer+"/post", "")
	if status != http.StatusOK {
		t.Fatalf("der Dienst gab %d: %s", status, rumpf)
	}

	ueberDenDienst := poststandAus(t, rumpf)

	if ueberDenDienst != vonHand {
		t.Errorf("die Flächen sagen %+v und %+v", ueberDenDienst, vonHand)
	}

	// **Und nichts über den Inhalt.** Ein Stand, der eine Nachricht oder
	// einen Vorgang nennte, wäre eine zweite Sicht auf den Handel.
	for _, wort := range []string{"nachricht", "vorgang", "angebot", "position"} {
		if strings.Contains(strings.ToLower(rumpf), wort) {
			t.Errorf("der Stand nennt %q: %s", wort, rumpf)
		}
	}
}

// poststand ist, was von einem Stand zurückkommt.
type poststand struct {
	Kennung          string `json:"kennung"`
	Offen            int    `json:"offen"`
	HoechsteVersuche int    `json:"hoechsteVersuche"`
}

// poststandAus liest die Antwort.
func poststandAus(t *testing.T, roh string) poststand {
	t.Helper()

	var stand poststand
	if fehler := json.Unmarshal([]byte(roh), &stand); fehler != nil {
		t.Fatalf("die Antwort deuten: %v — %s", fehler, roh)
	}

	return stand
}
