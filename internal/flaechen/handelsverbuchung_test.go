package flaechen_test

// **Die Verbuchung über jede Fläche** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Was hier gemessen wird
//
// **Dass beide Flächen dieselbe fachliche Antwort geben** (§16) — und zwar
// auch die **Ablehnung**. Ein Lauf, der nur den Erfolgsweg fährt, hat die
// Hälfte ausgelassen (§10), und die Ablehnung ist hier die interessantere:
// ohne bestätigte Übergabe wird nicht gebucht, und beide Flächen müssen das
// gleich sagen.
//
// **Und dass das Journal den Ausgangssaldo führt.** Der heutige Bestand steht
// als Einfuhr darin — keine erfundenen Käufe, kein Preis.
//
// # Und seit dem 2026-09-19 auch der geglückte Abschluss
//
// **Er steht nebenan** — `handelsleitung_test.go` führt einen Handel von
// `entwurf` bis `beidseitigVerbucht`, abwechselnd über beide Flächen
// (#1637).
// Hier stand bis dahin, dass es ihn nicht gibt: `service.Handelsleitung` aus
// `T1` hatte keinen Aufrufer, und ohne die Schritte davor kam kein Vorgang
// bis zur bestätigten Übergabe. **Das war ein Befund und keine Eigenschaft
// dieses Werkstücks**, und er ist behoben.
//
// **Was dieser Lauf misst, bleibt dasselbe**: dass beide Flächen dieselben
// fachlichen Fehler mit derselben Begründung abweisen, und dass das Journal
// über beide gleich zurückkommt.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// derUnbekannteHandel ist ein Vorgang, den es nicht gibt.
const derUnbekannteHandel = "v-gibtesnicht"

// dieVerbuchungsbefehle sind die drei Schritte, wie sie im Katalog stehen.
//
// **Als ganze Zeichenketten und nicht als Wortlisten**: `docs/FAEHIGKEITEN.md`
// nennt sie so, und ein Wächter hält beide Seiten zusammen — wer hier
// zerlegte, hätte einen Lauf, den niemand der Fähigkeit zuordnen kann.
var dieVerbuchungsbefehle = []string{
	"mse handel verbuchen",
	"mse handel abgleichen",
	"mse handel bestaetigen",
}

// derSchrittweg ist der Endpunkt, wie er im Katalog steht.
const derSchrittweg = "POST /handel/{vorgang}/schritt"

func TestDieVerbuchungWeistUeberJedeFlaecheGleichAb(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	// **Jeder der drei Schritte wird an jeder Fläche gleich abgewiesen** —
	// mit demselben unbekannten Vorgang, derselben Klasse und demselben Kern
	// (Zweitblick über Phase C, Befund 3: hier rief die Kommandozeile den
	// Befehl ohne Argument und prüfte eine Bedienungsklage, der Dienst einen
	// unbekannten Vorgang und nur „nicht 200" — zwei verschiedene Nein).
	// **Je Schritt seine Argumente**, an jeder Fläche dieselben: `verbuchen`
	// nimmt die Freigabe aus dem Fenster, `bestaetigen` Partei und Freigabe,
	// `abgleichen` nichts (`cmd/mse/handelsverbuchung.go`, `handelsargumente`).
	const erfundeneFreigabe = "f-erfunden"
	argumente := map[string][]string{
		"verbuchen":   {erfundeneFreigabe},
		"abgleichen":  nil,
		"bestaetigen": {"bgegenueber", erfundeneFreigabe},
	}
	fenster := map[string]func(*app.Bindungen) *app.Fehler{
		"verbuchen": func(b *app.Bindungen) *app.Fehler {
			return b.Handelsverbuchung(t.Context(), derUnbekannteHandel, erfundeneFreigabe).Fehler
		},
		"abgleichen": func(b *app.Bindungen) *app.Fehler {
			return b.Handelsabgleich(t.Context(), derUnbekannteHandel).Fehler
		},
		"bestaetigen": func(b *app.Bindungen) *app.Fehler {
			return b.Handelsbestaetigung(t.Context(), derUnbekannteHandel, "bgegenueber", erfundeneFreigabe).Fehler
		},
	}

	for _, befehl := range dieVerbuchungsbefehle {
		schritt := strings.Fields(befehl)[2]
		rumpf := `{"schritt":"` + schritt + `","wer":"bgegenueber","freigabe":"` + erfundeneFreigabe + `"}`

		abweisungGleich(t, befehl, abweisung{
			klasse: klasseNichtGefunden,
			kern:   "den Handelsvorgang gibt es nicht",
			cli: func() (string, int) {
				aus, klage, code := kommandozeile(t, umgebung,
					append(append(strings.Fields(befehl)[1:], derUnbekannteHandel), argumente[schritt]...)...)

				return aus + klage, code
			},
			ohneMCP: "MCP schreibt keinen Handel (FAEHIGKEITEN.md)",
			rest: func() (int, string) {
				return an.ruf(t, http.MethodPost,
					strings.ReplaceAll(pfadAus(derSchrittweg), "{vorgang}", derUnbekannteHandel), rumpf)
			},
			fenster: func() *app.Fehler { return fenster[schritt](fensterAus(t, umgebung)) },
		})
	}

	// **Und mit Vorgang sagt die Kommandozeile, was fehlt** — der Vorgang und
	// nicht der Schlüsselbund. Wer den Ausweis zuerst holte, ließe jemanden
	// einen Dialog wegklicken, um danach „gibt es nicht" zu lesen.
	ausgabe, klage, code := kommandozeile(t, umgebung,
		"--json", "handel", "verbuchen", derUnbekannteHandel)
	if code == 0 {
		t.Fatalf("die Kommandozeile buchte einen Handel, den es nicht gibt: %s",
			ausgabe)
	}

	if !strings.Contains(strings.ToLower(klage+ausgabe), "vorgang") {
		t.Errorf("die Klage nennt den Vorgang nicht: %s / %s", klage, ausgabe)
	}

	// **Ein unbekannter Schritt ist eine Beanstandung und kein 500.**
	fremd, rumpf := an.ruf(t, http.MethodPost,
		strings.ReplaceAll(pfadAus(derSchrittweg), "{vorgang}",
			derUnbekannteHandel),
		`{"schritt":"tanzen","wer":""}`)
	if fremd != http.StatusBadRequest && fremd != http.StatusUnprocessableEntity {
		t.Errorf("der fremde Schritt gab %d: %s", fremd, rumpf)
	}

	ausgabe, _, code = kommandozeile(t, umgebung,
		"--json", "handel", "tanzen", derUnbekannteHandel)
	if code == 0 {
		t.Errorf("die Kommandozeile nahm einen fremden Schritt: %s", ausgabe)
	}
}

