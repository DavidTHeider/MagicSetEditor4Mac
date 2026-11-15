// **Wie groß die vier Flächen sind — gezählt, nicht behauptet**
// (#1107).
//
// # Warum es diesen Wächter gibt
//
// Der Einleitungssatz von `docs/FAEHIGKEITEN.md` nennt vier Zahlen, und keine
// davon hat je jemand nachgezählt. Am 2026-08-25 standen dort „69 Bindungen,
// 14 MCP-Werkzeuge, 8 Befehle auf der Kommandozeile und 8 Endpunkte im
// Dienst" — tatsächlich waren es 99, 32, 26 und 31. Der Satz begründet, **wozu
// der Katalog da ist**; mit falschen Zahlen begründet er nichts.
//
// Der Wächter daneben (`faehigkeiten_test.go`) hält die Übersichtstabelle und
// hat diesen Satz nie angesehen. Genau die Gestalt, vor der §10 warnt: **ein
// Wächter, der die halbe Zusage hält, macht die andere unsichtbar.**
//
// # Was er nicht hält
//
// Ob die Zahl das Richtige zählt. Er zählt **exportierte** Methoden auf
// `*Bindungen`, `Name:`-Zeilen in den MCP-Werkzeugen, `case`-Zweige der
// Kommandozeile und `HandleFunc`-Zeilen des Dienstes — dass eine Bindung eine
// Fähigkeit ist und nicht zwei, sagt der Katalog und nicht diese Messung.
//
// **Exportiert seit dem 2026-08-29** (#1287). Vorher zählte er jede Methode
// auf `*Bindungen` und kam auf eine mehr, als Wails erzeugt; der Unterschied
// und seine Gegenprobe stehen in `bindungszaehler_test.go`.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// flaeche ist eine der vier Flächen samt der Art, sie zu zählen.
type flaeche struct {
	// wort ist, wie der Satz die Fläche nennt — der Text **hinter** der Zahl.
	wort string

	// verzeichnis wird ganz gelesen; stehen stattdessen dateien da, nur die.
	//
	// **Für die Kommandozeile sind es benannte Dateien**, und das ist keine
	// Bequemlichkeit: `case "…"` steht in `cmd/mse` auch in den
	// Unterbefehlen von `karte`, `schluesselwort` und `scryfall`. Über das
	// Verzeichnis gezählt kämen neun zu viel heraus, und die Zahl behauptete
	// etwas über *Befehle*, das sie über *Zweige* erhoben hat (§10, „Die
	// Reichweite einer Messung ist die Reichweite ihrer Aussage").
	//
	// **Es sind zwei seit dem 2026-08-27** (#1165): die Verteilung liegt in
	// `lauf.go` (am Set) und `verteilungen.go` (an Besitz, Paketdatei und
	// fremdem Dienst). Eine Datei allein zählte danach 34 statt 49 — dieselbe
	// Falle noch einmal, nur andersherum.
	// **Und seit dem 2026-09-05 ist es weder das eine noch das andere**
	// (#1351, Werkstück 10): `dateienMit` liest alle Dateien eines
	// Verzeichnisses, **die eine Marke tragen**.
	//
	// **Der Anlass ist derselbe Fehler zum dritten Mal.** `lauf.go` ging über
	// die Zeilengrenze aus §2, die Befehle über Menschen zogen nach
	// `leutebefehle.go` — und die Messung sagte danach **101** bei 127
	// Befehlen. Zwei Namen in einer Liste hatten wieder nicht mitwandern
	// können.
	//
	// **Eine Liste von Dateinamen ist die falsche Frage.** Was einen
	// Verteiler ausmacht, ist nicht sein Name, sondern dass er `switch befehl`
	// schreibt — und diese Marke wandert mit. Die Unterbefehle von `karte`,
	// `schluesselwort` und `scryfall` tragen sie nicht: sie verzweigen über
	// `rest[0]` und `wort`.
	verzeichnis string
	dateien     []string
	dateienMit  string
	muster      *regexp.Regexp

	// zaehler ersetzt das Muster, wo ein Muster die Aussage nicht trägt.
	//
	// **Für die Bindungen tut es das nicht.** `^func (b *Bindungen) ` zählt
	// auch `waehlenUndOeffnen`, und Wails bindet nur Exportiertes — die Zahl
	// war um eins zu hoch, und weil sie gewacht war, sah es niemand
	// (#1287).
	// Der Ersatz liest den Syntaxbaum; er steht in `bindungszaehler_test.go`
	// samt seiner Gegenprobe.
	zaehler func(*testing.T) int
}

