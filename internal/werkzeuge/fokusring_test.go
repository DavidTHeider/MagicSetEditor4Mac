// **Jedes Eingabefeld trägt seinen Fokusring — im Quelltext gezählt** (#828).
//
// # Warum hier und nicht im Frontend
//
// `barrierefreiheit.test.tsx` prüft den **gerenderten** Baum, und der zeigt
// nur, was eine Seite gerade hergibt: die Feldwahl der Statistik erscheint
// erst mit Daten, die Massenänderung erst mit einer Auswahl. **Beide hatten
// keinen Ring, und beide blieben grün** — gesehen hat es der Zweitblick über
// Phase O, indem er den Quelltext las.
//
// Diese Prüfung liest ihn auch. Sie ersetzt die andere nicht: dort geht es um
// den Baum, den ein Mensch vor sich hat, hier um jede Zeile, die je einen
// bekommen könnte.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// bedienelement findet den Anfang eines `input`, `select` oder `textarea`.
//
// **Ohne `button` und `a`.** Die tragen ihren Ring über `Knopf` und über
// eigene Klassensätze, und ein Muster über sie fände die vielen Fälle mit, in
// denen die Klasse aus einer Konstanten kommt — falscher Alarm, und ein
// Wächter, der falschen Alarm schlägt, ist schlechter als keiner.
var bedienelement = regexp.MustCompile(`<(input|select|textarea)\s`)

// attributeAb liest die Attribute eines Elements ab dem Anfang seines Namens.
//
// **Über die Klammertiefe und nicht über ein Muster.** Der erste Anlauf nahm
// `[^>]*?` bis zum nächsten `>` — und fand **drei** von fünfzig
// Bedienelementen: in JSX steht in fast jedem `onChange={(e) => …}`, und der
// Pfeil ist ein `>`. Gesehen hat es die Reichweitenprüfung unten, nicht das
// Lesen.
func attributeAb(inhalt string, ab int) string {
	tiefe := 0

	for i := ab; i < len(inhalt); i++ {
		switch inhalt[i] {
		case '{':
			tiefe++

		case '}':
			tiefe--

		case '>':
			if tiefe == 0 {
				return inhalt[ab:i]
			}
		}
	}

	return inhalt[ab:]
}

// hatRing sagt, ob ein Stück Quelltext eine Fokusangabe trägt.
var hatRing = regexp.MustCompile(`focus-visible:|focus:`)

// ringAmRahmen findet einen Ring, den ein **umgebendes** Element trägt.
//
// # Warum das gilt
//
// Ein zusammengesetztes Feld hat mehr als ein Element: das Quelltextfeld
// (#1164) legt eine Nummernspalte, eine Farbanzeige und die Eingabe
// übereinander. Ein Ring um die Eingabe allein ließe die Nummern draußen und
// säße innerhalb des Rahmens — er sähe aus wie ein Fehler. Der Ring gehört um
// das Ganze, und `focus-within:` ist genau dafür da.
//
// # Was diese Zeile nicht prüft
//
// **Dass das Element mit `focus-within:` dieses Feld wirklich umschließt.**
// Der Wächter liest Zeichen, keinen Baum. Wer beides in eine Datei schreibt,
// ohne dass eins im anderen liegt, kommt hier durch — dagegen steht die
// Prüfung am gerenderten Baum in `barrierefreiheit.test.tsx` und die
// Sichtprüfung (§3).
//
// # Deshalb gilt der Freibrief nur in einer Datei mit **einem** Bedienelement
//
// Der erste Anlauf suchte `focus-within:` **irgendwo** in der Datei. Damit gab
// ein einziger solcher Rahmen jedem ringlosen Feld darin frei — und der Satz
// darunter, ein Feld ohne umgebenden Ring falle weiterhin auf, war falsch.
// Angestrichen vom Zweitblick durch Codex über Phase 3.3.
//
// Steht nur **ein** Bedienelement in der Datei, fallen „irgendwo" und „um
// dieses herum" so weit zusammen, wie ein Zeichenmuster es sehen kann. Bei
// zweien kann es das nicht mehr, und dann zählt der Rahmen nicht.
//
// **Und `outline-none` bleibt Absicht.** Es steht dort, damit nicht zwei Ringe
// übereinanderliegen; ein Feld, das es ohne umgebenden Ring trägt, fällt in
// jeder Datei mit mehr als einem Bedienelement weiterhin auf.
//
// # Und `has-focus-visible:` zählt genauso
//
// Tailwind 4 kennt beides: `focus-within:` fragt `:focus-within`,
// `has-focus-visible:` fragt `:has(:focus-visible)`. Für einen Ring **um** ein
// Feld herum ist die zweite Form die genauere — sie geht nur an, wenn der
// Browser den Fokus für sichtbar hält, also bei Tastaturbedienung und nicht
// bei jedem Mausklick. Die Wahlkacheln einer Stiloption benutzen sie
// (#1225):
// ein Optionsfeld, das eine Maus anklickt, soll keinen Ring bekommen.
//
// **Für diesen Wächter ist es dieselbe Zusage** — ein umgebendes Element
// trägt den Ring —, und deshalb dieselbe Zeile und dieselbe Einschränkung auf
// **ein** Bedienelement je Datei.
var ringAmRahmen = regexp.MustCompile(
	`(focus-within|has-focus-visible):(outline|ring|border)`)

