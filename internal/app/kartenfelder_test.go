package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Feldbeschreibungen kommen im Frontend an**
// (#560).
//
// Der Karteneditor stellte jedes Feld als einzeiliges Textfeld dar — den
// Regeltext, die Seltenheit, die Manakosten. Was ein Feld ist, steht im
// Spielpaket und kam nie an: `WertDTO` trägt nur Name und Wert.
func TestKartenfelderNennenIhreForm(t *testing.T) {
	waehler := &stilwaehler{felder: []domain.Field{
		{
			Name: "rule text", Kind: domain.KindText, Mehrzeilig: true,
			Beschreibung: "Was die Karte tut.",
		},
		{Name: "rarity", Kind: domain.KindChoice, Choices: []string{"common", "rare"}},
	}}

	anwendung := app.NeueApp("")
	anwendung.MitStilwahl(waehler)

	antwort := anwendung.Kartenfelder(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Felder) != 2 {
		t.Fatalf("%d Felder", len(antwort.Felder))
	}

	regel := antwort.Felder[0]
	if regel.Art != "text" || !regel.Mehrzeilig ||
		regel.Beschreibung != "Was die Karte tut." {
		t.Errorf("%+v", regel)
	}

	seltenheit := antwort.Felder[1]
	if seltenheit.Art != "choice" || len(seltenheit.Wahlen) != 2 {
		t.Errorf("%+v", seltenheit)
	}

	// **Eine leere Liste und kein `nil`.** Ohne sie stünde im JSON `null`, und
	// das Frontend prüfte auf zwei Dinge statt auf eines.
	if regel.Wahlen == nil {
		t.Error("die Wahlen eines Textfeldes sind nil statt leer")
	}
}

// **Eine unlesbare Kennung wird abgewiesen.**
func TestKartenfelderWeisenEineUnlesbareKennungAb(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitStilwahl(&stilwaehler{})

	befund := anwendung.Kartenfelder(t.Context(), "abc").Fehler
	if befund == nil || befund.Feld != "id" {
		t.Errorf("%+v", befund)
	}
}

// **Und ein Lesefehler kommt als Befund heraus.**
func TestKartenfelderMeldenEinenLesefehler(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitStilwahl(&stilwaehler{fehler: errors.New("kein Spielpaket")})

	antwort := anwendung.Kartenfelder(t.Context(), "1")
	if antwort.Fehler == nil {
		t.Fatal("kein Befund")
	}

	if len(antwort.Felder) != 0 {
		t.Errorf("%d Felder trotz Fehler", len(antwort.Felder))
	}
}

// **Ein berechnetes Feld nennt seine Teile** —
// #696.
//
// `text` trägt `save value: false` und entsteht aus Regeltext und Flavortext.
// Der Editor bot es als Textkasten an; wer hineintippte, schrieb einen Wert, den
// die Vorlage nie liest. Jetzt kommt an, woraus es besteht — und die Oberfläche
// bietet **diese** Felder an.
func TestKartenfelderNennenIhreTeile(t *testing.T) {
	waehler := &stilwaehler{felder: []domain.Field{
		{
			Name: "text", Kind: domain.KindText, Gespeichert: false,
			Skript: `combined_editor(field1: card.rule_text, ` +
				`separator: "<line>\n</line>", field2: card.flavor_text)`,
		},
		{Name: "rule text", Kind: domain.KindText, Gespeichert: true},
		{Name: "flavor text", Kind: domain.KindText, Gespeichert: true},
	}}

	anwendung := app.NeueApp("")
	anwendung.MitStilwahl(waehler)

	antwort := anwendung.Kartenfelder(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if strings.Join(antwort.Felder[0].Teilfelder, "|") != "rule text|flavor text" {
		t.Errorf("Teile = %q", antwort.Felder[0].Teilfelder)
	}

	// **Eine leere Liste und kein `nil`** — dieselbe Begründung wie bei den
	// Wahlen.
	if antwort.Felder[1].Teilfelder == nil {
		t.Error("die Teile eines gewöhnlichen Feldes sind nil statt leer")
	}
}
