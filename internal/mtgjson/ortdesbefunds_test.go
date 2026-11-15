package mtgjson_test

// **Jeder Befund nennt sein Set** — der Befund des Zweitblicks über Phase 4.2
// zu #1289.
//
// # Was hier schiefging, und warum es zweimal schiefging
//
// Der Fuzzlauf fand am 2026-08-29 einen Befund **ohne Ort** und der Fall wurde
// behoben — im Schemaprüfer, über `ortVon`. Die **Klasse** blieb offen: die
// Befunde der Zuordnung bauen ihren Ort an drei weiteren Stellen, und alle drei
// nahmen `roh.Code` roh. Bei einem AllSets-Schlüssel `""` entstand daraus
// `" → uuid"` — ein Ort, der genauso wenig sagt wie gar keiner, und der am
// Fuzzwächter vorbeikam, weil der auf `Ort == ""` prüfte.
//
// Das ist die Gestalt aus §10: *„welche anderen Beispiele hätten denselben Weg
// genommen?"* — hier drei, und keines war betrachtet worden.
//
// # Was dieser Lauf hält
//
// Ein Ort beginnt nie mit dem Trenner. Das ist die schwächste Form der Zusage
// „der Befund nennt sein Set" und zugleich die, die sich ohne Kenntnis des
// Sets prüfen lässt — welchen Namen ein Set bekommt, sagen die Läufe daneben.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// ohneSetImOrt sagt, ob ein Ort sein Set verschweigt.
func ohneSetImOrt(ort string) bool {
	return ort == "" || strings.HasPrefix(strings.TrimSpace(ort), "→")
}

func TestJederBefundNenntSeinSet(t *testing.T) {
	t.Parallel()

	zuordnung, da := mtgjson.ZuordnungFuer("magic")
	if !da {
		t.Fatal("die Zuordnung für `magic` fehlt — der Lauf greift daneben")
	}

	for name, roh := range map[string]string{
		"ohne Schlüssel und ohne Code": `{"data":{"":{"cards":[{"name":"A","uuid":"u1","gibtsNicht":1}]}}}`,
		"ohne Code, aber mit Namen": `{"data":{"":{"name":"Probe",` +
			`"cards":[{"name":"A","uuid":"u1","gibtsNicht":1}]}}}`,
		"ein Feld je Karte": `{"data":{"":{"cards":[{"name":"A","uuid":"u1",` +
			`"convertedManaCost":"keine Zahl"}]}}}`,

		// **Leerraum ist kein Name**, und das Feld auch nicht — der Fund des
		// Fuzzlaufs vom 2026-08-30, minimiert. Der Ort lautete `"  → "`.
		"Setschlüssel und Feldname sind Leerraum": `{"data":{" ":{"cards":[{"":{}}]}}}`,

		// **Ein Schlüssel in anderer Schreibung** — der Fund der Fuzzkampagne
		// vom 2026-10-10 (C02, #2227), unverändert. Der Leser nahm „Code"
		// als `code` und las ein Set; die Prüfung las eine Abbildung von
		// Sets, fand darin `"":""` und meldete das **ohne Ort**.
		"Code in anderer Schreibung": `{"":"","Code":"0","00000":{"":{"":"","":"","":[]}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			quelle, fehler := mtgjson.LiesInhalt([]byte(roh), "probe.json", zuordnung)
			if fehler != nil {
				t.Fatalf("LiesInhalt: %v", fehler)
			}

			if len(quelle.Befunde) == 0 {
				t.Fatal("kein Befund — der Lauf greift daneben")
			}

			for _, befund := range quelle.Befunde {
				if ohneSetImOrt(befund.Ort) {
					t.Errorf("ein Befund ohne Set im Ort: %q — %q",
						befund.Ort, befund.Meldung)
				}

				if strings.HasSuffix(befund.Ort, "→ ") || strings.HasSuffix(befund.Ort, "→") {
					t.Errorf("ein Befund ohne Feld im Ort: %q — %q",
						befund.Ort, befund.Meldung)
				}
			}
		})
	}
}

// **Ein Set ohne Kürzel bekommt seinen Namen** und nicht den festen Satz.
func TestEinSetOhneKuerzelHeisstNachSeinemNamen(t *testing.T) {
	t.Parallel()

	zuordnung, _ := mtgjson.ZuordnungFuer("magic")

	quelle, fehler := mtgjson.LiesInhalt([]byte(
		`{"data":{"":{"name":"Probe","cards":[{"name":"A","uuid":"u1"}]}}}`),
		"probe.json", zuordnung)
	if fehler != nil {
		t.Fatalf("LiesInhalt: %v", fehler)
	}

	for _, befund := range quelle.Befunde {
		if !strings.HasPrefix(befund.Ort, "Probe") {
			t.Errorf("der Ort %q nennt den Namen des Sets nicht", befund.Ort)
		}
	}
}

// **Auch der Schreibweg nennt sein Set.**
//
// Der Nachlauf über Phase 4.2 hat angestrichen, dass die Klasse nur für die
// Leserfamilie geschlossen war: `setAus` fiel roh auf den Setnamen zurück, und
// ein Set ohne `code` und mit leerem Namen gab einen Ort, der mit dem Trenner
// beginnt.
func TestAuchDerSchreibwegNenntSeinSet(t *testing.T) {
	t.Parallel()

	zuordnung, da := mtgjson.ZuordnungFuer("magic")
	if !da {
		t.Fatal("die Zuordnung für `magic` fehlt — der Lauf greift daneben")
	}

	set := &domain.Set{
		Name: "  ",
		Cards: []*domain.Card{
			{Values: []domain.Value{{Field: "super type", Text: "Basic Land"}}},
		},
	}

	_, befunde := zuordnung.Schreibe(set)
	if len(befunde) == 0 {
		t.Fatal("kein Befund — der Lauf greift daneben")
	}

	for _, befund := range befunde {
		if ohneSetImOrt(befund.Ort) {
			t.Errorf("ein Befund ohne Set im Ort: %q — %q", befund.Ort, befund.Meldung)
		}

		if strings.HasSuffix(befund.Ort, "→ ") || strings.HasSuffix(befund.Ort, "→") {
			t.Errorf("ein Befund ohne Feld im Ort: %q — %q", befund.Ort, befund.Meldung)
		}
	}
}
