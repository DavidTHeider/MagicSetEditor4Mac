// **Die Ausführung an der Naht** — Werkstück `T2` von Phase 5.6
// (#1467).
//
// # Warum eine eigene Datei
//
// **Weil sie eine andere Frage beantwortet** (§2): nebenan steht der Weg einer
// Sammlung durch die Naht, hier steht die **eine** Eigenschaft, die seit T2
// drei Werte hat statt zwei — und zwar über alle vier Flächen hinweg (§16).
//
// # Was hier nicht steht
//
// **Ob die Kommandozeile dasselbe Wort schreibt.** Das prüft
// `internal/flaechen`; hier steht die Naht, an der jede Fläche hängt.
package app_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine geätzte Karte bleibt geätzt** — durch die Naht hin und zurück.
//
// **Das ist der ganze Anlass von T2**: `foil bool` konnte nur „ja" sagen, und
// eine geätzte Karte wurde damit zur Foil-Karte, sobald sie einmal durch die
// Naht gegangen war.
func TestEineGeaetzteKarteBleibtGeaetzt(t *testing.T) {
	naht := sammlungsnaht(t)
	ctx := probekontext(t)

	if _, fehler := naht.ExemplarAnlegen(ctx, app.ExemplarDTO{
		Printing: "a-druck", Menge: 1, Finish: string(domain.FinishEtched),
	}); fehler != nil {
		t.Fatalf("ExemplarAnlegen: %v", fehler)
	}

	gelesen, fehler := naht.Sammlung(ctx)
	if fehler != nil {
		t.Fatalf("Sammlung: %v", fehler)
	}

	if len(gelesen) != 1 || gelesen[0].Finish != string(domain.FinishEtched) {
		t.Errorf("im Bestand steht %+v", gelesen)
	}
}

// **Ohne Angabe ist eine Karte nonfoil, und sie sagt es auch.**
//
// **Der Nullwert wird beantwortet und nicht durchgereicht**: wer die leere
// Zeichenkette an die Oberfläche gäbe, hätte dort ein viertes Wort für etwas,
// wofür es drei gibt.
func TestOhneAngabeKommtNonfoilZurueck(t *testing.T) {
	naht := sammlungsnaht(t)
	ctx := probekontext(t)

	if _, fehler := naht.ExemplarAnlegen(ctx, app.ExemplarDTO{
		Printing: "a-druck", Menge: 1,
	}); fehler != nil {
		t.Fatalf("ExemplarAnlegen: %v", fehler)
	}

	gelesen, fehler := naht.Sammlung(ctx)
	if fehler != nil {
		t.Fatalf("Sammlung: %v", fehler)
	}

	if len(gelesen) != 1 || gelesen[0].Finish != string(domain.FinishNonfoil) {
		t.Errorf("im Bestand steht %+v", gelesen)
	}
}

// **Eine Ausführung, die es nicht gibt, wird am Feld abgewiesen.**
//
// **Am Feld und nicht als anonymer Hinweis** (§6): wer sich vertippt hat, soll
// es dort sehen, wo er getippt hat.
func TestEinExemplarMitUnbekannterAusfuehrungWirdAbgewiesen(t *testing.T) {
	naht := sammlungsnaht(t)

	_, fehler := naht.ExemplarAnlegen(probekontext(t), app.ExemplarDTO{
		Printing: "a-druck", Menge: 1, Finish: "glitzernd",
	})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	var befund *domain.FieldError
	if !errors.As(fehler, &befund) || befund.Field != "finish" {
		t.Errorf("der Befund ist %+v", fehler)
	}
}

// **Und die Sicht schneidet nach der dritten Ausführung.**
//
// **Sonst hätte der Bestand drei Werte und die Sicht zwei** — und eine
// geätzte Karte wäre über den Filter nicht mehr zu finden.
func TestDieSichtFindetDieGeaetzteKarte(t *testing.T) {
	naht := sammlungsnaht(t)
	ctx := probekontext(t)

	for _, dto := range []app.ExemplarDTO{
		{Printing: "a-druck", Menge: 1, Finish: string(domain.FinishEtched)},
		{Printing: "b-druck", Menge: 1, Finish: string(domain.FinishFoil)},
		{Printing: "c-druck", Menge: 1},
	} {
		if _, fehler := naht.ExemplarAnlegen(ctx, dto); fehler != nil {
			t.Fatalf("ExemplarAnlegen %s: %v", dto.Printing, fehler)
		}
	}

	for _, fall := range []struct {
		wie      domain.Ausfuehrung
		erwartet string
	}{
		{domain.AusfuehrungEtched, "a-druck"},
		{domain.AusfuehrungFoil, "b-druck"},
		{domain.AusfuehrungNonfoil, "c-druck"},
	} {
		t.Run(string(fall.wie), func(t *testing.T) {
			fenster, _, fehler := naht.Sammlungsfenster(ctx,
				domain.Sammlungssicht{Ausfuehrung: fall.wie}, 0, 10)
			if fehler != nil {
				t.Fatalf("Sammlungsfenster: %v", fehler)
			}

			if len(fenster) != 1 || fenster[0].Printing != fall.erwartet {
				t.Errorf("der Ausschnitt ist %+v", fenster)
			}
		})
	}
}

// **Wer beides schickt, hat `finish` gemeint.**
//
// **Die Regel steht in der Naht und nicht an jeder Fläche** (§16). Dieser Lauf
// ist der Grund, warum sie dort steht: MCP, REST und die Kommandozeile führen
// dieselben zwei Felder, und drei Flächen, die den Vorrang je selbst rechnen,
// rechnen ihn irgendwann verschieden.
func TestFinishSchlaegtDenAltenWahrheitswert(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		finish   string
		foil     bool
		erwartet string
	}{
		"nur der alte Wert, wahr":   {"", true, "foil"},
		"nur der alte Wert, falsch": {"", false, "nonfoil"},
		"beide, und sie stimmen":    {"foil", true, "foil"},

		// **Der Fall, für den es die Regel gibt**: ein Gegenüber schickt die
		// geätzte Ausführung und dazu den alten Wert, der sie nicht sagen
		// kann. Ohne den Vorrang wäre die Karte gewöhnlich.
		"beide, und sie streiten": {"etched", false, "etched"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if wie := app.AusfuehrungAus(fall.finish, fall.foil); wie != fall.erwartet {
				t.Errorf("es kam %q statt %q", wie, fall.erwartet)
			}
		})
	}
}
