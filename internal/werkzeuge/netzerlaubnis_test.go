package werkzeuge_test

// **Jeder Netzklient eines fremden Dienstes fragt vor jeder Anfrage** —
// Werkstück 8 von Phase 4.8
// (#1321).
//
// # Warum es diesen Wächter gibt
//
// `netz.Klient.MitErlaubnis` ist die Stelle, an der »aus heißt: kein einziger
// Aufruf« eingelöst wird — **vor** der Drossel und vor jeder Anfrage. Wer
// einen Klienten ohne sie baut, hat einen Dienst, den ein Schalter nicht mehr
// erreicht.
//
// **Genau das ist geschehen.** `Suchquellen` baute für magicthegathering.io,
// MTGJSON und Scrydex je einen blanken `netz.NeuerKlient(…)`; der Schalter
// wurde nur beim **Verdrahten** gelesen, und das Fenster verdrahtet einmal,
// beim Start. Gemessen in `internal/app/dienstschalterwirkt_test.go`: nach dem
// Abschalten ging die Suche hinaus und kam mit einer Karte zurück.
//
// Der Kommentar über `Suchquellen` sagte derweil, der Transport halte die
// Schalter „ohnehin". Er hielt sie für Scryfall.
//
// # Was er hält
//
// Jedes `netz.NeuerKlient(` in `internal/app` trägt in derselben Anweisung
// ein `.MitErlaubnis(` — oder steht in [ohneErlaubnis] mit einem Grund.
//
// # Was er nicht hält
//
// **Ob die Erlaubnis den richtigen Dienst prüft.** Er sieht den Aufruf, nicht
// sein Argument. Dass `dienstklient(wahl, domain.DienstMTGAPI)` auch
// magicthegathering.io meint, prüft der Lauf in `internal/app`.
//
// **Und ob sie überhaupt gelesen wird.** Er prüft, dass eine Erlaubnis
// **gesetzt** wird — nicht, dass der Weg nach draußen sie nimmt. Genau da lag
// die Lücke: `netz.Strom` rief `Do`, ohne zu fragen, und dieser Wächter war
// grün. Die andere Hälfte hält seit A3 (#1442)
// `internal/netz/erlaubnisvollstaendig_test.go`.
//
// Und er sieht nur `internal/app`. Wer das Netz sonst noch erreichen darf,
// zählt `internal/schichten/netz_test.go` auf.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ohneErlaubnis sind die Klienten, die bewusst keine Sperre tragen.
//
// **Eine Aufzählung mit Grund und kein Muster.** Wer einen Klienten hier
// einträgt, sagt, warum ein Schalter ihn nicht erreichen muss.
var ohneErlaubnis = map[string]string{
	// **Die beiden zentralen Bauwege** (N5, #1571). Sie bauen den
	// Transport nach dem eingestellten Netzweg und **nichts weiter**; wer
	// einen Dienst damit anspricht, hängt seine Erlaubnis daran — jede der
	// neun Stellen darüber tut es. Eine Sperre hier wäre die zweite, und
	// zwei Sperren über derselben Frage laufen auseinander.
	"netzklientMit": "Der zentrale Bauweg nach dem eingestellten Netzweg, " +
		"mit eigenen Fristen je Dienst — die Kopf- und Stillstandsfrist eines " +
		"KI-Anbieters ist eine Eigenschaft dieses Dienstes, der Proxyweg " +
		"nicht. Er baut den Transport und nichts weiter; wer einen Dienst " +
		"damit anspricht, hängt seine Erlaubnis daran, und alle vier tun es. " +
		"Eine Sperre hier wäre die zweite, und zwei Sperren über derselben " +
		"Frage laufen auseinander.",

	"FassungsnahtMit": "Die Fassungsprüfung hat einen eigenen Schalter — die " +
		"Einstellung `aktualisierung`, Vorgabe „aus“. Sie entscheidet, ob " +
		"überhaupt gefragt wird; der Klient wird eingerichtet und nicht " +
		"von selbst gefragt (§16, Regel 6). **Nachgesehen am 2026-09-02**, " +
		"weil genau diese Behauptung bei den Diensten falsch war: gelesen " +
		"wird sie zur Anfragezeit, in der Bindung selbst — " +
		"`aktualisierung.go`, NachNeuerFassungSehen, und " +
		"`aktualisierungholen.go`, zuHolen. Nicht beim Verdrahten.",
}

