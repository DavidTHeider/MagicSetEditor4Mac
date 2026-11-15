package domain_test

// **Was ein Modell belegen darf** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Der Satz, gegen den hier geprüft wird
//
// **„Unbekannte IDs/erfundene Regelnummern/URLs ergeben Validierungsbefund,
// keine vertrauenswürdige Quellenanzeige."** Eine erfundene Regelnummer ist
// kein Darstellungsproblem — sie fällt auf, bevor jemand den Satz liest.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zweiBelege sind die, die ein Paket vergeben hat.
func zweiBelege() []domain.Beleg {
	return []domain.Beleg{
		{
			Kennung: "B1", Quelle: domain.QuelleRegelwerk,
			Titel: "Regel 702.9b", Regelnummer: "702.9b",
			Ausschnitt: "Trample ändert, wie überschüssiger Kampfschaden zugewiesen wird.",
		},
		{
			Kennung: "B2", Quelle: domain.QuelleOracle,
			Titel: "Lightning Bolt", Bezugsadresse: "https://beispiel.invalid/o1",
			Ausschnitt: "Lightning Bolt deals 3 damage to any target.",
		},
	}
}

// **Eine Antwort, die nur vergebene Belege nennt, ist ohne Befund.**
func TestEineAntwortMitVergebenenBelegenIstOhneBefund(t *testing.T) {
	t.Parallel()

	befunde := domain.PruefeBelege(
		"Trampelschaden wird nach 702.9b zugewiesen [B1]; "+
			"Lightning Bolt macht 3 Schaden [B2].", zweiBelege())

	if len(befunde) != 0 {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}

// **Ein Beleg, den es nicht gibt, ist ein Befund.**
func TestEinUnbekannterBelegIstEinBefund(t *testing.T) {
	t.Parallel()

	befunde := domain.PruefeBelege("Das steht so in [B7].", zweiBelege())

	if len(befunde) != 1 || befunde[0].Art != domain.BefundUnbekannteKennung ||
		befunde[0].Stelle != "B7" {
		t.Fatalf("die Befunde sind %+v", befunde)
	}

	// **Zweimal derselbe ist ein Befund** — eine Liste, die mit der Zahl der
	// Erwähnungen wächst, sagt nichts über die Zahl der Fehler.
	zweimal := domain.PruefeBelege("[B7] und noch einmal [B7].", zweiBelege())
	if len(zweimal) != 1 {
		t.Fatalf("die Befunde sind %+v", zweimal)
	}
}

// **Eine Regelnummer, die in keinem Beleg steht, ist ein Befund.**
//
// **Und der Satz dazu ist vorsichtig**: sie mag es geben — belegt ist sie
// hier nicht. Eine Aussage über das Regelwerk kann dieses Programm nicht
// treffen, es hat keines.
func TestEineErfundeneRegelnummerIstEinBefund(t *testing.T) {
	t.Parallel()

	befunde := domain.PruefeBelege(
		"Nach Regel 123.4c gilt das nicht.", zweiBelege())

	if len(befunde) != 1 ||
		befunde[0].Art != domain.BefundErfundeneRegelnummer ||
		befunde[0].Stelle != "123.4c" {
		t.Fatalf("die Befunde sind %+v", befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "belegt ist sie hier nicht") {
		t.Errorf("die Meldung ist %q", befunde[0].Meldung)
	}

	// **Eine Zahl ohne Punkt ist keine Regelnummer.** „3 Schaden" darf nicht
	// als erfundene Regel gelten.
	if befunde := domain.PruefeBelege(
		"Das macht 3 Schaden.", zweiBelege()); len(befunde) != 0 {
		t.Errorf("eine Zahl wurde als Regelnummer gelesen: %+v", befunde)
	}
}

// **Eine Adresse, die in keinem Beleg steht, wird nicht als Quelle gezeigt.**
func TestEineErfundeneAdresseIstEinBefund(t *testing.T) {
	t.Parallel()

	befunde := domain.PruefeBelege(
		"Siehe https://erfunden.invalid/seite für Einzelheiten.", zweiBelege())

	if len(befunde) != 1 || befunde[0].Art != domain.BefundErfundeneAdresse {
		t.Fatalf("die Befunde sind %+v", befunde)
	}

	// **Die Adresse aus einem Beleg ist keine.**
	if befunde := domain.PruefeBelege(
		"Siehe https://beispiel.invalid/o1 [B2].",
		zweiBelege()); len(befunde) != 0 {
		t.Errorf("die belegte Adresse wurde beanstandet: %+v", befunde)
	}
}

// **Alle drei Arten kommen zusammen vor.**
func TestAlleDreiArtenKommenZusammenVor(t *testing.T) {
	t.Parallel()

	befunde := domain.PruefeBelege(
		"[B9] sagt nach 999.9z, siehe https://erfunden.invalid/x.",
		zweiBelege())

	if len(befunde) != 3 {
		t.Fatalf("die Befunde sind %+v", befunde)
	}

	arten := map[string]bool{}
	for _, befund := range befunde {
		arten[befund.Art] = true
	}

	for _, art := range []string{
		domain.BefundUnbekannteKennung, domain.BefundErfundeneRegelnummer,
		domain.BefundErfundeneAdresse,
	} {
		if !arten[art] {
			t.Errorf("die Art %q fehlt", art)
		}
	}
}

// **Und ein Satz, der sich auf nichts beruft, ist als solcher erkennbar.**
func TestEinSatzOhneBelegIstErkennbar(t *testing.T) {
	t.Parallel()

	if domain.NenntEinenBeleg("Das ist einfach so.") {
		t.Error("ein Satz ohne Beleg galt als belegt")
	}

	if !domain.NenntEinenBeleg("Das steht in [B1].") {
		t.Error("ein Satz mit Beleg galt als unbelegt")
	}
}

// **Eine Regelnummer im Ausschnitt eines Belegs zählt als belegt.**
//
// **Auch ohne eigenes Feld**: ein Ruling nennt die Regel im Text, und wer nur
// das Feld `Regelnummer` prüfte, beanstandete jede Nummer, die ein Ruling
// zitiert.
func TestEineRegelnummerImAusschnittZaehltAlsBelegt(t *testing.T) {
	t.Parallel()

	belege := []domain.Beleg{{
		Kennung: "B1", Quelle: domain.QuelleRuling,
		Ausschnitt: "Siehe dazu Regel 601.2 über das Wirken.",
	}}

	if befunde := domain.PruefeBelege(
		"Nach 601.2 geschieht das [B1].", belege); len(befunde) != 0 {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}
