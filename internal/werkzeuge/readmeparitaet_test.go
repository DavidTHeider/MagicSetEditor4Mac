// **Die Paritätszahlen im README sind gezählt** —
// #1287.
//
// # Der Befund
//
// Der zweite Absatz des README sagte am 2026-08-29 „115 vermessene Einheiten,
// davon 69 portiert und 28 angepasst — und acht, die noch fehlen". Die
// Tabellen von `docs/PARITAET.md` führten **116** Zeilen und **neun**
// `deferred`.
//
// Beide Nachbarn waren grün: `TestDieParitaetszaehlungStimmt` hält die
// Zählung **in** der Matrix, `TestDieZahlenImFliesstextStimmenAuch` den
// Fließtext **in** derselben Datei. Das README steht in keinem von beiden —
// und es ist die Datei, die ein Fremder zuerst liest. Genau die Gestalt aus
// §10: **ein Wächter, der die halbe Zusage hält, macht die andere
// unsichtbar.**
//
// # Was dieser Wächter hält
//
//   - Jede Zahl des Satzes steht für eine Marke der Matrix und stimmt mit
//     deren Zählung überein — Ziffer wie Zahlwort.
//   - **Das genannte Beispiel trägt die Marke, die der Satz behauptet.** „…
//     darunter das **Symbolfeld einer Karte**" ist nur so lange wahr, wie es
//     eine `deferred`-Zeile gibt, deren **benennende Aussage** diesen
//     Wortlaut trägt. Wird das Symbolfeld gebaut, ist der Satz rot — und nicht
//     bloß die Zahl daneben.
//
// **Die erste fette Aussage, nicht die ganze Zelle.** Der erste Anlauf suchte
// mit `strings.Contains` über die Zeile — und die `deferred`-Zeile zu `symbol`
// sagt in ihrem Erklärtext ausdrücklich „**Nicht zu verwechseln mit dem
// Setsymbol**: das ist gebaut". „darunter das **Setsymbol**" war damit grün,
// obwohl dieselbe Zeile das Gegenteil behauptet. Angestrichen vom Zweitblick
// über Phase 4.0 (Befund 4): eine Verneinung ist kein Beleg.
//
// # Was er nicht hält
//
// **Ob die Marke stimmt.** Dass eine Zeile `deferred` trägt und nicht
// `ported`, ist eine Aussage über Verhalten; das prüft ein Test, kein Wächter
// über eine Tabelle. Derselbe Satz steht über der Matrix selbst.
//
// **Und nicht den Rest des README.** Er greift genau an diesem einen Absatz.
package werkzeuge

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// paritaetszahlImReadme ordnet einer Formulierung ihre Marke zu.
type paritaetszahlImReadme struct {
	// muster fängt die Zahl — als Ziffer oder als Wort.
	muster *regexp.Regexp

	// marke ist die Marke aus `docs/PARITAET.md`, oder leer für die Gesamtzahl.
	marke string
}

// readmeParitaetszahlen sind die vier Angaben des Absatzes.
var readmeParitaetszahlen = []paritaetszahlImReadme{
	{regexp.MustCompile(`(\S+) vermessene Einheiten`), ""},
	{regexp.MustCompile(`(\S+) portiert`), "ported"},
	{regexp.MustCompile(`(\S+) angepasst`), "adapted"},
	{regexp.MustCompile(`(\S+) vertagte`), "deferred"},
}

// readmeBeispiel fängt „darunter das **Symbolfeld einer Karte**".
var readmeBeispiel = regexp.MustCompile(`darunter das \*\*([^*]+)\*\*`)

// TestDieParitaetszahlenImReadmeSindGezaehlt hält den Absatz.
func TestDieParitaetszahlenImReadmeSindGezaehlt(t *testing.T) {
	t.Parallel()

	zeilen, zaehlung := paritaetLesen(t)

	absatz := readmeParitaetsabsatz(t)

	for _, angabe := range readmeParitaetszahlen {
		fund := angabe.muster.FindStringSubmatch(absatz)
		if fund == nil {
			t.Errorf("der Absatz nennt nichts für %q — er ist umgeschrieben "+
				"worden, und der Wächter greift daneben", angabe.muster)

			continue
		}

		genannt, lesbar := zahlOderZahlwort(fund[1])
		if !lesbar {
			t.Errorf("%q ist weder Ziffer noch Zahlwort bis zwölf", fund[1])

			continue
		}

		erwartet := len(zeilen)
		if angabe.marke != "" {
			erwartet = zaehlung[angabe.marke]
		}

		if genannt != erwartet {
			t.Errorf("das README nennt %q (%d) für %q — die Matrix zählt %d",
				fund[1], genannt, angabe.muster, erwartet)
		}
	}

	pruefeBeispiel(t, absatz, zeilen)
}

