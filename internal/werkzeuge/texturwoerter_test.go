// Der Wächter über die Wörter der Spieltexturen — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Warum es ihn gibt
//
// **Weil es keine Bindung gibt, die die Wörter nennt.** Der erste Anlauf band
// eine `Texturgueten`, die dem Fenster die fünf Stufen aufzählt — und niemand
// rief sie: das Fenster weiß, welche Stufe es will, und schreibt sie hin. Eine
// Bindung ohne Aufrufer ist keine Anwendungsfunktion, sondern eine Zusage ohne
// Einlösung (§2), also ist sie gefallen.
//
// **Damit hängen die Wörter frei.** `guete="tisch"` in einer Komponente und
// `GueteTisch = "tisch"` in Go sind zwei Zeichenketten, die niemand
// aneinanderhält; wer die eine umbenennt, bekommt an der anderen keinen
// Fehler, sondern ein Bild, das nicht kommt. Also zählt hier jemand nach.
//
// # Was er nicht prüft
//
// **Ob die Stufe die richtige ist.** Dass der Tisch `tisch` will und nicht
// `miniatur`, ist eine Entscheidung der Seite; er sieht nur, dass das Wort
// eines ist, das es gibt.
//
// # Zwei Ausdrücke, und es waren einmal vier
//
// **Die Gestalt des Fensters hat sich geändert, und der Wächter wurde rot.**
// Beim ersten Anlauf nahm `Kartenbild` `seite="…"` und `guete="…"` als
// Eigenschaften, und der Haken drei Zeichenketten je Ruf; seit der Wunsch
// ein **Verbund** ist, gibt es beide Gestalten nicht mehr. Die Ausdrücke
// dafür sind gefallen, statt als tote Prüfung stehenzubleiben.
//
// **Dass das auffiel, ist die Selbstprüfung daneben** — sie verlangt von
// jedem Ausdruck, dass er etwas findet. Ohne sie wäre der Wächter still grün
// geblieben und hätte nichts mehr gehalten (§10).
//
// **Und er sieht keine gerechneten Wörter.** Ein `guete={stufe}` aus einer
// Variablen steht in keiner Zeichenkette — dann findet er nichts und sagt
// nichts. Solange die Seiten die Stufe hinschreiben, greift er; sobald eine
// sie ausrechnet, ist ihr Weg ungedeckt. Der Satz steht hier, damit das
// nächste `guete={…}` nicht für geprüft gehalten wird.
//
// # Und er liest nur den Texturweg
//
// **Weil `seite` in diesem Fenster zwei Dinge heißt.** Es ist die Seite einer
// Karte — und es ist die Seite, auf der jemand gerade steht: `seite: 'sets'`,
// `seite: 'turniere'`, `seite: 'konsole'`. Ein Wächter über alle Dateien
// strich beim ersten Lauf zehn solcher Wörter an, und keines davon war ein
// Befund. Gelesen werden deshalb die Dateien, die den Texturweg beim Namen
// nennen.
//
// **Der Preis steht hier**: eine neue Datei, die Kartenseiten nennt, ohne
// einen dieser Namen zu führen, sieht er nicht. Das ist der Tausch gegen zehn
// falsche Befunde, und er ist bewusst.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// texturweg sagt, ob eine Datei zum Texturweg gehört.
var texturweg = regexp.MustCompile(
	`Texturlauf|texturschluessel|Texturapi|Spieltextur|useSpieltexturen|` +
		`<Kartenbild`)

