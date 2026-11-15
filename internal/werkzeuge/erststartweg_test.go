// Der Wächter über den Erststart-Weg.
//
// Er steht an **drei** Orten: als Zettel im `.dmg` (`scripts/releasezettel.sh`), in
// `README.md` und in `docs/RELEASE.md`. Dieselbe Zusage an drei Orten läuft
// auseinander — das ist Punkt 4 aus „Was Phase 1 gelehrt hat", und beim
// Nachstellen für #62
// war sie schon auseinandergelaufen: `docs/RELEASE.md` nannte den Grund „von
// einem nicht verifizierten Entwickler", und das ist der Wortlaut vor macOS 15.
// Was macOS 26.6.1 wirklich zeigt, steht in `docs/bilder/`.
//
// # Was er nicht prüft
//
// **Ob der Wortlaut heute noch stimmt.** Er hält drei Dateien gegeneinander
// und gegen einen Satz, den jemand am 2026-08-16 abgelesen hat. Ändert Apple
// die Meldung, bleibt der Wächter grün und die Anleitung wird falsch — dagegen
// hilft nur, den Weg wieder nachzustellen (`docs/RELEASE.md`, Abschnitt
// „Nachgestellt, nicht angenommen").
//
// **Den letzten Schritt.** Was nach „Dennoch öffnen" erscheint, ist nicht
// nachgemessen; warum, steht in `docs/RELEASE.md`. Ein Wächter, der die halbe
// Zusage hält, macht die andere unsichtbar — also wird sie ausgesprochen.
//
// **Die Reihenfolge und die Vollständigkeit.** Er sucht Wortlaute und
// Stationen, wo immer sie stehen. Eine Anleitung, die die Schritte 3 und 5
// vertauschte oder einen ganz ausließe, ohne die Station zu verlieren, käme
// durch. Ein Wächter über die Reihenfolge müsste die drei Dateien in eine Form
// zwingen, die sie nicht teilen — der Zettel im `.dmg` ist reiner Text, die
// beiden anderen sind Markdown mit Bildern dazwischen.
//
// **Was in den Bildern steht.** Geprüft wird, dass es PNG-Dateien mit
// plausibler Größe sind und dass beide Dokumente sie einbinden. Ob darauf das
// Richtige zu sehen ist, sieht nur jemand, der hinsieht.
//
// **Das gebaute `.dmg`.** Der Zettel wird aus `releasezettel.sh` gelesen, nicht aus
// einem Abbild. Dass `hdiutil` ihn wirklich hineinlegt, sagt nur ein Bau — und
// ein Bau läuft nicht in der Pipeline ([0007](../../docs/entscheidungen/0007-kein-entfernter-lauf.md)).
//
// Die drei letzten Absätze hat der Zweitblick über Phase 8 angestrichen: sie
// fehlten, und damit sah der Wächter stärker aus, als er ist.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// abgelesen sind die Sätze, die auf dem Bildschirm standen.
//
// Jeder ist ein Bildschirmfoto in `docs/bilder/`, kein Gedächtnis: der
// Blockierhinweis und seine Begründung aus dem Warnfenster, der Hinweis aus
// *Datenschutz & Sicherheit* und die beiden Knöpfe.
//
// **Der gefährliche Knopf steht mit dabei.** „In den Papierkorb legen" ist im
// Warnfenster der obere und breitere — eine Anleitung, die ihn nicht nennt,
// überlässt es dem Zufall, welchen jemand drückt.
// **Die Stationen des Weges stehen mit dabei.** Bis zum Zweitblick über Phase 8
// prüfte diese Liste nur die Sätze aus den Fenstern; wer „nach Programme
// ziehen" oder „Datenschutz & Sicherheit" herausstrich, blieb grün. Ein Weg
// ohne seine Stationen ist keine Anleitung, sondern eine Zitatsammlung.
var abgelesen = []string{
	"Programme",
	"nicht geöffnet",
	"Apple konnte nicht überprüfen",
	"Datenschutz & Sicherheit",
	"Sicherheit",
	"wurde blockiert, um deinen Mac zu schützen",
	"In den Papierkorb legen",
	"Fertig",
	"Dennoch öffnen",
}

// veraltet ist der Wortlaut aus der Zeit vor macOS 15.
//
// Er hat sich in `docs/RELEASE.md` gehalten, weil er plausibel klingt: Apple
// hat ihn jahrelang so gezeigt. Genau deshalb steht er hier — eine falsche
// Begründung erkennt niemand beim Lesen, sondern erst vor dem Fenster.
var veraltet = []string{
	"nicht verifizierte",
	"unbekannten Entwickler",
}

