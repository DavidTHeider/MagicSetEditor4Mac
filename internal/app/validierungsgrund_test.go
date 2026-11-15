package app_test

// **Eine Validierung ohne Schlüssel sagt trotzdem, was schiefging** —
// #1601.
//
// # Der Befund, für den es diesen Lauf gibt
//
// **Gesehen am 2026-09-17 an der laufenden Anwendung**: eine Partie mit einem
// leeren Deck angefangen, und am Fenster stand „Es ist ein unerwarteter
// Fehler aufgetreten. Was genau, steht im Protokoll." **Im Protokoll stand
// nichts** — der einzige `ERROR`-Eintrag der Datei war zwei Tage alt.
//
// **Und das war kein Versehen im Protokoll**, sondern eine Zusage, die an der
// falschen Stelle stand: `TestErwarteteFehlerStehenNichtImProtokoll` hält
// ausdrücklich fest, dass eine Validierung **nicht** mitgeschrieben wird — ein
// Protokoll, das jede Eingabeprüfung führt, verdeckt den einen Eintrag, um
// dessentwillen es da ist. Falsch war der **Satz**: er verwies auf ein
// Protokoll, das nach eigener Entscheidung leer bleibt (§10).
//
// **Der Satz aus Go war die ganze Zeit da** und wurde weggeworfen: „ein
// Deckabzug ohne Karten ist kein Deck" stand in `meldung`, und der Katalog
// kennt den Schlüssel `unbekannt` — deshalb griff der Rückfall nicht, den
// `frontend/src/sprache/de/fehler.ts` in seinem Kopf zusagt.

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Validierung ohne eigenen Schlüssel trägt ihren Grund als Wert.**
func TestEineValidierungOhneSchluesselTraegtIhrenGrund(t *testing.T) {
	t.Parallel()

	fehler := app.Uebersetze(t.Context(), fmt.Errorf(
		"%w: ein Deckabzug ohne Karten ist kein Deck", domain.ErrValidation))

	switch {
	case fehler.Code != app.CodeValidierung:
		t.Errorf("der Code ist %q", fehler.Code)
	case fehler.Schluessel != domain.SchluesselValidierungsgrund:
		t.Errorf("der Schlüssel ist %q", fehler.Schluessel)
	case !strings.Contains(fehler.Werte["grund"], "kein Deck"):
		t.Errorf("der Grund steht nicht in den Werten: %+v", fehler.Werte)
	case !strings.Contains(fehler.Meldung, "kein Deck"):
		t.Errorf("die Meldung ist %q", fehler.Meldung)
	}
}

// **Und der Wortlaut von `ErrValidation` steht nicht darin.**
//
// **Weil er nichts erklärt.** „Validierung fehlgeschlagen" ist die Marke, an
// der `errors.Is` die Art erkennt — für einen Menschen ist sie ein Wort
// zuviel vor dem Satz, der ihm sagt, was zu tun ist (§8).
func TestDieMarkeStehtNichtImGrund(t *testing.T) {
	t.Parallel()

	fehler := app.Uebersetze(t.Context(), fmt.Errorf(
		"%w: ein Deckabzug ohne Karten ist kein Deck", domain.ErrValidation))

	if strings.Contains(fehler.Werte["grund"], "Validierung fehlgeschlagen") {
		t.Errorf("die Marke steht im Grund: %q", fehler.Werte["grund"])
	}
}

// **Und auch dann nicht, wenn sie mitten im Satz steht** — gemessen am
// 2026-09-17 an der laufenden Anwendung.
//
// # Warum das der Normalfall ist und nicht der Sonderfall
//
// **Weil jeder Aufrufer seinen eigenen Vorsatz davorsetzt.** Der Fachkern
// wirft `fmt.Errorf("%w: ein Deckabzug ohne Karten ist kein Deck", …)`, und
// der Aufbau darüber umschließt ihn mit `der Sitz "sitz1": %w`. Was am
// Fenster ankam, war:
//
//	Das ging nicht: der Sitz "sitz1": Validierung fehlgeschlagen:
//	ein Deckabzug ohne Karten ist kein Deck.
//
// **Gezählt am selben Tag**: 56 Stellen setzen die Marke nach vorn, 190
// umschließen einen Fehler mit eigenem Vorsatz. Der Lauf darüber prüfte die
// Minderheit — eine Vorrichtung, die enger ist als der Betrieb (§3).
//
// **Der frühere Kommentar nannte es eine Entscheidung** („wo sie nicht vorn
// steht, bleibt alles stehen"), und das war sie auch — nur ohne diese
// Messung. Eine Marke ist an jeder Stelle eine Marke: sie steht da, damit
// `errors.Is` greift, und für einen Menschen ist sie überall ein Wort zuviel.
func TestDieMarkeStehtAuchNichtInDerMitte(t *testing.T) {
	t.Parallel()

	fehler := app.Uebersetze(t.Context(), fmt.Errorf("der Sitz %q: %w",
		"sitz1", fmt.Errorf("%w: ein Deckabzug ohne Karten ist kein Deck",
			domain.ErrValidation)))

	grund := fehler.Werte["grund"]

	switch {
	case strings.Contains(grund, "Validierung fehlgeschlagen"):
		t.Errorf("die Marke steht mitten im Grund: %q", grund)

	// **Und die Gegenprobe**: ohne sie wäre ein Weg grün, der den Satz
	// einfach wegwirft (§10).
	case grund != `der Sitz "sitz1": ein Deckabzug ohne Karten ist kein Deck`:
		t.Errorf("der Grund lautet %q", grund)
	}
}

