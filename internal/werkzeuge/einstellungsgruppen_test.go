// **Jede Einstellung steht in genau einer Gruppe** — UX-04.AK-01
// (#1676).
//
// # Der Anlass
//
// **Der Entwurf schneidet anders als der Bestand.** Das Artboard
// `Einstellungen` zeigt neun Gruppen — Allgemein, Erscheinungsbild, Nutzer
// und Profil, Fachfunktionen, KI, Verbindungen, Dienste, Daten und
// Sicherung, Diagnose —, und `sprache` und `erscheinungsbild` liegen beide
// im Bereich `anwendung` und trotzdem in zweien davon. Die Zuordnung ist
// deshalb eine **Aufzählung je Schlüssel** in
// `frontend/src/seiten/einstellungsgruppen.ts`.
//
// # Was er prüft
//
// **Jeder Schlüssel aus [domain.Einstellungsreihe] steht dort**, und dort
// steht keiner, den es nicht gibt. Eine neue Einstellung erzwingt damit die
// Entscheidung, statt still unter „Allgemein" zu landen — der Rückfall ist
// für einen **fremden** Bestand gedacht (§8) und nicht für Vergesslichkeit.
//
// # Was er nicht prüft
//
// **Ob die Gruppe die richtige ist.** Er sieht, dass eine dasteht, nicht
// dass sie trifft — ob die Kostengrenze zur KI gehört oder zu den Diensten,
// entscheidet ein Mensch. Und er sieht die Tabelle, nicht die Seite: dass
// die Einstellungsseite danach anordnet, prüft
// `frontend/src/seiten/einstellungsgruppen.test.ts` und der Lauf über die
// Seite daneben.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gruppenzeile liest einen Eintrag der Tabelle: `schluessel: 'gruppe',` —
// mit und ohne Anführungszeichen um den Schlüssel.
var gruppenzeile = regexp.MustCompile(`(?m)^\s+'?([a-zA-Z][\w.]*)'?:\s*'(\w+)',`)

func TestJedeEinstellungStehtInGenauEinerGruppe(t *testing.T) {
	t.Parallel()

	zugeordnet := gruppentabelle(t)

	for _, schluessel := range domain.Einstellungsreihe() {
		if _, dabei := zugeordnet[schluessel]; !dabei {
			t.Errorf("%q steht in keiner Gruppe — eintragen in "+
				"frontend/src/seiten/einstellungsgruppen.ts", schluessel)
		}
	}
}

// **Und keine Gruppe führt einen Schlüssel, den es nicht gibt.** Ein
// Eintrag, der ins Leere zeigt, sieht aus wie eine Zuordnung und ist keine —
// er bliebe stehen, wenn eine Einstellung wegfällt.
func TestKeineGruppeFuehrtEinenFremdenSchluessel(t *testing.T) {
	t.Parallel()

	echte := make(map[string]bool, len(domain.Einstellungsreihe()))
	for _, schluessel := range domain.Einstellungsreihe() {
		echte[schluessel] = true
	}

	for schluessel := range gruppentabelle(t) {
		if !echte[schluessel] {
			t.Errorf("%q steht in einer Gruppe, aber in keiner Einstellung",
				schluessel)
		}
	}
}

// gruppentabelle liest die Zuordnung aus dem Quelltext des Fensters.
func gruppentabelle(t *testing.T) map[string]string {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join("..", "..", "frontend", "src",
		"seiten", "einstellungsgruppen.ts"))
	if err != nil {
		t.Fatalf("die Gruppentabelle lesen: %v", err)
	}

	// **Nur der Verbund und nicht die ganze Datei** — die Aufzählung der
	// neun Gruppen darüber sähe sonst wie neun Einträge aus.
	inhalt := string(roh)
	anfang := regexp.MustCompile(`nachGruppe[^=]*= \{`).FindStringIndex(inhalt)
	if anfang == nil {
		t.Fatal("der Verbund nachGruppe steht nicht in der Datei")
	}

	tabelle := map[string]string{}

	for _, treffer := range gruppenzeile.FindAllStringSubmatch(
		inhalt[anfang[1]:], -1) {
		tabelle[treffer[1]] = treffer[2]
	}

	if len(tabelle) == 0 {
		t.Fatal("in nachGruppe fand sich kein Eintrag — der Leser greift daneben")
	}

	return tabelle
}