func TestDasJournalFuehrtDenAusgangssaldoUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	ausgabe, _, code := kommandozeile(t, umgebung, "--json", "handelsjournal")
	if code != 0 {
		t.Fatalf("die Kommandozeile gab %d: %s", code, ausgabe)
	}

	vonHand := journalAus(t, ausgabe)

	// **Über den Dienst geht dasselbe** — ein Vorgang, den es nicht gibt,
	// hat ein leeres Journal, und das ist eine Auskunft und kein Fehler.
	status, rumpf := an.ruf(t, http.MethodGet,
		strings.ReplaceAll(pfadAus(derJournalweg), "{vorgang}",
			derUnbekannteHandel), "")
	if status != http.StatusOK {
		t.Fatalf("der Dienst gab %d: %s", status, rumpf)
	}

	if len(journalAus(t, rumpf)) != 0 {
		t.Errorf("ein Vorgang, den es nicht gibt, führt Zeilen: %s", rumpf)
	}

	// **Jede Zeile des Saldos ist eine Einfuhr ohne Vorgang und ohne Preis.**
	for _, zeile := range vonHand {
		if zeile.Anlass != "einfuhr" || zeile.Art != "zugang" {
			t.Errorf("die Saldozeile ist ein %q/%q", zeile.Art, zeile.Anlass)
		}

		if zeile.Vorgang != "" || zeile.Grund != "" {
			t.Errorf("die Saldozeile nennt %q/%q", zeile.Vorgang, zeile.Grund)
		}
	}

	// **Und kein Preis steht darin** — weder im Feldnamen noch als Wert.
	for _, wort := range []string{"preis", "waehrung", "erwerb"} {
		if strings.Contains(strings.ToLower(ausgabe), wort) {
			t.Errorf("das Journal nennt %q: %s", wort, ausgabe)
		}
	}
}

// derJournalweg ist der Endpunkt, wie er im Katalog steht.
const derJournalweg = "GET /handel/{vorgang}/journal"

// pfadAus schneidet den Pfad aus einer Katalogzeile.
//
// **Damit die Zeichenkette im Lauf dieselbe ist wie im Katalog** — zwei
// Schreibweisen liefen auseinander, und die im Katalog wäre die, an der
// niemand etwas merkt.
func pfadAus(katalogzeile string) string {
	return strings.Fields(katalogzeile)[1]
}

// journalzeile ist, was von einer Bewegung zurückkommt.
type journalzeile struct {
	ID       int64  `json:"id"`
	Art      string `json:"art"`
	Anlass   string `json:"anlass"`
	Vorgang  string `json:"vorgang"`
	Grund    string `json:"grund"`
	Printing string `json:"printing"`
	Menge    int    `json:"menge"`
}

// journalAus liest die Antwort.
func journalAus(t *testing.T, roh string) []journalzeile {
	t.Helper()

	var zeilen []journalzeile
	if fehler := json.Unmarshal([]byte(roh), &zeilen); fehler != nil {
		t.Fatalf("die Antwort deuten: %v — %s", fehler, roh)
	}

	return zeilen
}
