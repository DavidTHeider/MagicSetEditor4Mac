package service_test

// **Der Weg der Herkunft durch die Schichten** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Was hier gemessen wird
//
// **Dass die Herkunft ankommt, und zwar unverändert.** Zwischen dem
// Auftragswerk und der Ablage liegen zwei Stellen — `Feldweg` löst Werte auf,
// `Kartenaufsicht` prüft die Zugehörigkeit. Beide dürfen die Herkunft
// weiterreichen und keine von beiden darf sie verlieren.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Feldweg reicht Werte und Herkunft weiter.**
func TestDerFeldwegReichtDieHerkunftWeiter(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	weg := service.NeuerFeldweg(
		wegkarten{karte: &domain.CardRow{ID: 4, SetID: 3}},
		wegfelder{felder: m15felder()}, ablage)

	if _, _, fehler := weg.FelderVonAussenMitHerkunft(t.Context(), 4,
		[]domain.Value{{Field: "name", Text: "Blitzschlag"}},
		"ki", []domain.KIFeldherkunft{eineWegherkunft()}, nil, ""); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(ablage.feldherkuenfte) != 1 ||
		ablage.feldherkuenfte[0].Feld != "name" {
		t.Fatalf("angekommen ist %+v", ablage.feldherkuenfte)
	}

	// **Und die Marke am Feld ebenso** (#888): sie steht im Verlauf und sagt
	// „nicht getippt". Das eine ersetzt das andere nicht.
	if len(ablage.herkuenfte) != 1 || ablage.herkuenfte[0] != "ki" {
		t.Errorf("die Marken sind %v", ablage.herkuenfte)
	}
}

// **Ein Fehler beim Auflösen hält ihn auf.**
//
// **Vor dem Schreiben und nicht danach**: ein Wert, der sich nicht in die
// Teile zerlegen lässt, die der Bestand trägt, darf nicht halb ankommen.
func TestEinFehlerBeimAufloesenHaeltDenHerkunftswegAuf(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}

	if _, _, fehler := service.NeuerFeldweg(
		wegkarten{fehler: errors.New("die Karte klemmt")},
		wegfelder{felder: m15felder()}, ablage,
	).FelderVonAussenMitHerkunft(t.Context(), 4,
		[]domain.Value{{Field: "name", Text: "x"}},
		"ki", []domain.KIFeldherkunft{eineWegherkunft()}, nil, ""); fehler != nil {
		t.Fatalf("die unlesbare Vorlage hielt den Weg auf: %v", fehler)
	}

	// **Ohne lesbare Vorlage gehen die Werte roh hinaus** — dieselbe Duldung
	// wie beim gewöhnlichen Schreiben.
	if len(ablage.feldherkuenfte) != 1 {
		t.Errorf("angekommen ist %+v", ablage.feldherkuenfte)
	}
}

// **Ein unteilbarer Wert hält ihn dagegen auf.**
//
// **Vor dem Schreiben und nicht danach**: `type` rechnet seinen Trenner aus
// und lässt sich nicht spalten. Ein halb geschriebener Verbund wäre eine
// Karte, von der niemand sagen kann, welche Hälfte gilt.
func TestEinUnteilbarerWertHaeltDenHerkunftswegAuf(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}

	_, _, fehler := service.NeuerFeldweg(
		wegkarten{karte: &domain.CardRow{ID: 4, SetID: 3}},
		wegfelder{felder: m15felder()}, ablage,
	).FelderVonAussenMitHerkunft(t.Context(), 4,
		[]domain.Value{{Field: "type", Text: "Kreatur"}},
		"ki", []domain.KIFeldherkunft{eineWegherkunft()}, nil, "")
	if fehler == nil {
		t.Fatal("der unteilbare Wert ging durch")
	}

	if len(ablage.feldherkuenfte) != 0 {
		t.Errorf("geschrieben wurde trotzdem %+v", ablage.feldherkuenfte)
	}
}

// **Die Aufsicht prüft die Zugehörigkeit — und reicht dann weiter.**
func TestDieAufsichtPrueftVorDerHerkunft(t *testing.T) {
	t.Parallel()

	felder := &feldprobe{}
	aufsicht := service.NeueKartenaufsicht(
		&kartensicht{karten: map[int64]int64{4: 2}},
		&lebenszyklusprotokoll{}).MitFeldweg(felder, felder)

	if _, _, fehler := aufsicht.FelderVonAussenMitHerkunft(t.Context(), 2, 4,
		[]domain.Value{{Field: "name", Text: "Blitzschlag"}},
		"ki", []domain.KIFeldherkunft{eineWegherkunft()}, nil, ""); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(felder.feldherkuenfte) != 1 {
		t.Fatalf("angekommen ist %+v", felder.feldherkuenfte)
	}

	// **Und eine fremde Karte kommt gar nicht bis zum Feldweg.**
	if _, _, fehler := aufsicht.FelderVonAussenMitHerkunft(t.Context(), 9, 4,
		[]domain.Value{{Field: "name", Text: "x"}},
		"ki", []domain.KIFeldherkunft{eineWegherkunft()}, nil, ""); fehler == nil {
		t.Fatal("die fremde Karte wurde geschrieben")
	}

	if len(felder.feldherkuenfte) != 1 {
		t.Errorf("angekommen ist %+v — die fremde Karte kam durch",
			felder.feldherkuenfte)
	}
}

// eineWegherkunft baut eine vollständige Herkunft.
func eineWegherkunft() domain.KIFeldherkunft {
	return domain.KIFeldherkunft{
		Feld: "name", Anbieter: "ollama", Modell: "gemma4:8b-q4",
		Aufgabe: "Schreibe den Regeltext.", Auftrag: "ki-4-1",
		Zeitpunkt: time.Now().UTC(),
		Bezug:     domain.KIInhaltsbezug([]string{"Blitzschlag"}),
	}
}
