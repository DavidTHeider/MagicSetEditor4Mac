package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// wortpflegeprobe ist ein Double der Naht (§3).
type wortpflegeprobe struct {
	liste  service.Schluesselwortliste
	neu    int64
	fehler error

	setID, wortID int64
	feld, wert    string
	geloescht     bool

	// hilfen sind die Parameterhilfen (#1342).
	hilfen service.Parameterhilfen
}

func (p *wortpflegeprobe) Parameterhilfen(_ context.Context, setID, wortID int64) (
	service.Parameterhilfen, error,
) {
	p.setID, p.wortID = setID, wortID

	return p.hilfen, p.fehler
}

func (p *wortpflegeprobe) Liste(_ context.Context, setID int64) (
	service.Schluesselwortliste, error,
) {
	p.setID = setID

	return p.liste, p.fehler
}

func (p *wortpflegeprobe) Anlegen(_ context.Context, setID int64) (int64, error) {
	p.setID = setID

	return p.neu, p.fehler
}

func (p *wortpflegeprobe) Setzen(
	_ context.Context, setID, wortID int64, feld, wert string,
) error {
	p.setID, p.wortID, p.feld, p.wert = setID, wortID, feld, wert

	return p.fehler
}

func (p *wortpflegeprobe) Loeschen(_ context.Context, setID, wortID int64) error {
	p.setID, p.wortID, p.geloescht = setID, wortID, true

	return p.fehler
}

// mitWoertern verdrahtet eine Anwendung um diese Probe.
func mitWoertern(probe *wortpflegeprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitSchluesselwoertern(probe)

	return anwendung
}

// **Die Liste kommt als DTO an, samt der Unterscheidung fest/eigen** —
// #1062.
func TestDieSchluesselwortlisteKommtAn(t *testing.T) {
	probe := &wortpflegeprobe{liste: service.Schluesselwortliste{
		Woerter: []service.Schluesselwortsicht{
			{ID: 7, Wort: "Beben", Muster: "Beben", Modus: "custom", Gueltig: true},
			{Wort: "Fliegend", Muster: "Fliegend", Fest: true, Gueltig: true},
			{Wort: "Kaputt", Fest: true},
		},
		Modi: []string{"core", "custom"},
	}}

	antwort := mitWoertern(probe).Schluesselwoerter(probekontext(t), "3")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.setID != 3 {
		t.Errorf("gefragt wurde nach Set %d", probe.setID)
	}

	if len(antwort.Woerter) != 3 || len(antwort.Modi) != 2 {
		t.Fatalf("die Antwort ist %+v", antwort)
	}

	// **Nur die des Sets tragen eine Kennung.** Ohne sie ist der Schreibweg
	// gar nicht erst gangbar.
	if antwort.Woerter[0].ID != "7" || antwort.Woerter[0].Fest {
		t.Errorf("das erste ist %+v", antwort.Woerter[0])
	}

	if antwort.Woerter[1].ID != "" || !antwort.Woerter[1].Fest {
		t.Errorf("das zweite ist %+v", antwort.Woerter[1])
	}

	if antwort.Woerter[2].Gueltig {
		t.Errorf("das dritte gilt als gültig: %+v", antwort.Woerter[2])
	}
}

// **Anlegen gibt die Kennung des neuen zurück.**
func TestSchluesselwortAnlegenGibtDieKennung(t *testing.T) {
	probe := &wortpflegeprobe{neu: 42}

	antwort := mitWoertern(probe).SchluesselwortAnlegen(probekontext(t), "3")
	if antwort.Fehler != nil || antwort.WortID != "42" {
		t.Errorf("die Antwort ist %+v", antwort)
	}
}

