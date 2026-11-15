// **Ein offenes Sideboard trägt bis an jede Fläche** — Befund 1 des
// Nachlaufs über Phase 4.6
// (#1293).
//
// # Der Befund
//
// Die Korrekturrunde nach dem Zweitblick führte `Format.SideboardOffen` ein,
// weil `sideboardgroesse: 0` zwei Dinge zugleich sagte: „kein Sideboard" und
// „keine Grenze". CR 100.4b legt bei limitiertem Spiel die übrigen Karten des
// Pools ins Sideboard, **ohne eine Zahl zu nennen**.
//
// **Das Feld stand danach nur in `internal/domain`.** Die Tabelle kannte es
// nicht, der Leser gab es nicht zurück, das DTO trug es nicht. Nach dem ersten
// Schreiben war wieder `false` in Kraft — und `mse deck pruefen` sagte für ein
// limitiertes Deck mit drei Karten im Sideboard:
//
//	! Das Sideboard hat 3 Karten; "Draft" kennt kein Sideboard.
//
// **Beobachtet, nicht erschlossen.** Die Korrektur war in der laufenden
// Anwendung wirkungslos, während `internal/domain` grün war.
//
// # Warum dieser Test hier steht
//
// Weil kein Test einer einzelnen Schicht ihn gefunden hätte: `domain` war
// grün, `storage` fragte nicht danach, und `app` arbeitet gegen Doppel, die
// nichts ablegen. Erst der Weg **durch** Bestand und Fläche zeigt es.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// zweiterDruck ist eine zweite Druckkennung — fürs Sideboard.
//
// **Eine zweite und nicht dieselbe**: derselbe Druck in beiden Bereichen
// prüfte nebenbei die Höchstzahl gleicher Karten mit, und ein rotes Ergebnis
// hätte dann zwei Lesarten.
const zweiterDruck = "a1b2c3d4-0000-4000-8000-00000000000b"

// **Angelegt aus der Bauart, wieder gelesen — an jeder Fläche noch offen.**
func TestEinOffenesSideboardTraegtBisAnJedeFlaeche(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := deckumgebung(t)

		mussKommandozeile(t, umgebung, "format", "neu", "limitiert", "Draft", "magic")

		pruefeOffenesSideboard(t, mussKommandozeile(t, umgebung, "formate", "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := deckumgebung(t)

		kennungAusMCP(t, umgebung, "schreiben.format.ausbauart", map[string]any{
			"bauart": "limitiert", "name": "Draft", "spiel": "magic",
		})

		// **MCP gibt die Liste nackt** — ohne Hülle, anders als CLI und REST.
		pruefeOffenesSideboard(t,
			`{"formate":`+mussMCP(t, umgebung, "lesen.formate", map[string]any{})+`}`)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := deckumgebung(t)
		an := starteDienst(t, umgebung)

		kennungAusREST(t, an, "/formate",
			`{"bauart":"limitiert","name":"Draft","spiel":"magic"}`)

		rumpf := mussREST(t, an, http.MethodGet, "/formate", "")

		pruefeOffenesSideboard(t, `{"formate":`+rumpf+`}`)
	})
}

// **Und ein limitiertes Deck mit Sideboard bekommt kein falsches Nein.**
//
// Das ist die Abnahme, die zählt: nicht dass ein Feld da ist, sondern dass
// das **Urteil** sich ändert. Ein Feld, das durchgereicht wird und niemanden
// erreicht, wäre dieselbe Lücke eine Ebene weiter.
func TestEinLimitiertesDeckDarfEinSideboardHaben(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)

	deckID := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"deck", "neu", "Sealed", "magic"))

	formatID := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"format", "neu", "limitiert", "Draft", "magic"))

	mussKommandozeile(t, umgebung, "deck", "format", deckID, formatID)

	// **Vierzig im Hauptdeck, damit die Untergrenze nicht mitredet**, und
	// drei daneben — genau der Fall, den `mse deck pruefen` abwies.
	mussKommandozeile(t, umgebung, "deck", "legen", deckID, druckkennung, "40")
	mussKommandozeile(t, umgebung, "deck", "legen", deckID, zweiterDruck, "3", "sideboard")

	roh := mussKommandozeile(t, umgebung, "deck", "pruefen", deckID, "--json")

	var gelesen struct {
		Urteil struct {
			Groessen []struct {
				Meldung string `json:"meldung"`
			} `json:"groessen"`
		} `json:"urteil"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	for _, befund := range gelesen.Urteil.Groessen {
		if strings.Contains(befund.Meldung, "Sideboard") {
			t.Errorf("das Urteil beanstandet das Sideboard: %q", befund.Meldung)
		}
	}
}

// pruefeOffenesSideboard sichert, dass das Feld die Fläche erreicht — **und**
// dass die Zahl daneben null geblieben ist.
//
// **Beides**, denn genau diese zwei Werte sagten einmal dasselbe. Ein Test,
// der nur `sideboardOffen` prüfte, ließe eine Umsetzung durchgehen, die
// nebenbei eine erfundene Zahl einträgt.
func pruefeOffenesSideboard(t *testing.T, roh string) {
	t.Helper()

	var gelesen struct {
		Formate []struct {
			Name             string `json:"name"`
			Sideboardgroesse int    `json:"sideboardgroesse"`
			SideboardOffen   bool   `json:"sideboardOffen"`
		} `json:"formate"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	if len(gelesen.Formate) != 1 {
		t.Fatalf("%d Formate, erwartet eines (%s)", len(gelesen.Formate), roh)
	}

	format := gelesen.Formate[0]

	if !format.SideboardOffen {
		t.Errorf("%q kommt ohne offenes Sideboard zurück (%s)", format.Name, roh)
	}

	if format.Sideboardgroesse != 0 {
		t.Errorf("%q nennt eine Sideboardgröße von %d — die Bauart nennt keine",
			format.Name, format.Sideboardgroesse)
	}
}

