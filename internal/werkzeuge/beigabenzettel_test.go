// Der Zettel im `.dmg` sagt, was im `.dmg` liegt.
//
// # Warum das ein eigener Wächter ist
//
// `lizenz_test.go` hält die **Lizenz** an sieben Stellen gegeneinander. Hier
// steht die andere Hälfte derselben Zusage: nicht *unter welcher Lizenz*,
// sondern *was überhaupt* mitgeht. Beide Fragen an einem Wächter hießen, ihn
// rot zu machen, wenn eine von zweien fällt, ohne dass man sähe welche (§2).
//
// # Der Anlass
//
// Der Zettel behauptete bis zum 2026-08-23:
//
//	Kartenrahmen, Mana-Symbole, Schriften, Regeltexte und Marken von Wizards
//	of the Coast … werden von dieser Anwendung auch nicht mitgeliefert.
//
// **Und dasselbe Skript legte sie dreißig Zeilen später hinein.**
// `scripts/release.sh` ruft `scripts/mitliefern.sh`, und das kopiert acht
// Pakete des bezogenen Packs nach `Contents/Resources/vorlagen` — 167 MB,
// darunter `magic-modules.mse-include` mit den Rahmen und
// `magic-mana-large.mse-symbol-font` mit den Manasymbolen.
//
// Der Satz war einmal wahr. Er ist es seit
// [0037](../../docs/entscheidungen/0037-die-vorlagen-gehen-ins-buendel.md) nicht
// mehr, und niemand hat ihn nachgezogen
// (#894).
//
// # Was er nicht prüft
//
//   - **Das gebaute `.dmg`.** Gelesen werden die beiden Skripte, nicht ein
//     Abbild — ein Bau läuft nicht in der Pipeline
//     ([0007](../../docs/entscheidungen/0007-kein-entfernter-lauf.md)). Dass
//     `hdiutil` den Zettel wirklich hineinlegt, sagt nur ein Bau.
//   - **Ob die Aufzählung im Zettel vollständig ist.** Er prüft, dass der
//     Zettel die Beigabe **nicht verneint** und auf `DRITTANBIETER.md`
//     verweist. Ob jedes einzelne Paket genannt ist, prüft er nicht — die
//     Liste steht in `mitliefern.sh` und ändert sich dort.
//
// # Und die Release-Notiz gehört dazu
//
// `docs/RELEASE.md` führt einen Block **zum Übernehmen** — den Text, der in
// die Freigabe geht. Er sagte bis zum 2026-08-23 dasselbe Falsche wie der
// Zettel, und der Wächter sah ihn nicht: er las nur die Skripte
// (#919).
//
// **Zwei Texte, dieselbe Zusage, ein Wächter.** Wer die Vorlage übernimmt,
// veröffentlicht, was hier geprüft wird.
//   - **Ob `DRITTANBIETER.md` stimmt.** Er prüft, dass der Zettel dorthin
//     zeigt, nicht was dort steht.
package werkzeuge_test

import (
	"strings"
	"testing"
)

// verneinungen sind die Wendungen, mit denen der Zettel die Beigabe bestritt.
//
// **Die Wendung und nicht das Wort.** „mitgeliefert" darf im Zettel stehen —
// er soll ja sagen, dass etwas mitgeht. Gesucht wird die Verneinung.
var verneinungen = []string{
	"nicht mitgeliefert",
	"nicht mit mitgeliefert",
	"werden nicht mitgeliefert",
	"nicht enthalten",
	"liegen nicht bei",
}

// **Kein ausgelieferter Text verneint die Beigabe, solange der Freigabeweg sie
// einlegt.**
//
// Die Bedingung ist keine Zierde: fiele `mitliefern.sh` je aus dem
// Freigabeweg, wäre der Satz wieder wahr — und ein Wächter, der ihn dann
// verböte, erzwänge eine falsche Angabe.
func TestDerZettelVerneintKeineBeigabeDieDerFreigabewegEinlegt(t *testing.T) {
	if !freigabewegLegtBeigabenEin(t) {
		t.Skip("scripts/release.sh ruft mitliefern.sh nicht mehr — dann darf der Zettel die Beigabe verneinen")
	}

	texte := map[string]string{
		"der Zettel im .dmg":                   zettelImAbbild(t),
		"die Release-Notiz in docs/RELEASE.md": releasenotiz(t),
	}

	for name, text := range texte {
		for _, verneinung := range verneinungen {
			if strings.Contains(text, verneinung) {
				t.Errorf("%s sagt %q — der Freigabeweg legt die Vorlagen aber ein "+
					"(scripts/mitliefern.sh)", name, verneinung)
			}
		}
	}

	if !strings.Contains(texte["der Zettel im .dmg"], "DRITTANBIETER.md") {
		t.Error("der Zettel verweist nicht auf DRITTANBIETER.md — dort steht, " +
			"wem die Beigaben gehören")
	}
}