// texturfeld findet `seite:` und `guete:` als Feld eines Verbunds.
//
// **Bis zum Zeilenende und nicht bis zum ersten Anführungszeichen.** Der
// erste Anlauf verlangte die Zeichenkette **unmittelbar** hinter dem
// Doppelpunkt — und `seite: objekt.verdeckt ? 'zweite' : 'vorne'` steht so
// im Tisch und rutschte durch. Aufgefallen ist es an der Gegenprobe: ein
// erfundenes Wort dort ließ den Wächter grün (§10, 2026-09-15).
//
// **Und der Wert endet am nächsten Feld.** „Laut und falsch statt leise und
// falsch" stand hier als hinnehmbarer Preis — und wurde am selben Tag
// fällig: `{ …, seite: 'vorne', guete: 'tisch' }` gab `tisch` als
// **Kartenseite** ab, und der Wächter klagte über eine Güte. Ein Wächter,
// der bei richtigem Quelltext rot wird, wird abgeschaltet.
//
// **Der Wert endet am Komma**, und zwar im Ausdruck selbst.
//
// **Ein Schnitt daneben genügte nicht.** Der zweite Anlauf las die ganze
// Zeile und kürzte sie hinterher — und ließ damit das **zweite** Feld
// derselben Zeile ganz fallen: `[^\n]*` verschluckt es, und
// `FindAllStringSubmatch` sucht nur nicht überlappende Treffer. Eine falsche
// **Güte** neben einer richtigen Seite kam so durch; gemessen als grüne
// Gegenprobe am 2026-09-15.
//
// **Der Preis ist ein Wert mit Komma darin** — ein Ruf mit zwei Argumenten
// etwa. Der wird abgeschnitten, und was dahinter steht, sieht der Wächter
// nicht. Das ist die Richtung, in der ein Irrtum hier weh tut: er übersieht
// etwas, statt bei richtigem Quelltext rot zu werden.
var texturfeld = regexp.MustCompile(`\b(seite|guete):\s*([^,\n]*)`)

// texturvergleich findet `seite === '…'` und `guete !== '…'`.
//
// **Der vierte Ausdruck, und er kam mit einem Fall dazu, den die drei
// anderen nicht sahen**: `karteID === ” && seite !== 'ruecken'` in
// `useSpieltexturen` entscheidet, ob überhaupt geholt wird. Wer dort das
// Wort verschriebe, bekäme keinen Fehler, sondern einen Tisch ohne Rücken.
var texturvergleich = regexp.MustCompile(
	`\b(seite|guete)\s*[=!]==\s*('[^']*'|"[^"]*")`)

// zeichenkette zieht jedes Wort in Anführungszeichen heraus.
var zeichenkette = regexp.MustCompile(`'([^']*)'|"([^"]*)"`)

// **Jedes Wort, das das Fenster für eine Seite oder eine Stufe hinschreibt,
// gibt es in Go.**
func TestJedesTexturwortGibtEsInGo(t *testing.T) {
	seiten, gueten, _ := texturwoerter(t)

	for wort, orte := range seiten {
		// **Der Rücken ist ein gültiges Wort und keine Kartenseite.** Er
		// steht in jeder Antwort des Rückenwegs, und Läufe schreiben ihn
		// hin — abzuweisen wäre falsch, ihn zu einer Kartenseite zu machen
		// erst recht (#1459: „Kartenrückseite ist kein Synonym für die
		// zweite öffentliche Kartenseite").
		if wort == service.MarkeRuecken {
			continue
		}

		if !service.IstKartenseite(service.Kartenseite(wort)) {
			t.Errorf("%s nennt die Kartenseite %q — die gibt es in "+
				"`internal/service` nicht", strings.Join(orte, ", "), wort)
		}
	}

	for wort, orte := range gueten {
		if !service.IstTexturguete(service.Texturguete(wort)) {
			t.Errorf("%s nennt die Texturgüte %q — die gibt es in "+
				"`internal/service` nicht", strings.Join(orte, ", "), wort)
		}
	}
}