// pruefeBeispiel hält den benannten Fall gegen seine Marke.
func pruefeBeispiel(t *testing.T, absatz string, zeilen []paritaetszeile) {
	t.Helper()

	// **Ohne vertagte Zeile gibt es kein Beispiel zu nennen** — und dann
	// wäre die Forderung nach einem falsch.
	//
	// **Hier stand sie unbedingt**, und das war eine unausgesprochene
	// Annahme: der Wächter setzte voraus, dass die Matrix immer mindestens
	// eine `deferred`-Zeile führt. Seit Werkstück D der Phase 4.9
	// (#1324)
	// führt sie keine, und der Wächter konnte gar nicht mehr grün werden.
	//
	// **Umgekehrt gilt er weiter**: gibt es keine vertagte Zeile, darf der
	// Absatz auch keine als Beispiel nennen.
	vertagte := 0

	for _, zeile := range zeilen {
		if zeile.marke == "deferred" {
			vertagte++
		}
	}

	fund := readmeBeispiel.FindStringSubmatch(absatz)

	if vertagte == 0 {
		if fund != nil {
			t.Errorf("der Absatz nennt %q als vertagt — die Matrix führt "+
				"keine einzige vertagte Zeile", fund[1])
		}

		return
	}

	if fund == nil {
		t.Fatal("der Absatz nennt kein Beispiel mehr — dann fehlt ihm das " +
			"Anschauliche, und dieser Wächter greift daneben")
	}

	for _, zeile := range zeilen {
		if zeile.marke != "deferred" {
			continue
		}

		if alsGanzesWort(fund[1]).MatchString(ersteFetteAussage(zeile)) {
			return
		}
	}

	t.Errorf("das README nennt %q als vertagt — in der Matrix führt keine "+
		"»deferred«-Zeile diesen Wortlaut in ihrer benennenden Aussage — "+
		"dem fetten Satz, mit dem ihre Beschreibung beginnt. "+
		"Entweder ist es gebaut, oder der Satz braucht ein anderes Beispiel",
		fund[1])
}

// alsGanzesWort baut das Muster für ein Beispiel an Wortgrenzen.
//
// **Ohne sie belegt jedes Teilwort sich selbst.** „darunter das **Feld**" war
// grün, weil „Feld" in „Symbolfeld einer Karte" steckt — ein Beispiel, das
// keine Einheit benennt, sah aus wie eines, das eine benennt (Nachlauf über
// Phase 4.0, Befund 4).
func alsGanzesWort(beispiel string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(beispiel) + `\b`)
}

// benennendeAussage fängt `**…**` **am Anfang** einer Zelle.
var benennendeAussage = regexp.MustCompile(`^\*\*([^*]+)\*\*`)

// ersteFetteAussage gibt den Satz, mit dem eine Zeile sich selbst benennt.
//
// **Er steht am Anfang ihrer Beschreibung, und zwar wörtlich am Anfang.** Das
// ist keine Konvention, die dieser Wächter erfindet: die Matrix schreibt so,
// weil eine Zeile in einer Tabelle sonst nicht überflogen werden kann. Was
// danach kommt, sind Gründe, Einschränkungen und Abgrenzungen — und eine
// Abgrenzung nennt gerade das, was **nicht** gemeint ist.
//
// **„Am Anfang" seit dem 2026-08-29.** Vorher genügte der **erste** fette
// Abschnitt irgendwo in der Zeile, und die `deferred`-Zeile `Symbolwerte`
// schreibt „siehe Bereich 2 — das **Feld**, nicht die Setzeichnung". Damit war
// „darunter das **Feld**" ein gültiges Beispiel, obwohl es nichts benennt und
// obwohl der Satz daneben schon wieder eine Verneinung ist — dieselbe Klasse
// wie beim `Setsymbol`, nur eine Zeile weiter. Gefunden bei der Gegenprobe zum
// Nachlauf über Phase 4.0: **sie blieb grün, und das war der Befund.**
//
// Eine Zeile ohne benennende Aussage kann kein Beispiel liefern — und das ist
// richtig so: sie benennt sich nicht selbst.
func ersteFetteAussage(zeile paritaetszeile) string {
	if len(zeile.spalten) == 0 {
		return ""
	}

	letzte := strings.TrimSpace(zeile.spalten[len(zeile.spalten)-1])

	fund := benennendeAussage.FindStringSubmatch(letzte)
	if fund == nil {
		return ""
	}

	return fund[1]
}

// readmeParitaetsabsatz gibt den Absatz, der die Deckung beziffert.
func readmeParitaetsabsatz(t *testing.T) string {
	t.Helper()

	roh := gelesen(t, filepath.Join(wurzel, "README.md"))

	anfang := strings.Index(roh, "Wie weit die Deckung reicht")
	if anfang < 0 {
		t.Fatal("das README beziffert die Deckung nicht mehr")
	}

	ende := strings.Index(roh[anfang:], "\n\n")
	if ende < 0 {
		t.Fatal("der Absatz hört nicht auf — der Leser greift daneben")
	}

	return strings.ReplaceAll(roh[anfang:anfang+ende], "\n", " ")
}

// zahlOderZahlwort liest „9" wie „neun".
//
// **Beides, weil der Satz beides schreibt.** „116 vermessene Einheiten" ist
// eine Ziffer, „neun vertagte" ein Wort — und ein Wächter, der nur Ziffern
// läse, zwänge den Absatz in eine Schreibweise, die niemand lesen will.
func zahlOderZahlwort(roh string) (int, bool) {
	nackt := strings.Trim(roh, "*.,;:—")

	if zahl, err := strconv.Atoi(nackt); err == nil {
		return zahl, true
	}

	zahl, ist := zahlwoerter[strings.ToLower(nackt)]

	return zahl, ist
}
