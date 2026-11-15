// Der Wächter gegen den Wortlaut der Regelwerke im Repositorium.
//
// # Wofür
//
// **§11 sagt: nachgelesen wird, nichts wird übernommen.** Die Comprehensive
// Rules und die Magic Tournament Rules sind urheberrechtlich geschützt und
// haben keine Weitergabelizenz; zitiert wird die **Nummer** samt Fassung, nie
// der Satz.
//
// **Die Regel stand seit dem 2026-09-15 da und hat nicht getragen.** Am
// 2026-09-16 standen **vierzehn** abgeschriebene Sätze im Baum, alle älter als
// die Regel und keiner je wieder angesehen — dazu einer in einer Datei, deren
// eigener Kopf „Kein Wortlaut, keine Kopie" sagt. Eine Regel, an die sich
// niemand hält, ist keine Regel, sondern ein Vorwurf, den man später gegen
// jeden erheben kann (§9). **Deshalb ist die Vorgabe jetzt sicher, statt
// wiederholt zu werden** — dieselbe Wendung wie beim Anmeldebund (§3).
//
// # Was er prüft
//
// **Eine Nummer, die unmittelbar in ein Zitat führt.** `CR 205.4c: „Any land
// with the supertype …"` ist die Gestalt, die entsteht, wenn jemand
// abschreibt. Ein Zitat weiter unten im selben Absatz fängt er nicht — dafür
// müsste er verstehen, wovon der Absatz handelt.
//
// # Was er ausdrücklich durchlässt
//
// **Kurze Namen.** `CR 205.4c nennt den Supertyp „basic"` ist ein Begriff und
// kein Satz; Kartennamen wie „Dryad Arbor" ebenso. Die Grenze liegt bei
// [hoechstensWorte] Wörtern und ist eine Setzung — sie trennt den **Begriff**
// vom **Satz**, und ein Begriff ist nicht geschützt.
//
// **Und seine eigene Datei.** Sie muss die Gestalt enthalten, gegen die sie
// prüft; sonst hätte sie keine Gegenprobe (§10).
package werkzeuge_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// hoechstensWorte ist die Grenze zwischen Begriff und Satz.
const hoechstensWorte = 4

// englischeWoerter sind Wörter, die im Deutschen keine sind.
//
// **Wozu sie da sind**: geschützt ist der **Wortlaut** der Regelwerke, und der
// ist englisch. Eine **deutsche** Wiedergabe ist genau das, was §11 verlangt —
// sie darf in Anführungszeichen stehen, ohne ein Befund zu sein.
//
// **Die Liste ist bewusst kurz und bewusst eindeutig.** Wörter, die es in
// beiden Sprachen gibt — `in`, `an`, `all`, `so`, `was`, `man` —, stehen nicht
// darin: sie machten aus einem deutschen Satz einen englischen.
var englischeWoerter = map[string]bool{
	"the": true, "of": true, "and": true, "or": true, "to": true,
	"on": true, "at": true, "by": true, "is": true, "are": true,
	"were": true, "be": true, "been": true, "being": true, "have": true,
	"has": true, "had": true, "does": true, "did": true, "do": true,
	"not": true, "no": true, "any": true, "each": true, "more": true,
	"than": true, "that": true, "this": true, "these": true, "those": true,
	"with": true, "for": true, "from": true, "its": true, "their": true,
	"them": true, "they": true, "it": true, "as": true, "if": true,
	"when": true, "only": true, "may": true, "must": true, "can": true,
	"cannot": true, "other": true, "same": true, "used": true, "use": true,
	"rule": true, "rules": true, "card": true, "cards": true, "deck": true,
	"decks": true, "player": true, "players": true, "game": true,
	"games": true, "number": true, "names": true, "but": true,
	"which": true, "such": true, "unless": true, "otherwise": true,
	"applies": true, "contain": true, "printed": true, "before": true,
	"after": true,
}

// mindestensEnglisch ist, wie viele davon ein Zitat englisch machen.
const mindestensEnglisch = 2