// dieVierFlaechen sind die vier Flächen samt der Art, jede zu zählen.
//
// **Sie stehen als Funktion da und nicht im Lauf**, seit der Wächter über die
// Zahlen im ganzen Repositorium sie ebenfalls braucht
// (`flaechenzahlen_test.go`, #1287). Zwei Abschriften derselben Messung wären
// zwei Messungen, und die eine veraltete still.
func dieVierFlaechen() []flaeche {
	return []flaeche{
		{
			wort: "Bindungen am Fenster",
			zaehler: func(t *testing.T) int {
				return len(bindungenAusVerzeichnis(t, true))
			},
		},
		{
			// **Und die Namen, die erst zur Laufzeit entstehen.**
			//
			// `bewegungswerkzeug` baut `schreiben.zuruecknehmen` und
			// `…wiederherstellen` aus einer Fallunterscheidung; ein Muster über
			// `Name: "…"` sieht sie nicht. Der Zähler sagte deshalb 60 bei 62
			// Werkzeugen — und die Zahl im Satz war wieder eine Behauptung.
			// Angestrichen vom Zweitblick über Phase 3.4.
			//
			// **Gezählt wird deshalb jede Zeichenkette, die wie ein
			// Werkzeugname aussieht**, und über eine Menge: derselbe Name
			// kommt in `bewegung.go` zweimal vor (einmal als Vorgabe, einmal
			// im Zweig), und zwei Fundstellen sind kein zweites Werkzeug.
			// **Und mit Ziffern im Namen** (#1291). `lesen.mse2orte` fiel
			// durch `[a-z.]+`, und der Zähler sagte 88 bei 89 Werkzeugen —
			// dasselbe Muster stand an zwei weiteren Stellen und ließ es
			// ebenso fallen. Ein Wächter, der ein Zeichen nicht kennt, hält
			// seine Zusage für alle Namen ohne dieses Zeichen (§10).
			//
			// **Und der Bindestrich seit dem 2026-08-31** (#1340). Er ist
			// derselbe Fall ein zweites Mal: `schreiben.symbol-nachzeichnen`
			// fiel durch, und der Zähler sagte 117 bei 118. Das Werkzeug heißt
			// heute ohne Bindestrich — **117 von 118 Namen tun das**, und eine
			// Ausnahme wäre keine Konvention gewesen. Die Zeichenklasse steht
			// trotzdem hier: dass heute kein Name ihn trägt, ist der Grund,
			// warum es beim nächsten Mal wieder niemand sähe.
			// **Und was er nicht hält** — Befund 6 des Zweitblicks über die
			// Phasen 4.0 bis 4.9, am 2026-09-03.
			//
			// Er zählt **Zeichenketten im Quelltext**, nicht angemeldete
			// Werkzeuge. Eine passende Zeichenkette in einer Erklärung, einer
			// Hilfsfunktion oder einer Namensliste erfüllt ihn, ohne dass das
			// Werkzeug über `tools/list` erreichbar wäre.
			//
			// **Die Zusage selbst hält jetzt
			// `internal/flaechen/werkzeugzahl_test.go`**: er fragt den
			// **gebauten** Server. Dort ist sie zu halten und nirgends sonst
			// — `mcp.Werkzeuge(mcp.Naehte{})` gibt mit leeren Nähten 80 statt
			// 153, weil ganze Gruppen an einer Naht hängen.
			//
			// **Dieser hier bleibt trotzdem**, und zwar als billige
			// Vorwarnung: er läuft ohne gebautes Programm und fällt schon
			// beim Umbenennen auf.
			wort: "MCP-Werkzeuge", verzeichnis: "../mcp",
			muster: regexp.MustCompile(`"(?:lesen|schreiben)\.[a-z0-9.-]+"`),
		},
		{
			// **Ohne `case "":`** — „kein Befehl genannt" ist keiner.
			//
			// **Über das ganze Verzeichnis und nicht über `lauf.go`** (#1165):
			// die Verteilung liegt seit dem 2026-08-27 in **vier** Funktionen
			// — die am Set in `lauf.go`, die an Besitz, Paketdatei und fremdem
			// Dienst in `verteilungen.go`. Eine Messung über eine Datei sagte
			// danach „34" bei 49 Befehlen, und die Zahl im Satz wäre wieder
			// eine Behauptung. **Die Reichweite einer Messung ist die
			// Reichweite ihrer Aussage** (§10).
			//
			// **Am 2026-08-28 ist es fast wieder passiert** (#1169):
			// `lauf.go` ging mit der Bestandssicherung über die Zeilengrenze
			// aus §2, und der erste Schnitt zog vier `case`-Zweige in eine
			// eigene Weiche — die Messung sagte danach „58" bei 61 Befehlen.
			// Geschnitten wurde stattdessen bei `zerlege`: **wie** eine Zeile
			// zerfällt, ist eine andere Frage als **welche** Befehle es gibt,
			// und diese Naht wandert nicht.
			// **Und am 2026-09-05 ist es zum dritten Mal passiert** (#1351,
			// Werkstück 10): `token-loeschen` hob `lauf.go` über die Grenze,
			// die Befehle über Menschen zogen nach `leutebefehle.go`, und die
			// Messung sagte **101** bei 127. Gezählt wird seitdem nicht mehr
			// über zwei Namen, sondern über **jede Datei in `cmd/mse`, die
			// `switch befehl` schreibt** — das ist, was einen Verteiler
			// ausmacht, und es wandert mit ihm.
			wort:        "Befehle auf der Kommandozeile",
			verzeichnis: "../../cmd/mse",
			dateienMit:  "switch befehl {",
			muster:      regexp.MustCompile(`(?m)^\tcase "[a-z]`),
		},
		{
			wort: "Endpunkte im Dienst", verzeichnis: "../rest",
			muster: regexp.MustCompile(`weg\.HandleFunc\(`),
		},
	}
}