// klassenAusdruck findet `className={name}` — eine Klasse aus einer Größe.
var klassenAusdruck = regexp.MustCompile(
	`className=\{\s*([A-Za-zÄÖÜäöüß_][\wÄÖÜäöüß]*)`)

// ringUeberGroesse folgt einer Klassenangabe, die aus einer Größe kommt.
//
// **Der erste Anlauf zählte jede solche Angabe als Befund** — und meldete
// zwanzig, von denen **alle** falsch waren: `className={feld}` zeigt auf eine
// Konstante, die den Ring sehr wohl trägt. Ein Wächter, der falschen Alarm
// schlägt, ist schlechter als keiner (§10), und dieser hätte beim ersten Blick
// dazu geführt, dass man ihm nicht mehr glaubt.
//
// **Was er nicht auflöst, sagt er:** eine Größe, die aus einer anderen Datei
// kommt, findet er nicht — dort steht `unbekannt`, und der Aufrufer entscheidet.
// Heute gibt es keinen solchen Fall; käme einer, meldete sich der Zähler unten.
func ringUeberGroesse(inhalt, name string) (traegt, gefunden bool) {
	stelle := regexp.MustCompile(`(?:const|let)\s+` + regexp.QuoteMeta(name) + `\s*=`).
		FindStringIndex(inhalt)
	if stelle == nil {
		return false, false
	}

	// Bis zum Ende der Zuweisung: die nächste Zeile, die nicht eingerückt ist
	// oder mit einem Anführungszeichen beginnt, gehört nicht mehr dazu.
	rest := inhalt[stelle[1]:]
	if ende := strings.Index(rest, "\n\n"); ende > 0 {
		rest = rest[:ende]
	}

	return hatRing.MatchString(rest), true
}

func TestJedesEingabefeldTraegtSeinenFokusring(t *testing.T) {
	wurzel := filepath.Join("..", "..", "frontend", "src")

	var ohneRing []string

	gezaehlt, ungeloest, amRahmen := 0, 0, 0

	if err := filepath.WalkDir(wurzel, func(pfad string, eintrag os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if eintrag.IsDir() || !strings.HasSuffix(pfad, ".tsx") ||
			strings.HasSuffix(pfad, ".test.tsx") {
			return nil
		}

		inhalt, err := os.ReadFile(pfad) //nolint:gosec // Quelltext im Repository
		if err != nil {
			return err
		}

		text := string(inhalt)

		stellen := bedienelement.FindAllStringSubmatchIndex(text, -1)

		for _, stelle := range stellen {
			art := text[stelle[2]:stelle[3]]
			attribute := attributeAb(text, stelle[1])

			// **Ohne `className` ist es keine Aussage.** Ein Feld, dessen
			// Klassen von außen kommen, trägt seinen Ring dort — und wo das
			// nicht stimmt, findet es die Prüfung am gerenderten Baum.
			if !strings.Contains(attribute, "className") {
				continue
			}

			gezaehlt++

			if hatRing.MatchString(attribute) {
				continue
			}

			// **Einer Größe wird gefolgt.** `className={feld}` zeigt in fast
			// jedem Fall auf eine Konstante mit dem Ring darin.
			if treffer := klassenAusdruck.FindStringSubmatch(attribute); treffer != nil {
				traegt, gefunden := ringUeberGroesse(text, treffer[1])
				if traegt {
					continue
				}

				if !gefunden {
					ungeloest++

					continue
				}
			}

			// **Ein Ring am Rahmen zählt — bei einem einzigen Feld** (siehe
			// [ringAmRahmen]).
			if len(stellen) == 1 && ringAmRahmen.MatchString(text) {
				amRahmen++

				continue
			}

			ohneRing = append(ohneRing, filepath.Base(pfad)+" <"+art+">")
		}

		return nil
	}); err != nil {
		t.Fatalf("frontend/src nicht lesbar: %v", err)
	}

	// **Die Gegenprobe zur Reichweite** (§10): findet das Muster nichts, ist
	// der Lauf grün und sagt nichts.
	if gezaehlt < 10 {
		t.Fatalf("nur %d Bedienelemente gefunden — das Muster greift ins Leere", gezaehlt)
	}

	if len(ohneRing) > 0 {
		t.Errorf("%d Bedienelemente ohne Fokusring: %s",
			len(ohneRing), strings.Join(ohneRing, ", "))
	}

	// **Was nicht aufgelöst wurde, steht dabei** (§10). Eine Klassengröße aus
	// einer anderen Datei kann dieser Wächter nicht verfolgen; heute gibt es
	// keine, und wenn eine käme, sagt es diese Zahl.
	t.Logf("%d Bedienelemente geprüft, %d Klassengrößen nicht aufgelöst, "+
		"%d Ringe am Rahmen statt am Feld", gezaehlt, ungeloest, amRahmen)
}