// **`format setzen` nimmt das offene Sideboard nicht still weg.**
//
// # Die Falle, die dieser Test schließt
//
// `format setzen` nennt jeden Wert an seiner Stelle. Eine Marke, die **keine**
// Stelle hat, verschwindet dabei — wer den Namen eines limitierten Formats
// ändert, verlöre nebenbei das offene Sideboard, und nichts sagte es ihm.
//
// Die siebte Stelle heißt deshalb `<sideboard>` und nimmt **drei** Antworten:
// eine Zahl, `0` für „kein Sideboard" und `offen` für „ohne Grenze". Damit
// kann sie nicht vergessen werden — sie ist Pflicht.
func TestFormatSetzenBehaeltEinOffenesSideboard(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)

	formatID := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"format", "neu", "limitiert", "Draft", "magic"))

	mussKommandozeile(t, umgebung, "format", "setzen",
		formatID, "1", "Sealed", "magic", "40", "0", "offen", "0")

	pruefeOffenesSideboard(t, mussKommandozeile(t, umgebung, "formate", "--json"))
}

// **Und eine Zahl an derselben Stelle schließt es wieder.**
//
// Die Gegenprobe zum Test darüber: bliebe `offen` einmal gesetzt hängen,
// wäre der erste grün und die Marke unabstellbar.
func TestEineZahlSchliesstDasSideboardWieder(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)

	formatID := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"format", "neu", "limitiert", "Draft", "magic"))

	mussKommandozeile(t, umgebung, "format", "setzen",
		formatID, "1", "Sealed", "magic", "40", "0", "15", "0")

	roh := mussKommandozeile(t, umgebung, "formate", "--json")

	var gelesen struct {
		Formate []struct {
			Sideboardgroesse int  `json:"sideboardgroesse"`
			SideboardOffen   bool `json:"sideboardOffen"`
		} `json:"formate"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	if gelesen.Formate[0].SideboardOffen || gelesen.Formate[0].Sideboardgroesse != 15 {
		t.Errorf("das Sideboard ist %+v — erwartet 15 und nicht offen", gelesen.Formate[0])
	}
}

// **Auch über MCP bleibt es stehen, wenn niemand es nennt.**
//
// Dort gilt die andere Regel als an der Kommandozeile: „was nicht gesagt
// wird, bleibt". Ein Umbenennen, das die Marke abräumte, wäre derselbe
// stille Verlust — nur über eine andere Fläche.
func TestMCPBehaeltEinOffenesSideboardBeimUmbenennen(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)

	formatID := kennungAusMCP(t, umgebung, "schreiben.format.ausbauart", map[string]any{
		"bauart": "limitiert", "name": "Draft", "spiel": "magic",
	})

	mussMCP(t, umgebung, "schreiben.format.setzen", map[string]any{
		"id": formatID, "fassung": 1, "name": "Sealed",
	})

	pruefeOffenesSideboard(t,
		`{"formate":`+mussMCP(t, umgebung, "lesen.formate", map[string]any{})+`}`)
}

// **Und wer es nennt, kann es abstellen.**
func TestMCPKannDasOffeneSideboardAbstellen(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)

	formatID := kennungAusMCP(t, umgebung, "schreiben.format.ausbauart", map[string]any{
		"bauart": "limitiert", "name": "Draft", "spiel": "magic",
	})

	mussMCP(t, umgebung, "schreiben.format.setzen", map[string]any{
		"id": formatID, "fassung": 1, "sideboardOffen": false, "sideboardgroesse": 15,
	})

	roh := `{"formate":` + mussMCP(t, umgebung, "lesen.formate", map[string]any{}) + `}`

	var gelesen struct {
		Formate []struct {
			Sideboardgroesse int  `json:"sideboardgroesse"`
			SideboardOffen   bool `json:"sideboardOffen"`
		} `json:"formate"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	if gelesen.Formate[0].SideboardOffen || gelesen.Formate[0].Sideboardgroesse != 15 {
		t.Errorf("das Sideboard ist %+v — erwartet 15 und nicht offen", gelesen.Formate[0])
	}
}
