// **Jeder Abzug im `README.md` lässt sich neu machen** —
// #1271.
//
// # Wofür dieser Wächter da ist
//
// Die Abzüge entstehen aus einem Lauf, weil Bildschirmabzüge veralten und ein
// Abzug, den niemand neu machen kann, nicht neu gemacht wird. **Seit 0.6.0
// ist es die Hüllenprobe `readmebilder` am nativen Fenster**; bis dahin war
// es `scripts/abzuege.mjs` über einen Browser. Die Wege stehen in
// `scripts/abzuegeliste.mjs`.
//
// **Und trotzdem band das `README.md` zehn Abzüge ein, die es nicht kannte** —
// von Hand aufgenommen und damit an dem Tag festgeschrieben, an dem sie
// entstanden. Der Aufmacher zeigte die Kartenlasche in einer Anordnung, die
// es seit #1271 nicht mehr gibt.
//
// Dieser Wächter hält beide Richtungen:
//
//   - Jedes Bild, das `README.md` einbindet, entsteht aus dem Skript oder
//     steht in [vonHand] mit einem Grund.
//   - Jeder Weg im Skript wird eingebunden — sonst entsteht eine Datei, die
//     niemand sieht.
//
// # Was er nicht prüft
//
//   - **Ob der Abzug den heutigen Stand zeigt.** Er sieht, dass es einen Weg
//     gibt, nicht wann jemand ihn zuletzt gefahren hat. Das steht in
//     `docs/RELEASE.md` und ist Handarbeit — wie jeder Lauf, der ein Fenster
//     braucht.
//   - **Ob das Bild etwas erklärt.** Dafür gibt es keinen Maßstab außer dem
//     Ansehen.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// abzugImText findet `docs/bilder/<name>.png` im Fließtext.
var abzugImText = regexp.MustCompile(`docs/bilder/([a-z0-9-]+)\.png`)

// wegImSkript findet `name: '…'` in `scripts/abzuegeliste.mjs` — einfache
// oder doppelte Anführungszeichen: die Liste ist JavaScript, und ein
// Wechsel der Schreibweise machte den Wächter sonst blind statt rot.
var wegImSkript = regexp.MustCompile(`name:\s*['"]([a-z0-9-]+)['"]`)

// wegeblock grenzt die Suche auf `const wege = [ … ];` ein.
//
// **Nötig, weil `name:` auch anderswo vorkommt**: die Erscheinungsbilder
// werden über `{ name: "prefers-color-scheme", … }` emuliert, und ohne diese
// Klammer hielte der Wächter das für einen Abzug und verlangte eine Datei
// dafür.
var wegeblock = regexp.MustCompile(`(?s)(?:export )?const wege = \[(.*?)\n\];`)

// dunkleFassung ist der Anhang, den der dunkle Abzug trägt.
const dunkleFassung = "-dunkel"

// vonHand sind die Bilder ohne Weg im Skript — jedes mit seinem Grund.
//
// **Ein leerer Grund ist keiner**, wie bei jeder Ausnahmeliste in diesem
// Projekt.
var vonHand = map[string]string{
	"gatekeeper-blockiert": "ein Systemdialog von macOS, kein Fenster dieser " +
		"Anwendung — das Skript fährt einen Browser und sieht ihn nicht",
	"gatekeeper-dennoch-oeffnen": "dito: die Systemeinstellungen",
}