// nummerierterSchritt findet eine Zeile, die als Schritt einer Anleitung
// gelesen wird — in allen drei Dateien dieselbe Form.
var nummerierterSchritt = regexp.MustCompile(`(?m)^\s*\d+\.\s.*$`)

// zwischenraum fasst jede Folge aus Leerzeichen und Umbrüchen zusammen.
//
// **Sonst prüfte der Wächter den Zeilenumbruch mit.** Alle drei Dateien brechen
// bei 80 Zeichen um, und „Apple konnte nicht überprüfen" fällt dabei
// auseinander. Ein Wächter, der deshalb rot wird, erzwingt eine Zeile ohne
// Umbruch — er hielte die Form statt der Zusage.
var zwischenraum = regexp.MustCompile(`\s+`)

// befundeErststartweg prüft einen Text gegen den nachgemessenen Weg.
//
// Gibt die Befunde zurück, nicht einen Wahrheitswert: welcher Satz fehlt, ist
// die Auskunft, die jemandem beim Beheben hilft.
func befundeErststartweg(text string) []string {
	var befunde []string

	fortlaufend := zwischenraum.ReplaceAllString(text, " ")

	for _, satz := range abgelesen {
		if !strings.Contains(fortlaufend, satz) {
			befunde = append(befunde, "es fehlt der abgelesene Wortlaut: "+satz)
		}
	}

	for _, satz := range veraltet {
		if strings.Contains(fortlaufend, satz) {
			befunde = append(befunde, "veralteter Wortlaut vor macOS 15: "+satz)
		}
	}

	// **Rechtsklick darf vorkommen, aber nicht als Schritt.** Der Satz „das gibt
	// es nicht mehr" gehört in die Anleitung; die Handlung nicht. Eine Prüfung
	// auf das bloße Wort verböte beides und triebe den Hinweis heraus, der die
	// veraltete Anleitung als veraltet kennzeichnet.
	for _, zeile := range nummerierterSchritt.FindAllString(text, -1) {
		if strings.Contains(zeile, "Rechtsklick") {
			befunde = append(befunde, "Rechtsklick steht als Schritt: "+strings.TrimSpace(zeile))
		}
	}

	return befunde
}

// erststartnotiz holt den Zettel aus `scripts/releasezettel.sh`.
//
// Aus dem Skript und nicht aus einer Kopie: der Zettel entsteht dort, und eine
// zweite Fassung im Testverzeichnis prüfte sich selbst.
func erststartnotiz(t *testing.T) string {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join("..", "..", "scripts", "releasezettel.sh"))
	if err != nil {
		t.Fatalf("releasezettel.sh nicht lesbar: %v", err)
	}

	_, nach, gefunden := strings.Cut(string(roh), "<<'NOTIZ'\n")
	if !gefunden {
		t.Fatal("in releasezettel.sh steht kein Hier-Dokument NOTIZ mehr — hat sich der Zettel bewegt?")
	}

	notiz, _, gefunden := strings.Cut(nach, "\nNOTIZ\n")
	if !gefunden {
		t.Fatal("das Hier-Dokument NOTIZ ist nicht geschlossen")
	}

	return notiz
}

// **Der Name sagt, was geprüft wird, und nicht mehr.**
//
// Er hieß bis zum Zweitblick über Phase 8 „…StehtUeberallGleich". Das war eine
// stärkere Zusage als die Prüfung: verglichen werden Wortlaute und Stationen,
// nicht die Reihenfolge und nicht die Vollständigkeit des Weges. Ein Testname,
// der mehr behauptet als sein Rumpf, ist dieselbe Sorte Behauptung wie ein
// Dateikopf, den niemand nachprüft (`AGENTS.md` §10).
func TestJederOrtNenntDenAbgelesenenWortlaut(t *testing.T) {
	orte := []struct {
		name string
		text func(*testing.T) string
	}{
		{"der Zettel im .dmg", erststartnotiz},
		{"README.md", func(t *testing.T) string { return datei(t, "README.md") }},
		{"docs/RELEASE.md", func(t *testing.T) string { return datei(t, "docs", "RELEASE.md") }},
	}

	for _, ort := range orte {
		t.Run(ort.name, func(t *testing.T) {
			for _, befund := range befundeErststartweg(ort.text(t)) {
				t.Error(befund)
			}
		})
	}
}