// **Ein Fehler mit eigenem Schlüssel behält ihn.**
//
// **Die Gegenprobe** (§10): ohne sie sagte der Lauf darüber nur, dass
// **irgendein** Schlüssel gesetzt wird — und ein Weg, der jeden Fehler auf
// `validierungsgrund` stellte, käme genauso durch.
//
// **Und je Klasse einer** — NET-02, S14.NAT-02 (#1706): keine Route, keine
// Erlaubnis und eine gescheiterte Schlüsselablage sind drei Ursachen mit
// drei Ratschlägen.
func TestEinEigenerSchluesselBleibtStehen(t *testing.T) {
	t.Parallel()

	for ursache, erwartet := range map[error]domain.Schluessel{
		domain.ErrOhneNetzerlaubnis:    domain.SchluesselOhneNetzerlaubnis,
		domain.ErrOhneNetzroute:        domain.SchluesselOhneNetzroute,
		domain.ErrSchluesselbundAblage: domain.SchluesselSchluesselbundAblage,
	} {
		fehler := app.Uebersetze(t.Context(), fmt.Errorf("%w: kein Weg hinaus", ursache))

		if fehler.Schluessel != erwartet {
			t.Errorf("%v gab den Schlüssel %q", ursache, fehler.Schluessel)
		}
	}
}

// **Und ein Feldbefund bleibt ein Feldbefund.**
func TestEinFeldbefundBehaeltSeinFeld(t *testing.T) {
	t.Parallel()

	fehler := app.Uebersetze(t.Context(), domain.NewFieldError(
		"name", domain.SchluesselPflichtfeld, "leer"))

	switch {
	case fehler.Schluessel != domain.SchluesselPflichtfeld:
		t.Errorf("der Schlüssel ist %q", fehler.Schluessel)
	case fehler.Feld != "name":
		t.Errorf("das Feld ist %q", fehler.Feld)
	}
}

// **Ein unbekannter Fehler nennt den Griff ins Protokoll** — der letzte
// offene Punkt aus #1601.
//
// # Warum ein anderer Satz und nicht ein angehängter
//
// **Weil der Rückfallsatz auf das Protokoll verweist** („unter »Hilfe →
// Diagnose« siehst du sein Ende"), und ein Protokoll hat viele Zeilen. Die
// Vorgangskennung steht in jeder Zeile dieses Rufs — sie ist der Griff, mit
// dem ein Mensch die richtige findet. Ohne sie ist der Verweis ein Hinweis
// auf eine Datei, nicht auf einen Eintrag.
//
// **Und die Entscheidung fällt hier und nicht im Fenster**: das Fenster
// übersetzt einen Schlüssel, es sucht sich keinen aus (§9).
func TestEinUnbekannterFehlerNenntSeinenVorgang(t *testing.T) {
	t.Parallel()

	ctx := app.MitVorgang(t.Context(), "abc12345")

	fehler := app.Uebersetze(ctx, errors.New("irgendwas aus dem Unterbau"))

	switch {
	case fehler.Schluessel != domain.SchluesselUnbekanntMitVorgang:
		t.Errorf("der Schlüssel ist %q", fehler.Schluessel)
	case fehler.Werte["vorgang"] != "abc12345":
		t.Errorf("die Werte sind %v", fehler.Werte)
	}
}

// **Ohne Vorgang bleibt es beim alten Satz** — die erste Gegenprobe.
//
// **Ohne sie sagte der Lauf darüber wenig**: ein Weg, der **jeden**
// unbekannten Fehler umbenennt, käme genauso durch — und stellte dann einen
// Satz mit einem leeren Griff hin (§10).
func TestOhneVorgangBleibtDerAlteSchluessel(t *testing.T) {
	t.Parallel()

	fehler := app.Uebersetze(t.Context(), errors.New("irgendwas"))

	if fehler.Schluessel != domain.SchluesselUnbekannt {
		t.Errorf("der Schlüssel ist %q", fehler.Schluessel)
	}
}

// **Und ein eingeordneter Fehler bekommt ihn nicht** — die zweite.
//
// **Sonst trüge jede Meldung eine Hexzahl.** Die Kennung gehört zu dem einen
// Satz, der ohne sie ins Leere zeigt; an einem Satz, der selbst sagt, was zu
// tun ist, wäre sie Lärm.
func TestEinEingeordneterFehlerTraegtKeineKennungImSatz(t *testing.T) {
	t.Parallel()

	ctx := app.MitVorgang(t.Context(), "abc12345")

	fehler := app.Uebersetze(ctx, fmt.Errorf(
		"%w: gibt es nicht", domain.ErrNotFound))

	switch {
	case fehler.Schluessel != domain.SchluesselNichtGefunden:
		t.Errorf("der Schlüssel ist %q", fehler.Schluessel)
	case fehler.Werte["vorgang"] != "":
		t.Errorf("die Werte sind %v", fehler.Werte)
	}
}
