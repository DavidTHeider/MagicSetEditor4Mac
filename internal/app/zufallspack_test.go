package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// packprobe ist ein Double der Naht (§3).
type packprobe struct {
	arten   []service.Packangebot
	karten  []*domain.CardRow
	befunde []domain.Befund
	fehler  error

	setID int64
	art   string
	same  uint64
}

func (p *packprobe) Angebot(_ context.Context, setID int64) (
	[]service.Packangebot, []domain.Befund, error,
) {
	p.setID = setID

	return p.arten, p.befunde, p.fehler
}

func (p *packprobe) Ziehen(_ context.Context, setID int64, art string, same uint64) (
	[]*domain.CardRow, []domain.Befund, error,
) {
	p.setID, p.art, p.same = setID, art, same

	return p.karten, p.befunde, p.fehler
}

// mitPack verdrahtet eine Anwendung um diese Probe.
func mitPack(probe *packprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitPack(probe)

	return anwendung
}

// **Das Angebot kommt als DTO an** —
// #1072.
func TestDasPackangebotKommtAn(t *testing.T) {
	probe := &packprobe{arten: []service.Packangebot{
		{Name: "booster", Waehlbar: true, Karten: 12},
		{Name: "common", Karten: 8},
	}}

	antwort := mitPack(probe).Packarten(probekontext(t), "3")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.setID != 3 {
		t.Errorf("gefragt wurde nach Set %d", probe.setID)
	}

	if len(antwort.Arten) != 2 || !antwort.Arten[0].Waehlbar ||
		antwort.Arten[0].Karten != 12 {
		t.Errorf("das Angebot ist %+v", antwort.Arten)
	}
}

// **Der Same reist als Text** — über die Wails-Bindung verlöre ein `uint64`
// als JavaScript-Zahl oberhalb von 2⁵³ die letzten Stellen, und ein Same, der
// beim Hinweg schon ein anderer ist, wäre keine Wiederholbarkeit.
func TestDerSameReistAlsText(t *testing.T) {
	probe := &packprobe{karten: []*domain.CardRow{{ID: 7, Number: 1}}}

	antwort := mitPack(probe).PackZiehen(probekontext(t), "3", "booster", "18446744073709551615")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.same != 18446744073709551615 {
		t.Errorf("angekommen ist %d", probe.same)
	}

	if probe.art != "booster" || len(antwort.Karten) != 1 || antwort.Karten[0].ID != "7" {
		t.Errorf("gezogen wurde %+v", antwort.Karten)
	}
}

// **Ein leerer Same heißt null** — ein gültiger Same, und wer keinen angibt,
// bekommt trotzdem eine Ziehung, die sich wiederholen lässt.
func TestEinLeererSameIstNull(t *testing.T) {
	probe := &packprobe{}

	if antwort := mitPack(probe).PackZiehen(
		probekontext(t), "3", "booster", ""); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.same != 0 {
		t.Errorf("angekommen ist %d", probe.same)
	}
}

// **Jeder Fehlerweg meldet, und keiner stürzt ab.**
//
// **Leere Listen und kein `nil`**: über die Wails-Bindung stünde dort sonst
// `null`, und das Frontend prüfte auf zwei Dinge statt auf eines.
func TestDiePackbindungenWeisenAb(t *testing.T) {
	kaputt := &packprobe{fehler: errors.New("die Platte ist voll")}

	t.Run("das Angebot", func(t *testing.T) {
		for _, fall := range []struct {
			kennung string
			probe   *packprobe
		}{{"", &packprobe{}}, {"abc", &packprobe{}}, {"3", kaputt}} {
			antwort := mitPack(fall.probe).Packarten(probekontext(t), fall.kennung)

			if antwort.Fehler == nil {
				t.Errorf("%q kam durch", fall.kennung)
			}

			if antwort.Arten == nil || antwort.Befunde == nil {
				t.Errorf("%q gab nil statt leerer Listen", fall.kennung)
			}
		}

		if app.NeueApp("").Packarten(probekontext(t), "3").Fehler == nil {
			t.Error("ohne Naht kam es durch")
		}
	})

	t.Run("das Ziehen", func(t *testing.T) {
		faelle := []struct {
			kennung, same string
			probe         *packprobe
		}{
			{"", "1", &packprobe{}},
			{"3", "keine Zahl", &packprobe{}},
			{"3", "-1", &packprobe{}},
			{"3", "1", kaputt},
		}

		for _, fall := range faelle {
			antwort := mitPack(fall.probe).PackZiehen(
				probekontext(t), fall.kennung, "booster", fall.same)

			if antwort.Fehler == nil {
				t.Errorf("%q/%q kam durch", fall.kennung, fall.same)
			}

			if antwort.Karten == nil || antwort.Befunde == nil {
				t.Errorf("%q gab nil statt leerer Listen", fall.kennung)
			}
		}

		if app.NeueApp("").PackZiehen(probekontext(t), "3", "booster", "1").Fehler == nil {
			t.Error("ohne Naht kam es durch")
		}
	})
}

// **Die Befunde kommen mit** — eine Packart, deren Filter sich nicht ausrechnen
// ließ, zieht nichts, und niemand sähe warum (§8).
func TestDiePackbefundeKommenMit(t *testing.T) {
	probe := &packprobe{befunde: []domain.Befund{
		{Ort: "kaputt", Meldung: "der Filter ließ sich nicht ausrechnen"},
	}}

	if antwort := mitPack(probe).Packarten(probekontext(t), "3"); len(antwort.Befunde) != 1 {
		t.Errorf("die Befunde sind %+v", antwort.Befunde)
	}

	if antwort := mitPack(probe).PackZiehen(
		probekontext(t), "3", "booster", "1"); len(antwort.Befunde) != 1 {
		t.Errorf("die Befunde sind %+v", antwort.Befunde)
	}
}
