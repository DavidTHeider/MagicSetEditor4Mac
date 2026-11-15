// **Ein Exemplar ist ein Stück Papier** — Werkstück 2 von Phase 3.4
// (#1165).
//
// Was es ist und warum es **neben** der Karte steht, sagt
// [0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md).
package domain_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gutesExemplar gibt ein gültiges Exemplar, das jeder Fall abwandelt.
func gutesExemplar() domain.Exemplar {
	return domain.Exemplar{
		Printing: "8a1b0d5e-1e0f-4c0f-8b39-3d3a2f9a1c2d",
		Menge:    3,
		Zustand:  domain.ZustandNearMint,
		Sprache:  "en",
		Finish:   domain.FinishNonfoil,
	}
}

// **Ein gültiges Exemplar kommt durch.**
func TestEinGueltigesExemplarKommtDurch(t *testing.T) {
	t.Parallel()

	if befunde := gutesExemplar().Validate(); len(befunde) != 0 {
		t.Errorf("es gab Befunde: %+v", befunde)
	}
}

// **Was fehlen oder falsch sein kann, wird benannt — je Feld.**
//
// **Ein Befund je Feld und nicht eine Sammelmeldung**: §6 verlangt, dass ein
// Fehler am verursachenden Feld erscheint, und dafür muss er wissen, welches
// das ist.
func TestExemplarValidateBefunde(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		aendern func(*domain.Exemplar)
		feld    string
	}{
		"ohne Druck":            {func(e *domain.Exemplar) { e.Printing = "" }, "printing"},
		"Druck nur Leerraum":    {func(e *domain.Exemplar) { e.Printing = "   " }, "printing"},
		"Menge null":            {func(e *domain.Exemplar) { e.Menge = 0 }, "menge"},
		"Menge negativ":         {func(e *domain.Exemplar) { e.Menge = -1 }, "menge"},
		"Zustand unbekannt":     {func(e *domain.Exemplar) { e.Zustand = "prima" }, "zustand"},
		"Sprache leer":          {func(e *domain.Exemplar) { e.Sprache = "" }, "sprache"},
		"Erwerbspreis negativ":  {func(e *domain.Exemplar) { e.Erwerbspreis = -1 }, "erwerbspreis"},
		"Erwerbswährung fehlt":  {func(e *domain.Exemplar) { e.Erwerbspreis = 100 }, "erwerbswaehrung"},
		"Erwerbswährung allein": {func(e *domain.Exemplar) { e.Erwerbswaehrung = "EUR" }, "erwerbspreis"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			exemplar := gutesExemplar()
			fall.aendern(&exemplar)

			befunde := exemplar.Validate()
			if len(befunde) != 1 {
				t.Fatalf("es gab %d Befunde: %+v", len(befunde), befunde)
			}

			if befunde[0].Field != fall.feld {
				t.Errorf("der Befund hängt an %q statt an %q",
					befunde[0].Field, fall.feld)
			}

			if !errors.Is(befunde[0], domain.ErrValidation) {
				t.Errorf("der Befund ist kein ErrValidation: %v", befunde[0])
			}
		})
	}
}

// **Ein Zustand ohne Angabe ist `nearMint` — und nicht leer.**
//
// # Warum eine Vorgabe und keine Pflicht
//
// Wer eine Sammlung von Hand einträgt, will die Menge tippen und nicht bei
// jeder Zeile über den Zustand nachdenken. **Der häufigste Fall ist die
// Vorgabe**; wer es genauer weiß, ändert ihn.
func TestEinExemplarOhneZustandIstNearMint(t *testing.T) {
	t.Parallel()

	exemplar := gutesExemplar()
	exemplar.Zustand = ""

	exemplar.Normalisieren()

	if exemplar.Zustand != domain.ZustandNearMint {
		t.Errorf("der Zustand ist %q", exemplar.Zustand)
	}
}

// **Und eine Sprache ohne Angabe ist `en`** — die Sprache, in der Scryfall
// einen Druck ohne weitere Angabe führt.
func TestEinExemplarOhneSpracheIstEnglisch(t *testing.T) {
	t.Parallel()

	exemplar := gutesExemplar()
	exemplar.Sprache = ""

	exemplar.Normalisieren()

	if exemplar.Sprache != "en" {
		t.Errorf("die Sprache ist %q", exemplar.Sprache)
	}
}

// **Leerraum um den Druck geht weg.**
//
// Eine Kennung, die aus einer Zwischenablage kommt, trägt oft einen
// Zeilenumbruch — und zwei Exemplare desselben Drucks, von denen eines ein
// Leerzeichen mehr hat, wären zwei Zeilen statt einer.
func TestEinExemplarGlaettetSeinenDruck(t *testing.T) {
	t.Parallel()

	exemplar := gutesExemplar()
	exemplar.Printing = "  " + exemplar.Printing + "\n"

	exemplar.Normalisieren()

	if exemplar.Printing != gutesExemplar().Printing {
		t.Errorf("der Druck ist %q", exemplar.Printing)
	}
}

// **Die Zustände sind eine Aufzählung, und sie ist die von Cardmarket.**
func TestDieZustaendeSindDieVonCardmarket(t *testing.T) {
	t.Parallel()

	erwartet := []domain.Zustand{
		domain.ZustandMint, domain.ZustandNearMint, domain.ZustandExcellent,
		domain.ZustandGood, domain.ZustandLightPlayed, domain.ZustandPlayed,
		domain.ZustandPoor,
	}

	gefunden := domain.Zustaende()

	if len(gefunden) != len(erwartet) {
		t.Fatalf("es sind %d Zustände: %v", len(gefunden), gefunden)
	}

	for i, zustand := range erwartet {
		if gefunden[i] != zustand {
			t.Errorf("an Stelle %d steht %q statt %q", i, gefunden[i], zustand)
		}
	}
}

// **Die Liste ist eine Kopie.** Wer sie ändert, ändert nicht die Aufzählung.
func TestDieZustandslisteIstEineKopie(t *testing.T) {
	t.Parallel()

	erste := domain.Zustaende()
	erste[0] = "verbogen"

	if domain.Zustaende()[0] != domain.ZustandMint {
		t.Error("die Aufzählung hat sich ändern lassen")
	}
}

// **Ein Wunsch ist kein Besitz** — Vermerk
// [0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md).
//
// # Warum das eine Methode ist und kein `!e.Wunsch` an dreißig Stellen
//
// Weil die Regel an **einer** Stelle gehört. Sie hat drei Folgen — das
// Portfolio zählt Wünsche nicht, „besitze ich das?" sagt Nein, und die
// Eindeutigkeit nimmt sie auf —, und wer sie verteilt, hat sie irgendwann an
// einer davon anders.
func TestEinWunschIstKeinBesitz(t *testing.T) {
	t.Parallel()

	if !(domain.Exemplar{}).Besitzt() {
		t.Error("eine Zeile ohne Wunschmarke gilt als kein Besitz")
	}

	if (domain.Exemplar{Wunsch: true}).Besitzt() {
		t.Error("ein Wunsch gilt als Besitz")
	}
}

// **Der Lagerort wird geglättet wie jeder andere Freitext.**
func TestDerLagerortWirdGeglaettet(t *testing.T) {
	t.Parallel()

	exemplar := domain.Exemplar{Lagerort: "  Kiste 4  "}
	exemplar.Normalisieren()

	if exemplar.Lagerort != "Kiste 4" {
		t.Errorf("Lagerort %q", exemplar.Lagerort)
	}
}