// **Setzen und Löschen reichen beide Kennungen durch.**
func TestSchluesselwortSetzenUndLoeschen(t *testing.T) {
	probe := &wortpflegeprobe{}
	anwendung := mitWoertern(probe)

	if antwort := anwendung.SchluesselwortSetzen(
		probekontext(t), "3", "7", "muster", "Beben"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.setID != 3 || probe.wortID != 7 ||
		probe.feld != "muster" || probe.wert != "Beben" {
		t.Errorf("angekommen ist %+v", probe)
	}

	if antwort := anwendung.SchluesselwortLoeschen(
		probekontext(t), "3", "7"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !probe.geloescht {
		t.Error("gelöscht wurde nichts")
	}
}

// **Jede Kennung wird geprüft, und jeder Fehler kommt durch.**
//
// **Eine leere Liste und kein `nil`**: über die Wails-Bindung stünde dort
// sonst `null`, und das Frontend prüfte auf zwei Dinge statt auf eines.
func TestDieSchluesselwortbindungenWeisenAb(t *testing.T) {
	kaputt := &wortpflegeprobe{fehler: errors.New("die Platte ist voll")}

	t.Run("die Liste", func(t *testing.T) {
		for _, fall := range []struct {
			kennung string
			probe   *wortpflegeprobe
		}{
			{"", &wortpflegeprobe{}},
			{"abc", &wortpflegeprobe{}},
			{"3", kaputt},
		} {
			antwort := mitWoertern(fall.probe).Schluesselwoerter(probekontext(t), fall.kennung)

			if antwort.Fehler == nil {
				t.Errorf("%q kam durch", fall.kennung)
			}

			if antwort.Woerter == nil || antwort.Modi == nil {
				t.Errorf("%q gab nil statt leerer Listen", fall.kennung)
			}
		}
	})

	t.Run("das Anlegen", func(t *testing.T) {
		if mitWoertern(&wortpflegeprobe{}).SchluesselwortAnlegen(
			probekontext(t), "x").Fehler == nil {
			t.Error("die unbrauchbare Kennung kam durch")
		}

		if mitWoertern(kaputt).SchluesselwortAnlegen(probekontext(t), "3").Fehler == nil {
			t.Error("der Fehler kam nicht durch")
		}
	})

	t.Run("das Setzen", func(t *testing.T) {
		for _, fall := range [][2]string{{"", "7"}, {"3", ""}, {"3", "x"}} {
			if mitWoertern(&wortpflegeprobe{}).SchluesselwortSetzen(
				probekontext(t), fall[0], fall[1], "muster", "X").Fehler == nil {
				t.Errorf("%v kam durch", fall)
			}
		}

		if mitWoertern(kaputt).SchluesselwortSetzen(
			probekontext(t), "3", "7", "muster", "X").Fehler == nil {
			t.Error("der Fehler kam nicht durch")
		}
	})

	t.Run("das Löschen", func(t *testing.T) {
		for _, fall := range [][2]string{{"", "7"}, {"3", ""}} {
			if mitWoertern(&wortpflegeprobe{}).SchluesselwortLoeschen(
				probekontext(t), fall[0], fall[1]).Fehler == nil {
				t.Errorf("%v kam durch", fall)
			}
		}

		if mitWoertern(kaputt).SchluesselwortLoeschen(probekontext(t), "3", "7").Fehler == nil {
			t.Error("der Fehler kam nicht durch")
		}
	})
}

// **Und die Naht führt beide Seiten zusammen** — Lesen aus dem Setleser,
// Schreiben aus der Ablage.
//
// `service` soll nicht wissen, dass es zwei sind; ohne diesen Lauf stünde die
// Durchreiche ungeprüft da.
func TestDieSchluesselwortnahtReichtBeideSeitenDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probe", Game: "probe"}, nil)
	if err != nil {
		t.Fatalf("SetAnlegen: %v", err)
	}

	pflege, nahtfehler := app.SchluesselwortpflegeUeber(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	neu, anlegefehler := pflege.Anlegen(probekontext(t), setID)
	if anlegefehler != nil {
		t.Fatalf("Anlegen: %v", anlegefehler)
	}

	if setzfehler := pflege.Setzen(
		probekontext(t), setID, neu, "muster", "Beben"); setzfehler != nil {
		t.Fatalf("Setzen: %v", setzfehler)
	}

	liste, lesefehler := pflege.Liste(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("Liste: %v", lesefehler)
	}

	if len(liste.Woerter) != 1 || liste.Woerter[0].Muster != "Beben" {
		t.Fatalf("die Liste ist %+v", liste.Woerter)
	}

	if loeschfehler := pflege.Loeschen(probekontext(t), setID, neu); loeschfehler != nil {
		t.Fatalf("Löschen: %v", loeschfehler)
	}

	liste, lesefehler = pflege.Liste(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("Liste: %v", lesefehler)
	}

	if len(liste.Woerter) != 0 {
		t.Errorf("nach dem Löschen: %+v", liste.Woerter)
	}
}

// **Ohne Zuhause kommt die Pflege nicht zusammen**, und ein Lesefehler kommt
// durch.
//
// Zwei Zweige, die im Betrieb entscheiden, ob die Anwendung es meldet (§3).
func TestDieSchluesselwortpflegeMeldetIhreNahtfehler(t *testing.T) {
	t.Run("ohne Zuhause", func(t *testing.T) {
		t.Setenv("HOME", "")

		if _, fehler := app.SchluesselwortpflegeUeber(
			frischerBestand(t)); fehler == nil {
			t.Error("kein Fehler ohne Zuhause")
		}
	})

	t.Run("ein Lesefehler kommt durch", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		db := frischerBestand(t)

		pflege, fehler := app.SchluesselwortpflegeUeber(db)
		if fehler != nil {
			t.Fatalf("Naht: %v", fehler)
		}

		// **Ohne die Tabelle scheitert das Lesen** — der Weg, auf dem der
		// Fehler durch die Naht kommt.
		if _, err := db.ExecContext(
			probekontext(t), `DROP TABLE set_keywords`); err != nil {
			t.Fatalf("Tabelle entfernen: %v", err)
		}

		if _, err := pflege.Liste(probekontext(t), 1); err == nil {
			t.Error("das Lesen ohne Tabelle kam durch")
		}

		// **Und die Parameterhilfen gehen denselben Weg** (#1342): sie lesen
		// die Liste, bevor sie ein Wort suchen.
		if _, err := pflege.Parameterhilfen(probekontext(t), 1, 2); err == nil {
			t.Error("die Parameterhilfen ohne Tabelle kamen durch")
		}
	})
}
