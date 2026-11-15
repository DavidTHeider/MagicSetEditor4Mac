package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// wortlistenbauer ist ein Planbauer, der eine feste Auswahl gibt.
type wortlistenbauer struct {
	ohneWahlbilder

	bereiche []service.Wortlistenbereich
	wert     string
	fehler   error

	feld    string
	bereich int
	weg     []int
}

func (*wortlistenbauer) Kartenplan(context.Context, *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	return &render.Plan{Breite: 2, Hoehe: 2}, nil, nil
}

func (*wortlistenbauer) Kartenpalette(context.Context, *domain.CardRow, string) (
	[]service.Paletteneintrag, bool, []domain.Befund, error,
) {
	return nil, false, nil, nil
}

func (w *wortlistenbauer) Kartenwortlisten(context.Context, *domain.CardRow, string) (
	[]service.Wortlistenbereich, []domain.Befund, error,
) {
	return w.bereiche, nil, w.fehler
}

func (w *wortlistenbauer) Kartenwortwahl(
	_ context.Context, _ *domain.CardRow, feld string, bereich int, weg []int,
) (string, []domain.Befund, error) {
	w.feld, w.bereich, w.weg = feld, bereich, weg

	return w.wert, nil, w.fehler
}

// mitWortlisten verdrahtet eine Anwendung um diesen Planbauer.
func mitWortlisten(
	bauer *wortlistenbauer, leser karteprobe, schreiber *schreibprobe,
) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitEditor(schreiber, leser, bauer, render.NeuerRasterer(ohneBilder{}))

	return anwendung
}

// eineAuswahl baut einen Bereich mit einem Präfix, einem Punkt und einer Gruppe.
func eineAuswahl() []service.Wortlistenbereich {
	unten := service.Wortmenue{
		Gewaehlt:  service.KeineWahl,
		Eintraege: []service.Wortmenueeintrag{{Name: "Plane"}},
	}

	return []service.Wortlistenbereich{{
		Liste: "type-en",
		Text:  "Legendary Creature",
		Menue: service.Wortmenue{
			Kaestchen: true,
			Gewaehlt:  1,
			Eintraege: []service.Wortmenueeintrag{
				{Name: "Legendary ", IstPraefix: true, ZeileDarunter: true, Aktiv: true},
				{Name: "Creature", Aktiv: true},
				{Name: "Nichtstandard", IstGruppe: true, Untermenue: &unten},
			},
		},
	}}
}

// **Die Auswahl kommt als DTO an, samt Untermenü** —
// #1074.
func TestWortlistenGebenDieAuswahlWeiter(t *testing.T) {
	anwendung := mitWortlisten(&wortlistenbauer{bereiche: eineAuswahl()},
		karteprobe{karte: &domain.CardRow{ID: 1}}, &schreibprobe{})

	antwort := anwendung.Wortlisten(t.Context(), "1", "type")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Bereiche) != 1 {
		t.Fatalf("%d Bereiche", len(antwort.Bereiche))
	}

	bereich := antwort.Bereiche[0]
	if bereich.Liste != "type-en" || bereich.Text != "Legendary Creature" ||
		!bereich.Kaestchen || bereich.Gewaehlt != 1 {
		t.Errorf("der Bereich ist %+v", bereich)
	}

	if len(bereich.Eintraege) != 3 {
		t.Fatalf("%d Einträge", len(bereich.Eintraege))
	}

	erster := bereich.Eintraege[0]
	if !erster.IstPraefix || !erster.ZeileDarunter || !erster.Aktiv {
		t.Errorf("der erste Eintrag ist %+v", erster)
	}

	// **Eine Gruppe ist ein Eintrag mit Kindern** und trägt keine eigene
	// Marke; zwei Angaben über dasselbe liefen auseinander.
	if len(bereich.Eintraege[2].Kinder) != 1 ||
		bereich.Eintraege[2].Kinder[0].Name != "Plane" {
		t.Errorf("das Untermenü ist %+v", bereich.Eintraege[2])
	}

	if bereich.Eintraege[1].Kinder != nil {
		t.Errorf("ein gewöhnlicher Eintrag hat Kinder: %+v", bereich.Eintraege[1])
	}
}

// **Eine unlesbare Kennung, eine fehlende Karte und ein gescheiterter Planbau
// geben eine leere Liste und einen Fehler.**
//
// **Eine leere Liste und kein `nil`**, wie bei der Symbolpalette: über die
// Wails-Bindung stünde dort sonst `null`, und das Frontend prüfte auf zwei
// Dinge statt auf eines.
func TestWortlistenMeldenIhreFehler(t *testing.T) {
	faelle := []struct {
		name    string
		kennung string
		leser   karteprobe
		bauer   *wortlistenbauer
	}{
		{
			name: "keine Zahl", kennung: "abc",
			leser: karteprobe{karte: &domain.CardRow{ID: 1}}, bauer: &wortlistenbauer{},
		},
		{
			name: "die Karte fehlt", kennung: "1",
			leser: karteprobe{fehler: domain.ErrNotFound}, bauer: &wortlistenbauer{},
		},
		{
			name: "der Planbau scheitert", kennung: "1",
			leser: karteprobe{karte: &domain.CardRow{ID: 1}},
			bauer: &wortlistenbauer{fehler: errors.New("die Platte ist voll")},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			anwendung := mitWortlisten(fall.bauer, fall.leser, &schreibprobe{})

			antwort := anwendung.Wortlisten(t.Context(), fall.kennung, "type")

			if antwort.Fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}

			if antwort.Bereiche == nil {
				t.Error("nil statt einer leeren Liste")
			}
		})
	}
}