// uebernahmemarke steht in `docs/RELEASE.md` unmittelbar vor dem Block, der in
// die Freigabe geht.
const uebernahmemarke = "Der Block steht deshalb hier zum Übernehmen."

// releasenotiz holt den Block, der in `docs/RELEASE.md` zum Übernehmen steht.
//
// **Nur der eine Block und nicht jedes Blockzitat.** Der erste Anlauf sammelte
// alle Zeilen mit `>` — und fing damit den **erklärenden** Kasten mit, der den
// alten, falschen Wortlaut **zitiert** (#894).
// Ein Wächter, der die Erinnerung an einen Fehler für den Fehler hält, zwänge
// dazu, die Erinnerung zu löschen.
//
// Geschnitten wird deshalb an [uebernahmemarke] — dem Satz, der den Block
// ankündigt.
func releasenotiz(t *testing.T) string {
	t.Helper()

	_, nach, gefunden := strings.Cut(lies(t, "docs/RELEASE.md"), uebernahmemarke)
	if !gefunden {
		t.Fatalf("docs/RELEASE.md führt %q nicht mehr — der Wächter prüft dann nichts",
			uebernahmemarke)
	}

	var gesammelt []string

	for _, zeile := range strings.Split(nach, "\n") {
		if strings.HasPrefix(zeile, ">") {
			gesammelt = append(gesammelt, strings.TrimPrefix(zeile, "> "))

			continue
		}

		// **Der Block endet an der ersten Zeile, die kein Zitat ist** — und
		// Leerzeilen dazwischen gehören dazu.
		if len(gesammelt) > 0 && strings.TrimSpace(zeile) != "" {
			break
		}
	}

	if len(gesammelt) == 0 {
		t.Fatal("hinter der Übernahmemarke steht kein Blockzitat mehr")
	}

	return strings.Join(gesammelt, "\n")
}

// Die Gegenprobe: der Wächter muss bei der alten Behauptung rot werden.
//
// **Ohne sie wäre er eine Behauptung** (§10). Und die Gegenprobe muss das
// Verhalten wirklich verschieben — deshalb wird hier der Satz eingesetzt, der
// bis zum 2026-08-23 wirklich dastand, nicht ein erfundener.
func TestDerBeigabenwaechterSchlaegtAn(t *testing.T) {
	faelle := []struct {
		name   string
		zettel string
		fehlt  string
	}{
		{
			name: "die Verneinung, die bis zum 2026-08-23 dastand",
			zettel: "Lizenz: GPL-2.0-or-later — mit Ausnahme von Inhalten Dritter. " +
				"Kartenrahmen, Mana-Symbole, Schriften, Regeltexte und Marken von " +
				"Wizards of the Coast gehören nicht dazu und werden von dieser " +
				"Anwendung auch nicht mitgeliefert. Einzelheiten in DRITTANBIETER.md.",
			fehlt: "nicht mitgeliefert",
		},
		{
			name:   "kein Verweis auf DRITTANBIETER.md",
			zettel: "Lizenz: GPL-2.0-or-later. Die Vorlagen liegen bei.",
			fehlt:  "DRITTANBIETER.md",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if !befundeBeigaben(fall.zettel) {
				t.Errorf("der Wächter bleibt still, obwohl %q der Befund wäre", fall.fehlt)
			}
		})
	}
}

// befundeBeigaben sagt, ob ein Zettel gegen die Zusage verstößt.
//
// Ausgelagert, damit die Gegenprobe denselben Weg geht wie der Wächter — eine
// zweite Prüfung daneben prüfte sich selbst.
func befundeBeigaben(zettel string) bool {
	for _, verneinung := range verneinungen {
		if strings.Contains(zettel, verneinung) {
			return true
		}
	}

	return !strings.Contains(zettel, "DRITTANBIETER.md")
}

// freigabewegLegtBeigabenEin sagt, ob `release.sh` die Vorlagen einlegt.
func freigabewegLegtBeigabenEin(t *testing.T) bool {
	t.Helper()

	return strings.Contains(lies(t, "scripts/release.sh"), "mitliefern.sh")
}

// zettelImAbbild holt den Zettel aus `scripts/releasezettel.sh`.
//
// **Aus dem Skript und nicht aus einer Kopie**: er entsteht dort, und eine
// zweite Fassung im Testverzeichnis prüfte sich selbst.
func zettelImAbbild(t *testing.T) string {
	t.Helper()

	_, nach, gefunden := strings.Cut(lies(t, "scripts/releasezettel.sh"), "<<'NOTIZ'\n")
	if !gefunden {
		t.Fatal("in releasezettel.sh steht kein Hier-Dokument NOTIZ mehr — hat sich der Zettel bewegt?")
	}

	zettel, _, gefunden := strings.Cut(nach, "\nNOTIZ\n")
	if !gefunden {
		t.Fatal("das Hier-Dokument NOTIZ ist nicht geschlossen")
	}

	return zettel
}
