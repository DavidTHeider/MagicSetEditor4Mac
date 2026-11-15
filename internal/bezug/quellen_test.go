package bezug_test

import (
	"regexp"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
)

// **Die Bezugsquellenliste** —
// #263.
//
// Der Paketverwalter kann seit Phase 7 beziehen; was fehlte, ist die letzte
// Meile: heute muss jemand eine Adresse abtippen. Die Liste ist die Antwort,
// und sie steht **hier** und nicht im Fenster, weil sie geprüft gehört — gegen
// die Regel „ein Zweig ist kein Stand" (#56) und gegen `DRITTANBIETER.md`.
//
// **Mitgeliefert wird nichts.** Die Dateien landen beim Beziehen im
// Benutzerverzeichnis, nicht im Repository und nicht im `.dmg` (§11).
func TestJedeQuelleTraegtWasEinBezugBraucht(t *testing.T) {
	t.Parallel()

	quellen := bezug.Quellen()
	if len(quellen) == 0 {
		t.Fatal("die Liste ist leer")
	}

	for _, quelle := range quellen {
		t.Run(quelle.Name, func(t *testing.T) {
			t.Parallel()

			if quelle.Adresse == "" || quelle.Stand == "" || quelle.Schluessel == "" {
				t.Errorf("%+v", quelle)
			}

			// **Die Größe ist keine Zier.** Zwischen der kleinsten und der
			// größten Quelle liegt der Faktor fünf; wer das nicht sieht,
			// startet einen Bezug über Stunden, weil er den Namen für einen
			// Hinweis hielt.
			if quelle.Megabyte <= 0 {
				t.Errorf("%s nennt keine Größe", quelle.Name)
			}
		})
	}
}

// **Jeder Stand ist ein Commit, kein Zweig** (#56).
//
// `bezug.standPruefen` weist einen Zweig ab — aber erst beim Bezug, über eine
// Netzanfrage. Eine Liste, die von uns kommt, soll gar nicht erst eine
// enthalten: der Fehler fiele sonst beim Anwender auf, nicht hier.
func TestJederStandIstEinCommit(t *testing.T) {
	t.Parallel()

	commit := regexp.MustCompile(`^[0-9a-f]{40}$`)

	for _, quelle := range bezug.Quellen() {
		if !commit.MatchString(quelle.Stand) {
			t.Errorf("%s: %q ist kein Commit", quelle.Name, quelle.Stand)
		}
	}
}

// **Kein Name zweimal.** Die Oberfläche zeigt sie als Liste und braucht einen
// Schlüssel; zwei gleiche Namen ergäben zwei Einträge, von denen einer nie
// anklickbar ist.
func TestKeineQuelleKommtZweimalVor(t *testing.T) {
	t.Parallel()

	gesehen := make(map[string]bool)

	for _, quelle := range bezug.Quellen() {
		if gesehen[quelle.Name] {
			t.Errorf("%s steht zweimal", quelle.Name)
		}

		gesehen[quelle.Name] = true

		if gesehen[quelle.Schluessel] {
			t.Errorf("%s: der Schlüssel %q steht zweimal", quelle.Name, quelle.Schluessel)
		}

		gesehen[quelle.Schluessel] = true
	}
}

// **`M15-Magic-Pack` steht darin** — so nennt es
// #263
// ausdrücklich: es ist das, was ein neuer Nutzer als Erstes braucht.
func TestDasM15PackStehtInDerListe(t *testing.T) {
	t.Parallel()

	for _, quelle := range bezug.Quellen() {
		if quelle.Name == "M15-Magic-Pack" {
			return
		}
	}

	t.Error("M15-Magic-Pack fehlt")
}

// **Kein Unterverzeichnis — und das ist der Kern von #263.**
//
// Die Packs tragen ihre Vorlagen unter `data` und ihre **Schriften** daneben
// (`Magic - Fonts`, im installierten `M15-Magic-Pack` 131 Dateien, 26 MB).
// Die Vorgabe, die bis hierher im Formular stand, war `data` — damit kämen die
// Vorlagen und nicht ihre Schriften, und jede Karte zeigte einen Befund
// „Schrift fehlt", ohne dass jemand sähe, warum.
//
// `internal/schriften` durchsucht die **ganze** Paketwurzel
// (`schriften/pakete.go`), also findet es sie, sobald sie da sind. Was fehlte,
// war das Herholen.
func TestKeineQuelleLaesstIhreSchriftenZurueck(t *testing.T) {
	t.Parallel()

	for _, quelle := range bezug.Quellen() {
		if quelle.Unterverzeichnis != "" {
			t.Errorf("%s holt nur %q — die Schriften daneben blieben zurück",
				quelle.Name, quelle.Unterverzeichnis)
		}
	}
}