func TestJederNetzklientDerAnwendungFragtVorher(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join("..", "app")

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("internal/app lesen: %v", err)
	}

	gesehen := 0

	for _, eintrag := range eintraege {
		name := eintrag.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		gesehen += pruefeDatei(t, filepath.Join(wurzel, name), name)
	}

	// **Die Gegenprobe zur Aufzählung**: findet der Wächter gar keinen
	// Klienten, ist er grün und sagt nichts.
	//
	// **Sechs sind es in seinem Messbereich** — der gemeinsame Helfer für die
	// drei fremden Dienste, Scryfall, die Fassungsprüfung und seit A3
	// (#1442) der Bulkbezug der Preise bei MTGJSON. Hier stand einmal
	// „vier"; das war die Zahl der **Dienste**, nicht die der Bauzeilen, und
	// niemand hätte es gemerkt, weil die Zusicherung nur „größer null"
	// lautete. Angestrichen vom Nachlauf über Phase 4.8.
	//
	// **Jetzt steht die Zahl in der Zusicherung** und veraltet nicht mehr
	// still: wer einen Bauplatz hinzufügt, ändert sie hier mit und sagt
	// damit, dass er es wollte.
	// **Und seit A7 (#1446) die fünfte**: die lokale KI-Laufzeit. Dass ihre
	// Adresse auf der Schleife liegt, nimmt sie aus dieser Zählung nicht
	// heraus — „Loopback allein beweist keine lokale Inferenz", und ihr
	// Riegel fragt zwei Dinge: den Schalter und »nur lokal«.
	//
	// **Die sechste kam mit A8 (#1447)**: Anthropic. Ihr Riegel fragt
	// dieselben zwei — den Dienstschalter und »nur lokal« —, und unter der
	// Regel hält er schon die Modellliste auf.
	//
	// **Die siebte mit A9 (#1448)**: OpenAI. Seit ihr teilen sich die
	// Cloudanbieter **einen** Riegel (`pruefeCloudanbieter`) — zwei
	// gleichlautende liefen beim ersten Umbau auseinander. Gezählt werden
	// trotzdem die **Bauzeilen**, nicht die Riegel: jede Zeile ist ein
	// eigener Klient mit eigenen Fristen.
	//
	// **Die achte mit A10 (#1449)**: xAI. Eigene Zeile, eigene Fristen,
	// derselbe Riegel.
	//
	// **Die neunte mit A15 (#1514)**: die Einbettung für den Suchindex. Sie
	// geht an dieselbe Laufzeit wie die fünfte und trägt denselben Riegel —
	// eine **eigene Zeile**, weil ihre Stillstandsfrist eine andere ist: ein
	// kaltes Modell lädt seine Gewichte, bevor es den ersten Vektor gibt.
	// **Die zehnte kam mit N5 (#1571)**: der zentrale Bauweg in
	// `netzvorgaben.go`. Er trägt selbst keine Sperre — er baut den
	// Transport, und die Erlaubnis hängt der Aufrufer daran. Er steht
	// deshalb mit Grund in [ohneErlaubnis].
	if gesehen != 10 {
		t.Errorf("in internal/app stehen %d Bauzeilen für einen Netzklienten, "+
			"erwartet 10 — kommt eine dazu, gehört sie in diese Zahl und in "+
			"die Überlegung, ob sie eine Sperre braucht", gesehen)
	}
}