// **Jeder der drei Ausdrücke findet etwas.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil die Gegenprobe sonst grün bleibt.** Der erste Anlauf prüfte die
// **Summe** — mindestens zwei Seiten, mindestens zwei Stufen —, und als ein
// Ausdruck absichtlich zerstört wurde, deckten die beiden anderen ihn zu: der
// Wächter war grün, obwohl ein Drittel von ihm tot war. Eine Gegenprobe, die
// nichts ändert, ist keine (§10), und eine Summe, die drei Quellen zusammenzieht,
// macht genau das.
//
// Gezählt wird deshalb je Ausdruck. Wer einen von ihnen bricht, sieht ihn rot.
func TestJederTexturausdruckFindetEtwas(t *testing.T) {
	_, _, gezaehlt := texturwoerter(t)

	for _, name := range []string{"feld", "vergleich"} {
		if gezaehlt[name] == 0 {
			t.Errorf("der Ausdruck %q findet nichts — er ist tot, und was er "+
				"halten sollte, hält niemand", name)
		}
	}
}

// texturwoerter liest `frontend/src` und gibt, was dort als Seite und als Güte
// steht — je Wort die Dateien, die es nennen, dazu die Treffer je Ausdruck.
//
// **Läufe zählen mit.** Ein Test, der eine Stufe nennt, die es nicht gibt,
// prüft eine Welt, die es nicht gibt (§3) — und er ist grün, weil die
// Vorrichtung jede Zeichenkette annimmt.
func texturwoerter(t *testing.T) (
	seiten, gueten map[string][]string, gezaehlt map[string]int,
) {
	t.Helper()

	seiten = map[string][]string{}
	gueten = map[string][]string{}
	gezaehlt = map[string]int{}

	wurzel := filepath.Join("..", "..", "frontend", "src")

	fehler := filepath.WalkDir(wurzel, func(pfad string, eintrag os.DirEntry,
		err error,
	) error {
		if err != nil {
			return err
		}

		if eintrag.IsDir() || (!strings.HasSuffix(pfad, ".ts") &&
			!strings.HasSuffix(pfad, ".tsx")) {
			return nil
		}

		inhalt, err := os.ReadFile(pfad) //nolint:gosec // Quelltext im Repository
		if err != nil {
			return err
		}

		if !texturweg.Match(inhalt) {
			return nil
		}

		ort, err := filepath.Rel(wurzel, pfad)
		if err != nil {
			return err
		}

		sammle(string(inhalt), ort, seiten, gueten, gezaehlt)

		return nil
	})
	if fehler != nil {
		t.Fatalf("frontend/src nicht lesbar: %v", fehler)
	}

	return seiten, gueten, gezaehlt
}

// sammle trägt die Wörter einer Datei in die beiden Sammlungen ein.
func sammle(inhalt, ort string, seiten, gueten map[string][]string,
	gezaehlt map[string]int,
) {
	eintragen := func(wohin map[string][]string, wort string) {
		if wort == "" {
			return
		}

		for _, schon := range wohin[wort] {
			if schon == ort {
				return
			}
		}

		wohin[wort] = append(wohin[wort], ort)
	}

	fuer := func(name, roh string) {
		wohin := gueten
		if name == "seite" {
			wohin = seiten
		}

		for _, wort := range woerterIn(roh) {
			eintragen(wohin, wort)
		}
	}

	for _, treffer := range texturfeld.FindAllStringSubmatch(inhalt, -1) {
		gezaehlt["feld"]++

		fuer(treffer[1], treffer[2])
	}

	for _, treffer := range texturvergleich.FindAllStringSubmatch(inhalt, -1) {
		gezaehlt["vergleich"]++

		fuer(treffer[1], treffer[2])
	}
}

// woerterIn zieht jede Zeichenkette aus einem Ausdruck.
func woerterIn(roh string) []string {
	treffer := zeichenkette.FindAllStringSubmatch(roh, -1)
	gefunden := make([]string, 0, len(treffer))

	for _, treffer := range treffer {
		wort := treffer[1]
		if wort == "" {
			wort = treffer[2]
		}

		gefunden = append(gefunden, wort)
	}

	return gefunden
}
