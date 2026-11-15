package app

// **Ein Symbol der Bibliothek in den Fassungen eines Sets** — DOM-02.AK-02
// (#1697):
// die Größen- und Farbvorschau, gezeichnet vom Rasterer der Karten.
//
// **Ein innerer Lauf**, weil die Vorschau Rasterer und PNG-Kodierung der
// Anwendung braucht — und beide Nähte scheitern im Betrieb nicht.

import (
	"context"
	"errors"
	"image"
	"io"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// festeFassungen nennt immer dieselben Fassungen.
type festeFassungen struct {
	fassungen []msepaket.Symbolfassung
	fehler    error
}

func (f festeFassungen) Symbolfassungen(context.Context, int64) ([]msepaket.Symbolfassung, error) {
	return f.fassungen, f.fehler
}

// keineBilder ist eine Bildquelle ohne Bilder — die Vorschau braucht keine.
type keineBilder struct{}

func (keineBilder) Bild(string) (image.Image, error) { return nil, errors.New("keine Bilder") }

// klemmenderRasterer scheitert an jedem Plan — alles andere kann er.
type klemmenderRasterer struct{ *render.Rasterer }

func (klemmenderRasterer) Zeichne(context.Context, *render.Plan, float64, render.Guete) (*image.NRGBA, []render.Befund, error) {
	return nil, nil, errors.New("klemmt")
}

// mitFassungen baut die gebundene Anwendung mit einem Symbol und einer Quelle
// der Fassungen.
func mitFassungen(t *testing.T, quelle festeFassungen) (anwendung *App, id, setID string) {
	t.Helper()

	var gewaehlt string

	bindungen, setID := mitSymbolbibliothek(t, &gewaehlt, nil)
	anwendung = bindungen.app(t.Context())

	bibliothek, ist := anwendung.symbolbibliothek.(*service.Symbolbibliothek)
	if !ist {
		t.Fatalf("die Bibliothek ist %T", anwendung.symbolbibliothek)
	}

	anwendung.MitSymbolbibliothek(bibliothek.MitFassungen(quelle))
	anwendung.rasterer = render.NeuerRasterer(keineBilder{})

	return anwendung, bindungen.BibliothekssymbolAufnehmen(t.Context(), setID, "Krone").ID, setID
}

func TestDieFassungenZeigenDasSymbolInDreiGroessen(t *testing.T) {
	anwendung, id, setID := mitFassungen(t, festeFassungen{fassungen: []msepaket.Symbolfassung{
		{Name: "common", Feld: "symbol", Fuellfarbe: "black", Randfarbe: "white", Randbreite: 0.1},
		{Name: "rare", Feld: "symbol", Fuellart: "linear gradient", Fuellfarbe1: "gold", Fuellfarbe2: "brown"},
	}})

	antwort := NeueBindungen(anwendung).BibliothekssymbolFassungen(t.Context(), id, setID)
	if antwort.Fehler != nil || len(antwort.Fassungen) != 2 || antwort.Fassungen[1].Name != "rare" {
		t.Fatalf("die Fassungen: %+v", antwort)
	}

	for _, fassung := range antwort.Fassungen {
		if len(fassung.Bilder) != len(Fassungsgroessen) {
			t.Fatalf("%s: %d Bilder", fassung.Name, len(fassung.Bilder))
		}

		for stelle, bild := range fassung.Bilder {
			if bild.Groesse != Fassungsgroessen[stelle] || !strings.HasPrefix(bild.Vorschau, "data:image/png;base64,") {
				t.Errorf("%s: %+v", fassung.Name, bild)
			}
		}
	}
}

// **Was sich nicht zeichnen oder packen lässt, hat eine leere Vorschau** —
// die Fassung bleibt, wie bei der Symbolpalette.
func TestEineFassungOhneBildBleibtStehen(t *testing.T) {
	anwendung, id, setID := mitFassungen(t, festeFassungen{fassungen: []msepaket.Symbolfassung{
		{Name: "common", Feld: "symbol"},
	}})

	anwendung.rasterer = klemmenderRasterer{render.NeuerRasterer(keineBilder{})}

	if antwort := anwendung.BibliothekssymbolFassungen(anwendung.fenster, id, setID); antwort.Fehler != nil ||
		antwort.Fassungen[0].Bilder[0].Vorschau != "" {
		t.Errorf("klemmender Rasterer: %+v", antwort)
	}

	anwendung.rasterer = render.NeuerRasterer(keineBilder{})
	anwendung.packen = func(io.Writer, image.Image) error { return errors.New("voll") }

	if antwort := anwendung.BibliothekssymbolFassungen(anwendung.fenster, id, setID); antwort.Fehler != nil ||
		antwort.Fassungen[0].Bilder[0].Vorschau != "" {
		t.Errorf("volle Platte: %+v", antwort)
	}
}

func TestDieFassungenSagenWasFehlt(t *testing.T) {
	klemmt := errors.New("klemmt")
	anwendung, id, setID := mitFassungen(t, festeFassungen{fehler: klemmt})
	ctx := anwendung.fenster

	for name, antwort := range map[string]SymbolfassungenAntwort{
		"klemmende Quelle": anwendung.BibliothekssymbolFassungen(ctx, id, setID),
		"Kennung":          anwendung.BibliothekssymbolFassungen(ctx, "x", setID),
		"Set":              anwendung.BibliothekssymbolFassungen(ctx, id, "x"),
	} {
		if antwort.Fehler == nil || antwort.Fassungen == nil {
			t.Errorf("%s: %+v", name, antwort)
		}
	}

	anwendung.rasterer = nil
	if antwort := anwendung.BibliothekssymbolFassungen(ctx, id, setID); antwort.Fehler == nil {
		t.Errorf("ohne Rasterer: %+v", antwort)
	}

	ohne := NeueApp("")
	if antwort := ohne.BibliothekssymbolFassungen(probekontext(t), "1", "1"); antwort.Fehler == nil {
		t.Errorf("ohne Bibliothek: %+v", antwort)
	}
}

// **Auch die Bearbeitungswege sagen, dass die Bibliothek fehlt.**
func TestOhneBibliothekSagenAuchDieBearbeitungswegeEs(t *testing.T) {
	anwendung := NeueApp("")
	ctx := probekontext(t)

	for name, befund := range map[string]*Fehler{
		"Teile":        anwendung.BibliothekssymbolTeile(ctx, "1").Fehler,
		"Sichern":      anwendung.BibliothekssymbolSichern(ctx, "1", nil).Fehler,
		"Zurücknehmen": anwendung.BibliothekssymbolZuruecknehmen(ctx, "1").Fehler,
		"Nachziehen":   anwendung.BibliothekssymbolNachziehen(ctx, "1").Fehler,
	} {
		if befund == nil || befund.Code != CodeValidierung {
			t.Errorf("%s: %+v", name, befund)
		}
	}
}

// zeichnungslos findet das Symbol, aber nicht seine Zeichnung.
type zeichnungslos struct{ Symbolbibliotheksfuehrer }

func (zeichnungslos) Zeichnung(context.Context, int64) (msepaket.Symbolzeichnung, error) {
	return msepaket.Symbolzeichnung{}, errors.New("unlesbar")
}

// **Die Bearbeitung am Fenster** — dieselben Wege über die Bindungen, und
// ein Symbol, dessen Zeichnung sich nicht lesen lässt.
func TestDieBearbeitungGehtUeberDieBindungen(t *testing.T) {
	anwendung, id, _ := mitFassungen(t, festeFassungen{})
	bindungen := NeueBindungen(anwendung)

	teile := bindungen.BibliothekssymbolTeile(t.Context(), id)
	if teile.Fehler != nil || len(teile.Teile) == 0 {
		t.Fatalf("die Teile: %+v", teile)
	}

	kreis := SymbolGrundform(SymbolgrundformDTO{Art: "Kreis"}).Teile

	for name, fehler := range map[string]*Fehler{
		"sichern":      bindungen.BibliothekssymbolSichern(t.Context(), id, kreis).Fehler,
		"nachziehen":   bindungen.BibliothekssymbolNachziehen(t.Context(), id).Fehler,
		"zurücknehmen": bindungen.BibliothekssymbolZuruecknehmen(t.Context(), id).Fehler,
	} {
		if fehler != nil {
			t.Errorf("%s: %+v", name, fehler)
		}
	}

	anwendung.MitSymbolbibliothek(zeichnungslos{anwendung.symbolbibliothek})

	if antwort := bindungen.BibliothekssymbolTeile(t.Context(), id); antwort.Fehler == nil || antwort.Teile == nil {
		t.Errorf("unlesbar: %+v", antwort)
	}
}