// bauzeile sagt, ob hier ein Netzklient entsteht.
//
// # Warum drei Formen und nicht eine
//
// **Weil N5 (#1571)
// eine zweite und dritte eingeführt hat**: `netz.NeuerKlientMit` nimmt den
// eingestellten Proxyweg entgegen, und `App.netzklientMit` reicht ihn aus den
// Einstellungen hinein. Dieser Wächter suchte nach `netz.NeuerKlient(` und
// fand danach **null** — er maß nichts mehr und wäre grün gewesen, hätte die
// Zusicherung darunter nicht auf einer Zahl bestanden.
//
// **Genau davor warnt §10**: die Reichweite einer Messung ist die Reichweite
// ihrer Aussage. Hier hat die Zahl den Wächter gerettet.
func bauzeile(zeile string) bool {
	for _, form := range []string{
		"netz.NeuerKlient(", "netz.NeuerKlientMit(",
		"a.netzklient(", "a.netzklientMit(",
	} {
		if strings.Contains(zeile, form) {
			return true
		}
	}

	return false
}

// funktionskopf erkennt den Beginn einer Funktion.
var funktionskopf = regexp.MustCompile(`^func (?:\([^)]*\) )?([A-Za-z_][A-Za-z0-9_]*)`)

// pruefeDatei sieht jede Bauzeile einer Datei an und gibt ihre Zahl.
func pruefeDatei(t *testing.T, pfad, name string) int {
	t.Helper()

	roh, err := os.ReadFile(pfad) //nolint:gosec // ein fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s lesen: %v", name, err)
	}

	zeilen := strings.Split(string(roh), "\n")
	funktion := ""
	gefunden := 0

	for nummer, zeile := range zeilen {
		if kopf := funktionskopf.FindStringSubmatch(zeile); kopf != nil {
			funktion = kopf[1]
		}

		// **Kommentare zählen nicht.** Diese Datei und `dienstsuchnaht.go`
		// nennen den Aufruf im Fließtext; ein Wächter, der seinen eigenen
		// Kommentar misst, misst sich selbst (§10).
		if strings.HasPrefix(strings.TrimSpace(zeile), "//") {
			continue
		}

		if !bauzeile(zeile) {
			continue
		}

		gefunden++

		if erlaubtOhne(t, name, funktion) {
			continue
		}

		if !traegtErlaubnis(zeilen, nummer) {
			t.Errorf("%s:%d — der Klient in %q trägt kein .MitErlaubnis(; "+
				"ein Schalter erreicht ihn dann nicht mehr, sobald die Naht "+
				"einmal gebaut ist", name, nummer+1, funktion)
		}
	}

	return gefunden
}

// erlaubtOhne sagt, ob eine Funktion aufgezählt ist.
func erlaubtOhne(t *testing.T, name, funktion string) bool {
	t.Helper()

	grund, aufgezaehlt := ohneErlaubnis[funktion]
	if !aufgezaehlt {
		return false
	}

	if len(grund) < 40 {
		t.Errorf("%s: %q steht in ohneErlaubnis, aber der Grund ist zu kurz "+
			"— eine Ausnahme ohne Begründung ist eine Lücke", name, funktion)
	}

	return true
}

// traegtErlaubnis sieht in der Anweisung nach, die hier beginnt.
//
// **Sechs Zeilen weit** — so lang ist die längste dieser Bauzeilen heute, und
// weiter zu sehen hieße, eine Erlaubnis aus einer anderen Anweisung zu
// akzeptieren.
//
// **Ohne den führenden Punkt gesucht**, und das ist keine Nachlässigkeit: die
// Kette ist umbrochen, der Punkt steht am **Ende** der Zeile davor, und
// `.MitErlaubnis(` kommt als Zeichenkette nirgends vor. Der erste Anlauf
// suchte danach und meldete beide Klienten als ungesichert — ein Instrument,
// das den Fehler erfindet, den es finden soll.
//
// **Kommentarzeilen fliegen dafür heraus**: sonst genügte es, den Namen im
// Fließtext zu erwähnen.
func traegtErlaubnis(zeilen []string, ab int) bool {
	bis := min(ab+6, len(zeilen))

	var rumpf []string

	for _, zeile := range zeilen[ab:bis] {
		if !strings.HasPrefix(strings.TrimSpace(zeile), "//") {
			rumpf = append(rumpf, zeile)
		}
	}

	return strings.Contains(strings.Join(rumpf, "\n"), "MitErlaubnis(")
}
