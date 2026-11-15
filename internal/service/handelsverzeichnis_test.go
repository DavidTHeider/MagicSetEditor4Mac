package service_test

// **Die Liste der Handelsvorgänge und ihre Entwürfe** — TRADE-01
// (#1715).

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// verzeichnisprobe ist die Ablage im Arbeitsspeicher.
type verzeichnisprobe struct {
	zeilen      []domain.Handelsvorgangzeile
	ersetzt     []domain.Handelsangebot
	geschrieben []domain.Handelsangebot
	geloescht   []string
	archiv      map[string]bool
	klagt       error

	// schreibklage trifft die schreibenden Wege — nach einem geglückten
	// Lesen.
	schreibklage error
}

func (v *verzeichnisprobe) Uebersicht(context.Context) ([]domain.Handelsvorgangzeile, error) {
	return v.zeilen, v.klagt
}

func (v *verzeichnisprobe) Vorgang(_ context.Context, kennung string) (domain.Handelsangebot, error) {
	for _, zeile := range v.zeilen {
		if zeile.Angebot.Vorgang == kennung {
			return zeile.Angebot, v.klagt
		}
	}

	return domain.Handelsangebot{}, domain.MitGrund(domain.ErrNotFound, nil, "den Handelsvorgang gibt es nicht")
}

func (v *verzeichnisprobe) Fortschreiben(_ context.Context, _, angebot domain.Handelsangebot) error {
	v.geschrieben = append(v.geschrieben, angebot)

	return v.schreibklage
}

func (v *verzeichnisprobe) EntwurfErsetzen(_ context.Context, angebot domain.Handelsangebot) error {
	v.ersetzt = append(v.ersetzt, angebot)

	return v.schreibklage
}

func (v *verzeichnisprobe) EntwurfLoeschen(_ context.Context, vorgang string) error {
	v.geloescht = append(v.geloescht, vorgang)

	return v.schreibklage
}

func (v *verzeichnisprobe) Archivieren(_ context.Context, vorgang string, an bool) error {
	if v.archiv == nil {
		v.archiv = map[string]bool{}
	}

	v.archiv[vorgang] = an

	return v.schreibklage
}

// verzeichniskontakte kennt zwei Gegenüber.
type verzeichniskontakte struct{ klagt error }

func (k verzeichniskontakte) Liste(context.Context) ([]domain.Peerkontakt, error) {
	return []domain.Peerkontakt{
		{Kennung: "bdu", EigenerName: "Turfer", SelbstgewaehlterName: "b"},
		{Kennung: "cer", SelbstgewaehlterName: "Cer"},
	}, k.klagt
}

// verzeichnismarken trägt eine Marke an v1.
type verzeichnismarken struct{ klagt error }

func (m verzeichnismarken) JeObjekt(_ context.Context, art string) (map[string][]string, error) {
	if art != string(domain.MarkenartHandel) {
		return nil, errors.New("falsche Art")
	}

	return map[string][]string{"v1": {"messe"}}, m.klagt
}

// verzeichnisangebot ist ein Vorgang zwischen aich (Initiator) und `mit`.
func verzeichnisangebot(vorgang, mit string, zustand domain.Handelszustand) domain.Handelsangebot {
	return domain.Handelsangebot{
		Vertragsfassung: domain.Handelsvertragsfassung, Vorgang: vorgang, Revision: 1,
		Art: domain.HandelTausch, Initiator: "aich", Parteien: []string{"aich", mit},
		Zustand: zustand,
		Positionen: []domain.Handelsposition{{
			Von: "aich", Printing: "abc", Menge: 1, Guete: "NM", Sprache: "de", Finish: "nonfoil",
		}},
	}
}

func verzeichnis(zeilen ...domain.Handelsvorgangzeile) (*service.Handelsverzeichnis, *verzeichnisprobe) {
	ablage := &verzeichnisprobe{zeilen: zeilen}

	return service.NeuesHandelsverzeichnis(ablage, verzeichniskontakte{}, verzeichnismarken{}).
		MitKennungen(func() string { return "neu-1" }), ablage
}