// regelnummer findet einen Verweis auf eines der beiden Regelwerke.
//
// **Auch den Namen und nicht nur die Nummer**: `*Magic: The Gathering
// Comprehensive Rules*, „effective as of …"` stand in
// `internal/domain/standardland.go` und trug keine Nummer — die erste Fassung
// dieses Wächters ließ es durch.
var regelnummer = regexp.MustCompile(
	`\b(?:CR \d|MTR\b|Comprehensive Rules|Tournament Rules)`)

// vorlauf ist, wie weit vor einem Zitat ein Verweis stehen darf.
//
// **Hundertfünfzig Zeichen, und das ist gemessen und nicht geraten**: die
// erste Fassung ließ vierzig zu und **auf derselben Zeile** — sie fand 37
// Stellen und ließ **sechs** durch, weil die Nummer eine Zeile höher stand
// oder ein Halbsatz dazwischen. Ein Wächter, der die halbe Zusage hält, macht
// die andere unsichtbar (§10).
const vorlauf = 150

// zitatenden sind die schließenden Anführungszeichen, die hier vorkommen.
//
// **Zwei und nicht eines**: geschrieben wird `„…"` mit deutschem Anfang und
// geradem Schluss, gelegentlich aber auch mit typografischem. Wer nur eines
// kennte, ließe die Hälfte durch — und eine Zusage, die zur Hälfte hält, ist
// schlimmer als keine (§10).
var zitatenden = []string{"\u201c", "\""}

func TestKeinRegelwortlautImRepositorium(t *testing.T) {
	befunde := []string{}

	for _, pfad := range quellenUndTexte(t) {
		if strings.HasSuffix(pfad, "regelwortlaut_test.go") {
			continue
		}

		for _, zitat := range zitateNachRegelnummern(gelesen(t, pfad)) {
			befunde = append(befunde,
				pfad+": "+zitatzeile(zitat))
		}
	}

	if len(befunde) > 0 {
		t.Errorf("%s — §11 verlangt die Nummer, nicht den Satz:\n%s",
			abgeschrieben(len(befunde)), strings.Join(befunde, "\n"))
	}
}

// abgeschrieben benennt die Zahl der Funde in lesbarem Deutsch.
func abgeschrieben(anzahl int) string {
	if anzahl == 1 {
		return "ein abgeschriebener Regelsatz"
	}

	return fmt.Sprintf("%d abgeschriebene Regelsätze", anzahl)
}

// TestDerWortlautwaechterGreift ist die Gegenprobe (§10).
func TestDerWortlautwaechterGreift(t *testing.T) {
	faelle := []struct {
		name   string
		text   string
		findet int
	}{
		{
			"ein abgeschriebener Satz",
			"// CR 100.4a: „In constructed play, a sideboard may " +
				"contain no more than\n// fifteen cards.“",
			1,
		},
		{
			"ein Begriff ist kein Satz",
			"// CR 205.4c nennt den Supertyp „basic“.",
			0,
		},
		{
			"eine Beschreibung mit Nummer",
			"// **Eine Bibliothek ist eine geordnete Zone** — CR 401.1.",
			0,
		},
		{
			"eine deutsche Wiedergabe ist kein Wortlaut",
			"// CR 903.5e: die Regel sagt „Commanderpartien kennen kein " +
				"Sideboard“.",
			0,
		},
		{
			"ein Zitat ohne Regelnummer",
			"// Das Issue sagt: „Ersten automatischen Teilumfang " +
				"konkret festschreiben.“",
			0,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if gefunden := len(zitateNachRegelnummern(fall.text)); gefunden !=
				fall.findet {
				t.Fatalf("erwartet %d Befunde, gefunden %d",
					fall.findet, gefunden)
			}
		})
	}
}