func TestJederAbzugImReadmeHatSeinenWeg(t *testing.T) {
	t.Parallel()

	// **Die Liste steht seit dem 2026-09-03 in `abzuegeliste.mjs`** —
	// `abzuege.mjs` wuchs über 400 Zeilen (§2), als die Einführung einen
	// eigenen Abzug bekam (#1325).
	block := wegeblock.FindStringSubmatch(lies(t, "scripts/abzuegeliste.mjs"))
	if block == nil {
		t.Fatal("scripts/abzuegeliste.mjs hat kein `wege = [ … ];` — heißt " +
			"die Liste noch so?")
	}

	wege := map[string]bool{}
	for _, treffer := range wegImSkript.FindAllStringSubmatch(block[1], -1) {
		wege[treffer[1]] = true
	}

	if len(wege) == 0 {
		t.Fatal("scripts/abzuegeliste.mjs nennt keinen einzigen Weg — heißt das " +
			"Feld noch `name`?")
	}

	// **Hell und dunkel sind ein Abzug.** `…-dunkel.png` entsteht aus
	// demselben Weg; wer beide zählte, verlangte zweimal denselben Eintrag.
	eingebunden := map[string]bool{}
	for _, treffer := range abzugImText.FindAllStringSubmatch(
		lies(t, "README.md"), -1) {
		eingebunden[strings.TrimSuffix(treffer[1], dunkleFassung)] = true
	}

	if len(eingebunden) == 0 {
		t.Fatal("README.md bindet kein einziges Bild ein — stimmt das Muster noch?")
	}

	var ohneWeg []string

	for bild := range eingebunden {
		grund, ausgenommen := vonHand[bild]

		switch {
		case wege[bild] && ausgenommen:
			t.Errorf("%s steht im Skript **und** in vonHand — eines von beiden "+
				"ist falsch", bild)
		case wege[bild]:
		case ausgenommen && strings.TrimSpace(grund) == "":
			t.Errorf("%s ist ausgenommen, ohne dass ein Grund dasteht", bild)
		case ausgenommen:
		default:
			ohneWeg = append(ohneWeg, bild)
		}
	}

	if len(ohneWeg) > 0 {
		sort.Strings(ohneWeg)
		t.Errorf("diese Bilder bindet README.md ein, ohne dass ein Weg sie "+
			"erzeugt: %s — von Hand aufgenommen heißt: beim nächsten Umbau "+
			"veraltet und niemand merkt es", strings.Join(ohneWeg, ", "))
	}

	// **Und andersherum**: ein Weg ohne Einbindung erzeugt eine Datei, die
	// niemand sieht — und kostet bei jedem Lauf Zeit.
	for weg := range wege {
		if !eingebunden[weg] {
			t.Errorf("scripts/abzuegeliste.mjs erzeugt %s.png, README.md bindet es "+
				"nicht ein", weg)
		}
	}

	// **Und die Dateien liegen wirklich da** — hell wie dunkel.
	for weg := range wege {
		for _, anhang := range []string{"", dunkleFassung} {
			pfad := filepath.Join("..", "..", "docs", "bilder", weg+anhang+".png")
			if _, err := os.Stat(pfad); err != nil {
				t.Errorf("%s%s.png fehlt — der Weg steht im Skript, der Lauf "+
					"ist nicht gefahren", weg, anhang)
			}
		}
	}

	// **Und die Ausnahmeliste führt kein Gespenst.**
	for bild := range vonHand {
		if !eingebunden[bild] {
			t.Errorf("vonHand nennt %s — README.md bindet es nicht (mehr) ein", bild)
		}
	}

	// **Und der Satz darüber nennt dieselbe Zahl** — ergänzt am 2026-09-07.
	//
	// # Warum der Wächter das vorher nicht hielt
	//
	// **Er zählte die Wege und sagte die Zahl niemandem.** Im `README.md`
	// stand „fährt 19 Wege durch die Anwendung", während es 22 waren und
	// danach 27 — die Zahl ist über drei Phasen mitgelaufen, ohne dass etwas
	// rot wurde. Genau die Gestalt, gegen die §16 den Flächenzahlenwächter
	// hat: eine Zahl in Fließtext steht unter keiner Aufsicht, und gerade weil
	// der Absatz daneben gewacht war, las sie niemand nach.
	//
	// **Es ist derselbe Wächter und keiner daneben** (§10): die Zusage wird
	// dort geprüft, wo sie eingelöst wird, und wer einen Weg hinzufügt, sieht
	// beide Hälften zugleich.
	wegezahl := regexp.MustCompile(`fährt\s*\n?\s*(\d+) Wege durch die Anwendung`)

	satz := wegezahl.FindStringSubmatch(lies(t, "README.md"))
	if satz == nil {
		t.Fatal("README.md sagt nicht mehr, wie viele Wege das Skript fährt — " +
			"heißt der Satz noch so?")
	}

	genannt, wandelfehler := strconv.Atoi(satz[1])
	if wandelfehler != nil {
		t.Fatalf("%q ist keine Zahl: %v", satz[1], wandelfehler)
	}

	if genannt != len(wege) {
		t.Errorf("README.md nennt %d Wege, im Skript stehen %d", genannt, len(wege))
	}

	t.Logf("%d Wege im Skript, %d Bilder im README, %d von Hand",
		len(wege), len(eingebunden), len(vonHand))
}