// Die Gegenprobe: der Wächter muss bei jedem Verstoß rot werden.
//
// Ohne sie wäre er eine Behauptung — genau die Sorte, die dieses Repository
// zweimal teuer bezahlt hat (`AGENTS.md` §10).
func TestDerErststartwegWaechterSchlaegtAn(t *testing.T) {
	gut := erststartnotiz(t)

	faelle := []struct {
		name string
		text string
	}{
		{
			"ein fehlender abgelesener Satz",
			strings.ReplaceAll(gut, "Dennoch öffnen", "Trotzdem starten"),
		},
		{
			"der gefährliche Knopf ist nicht genannt",
			strings.ReplaceAll(gut, "In den Papierkorb legen", "…"),
		},
		{
			"der veraltete Grund",
			gut + "\nmacOS meldet, sie stamme von einem nicht verifizierten Entwickler.\n",
		},
		{
			"Rechtsklick als Schritt",
			gut + "\n  7. Rechtsklick auf die App und \"Öffnen\" wählen.\n",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if len(befundeErststartweg(fall.text)) == 0 {
				t.Error("der Wächter lässt es durch")
			}
		})
	}

	// **Und die andere Richtung**: eine Schranke, die auf alles anspringt, ist
	// so wenig wert wie eine, die nie anspringt (Phase 5, Punkt 6).
	if befunde := befundeErststartweg(gut); len(befunde) != 0 {
		t.Errorf("der unveränderte Zettel gilt als falsch: %v", befunde)
	}

	// **Der Hinweis auf den veralteten Weg bleibt erlaubt.** Er nennt
	// Rechtsklick, aber nicht als Schritt — verbietet der Wächter ihn, treibt er
	// die Warnung aus der Anleitung, die vor der veralteten Anleitung warnt.
	hinweis := gut + "\nDas frühere Öffnen über Rechtsklick gibt es nicht mehr.\n"
	if befunde := befundeErststartweg(hinweis); len(befunde) != 0 {
		t.Errorf("der Hinweis auf den alten Weg gilt als Verstoß: %v", befunde)
	}
}

// Die Bildschirmfotos gehören zum Beleg und werden mit ausgeliefert.
//
// Ein Verweis in `README.md` auf eine Datei, die es nicht gibt, ist auf GitHub
// ein leerer Rahmen — und der sieht aus wie ein Fehler des Lesers.
func TestDieBildschirmfotosLiegenDa(t *testing.T) {
	bilder := []string{
		"gatekeeper-blockiert.png",
		"gatekeeper-dennoch-oeffnen.png",
	}

	for _, bild := range bilder {
		t.Run(bild, func(t *testing.T) {
			pfad := filepath.Join("..", "..", "docs", "bilder", bild)

			roh, err := os.ReadFile(pfad) //nolint:gosec // der Pfad steht im Test
			if err != nil {
				t.Fatalf("%s fehlt: %v", pfad, err)
			}

			// **Ein PNG, und keins von null Byte.** Vorher stand hier ein
			// `os.Stat`, und der Zweitblick über Phase 8 hat gefragt, was eine
			// leere Datei daran hindert durchzukommen: nichts. Die Kennung sind
			// die acht Bytes, mit denen jedes PNG anfängt.
			if !strings.HasPrefix(string(roh), "\x89PNG\r\n\x1a\n") {
				t.Errorf("%s ist kein PNG", pfad)
			}

			// **20 KB, weil ein Bildschirmfoto eines Fensters darunter nicht
			// liegt.** Die beiden echten wiegen 56 und 139 KB; eine Fläche in
			// einer Farbe wiegt ein paar hundert Byte. Die Schranke trennt also
			// „ein Fenster" von „ein Platzhalter" — mehr kann sie nicht, und
			// mehr behauptet sie nicht.
			if len(roh) < 20*1024 {
				t.Errorf("%s wiegt nur %d Byte — ein Platzhalter?", pfad, len(roh))
			}

			for _, ort := range []string{"README.md", filepath.Join("docs", "RELEASE.md")} {
				if !strings.Contains(datei(t, strings.Split(ort, string(filepath.Separator))...), bild) {
					t.Errorf("%s verweist nicht auf %s", ort, bild)
				}
			}
		})
	}
}

// datei liest eine Datei aus der Projektwurzel.
func datei(t *testing.T, teile ...string) string {
	t.Helper()

	pfad := filepath.Join(append([]string{"..", ".."}, teile...)...)

	roh, err := os.ReadFile(pfad) //nolint:gosec // die Pfade stehen im Test, nicht in einer Eingabe
	if err != nil {
		t.Fatalf("%s nicht lesbar: %v", pfad, err)
	}

	return string(roh)
}
