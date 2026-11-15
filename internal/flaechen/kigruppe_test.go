// **Die atomare Gruppe über jede Fläche ohne Fenster** — Werkstück A13.1b von
// Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass alle drei Flächen dieselbe Absage geben**, wenn es den Auftrag nicht
// gibt — und zwar **bevor** irgendetwas geschrieben ist.
//
// **Und dass ein Ruf durch die ganze Kette trägt**: ein Batch über zwei
// Karten, danach die Gruppe, danach stehen beide Werte da. Das ist die
// Zusage, für die es A13.1b gibt.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine Gruppe mit einem erfundenen Auftrag scheitert überall gleich.**
func TestEineGruppeUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "eine Gruppe ohne Auftrag", abweisung{
		klasse: klasseNichtGefunden,
		// **Der Kern ist der übersetzte Satz und nicht der innere.**
		// `service` sagt „den Auftrag %q gibt es nicht"; was ein Mensch
		// liest, macht `app.Uebersetze` daraus — und geprüft wird, was
		// ankommt (§8).
		kern: "kein Vorschlag",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).KIGruppeUebernehmen(t.Context(), []app.KIGruppenwahlDTO{
				{Kennung: "gibtesnicht", Felder: []string{"name"}},
			}).Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"ki-gruppe", "gibtesnicht=name")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.ki.gruppe",
				map[string]any{
					"wahl": []map[string]any{
						{"kennung": "gibtesnicht", "felder": []string{"name"}},
					},
				})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/ki/gruppe",
				`{"wahl":[{"kennung":"gibtesnicht","felder":["name"]}]}`)
		},
	})
}

// **Und ein wirklich gefahrener Weg: Batch, Gruppe, zwei geschriebene
// Karten.**
//
// # Was dieser Lauf hält, den darüber nicht hält
//
// **Dass die Kette trägt.** Der darüber prüft die Absage; hier geht ein Ruf
// durch alles: Auftragswerk, Kostenwache, Modell, Schemaprüfung, die
// Aufsicht über jede Karte und **eine** Transaktion.
//
// **Und dass beide Karten danach ihren Wert tragen** — nicht eine.
func TestEineErzeugteGruppeSchreibtBeideKarten(t *testing.T) {
	t.Parallel()

	umgebung := mitEigenerSaat(t)
	laufzeit := starteLaufzeit(t, derVorschlag)

	kiEinschalten(t, umgebung, laufzeit.adresse)

	setID, _ := eineAbnahmekarte(t, umgebung)

	// **Zwei Karten desselben Sets** — eine Gruppe über eine Karte prüfte die
	// Schleife nicht, und zwei aus **verschiedenen** Sets stünden für 0109,
	// nicht für diesen Lauf.
	karten := kartenzeilenAus(t, mussKommandozeile(t, umgebung,
		"karten", setID, "--json"))
	if len(karten) < 2 {
		t.Fatalf("das Probeset hat %d Karten — für eine Gruppe braucht es zwei",
			len(karten))
	}

	erste, zweite := zahl(t, karten[0].ID), zahl(t, karten[1].ID)

	agent := starteMCPProzess(t, umgebung)

	roh := agent.muss("schreiben.ki.batch", map[string]any{
		"karteIds": []int64{erste, zweite},
		"anbieter": "ollama",
		"alias":    probemodell,
		"auftrag":  derAuftrag,
	})

	var gelesenerBatch struct {
		Ziele []struct {
			Kennung string `json:"kennung"`
			Fehler  string `json:"fehler"`
		} `json:"ziele"`
	}

	if fehler := json.Unmarshal([]byte(roh), &gelesenerBatch); fehler != nil {
		t.Fatalf("den Batch lesen: %v — %s", fehler, roh)
	}

	if len(gelesenerBatch.Ziele) != 2 {
		t.Fatalf("der Batch gab %d Ziele: %s", len(gelesenerBatch.Ziele), roh)
	}

	kennungen := make([]string, 0, 2)

	for _, ziel := range gelesenerBatch.Ziele {
		if ziel.Fehler != "" {
			t.Fatalf("ein Ziel scheiterte: %s", ziel.Fehler)
		}

		kennungen = append(kennungen, ziel.Kennung)
	}

	wahl := make([]map[string]any, 0, 2)
	for _, kennung := range kennungen {
		wahl = append(wahl, map[string]any{
			"kennung": kennung, "felder": []string{dasFeld},
		})
	}

	antwort := agent.muss("schreiben.ki.gruppe", map[string]any{"wahl": wahl})

	if !strings.Contains(antwort, `"karten": 2`) {
		t.Errorf("die Gruppe schrieb nicht zwei Karten:\n%s", antwort)
	}

	// **Und beide tragen den Wert.** Das ist der Unterschied zu „die Antwort
	// sagt zwei": hier steht, was im Bestand liegt.
	gelesen := agent.muss("lesen.karten", map[string]any{"setID": zahl(t, setID)})

	for _, karte := range []int64{erste, zweite} {
		if wert := feldwertMCP(t, gelesen, karte); wert != "Aus dem Modell" {
			t.Errorf("Karte %d trägt %q", karte, wert)
		}
	}
}
