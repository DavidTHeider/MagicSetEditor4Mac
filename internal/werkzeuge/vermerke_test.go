// Der Wächter über die Liste der Vermerke.
//
// `docs/entscheidungen/README.md` führt eine Tabelle aller Entscheidungen. Sie
// stand am 2026-08-16 bei **0011**, während acht Vermerke daneben lagen; die
// Datei sagt das selbst und schließt mit „Bis das jemand baut, ist der Bestand
// von Hand nachzuhalten." Am 2026-08-19 fehlten wieder zwei — 0027 und 0028.
//
// Eine Liste, die niemand nachzählt, ist keine Liste, sondern eine Behauptung
// über eine Liste (§10).
//
// # Was er nicht prüft
//
// **Ob die Beschreibung stimmt.** Er zählt Nummern, nicht Sätze. Eine Zeile
// „0029 — irgendwas" ist ihm recht.
//
// **Ob der Stand stimmt.** „angenommen" gegen „verworfen" steht im Vermerk und
// in der Tabelle; welcher von beiden gilt, weiß nur, wer ihn gelesen hat.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestJederVermerkStehtInDerListe(t *testing.T) {
	t.Parallel()

	verzeichnis := filepath.Join("..", "..", "docs", "entscheidungen")

	eintraege, err := os.ReadDir(verzeichnis)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	liste, err := os.ReadFile( //nolint:gosec // Dokumentation im Repository
		filepath.Join(verzeichnis, "README.md"))
	if err != nil {
		t.Fatalf("die Liste fehlt: %v", err)
	}

	vermerke := vermerknummern(eintraege)
	if len(vermerke) == 0 {
		t.Fatal("kein einziger Vermerk gefunden — der Wächter prüft nichts")
	}

	for nummer, datei := range vermerke {
		// **Auf den Dateinamen und nicht auf die Nummer allein.** Eine Zeile
		// `| 0029 | … |` ohne Verweis sähe wie ein Eintrag aus und führte
		// nirgendwohin.
		if !strings.Contains(string(liste), "("+datei+")") {
			t.Errorf("%s (%s) fehlt in docs/entscheidungen/README.md", nummer, datei)
		}
	}
}

// **Und die Liste erfindet keine.** Eine Zeile auf einen gelöschten Vermerk
// ist ein Verweis ins Leere — dieselbe Klasse, andere Richtung.
func TestDieListeNenntNurVorhandeneVermerke(t *testing.T) {
	t.Parallel()

	verzeichnis := filepath.Join("..", "..", "docs", "entscheidungen")

	liste, err := os.ReadFile( //nolint:gosec // Dokumentation im Repository
		filepath.Join(verzeichnis, "README.md"))
	if err != nil {
		t.Fatalf("die Liste fehlt: %v", err)
	}

	verweis := regexp.MustCompile(`\((\d{4}-[a-z0-9-]+\.md)\)`)

	treffer := verweis.FindAllStringSubmatch(string(liste), -1)
	if len(treffer) == 0 {
		t.Fatal("die Liste nennt keinen einzigen Vermerk — der Wächter prüft nichts")
	}

	for _, gefunden := range treffer {
		// **`filepath.Base` gegen den Pfaddurchgriff.** Der Name kommt aus einer
		// Datei; das Muster oben lässt keinen Schrägstrich zu, aber ein Wächter
		// soll sich nicht darauf verlassen, dass ein anderes Muster ihn schützt.
		ziel := filepath.Join(verzeichnis, filepath.Base(gefunden[1]))
		if _, fehler := os.Stat(ziel); fehler != nil {
			t.Errorf("die Liste nennt %s — die Datei gibt es nicht", gefunden[1])
		}
	}
}

// vermerknummern gibt je Vermerk seine Nummer und seinen Dateinamen.
func vermerknummern(eintraege []os.DirEntry) map[string]string {
	name := regexp.MustCompile(`^(\d{4})-[a-z0-9-]+\.md$`)
	gefunden := map[string]string{}

	for _, eintrag := range eintraege {
		if treffer := name.FindStringSubmatch(eintrag.Name()); treffer != nil {
			gefunden[treffer[1]] = eintrag.Name()
		}
	}

	return gefunden
}

// **Jede Zeile der Liste hat drei Spalten.**
//
// # Was [TestJederVermerkStehtInDerListe] nicht hält
//
// Er zählt Nummern. Eine Zeile mit **vier** Spalten — Verweis, Titel, Datum,
// Status — steht drin, verweist richtig und macht ihn grün; gerendert wird
// sie trotzdem falsch, denn die Kopfzeile nennt drei.
//
// Genau das ist am 2026-09-01 aufgefallen: Vermerk 0061 stand seit einem Tag
// mit vier Spalten da, und 0062 hat den Fehler kopiert. Ein Wächter, der die
// halbe Zusage hält, macht die andere unsichtbar (§10) — deshalb steht die
// andere Hälfte jetzt hier.
func TestJedeVermerkzeileHatDreiSpalten(t *testing.T) {
	t.Parallel()

	roh, err := os.ReadFile( //nolint:gosec // Dokumentation im Repository
		filepath.Join("..", "..", "docs", "entscheidungen", "README.md"))
	if err != nil {
		t.Fatalf("die Liste fehlt: %v", err)
	}

	var geprueft int

	for nummer, zeile := range strings.Split(string(roh), "\n") {
		if !strings.HasPrefix(zeile, "| [00") {
			continue
		}

		geprueft++

		// **Die führende und die schließende Pipe zählen mit** — eine Zeile
		// mit drei Spalten hat vier Trennzeichen.
		if felder := strings.Count(zeile, "|"); felder != 4 {
			t.Errorf("Zeile %d hat %d Spalten statt 3: %.80s…",
				nummer+1, felder-1, zeile)
		}
	}

	// **Die Gegenprobe gegen den Wächter selbst**: findet er keine Zeile,
	// prüft er nichts und ist trotzdem grün.
	if geprueft == 0 {
		t.Fatal("keine einzige Vermerkzeile gefunden — der Wächter prüft nichts")
	}
}