// TestDieFlaechengroessenImSatzSindGezaehlt hält den Einleitungssatz.
func TestDieFlaechengroessenImSatzSindGezaehlt(t *testing.T) {
	t.Parallel()

	flaechen := dieVierFlaechen()

	satz := einleitungssatz(t)

	for _, was := range flaechen {
		t.Run(was.wort, func(t *testing.T) {
			t.Parallel()

			gezaehlt := flaechentreffer(t, was)

			genannt := flaechenzahlVor(t, satz, was.wort)
			if genannt != gezaehlt {
				t.Errorf("der Satz nennt %d %s — gezählt sind %d",
					genannt, was.wort, gezaehlt)
			}
		})
	}
}

// flaechentreffer zählt die Fundstellen eines Musters.
//
// **Ohne Testdateien.** Ein Double, das eine Bindung nachbildet, ist keine
// Bindung — und eine Messung, die ihn mitzählte, wüchse mit den Tests.
func flaechentreffer(t *testing.T, was flaeche) int {
	t.Helper()

	if was.zaehler != nil {
		return was.zaehler(t)
	}

	if len(was.dateien) > 0 {
		funde := 0
		for _, datei := range was.dateien {
			funde += len(was.muster.FindAllString(gelesen(t, datei), -1))
		}

		return funde
	}

	eintraege, err := os.ReadDir(was.verzeichnis)
	if err != nil {
		t.Fatalf("%s lesen: %v", was.verzeichnis, err)
	}

	// **Über eine Menge**: ein Werkzeugname, der zweimal angemeldet wäre, ist
	// ein Fehler und keine zwei Werkzeuge.
	gesehen := map[string]bool{}
	funde := 0

	for _, eintrag := range eintraege {
		name := eintrag.Name()
		if eintrag.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") {
			continue
		}

		inhalt := gelesen(t, filepath.Join(was.verzeichnis, name))

		// **Und die Marke, wenn eine verlangt ist** — siehe [flaeche.dateienMit].
		if was.dateienMit != "" && !strings.Contains(inhalt, was.dateienMit) {
			continue
		}

		for _, fund := range was.muster.FindAllString(inhalt, -1) {
			funde++
			gesehen[fund] = true
		}
	}

	// **Nur die MCP-Werkzeuge tragen ihren Namen im Fund.** Bei den Bindungen
	// ist `func (b *Bindungen) ` für jede dieselbe Zeichenkette; über eine
	// Menge gezählt wären es eine.
	if strings.Contains(was.wort, "MCP") {
		return len(gesehen)
	}

	return funde
}

// einleitungssatz gibt den Absatz, der die vier Zahlen nennt.
func einleitungssatz(t *testing.T) string {
	t.Helper()

	roh, err := os.ReadFile(katalogpfad) //nolint:gosec // ein fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s lesen: %v", katalogpfad, err)
	}

	text := strings.ReplaceAll(string(roh), "\n", " ")

	anfang := strings.Index(text, "es gibt vier Flächen")
	if anfang < 0 {
		t.Fatal("der Einleitungssatz nennt die vier Flächen nicht mehr")
	}

	return text[anfang:min(anfang+400, len(text))]
}

// zahlVor liest die Zahl, die vor einem Wort steht.
func flaechenzahlVor(t *testing.T, satz, wort string) int {
	t.Helper()

	muster := regexp.MustCompile(`(\d+) ` + regexp.QuoteMeta(wort))

	fund := muster.FindStringSubmatch(satz)
	if fund == nil {
		t.Fatalf("der Satz nennt keine Zahl vor %q: %s", wort, satz)
	}

	zahl, err := strconv.Atoi(fund[1])
	if err != nil {
		t.Fatalf("%q ist keine Zahl: %v", fund[1], err)
	}

	return zahl
}

// gelesen gibt den Inhalt einer Quelldatei dieses Repositoriums.
func gelesen(t *testing.T, pfad string) string {
	t.Helper()

	roh, err := os.ReadFile(pfad) //nolint:gosec // Quelldateien dieses Repositoriums
	if err != nil {
		t.Fatalf("%s lesen: %v", pfad, err)
	}

	return string(roh)
}
