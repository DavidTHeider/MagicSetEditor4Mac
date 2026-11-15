package domain_test

// **Auch ein KI-Anbieter hat einen Vertrag** — Werkstück A8 von Phase 5.3
// (#1447).
//
// # Warum er nicht fehlen darf
//
// Anthropic hat seit A8 einen **Schalter** wie jeder Kartendienst. Ohne
// Vertrag stünde er in den Einstellungen da, und der Kasten daneben — »was
// jeder Dienst hier kann und unter welchen Bedingungen« — schwiege über ihn.
// Die Hilfe verweist auf genau diesen Kasten; ein Verweis auf eine Leerstelle
// ist eine Zusage, die ihr Rumpf nicht hält (§10).

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestAnthropicHatEinenVertragFuerDenVorschlag(t *testing.T) {
	t.Parallel()

	vertrag, steht := domain.VertragFuer(domain.DienstAnthropic, domain.FrageKIVorschlag)
	if !steht {
		t.Fatal("Anthropic hat keinen Vertrag für den Vorschlag")
	}

	if vertrag.Basisadresse != "https://api.anthropic.com" {
		t.Errorf("die Basisadresse ist %q", vertrag.Basisadresse)
	}

	// **Beide Wege der Positivliste**, und keiner mehr: was das Blatt rufen
	// kann, steht in `anthropic.Wege()`, und der Wächter in
	// `internal/werkzeuge` hält sie gegen den Quelltext.
	if !slices.Equal(vertrag.Endpunkte, []string{"/v1/models", "/v1/messages"}) {
		t.Errorf("die Endpunkte sind %v", vertrag.Endpunkte)
	}

	// **Nie gegen sein Original gelaufen**, und das steht als `nil` da und
	// nicht als beschönigende Angabe: ein Lauf brauchte ein Konto mit
	// Guthaben, und dieses Programm hat keines.
	if vertrag.Nachweis != nil {
		t.Errorf("der Vertrag trägt einen Nachweis: %+v", vertrag.Nachweis)
	}
}

// **Der Vorschlag ist mehrstimmig und bekommt darum keine Wahl.**
//
// # Warum das die richtige Einordnung ist
//
// Bei der **Suche** antwortet genau einer, und zwei Trefferlisten wären ein
// Widerspruch. Beim **Vorschlag** stehen die Modelle aller eingeschalteten
// Anbieter nebeneinander im Katalog, und wer fragt, nennt Anbieter **und**
// Modell in derselben Bedienhandlung. Eine Einstellung »frage.kivorschlag«
// wäre ein Schalter, den keine Stelle im Code liest (§10).
func TestDerVorschlagBekommtKeineWahl(t *testing.T) {
	t.Parallel()

	if !domain.FrageKIVorschlag.Mehrstimmig() {
		t.Error("der Vorschlag gilt als einstimmig")
	}

	if slices.Contains(domain.Fragen(), domain.FrageKIVorschlag) {
		t.Error("»frage.kivorschlag« steht zur Wahl")
	}

	if slices.Contains(domain.Einstellungsreihe(),
		domain.FrageEinstellung(domain.FrageKIVorschlag)) {
		t.Error("»frage.kivorschlag« hat eine Einstellung")
	}
}
