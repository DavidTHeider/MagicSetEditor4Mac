// **Die Balanceanalyse über jede Fläche ohne Fenster** — Werkstück A13.2b von
// Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass das Format an jeder Fläche ankommt** — und dass eine erfundene Art
// überall dieselbe Absage bekommt, bevor irgendetwas geschieht.
//
// **Und dass der Weg wirklich trägt** — der zweite Lauf unten schickt einen
// Ruf durch die ganze Kette, mit einem antwortenden Modell aus
// `kilaufzeitstoff_test.go`.
//
// # Und was nicht
//
// **Ein Vergleich gegen einen gefüllten Index.** Der unten läuft **ohne**
// gebauten Index, und das ist Absicht: dann vergibt der Kontextdienst keinen
// Beleg, und die erfundene Kennung des Modells wird zum Befund. Wie ein
// gebauter Index aussieht, steht in `indexlauf_test.go`.
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine Balancefrage ohne Laufzeit scheitert überall gleich.**
func TestEineBalancefrageUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "eine Balancefrage ohne Laufzeit", abweisung{
		klasse: klasseValidierung,
		kern:   "eingeschaltet",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).KIEinschaetzen(t.Context(), app.KIEinschaetzungsbezug{
				Art: string(domain.EinschaetzungBalance), KarteID: 1, Anbieter: "ollama", Alias: "gemma4:8b",
				Frage: "Wie steht sie da?", Format: "Commander",
			}).Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"ki-einschaetzung", "1", string(domain.EinschaetzungBalance),
				"ollama", "gemma4:8b", "Wie steht sie da?", "Commander")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.ki.einschaetzung",
				map[string]any{
					"karteId": 1, "art": string(domain.EinschaetzungBalance),
					"anbieter": "ollama", "alias": "gemma4:8b",
					"frage": "Wie steht sie da?", "format": "Commander",
				})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/karten/1/ki-einschaetzung",
				`{"art":"balance","anbieter":"ollama","alias":"gemma4:8b",`+
					`"frage":"Wie steht sie da?","format":"Commander"}`)
		},
	})
}

// derVergleich ist, was das Probemodell auf eine Balancefrage antwortet.
//
// **Mit einer Kennung, die es nicht gibt.** Ohne gebauten Index vergibt der
// Kontextdienst keinen einzigen Beleg — und **genau das ist der Fall, der
// zählt**: ein Modell, das „[B1]" schreibt, bekommt damit keinen Beleg,
// sondern einen Befund.
const derVergleich = `{"einschaetzung":"Sie steht zwischen beiden [B1].",` +
	`"einschraenkungen":["Nur die mitgegebenen Stellen."],` +
	`"vergleiche":[{"beleg":"B1","name":"Funkenwurf",` +
	`"unterschiede":["ein Mana billiger"]}],` +
	`"nichtBetrachtet":["Zusammenspiel mit Manaflut"]}`

// **Eine wirklich erzeugte Balanceanalyse über MCP.**
//
// # Was dieser Lauf hält, den Lauf darüber nicht hält
//
// **Dass der Weg trägt.** Der darüber prüft die Absage; hier geht ein Ruf
// durch die ganze Kette: Kontextdienst, Modell, Belegprüfung, Antwort.
//
// **Dass das Format das Modell erreicht.** Der Kasten zeigt es; ob es im
// Auftrag steht, sagt nur der Rumpf, den die Laufzeit gesehen hat.
//
// **Und dass eine erfundene Kennung ein Befund wird, kein Beleg** — dieselbe
// Zusage wie beim Regelhinweis, und hier an der Fläche gemessen.
func TestEineErzeugteBalanceanalyseUeberMCP(t *testing.T) {
	t.Parallel()

	umgebung := mitEigenerSaat(t)
	laufzeit := starteLaufzeit(t, derVergleich)

	kiEinschalten(t, umgebung, laufzeit.adresse)

	_, karteID := eineAbnahmekarte(t, umgebung)

	agent := starteMCPProzess(t, umgebung)

	antwort := agent.muss("schreiben.ki.einschaetzung", map[string]any{
		"karteId": karteID, "art": string(domain.EinschaetzungBalance),
		"anbieter": "ollama", "alias": probemodell,
		"frage": "Wie steht sie da?", "format": "Commander",
	})

	// **Die Strategie kommt vom Programm** — samt dem Modus, in dem wirklich
	// gesucht wurde, und dem Satz über den fehlenden Reranker.
	for _, teil := range []string{
		`"strategie"`, "kein Reranker", `"format": "Commander"`,
		`"nichtBetrachtet"`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}

	// **Und die erfundene Kennung steht als Befund.**
	if !strings.Contains(antwort, `"belegt": false`) {
		t.Errorf("die Antwort gilt als belegt:\n%s", antwort)
	}

	if !strings.Contains(antwort, "Funkenwurf") {
		t.Errorf("der Befund nennt die Vergleichskarte nicht:\n%s", antwort)
	}

	// **Das Format stand im Auftrag** — nicht nur in der Antwort.
	rumpf := laufzeit.rumpftext(t)
	if !strings.Contains(rumpf, "Format: Commander") {
		t.Errorf("der Auftrag nennt das Format nicht: %s", rumpf)
	}

	// **Und nirgends steht eine Zahl über die Balance.** Es gibt kein Feld
	// dafür ([domain.Balancefelder]); dieser Lauf hält, dass auch an der
	// Fläche keines nachwächst.
	for _, wort := range []string{`"wert"`, `"punkte"`, `"rang"`, `"score"`} {
		if strings.Contains(strings.ToLower(antwort), wort) {
			t.Errorf("die Antwort trägt %s:\n%s", wort, antwort)
		}
	}
}
