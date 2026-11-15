package domain_test

// **Die Unterschriftsform der Fassung 1 steht fest** — Werkstück `T1` von
// Phase 5.6
// (#1466),
// entschieden in [0117](../../docs/entscheidungen/0117-was-eine-handelsfassung-bindet.md).
//
// # Warum ein Goldlauf und kein Feldvergleich
//
// **Weil die Zusage die ganze Form ist und nicht ihre Teile.** Ein Lauf, der
// Feld für Feld prüfte, ginge bei einem **zusätzlichen** Feld durch — und
// genau das ist der Fall, den dieser hier fangen soll: wer ein Feld
// hinzufügt, ändert jeden Abdruck jeder Revision, die je geschrieben wurde.
//
// **Wird dieser Lauf rot, ist das kein Fehler, sondern eine Frage**: ist die
// Änderung gewollt? Dann steigt [domain.Handelsvertragsfassung], die Form der
// Fassung 1 bleibt hier stehen, und daneben kommt die neue.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Form der Fassung 1, Feld für Feld und in dieser Reihenfolge.**
func TestDieUnterschriftsformDerFassungEinsStehtFest(t *testing.T) {
	t.Parallel()

	if domain.Handelsvertragsfassung != 1 {
		t.Fatalf("dieser Goldlauf gilt der Fassung 1, es ist die %d — "+
			"die neue Form gehört daneben und nicht an diese Stelle",
			domain.Handelsvertragsfassung)
	}

	gold := strings.Join([]string{
		"handelsangebot", // das Domänenwort
		"1",              // Vertragsfassung
		"v1",             // Vorgang
		"1",              // Revision
		"",               // Eltern
		"tausch",         // Art
		"anna",           // Initiator
		"anna,bert",      // Parteien
		"false",          // BewertungGeteilt
		"1",              // Zahl der Positionen

		"anna",              // Position: Von
		"abc-123",           // Printing
		"0",                 // DruckID
		"2",                 // Menge
		"NM",                // Guete
		"de",                // Sprache
		"nonfoil",           // Finish
		"abdruck-der-zeile", // Abbild
	}, "\x00")

	if form := string(einAngebot().Unterschriftsform()); form != gold {
		t.Errorf("die Form lautet\n  %q\nerwartet war\n  %q",
			form, gold)
	}
}

// **Und mit geteilter Bewertung kommt genau ein Feld dazu.**
//
// **Am Ende der Position und nicht irgendwo** — eine Form, in der ein Feld
// seine Stelle wechselt, ist eine andere Form.
func TestDieGeteilteBewertungKommtAnsEndeDerPosition(t *testing.T) {
	t.Parallel()

	angebot := einAngebot()
	angebot.BewertungGeteilt = true
	angebot.Positionen[0].Bewertung = 4_200

	form := string(angebot.Unterschriftsform())

	if !strings.HasSuffix(form, "\x00abdruck-der-zeile\x004200") {
		t.Errorf("die Form endet mit %q", form[max(0, len(form)-40):])
	}

	// **Und der Schalter steht im Kopf** — sonst sagte die Unterschrift
	// nicht, worüber sie geschlossen wurde.
	if !strings.Contains(form, "\x00true\x00") {
		t.Error("der Schalter für die geteilte Bewertung steht nicht in der Form")
	}
}