// **Eine Wahl rechnet und schreibt in einem Aufruf** (§7, Regel 4).
func TestEineWahlSchreibtDenGerechnetenWert(t *testing.T) {
	bauer := &wortlistenbauer{wert: "<word-list-type-en>Instant</word-list-type-en>"}
	schreiber := &schreibprobe{}

	anwendung := mitWortlisten(bauer, karteprobe{karte: &domain.CardRow{ID: 1}}, schreiber)

	antwort := anwendung.WortWaehlen(t.Context(), "2", "1", "type", 0, []int{3})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if bauer.feld != "type" || bauer.bereich != 0 || len(bauer.weg) != 1 || bauer.weg[0] != 3 {
		t.Errorf("der Planbauer bekam %q / %d / %v", bauer.feld, bauer.bereich, bauer.weg)
	}

	if schreiber.karteID != 1 || schreiber.feld != "type" || schreiber.wert != bauer.wert {
		t.Errorf("geschrieben wurde %d / %q / %q",
			schreiber.karteID, schreiber.feld, schreiber.wert)
	}
}

// **Und sie prüft, was sie bekommt, bevor sie schreibt.**
func TestEineWahlPrueftIhreAngaben(t *testing.T) {
	faelle := []struct {
		name           string
		setID, karteID string
		feld           string
		leser          karteprobe
		bauer          *wortlistenbauer
	}{
		{
			name: "ohne Set", setID: "", karteID: "1", feld: "type",
			leser: karteprobe{karte: &domain.CardRow{ID: 1}}, bauer: &wortlistenbauer{},
		},
		{
			name: "ohne Karte", setID: "2", karteID: "", feld: "type",
			leser: karteprobe{karte: &domain.CardRow{ID: 1}}, bauer: &wortlistenbauer{},
		},
		{
			name: "ohne Feldnamen", setID: "2", karteID: "1", feld: "",
			leser: karteprobe{karte: &domain.CardRow{ID: 1}}, bauer: &wortlistenbauer{},
		},
		{
			name: "die Karte fehlt", setID: "2", karteID: "1", feld: "type",
			leser: karteprobe{fehler: domain.ErrNotFound}, bauer: &wortlistenbauer{},
		},
		{
			name: "der Weg führt ins Leere", setID: "2", karteID: "1", feld: "type",
			leser: karteprobe{karte: &domain.CardRow{ID: 1}},
			bauer: &wortlistenbauer{fehler: domain.ErrValidation},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			schreiber := &schreibprobe{}
			anwendung := mitWortlisten(fall.bauer, fall.leser, schreiber)

			antwort := anwendung.WortWaehlen(
				probekontext(t), fall.setID, fall.karteID, fall.feld, 0, []int{0})

			if antwort.Fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}

			// **Und nichts wurde geschrieben.** Ein Wert, der nach einem
			// abgewiesenen Klick trotzdem in der Karte stünde, wäre schlimmer
			// als die Meldung.
			if schreiber.feld != "" {
				t.Errorf("geschrieben wurde %q = %q", schreiber.feld, schreiber.wert)
			}
		})
	}
}

// **Ein gescheitertes Schreiben kommt durch.**
func TestEineWahlMeldetEinGescheitertesSchreiben(t *testing.T) {
	schreiber := &schreibprobe{fehler: domain.ErrConflict}

	anwendung := mitWortlisten(&wortlistenbauer{wert: "X"},
		karteprobe{karte: &domain.CardRow{ID: 1}}, schreiber)

	if antwort := anwendung.WortWaehlen(
		probekontext(t), "2", "1", "type", 0, []int{0}); antwort.Fehler == nil {
		t.Error("der Fehler kam nicht durch")
	}
}

// ohneWortlisten ist die Antwort eines Planbauers, den keine Wortliste
// interessiert — **eingebettet und nicht je Probe geschrieben**: fünf gleiche
// Rümpfe in fünf Dateien liefen beim nächsten Schnittstellenzusatz auseinander.
type ohneWortlisten struct{}

func (ohneWortlisten) Kartenwortlisten(context.Context, *domain.CardRow, string) (
	[]service.Wortlistenbereich, []domain.Befund, error,
) {
	return nil, nil, nil
}

func (ohneWortlisten) Kartenwortwahl(
	context.Context, *domain.CardRow, string, int, []int,
) (string, []domain.Befund, error) {
	return "", nil, nil
}

// ohneWahlbilder ist die Antwort eines Planbauers, den die Wahlbilder einer
// Stiloption nicht interessieren
// (#1225).
//
// **Eingebettet und nicht je Probe geschrieben** — derselbe Grund wie bei
// [ohneWortlisten], und derselbe Ort, damit beide zusammen fortgeschrieben
// werden.
type ohneWahlbilder struct{}

func (ohneWahlbilder) Optionsbilder(context.Context, int64, string) (
	[]service.Optionsbild, []domain.Befund, error,
) {
	return nil, nil, nil
}
