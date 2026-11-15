// **Das Fenster der Sammlung an der Naht** — Werkstück
// #1500.
//
// # Warum eine eigene Datei
//
// `sammlungfuerflaechen_test.go` steht an der Zeilengrenze (§2), und der
// Schnitt folgt der Frage: dort der ganze Weg einer Sammlung, hier der
// **Ausschnitt** — die Fähigkeit, die Kommandozeile, MCP und REST seit #1500
// gemeinsam mit dem Fenster haben (§16).
package app_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Sicht wirkt, der Ausschnitt schneidet, und beide Zahlen stimmen.**
//
// Drei Exemplare, eines davon Foil: die Sicht sieht zwei, der Ausschnitt gibt
// eines, und `insgesamt` nennt weiter alle drei — sonst stünde in der Fußzeile
// „1 von 1", wo jemand drei Karten besitzt.
func TestDasSammlungsfensterSchneidetUndZaehlt(t *testing.T) {
	naht := sammlungsnaht(t)

	for _, dto := range []app.ExemplarDTO{
		{Printing: "b-druck", Menge: 1, Notiz: "Zauber"},
		{Printing: "a-druck", Menge: 4, Notiz: "Zauber"},
		{
			Printing: "c-druck", Menge: 2,
			Finish: string(domain.FinishFoil), Notiz: "Zauber",
		},
	} {
		if _, fehler := naht.ExemplarAnlegen(probekontext(t), dto); fehler != nil {
			t.Fatalf("ExemplarAnlegen %s: %v", dto.Printing, fehler)
		}
	}

	fenster, zahlen, fehler := naht.Sammlungsfenster(probekontext(t),
		domain.Sammlungssicht{
			Suche:       "Zauber",
			Ausfuehrung: domain.AusfuehrungNonfoil,
			Sortierung:  domain.SortierungMenge,
		}, 0, 1)
	if fehler != nil {
		t.Fatalf("Sammlungsfenster: %v", fehler)
	}

	if zahlen.Gesamt != 2 || zahlen.Insgesamt != 3 {
		t.Errorf("gesehen %d, insgesamt %d — erwartet waren 2 und 3", zahlen.Gesamt, zahlen.Insgesamt)
	}

	// **Nach Menge absteigend**: das Stück mit vieren steht vorn.
	if len(fenster) != 1 || fenster[0].Printing != "a-druck" {
		t.Errorf("der Ausschnitt ist %+v", fenster)
	}
}

// **Eine Sicht, die es nicht gibt, wird abgewiesen — vor dem Lesen.**
func TestDasSammlungsfensterWeistEineUnbekannteSichtAb(t *testing.T) {
	naht := sammlungsnaht(t)

	_, _, fehler := naht.Sammlungsfenster(probekontext(t),
		domain.Sammlungssicht{Ausfuehrung: "vielleicht"}, 0, 10)

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("abgewiesen wurde mit %v", fehler)
	}
}

// **Und ein geschlossener Bestand meldet sich auch hier.**
func TestDasSammlungsfensterMeldetEinenGeschlossenenBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	naht, err := app.SammlungsnahtUeber(db)
	if err != nil {
		t.Fatalf("SammlungsnahtUeber: %v", err)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, _, fehler := naht.Sammlungsfenster(probekontext(t),
		domain.Sammlungssicht{}, 0, 10); fehler == nil {
		t.Error("der geschlossene Bestand nahm den Aufruf an")
	}
}