func TestDieListeSiehtNurEigeneVorgaengeMitLageUndNamen(t *testing.T) {
	t.Parallel()

	fremd := verzeichnisangebot("fremd", "bdu", domain.HandelAngeboten)
	fremd.Initiator, fremd.Parteien = "xyz", []string{"xyz", "bdu"}

	liste, _ := verzeichnis(
		domain.Handelsvorgangzeile{Angebot: verzeichnisangebot("v1", "bdu", domain.HandelEntwurf)},
		domain.Handelsvorgangzeile{Angebot: verzeichnisangebot("v2", "cer", domain.HandelBeidseitigVerbucht)},
		domain.Handelsvorgangzeile{Angebot: verzeichnisangebot("v3", "dan", domain.HandelAngeboten)},
		domain.Handelsvorgangzeile{Angebot: fremd},
	)

	zeilen, fehler := liste.Liste(t.Context(), "aich", domain.Handelsfilter{})
	if fehler != nil {
		t.Fatalf("Liste: %v", fehler)
	}

	if len(zeilen) != 3 {
		t.Fatalf("%d Zeilen, erwartet 3: %+v", len(zeilen), zeilen)
	}

	for _, fall := range []struct {
		zeile     domain.Handelsvorgangzeile
		lage      domain.Handelslage
		naechster domain.Handelsschritt
		partner   domain.Handelspartner
	}{
		// **Der eigene Name vor dem selbstgewählten** — so kennt man ihn.
		{zeilen[0], domain.LageDran, domain.SchrittAnbieten, domain.Handelspartner{Kennung: "bdu", Name: "Turfer"}},
		{zeilen[1], domain.LageAbgeschlossen, "", domain.Handelspartner{Kennung: "cer", Name: "Cer"}},
		// **Ein Unbekannter bleibt ohne Namen** — geraten wird keiner.
		{zeilen[2], domain.LageDran, domain.SchrittZustimmen, domain.Handelspartner{Kennung: "dan"}},
	} {
		if fall.zeile.Lage != fall.lage || fall.zeile.Naechster != fall.naechster ||
			!slices.Equal(fall.zeile.Gegenueber, []domain.Handelspartner{fall.partner}) {
			t.Errorf("%s: %+v", fall.zeile.Angebot.Vorgang, fall.zeile)
		}
	}

	if !slices.Equal(zeilen[0].Marken, []string{"messe"}) {
		t.Errorf("die Marke fehlt: %+v", zeilen[0])
	}

	gefiltert, fehler := liste.Liste(t.Context(), "aich", domain.Handelsfilter{Lagen: []domain.Handelslage{domain.LageAbgeschlossen}})
	if fehler != nil || len(gefiltert) != 1 || gefiltert[0].Angebot.Vorgang != "v2" {
		t.Errorf("gefiltert: %+v", gefiltert)
	}
}

func TestDieListeMeldetIhreFehler(t *testing.T) {
	t.Parallel()

	klage := errors.New("klemmt")

	for _, fall := range []struct {
		name   string
		baue   func() *service.Handelsverzeichnis
		filter domain.Handelsfilter
	}{
		{"ein ungültiger Filter", func() *service.Handelsverzeichnis {
			v, _ := verzeichnis()

			return v
		}, domain.Handelsfilter{Ordnung: "zufall"}},
		{"die Ablage", func() *service.Handelsverzeichnis {
			return service.NeuesHandelsverzeichnis(&verzeichnisprobe{klagt: klage}, verzeichniskontakte{}, verzeichnismarken{})
		}, domain.Handelsfilter{}},
		{"die Kontakte", func() *service.Handelsverzeichnis {
			return service.NeuesHandelsverzeichnis(&verzeichnisprobe{}, verzeichniskontakte{klagt: klage}, verzeichnismarken{})
		}, domain.Handelsfilter{}},
		{"die Marken", func() *service.Handelsverzeichnis {
			return service.NeuesHandelsverzeichnis(&verzeichnisprobe{}, verzeichniskontakte{}, verzeichnismarken{klagt: klage})
		}, domain.Handelsfilter{}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := fall.baue().Liste(t.Context(), "aich", fall.filter); fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}