// zitateNachRegelnummern gibt jedes Zitat, das unmittelbar auf eine
// Regelnummer folgt und länger ist als ein Begriff.
func zitateNachRegelnummern(text string) []string {
	gefunden := []string{}

	for _, stelle := range alleAnfuehrungen(text) {
		if !regelnummer.MatchString(text[max(0, stelle-vorlauf):stelle]) {
			continue
		}

		rest := text[stelle+len("„"):]

		schluss := erstesEnde(rest)
		if schluss < 0 {
			continue
		}

		zitat := rest[:schluss]
		if worte(zitat) > hoechstensWorte && istEnglisch(zitat) {
			gefunden = append(gefunden, zitat)
		}
	}

	return gefunden
}

// erstesEnde gibt die Stelle des ersten schließenden Anführungszeichens.
func erstesEnde(rest string) int {
	frueheste := -1

	for _, ende := range zitatenden {
		stelle := strings.Index(rest, ende)
		if stelle < 0 {
			continue
		}

		if frueheste < 0 || stelle < frueheste {
			frueheste = stelle
		}
	}

	return frueheste
}

// istEnglisch sagt, ob ein Zitat die geschützte Sprache trägt.
func istEnglisch(zitat string) bool {
	gezaehlt := 0

	for _, stueck := range strings.Fields(strings.ToLower(zitat)) {
		if englischeWoerter[strings.Trim(stueck, ".,;:()*„“\"'\u2018\u2019")] {
			gezaehlt++
		}
	}

	return gezaehlt >= mindestensEnglisch
}

// alleAnfuehrungen gibt jede Stelle, an der ein Zitat aufgeht.
func alleAnfuehrungen(text string) []int {
	stellen := []int{}

	for stelle := 0; ; {
		weiter := strings.Index(text[stelle:], "„")
		if weiter < 0 {
			return stellen
		}

		stellen = append(stellen, stelle+weiter)
		stelle += weiter + len("„")
	}
}

// worte zählt, was in einem Zitat ein Wort ist — Kommentarstriche nicht.
func worte(zitat string) int {
	gezaehlt := 0

	for _, stueck := range strings.Fields(zitat) {
		if strings.Trim(stueck, "/*-— \t") != "" {
			gezaehlt++
		}
	}

	return gezaehlt
}

// zitatzeile macht aus einem Zitat eine Fundstelle, die in eine Zeile passt.
func zitatzeile(zitat string) string {
	eine := strings.Join(strings.Fields(zitat), " ")
	if len(eine) > 70 {
		eine = eine[:70] + " …"
	}

	return eine
}

// quellenUndTexte gibt jede Datei, in der ein Zitat stehen könnte.
func quellenUndTexte(t *testing.T) []string {
	t.Helper()

	pfade := []string{}

	for _, wurzel := range []string{
		filepath.Join("..", ".."),
	} {
		fehler := filepath.WalkDir(wurzel, func(
			pfad string, eintrag os.DirEntry, fehler error,
		) error {
			if fehler != nil {
				return fehler
			}

			if eintrag.IsDir() {
				return uebergeheVerzeichnis(eintrag.Name())
			}

			// **Ein Verweis wird nicht gelesen** — C06 (#2231): `.go` und
			// `frontend/node_modules` sind in einem Arbeitsbaum Verweise auf
			// den Hauptbaum, und `uebergeheVerzeichnis` sieht nur Verzeichnisse.
			if !eintrag.Type().IsRegular() {
				return nil
			}

			if zaehltFuerWortlaut(pfad) {
				pfade = append(pfade, pfad)
			}

			return nil
		})
		if fehler != nil {
			t.Fatalf("den Baum ablaufen: %v", fehler)
		}
	}

	return pfade
}

// uebergeheVerzeichnis hält den Lauf aus dem heraus, was nicht uns gehört.
func uebergeheVerzeichnis(name string) error {
	switch name {
	case "node_modules", ".git", ".go", "build", "dist", "wailsjs", "graft":
		return filepath.SkipDir
	}

	return nil
}

// zaehltFuerWortlaut sagt, ob eine Datei geprüft wird.
func zaehltFuerWortlaut(pfad string) bool {
	for _, endung := range []string{".go", ".md", ".ts", ".tsx"} {
		if strings.HasSuffix(pfad, endung) {
			return true
		}
	}

	return false
}
