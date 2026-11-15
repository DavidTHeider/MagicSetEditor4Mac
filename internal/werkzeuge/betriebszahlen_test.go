// **Zwei Zahlenfamilien im Fließtext, die niemand hielt** — der zweite
// Zweitblick über die Freigabe 0.3.0, Befund 9
// (#1170).
//
// # Der Befund
//
// `README.md` und `docs/ADMINISTRATION.md` nannten **16** Migrationen; es sind
// **21**. Der vorhandene README-Zahlenwächter blieb grün, weil sein Ausdruck
// ausschließlich die vier Schnittstellenzeilen liest — eine Zusage, die
// **daneben** stand, war für ihn unsichtbar.
//
// Und die Zahl der mitgelieferten Pakete stand als „319 der 327" in drei
// Dateien. **Acht** ist die einzige davon, die dieses Repositorium selbst
// bestimmt; die anderen beiden sind Eigenschaften eines fremden Packs und
// wandern mit ihm.
//
// # Was hier gehalten wird
//
// Die Migrationszahl gegen das Verzeichnis, und die Zahl der mitgelieferten
// Pakete gegen `scripts/mitliefern.sh`.
//
// # Was **nicht** gehalten wird, und es steht hier, statt still zu bleiben
//
// **Die Größe des fremden Packs.** 324 und 526 sind Messungen an Verzeichnissen,
// die nicht in diesem Repositorium liegen (§11) — ein Wächter dafür bräuchte
// sie, und dann prüfte er auf einem Rechner, auf dem sie zufällig da sind.
// Was die Dateien deshalb tragen, ist das **Datum** der Zählung; das ist die
// ehrliche Form (§10).
//
// **Und das ist gemessen, nicht behauptet.** Die Gegenprobe am 2026-08-28:
// `**316 der 324**` zurück auf `**319 der 327**` gedreht lässt diesen Lauf
// **grün** — die Differenz ist beide Male acht. Genau der Fehler, den Codex
// gemeldet hat, ginge daran vorbei; was hier hält, ist allein die Zahl der
// mitgelieferten Pakete. Die Migrationszahl daneben geht bei derselben Probe
// rot.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// migrationszahl findet „**21** Schritte" und „**21 Migrationsschritte**".
var migrationszahl = regexp.MustCompile(`\*\*(\d+)\*\* Schritte|\*\*(\d+) Migrationsschritte\*\*`)

func TestDieGenannteMigrationszahlStimmt(t *testing.T) {
	t.Parallel()

	eintraege, err := os.ReadDir(filepath.Join("..", "storage", "migrationen"))
	if err != nil {
		t.Fatalf("die Migrationen lesen: %v", err)
	}

	// **Die Schritte der Grundlinie zählen mit** (Vermerk 0189): sie liegen
	// nicht mehr als Datei vor, ein Bestand führt sie aber wie jede andere.
	umfang, err := os.ReadFile(filepath.Join("..", "storage", "grundlinie", "umfang.txt"))
	if err != nil {
		t.Fatalf("den Umfang der Grundlinie lesen: %v", err)
	}

	wirklich := 0

	for _, zeile := range strings.Split(string(umfang), "\n") {
		if zeile != "" && !strings.HasPrefix(zeile, "#") {
			wirklich++
		}
	}

	for _, eintrag := range eintraege {
		if strings.HasSuffix(eintrag.Name(), ".sql") {
			wirklich++
		}
	}

	for _, datei := range []string{"README.md", "docs/ADMINISTRATION.md"} {
		funde := migrationszahl.FindAllStringSubmatch(lies(t, datei), -1)
		if len(funde) == 0 {
			t.Errorf("%s nennt keine Migrationszahl mehr — sie stand dort, "+
				"und ein Wächter, der nichts findet, hält nichts", datei)

			continue
		}

		for _, fund := range funde {
			genannt := fund[1] + fund[2]

			if zahl, _ := strconv.Atoi(genannt); zahl != wirklich { //nolint:errcheck // ein Nichttreffer gibt 0 und ist ungleich
				t.Errorf("%s nennt %s Migrationen, es sind %d",
					datei, genannt, wirklich)
			}
		}
	}
}

// mitgeliefert findet „316 der 324" und die Zahl 8 dazwischen im Fließtext.
var mitgeliefert = regexp.MustCompile(`\*\*(\d+) der (\d+)\*\* Pakete`)

func TestDieZahlDerMitgeliefertenPaketeStimmt(t *testing.T) {
	t.Parallel()

	skript := lies(t, "scripts/mitliefern.sh")

	anfang := strings.Index(skript, "PAKETE=(")
	if anfang < 0 {
		t.Fatal("mitliefern.sh führt keine Paketliste mehr")
	}

	liste, _, geschlossen := strings.Cut(skript[anfang:], ")")
	if !geschlossen {
		t.Fatal("die Paketliste in mitliefern.sh ist nicht geschlossen")
	}

	gebuendelt := strings.Count(liste, ".mse-")

	if gebuendelt == 0 {
		t.Fatal("die Paketliste ist leer")
	}

	for _, datei := range []string{"README.md", "docs/RELEASE.md"} {
		funde := mitgeliefert.FindAllStringSubmatch(lies(t, datei), -1)
		if len(funde) == 0 {
			t.Errorf("%s nennt nicht mehr, wie viele Pakete bezogen werden", datei)

			continue
		}

		for _, fund := range funde {
			bezogen, _ := strconv.Atoi(fund[1]) //nolint:errcheck // der Ausdruck lässt nur Ziffern durch
			gesamt, _ := strconv.Atoi(fund[2])  //nolint:errcheck // dito

			if gesamt-bezogen != gebuendelt {
				t.Errorf("%s sagt %d der %d würden bezogen — das sind %d "+
					"mitgelieferte, und mitliefern.sh führt %d",
					datei, bezogen, gesamt, gesamt-bezogen, gebuendelt)
			}
		}
	}
}
