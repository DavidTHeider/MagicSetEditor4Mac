// **Die Sammlungssichten für alle vier Flächen** — Werkstück 5 von Phase 4.5
// (#1292,
// Vermerk [0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
package app_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// geordneteNaht baut die Sichten über einem Bestand mit Besitz und Wunsch.
func geordneteNaht(t *testing.T) *app.Sammlungssichtennaht {
	t.Helper()

	db := frischerBestand(t)

	sammlung, err := app.SammlungsnahtUeber(db)
	if err != nil {
		t.Fatalf("SammlungsnahtUeber: %v", err)
	}

	for _, dto := range []app.ExemplarDTO{
		{
			Printing: "probe-druck", Menge: 4, Zustand: "nearMint",
			Sprache: "en", Lagerort: "Kiste 4",
		},
		{
			Printing: "probe-druck", Menge: 2, Zustand: "played",
			Sprache: "en", Wunsch: true,
		},
	} {
		if _, fehler := sammlung.ExemplarAnlegen(probekontext(t), dto); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}
	}

	return app.SichtenUeber(db)
}

// **Der ganze Weg: Besitz, Lagerorte, Dubletten, Wünsche.**
func TestDieSichtennahtFuehrtAlleVierWege(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	naht := geordneteNaht(t)

	stand, err := naht.Besitz(probekontext(t), "probe-druck")
	if err != nil {
		t.Fatalf("Besitz: %v", err)
	}

	if stand.Stueck != 4 || stand.Gewuenscht != 2 || !stand.Besitzt {
		t.Errorf("der Besitzstand ist %+v", stand)
	}

	orte, err := naht.Lagerorte(probekontext(t))
	if err != nil {
		t.Fatalf("Lagerorte: %v", err)
	}

	if len(orte) != 1 || orte[0].Lagerort != "Kiste 4" || orte[0].Stueck != 4 {
		t.Errorf("die Lagerorte sind %+v", orte)
	}

	dubletten, err := naht.Dubletten(probekontext(t), 1)
	if err != nil {
		t.Fatalf("Dubletten: %v", err)
	}

	if len(dubletten) != 1 || dubletten[0].Ueberzaehlig != 3 {
		t.Errorf("die Dubletten sind %+v", dubletten)
	}

	wuensche, err := naht.Wuensche(probekontext(t))
	if err != nil {
		t.Fatalf("Wuensche: %v", err)
	}

	if len(wuensche) != 1 || !wuensche[0].Wunsch {
		t.Errorf("die Wünsche sind %+v", wuensche)
	}
}

// **Ein geschlossener Bestand macht aus jedem der vier Wege einen Fehler.**
func TestEinGeschlossenerBestandAnDerSichtennaht(t *testing.T) {
	db := frischerBestand(t)
	naht := app.SichtenUeber(db)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, err := naht.Besitz(probekontext(t), "abc"); err == nil {
		t.Error("Besitz gab keinen Fehler")
	}

	if _, err := naht.Lagerorte(probekontext(t)); err == nil {
		t.Error("Lagerorte gab keinen Fehler")
	}

	if _, err := naht.Dubletten(probekontext(t), 1); err == nil {
		t.Error("Dubletten gab keinen Fehler")
	}

	if _, err := naht.Wuensche(probekontext(t)); err == nil {
		t.Error("Wuensche gab keinen Fehler")
	}
}

// **Ohne eingerichtete Naht sagen alle vier Bindungen es selbst.**
func TestOhneEingerichteteSichten(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if anwendung.SammlungBesitz(t.Context(), "abc").Fehler == nil {
		t.Error("SammlungBesitz gab keinen Fehler")
	}

	if anwendung.SammlungLagerorte(t.Context()).Fehler == nil {
		t.Error("SammlungLagerorte gab keinen Fehler")
	}

	if anwendung.SammlungDubletten(t.Context(), 1).Fehler == nil {
		t.Error("SammlungDubletten gab keinen Fehler")
	}

	if anwendung.SammlungWuensche(t.Context()).Fehler == nil {
		t.Error("SammlungWuensche gab keinen Fehler")
	}
}

// **Ein Fehler der Naht kommt an jeder Bindung an.**
func TestEinFehlerDerSichten(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitSammlungssichten(stolperndeSichten{})

	bindungen := app.NeueBindungen(anwendung)

	if bindungen.SammlungBesitz(t.Context(), "abc").Fehler == nil {
		t.Error("SammlungBesitz gab keinen Fehler")
	}

	if bindungen.SammlungLagerorte(t.Context()).Fehler == nil {
		t.Error("SammlungLagerorte gab keinen Fehler")
	}

	if bindungen.SammlungDubletten(t.Context(), 1).Fehler == nil {
		t.Error("SammlungDubletten gab keinen Fehler")
	}

	if bindungen.SammlungWuensche(t.Context()).Fehler == nil {
		t.Error("SammlungWuensche gab keinen Fehler")
	}
}

// **Und der geglückte Weg über die Bindungen.**
func TestDerGeglueckteSichtenweg(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitSammlungssichten(willigeSichten{})

	bindungen := app.NeueBindungen(anwendung)

	if bindungen.SammlungBesitz(t.Context(), "abc").Besitz == nil {
		t.Error("SammlungBesitz gab keinen Stand")
	}

	if bindungen.SammlungLagerorte(t.Context()).Fehler != nil {
		t.Error("SammlungLagerorte gab einen Fehler")
	}

	if bindungen.SammlungDubletten(t.Context(), 1).Fehler != nil {
		t.Error("SammlungDubletten gab einen Fehler")
	}

	if bindungen.SammlungWuensche(t.Context()).Fehler != nil {
		t.Error("SammlungWuensche gab einen Fehler")
	}
}

// stolperndeSichten scheitert auf allen vier Wegen.
type stolperndeSichten struct{}

func (stolperndeSichten) Besitz(context.Context, string) (*app.BesitzDTO, error) {
	return nil, errors.New("die Platte klemmt")
}

func (stolperndeSichten) Lagerorte(context.Context) ([]app.LagerortDTO, error) {
	return nil, errors.New("die Platte klemmt")
}

func (stolperndeSichten) Dubletten(context.Context, int) ([]app.DublettenDTO, error) {
	return nil, errors.New("die Platte klemmt")
}

func (stolperndeSichten) Wuensche(context.Context) ([]app.ExemplarDTO, error) {
	return nil, errors.New("die Platte klemmt")
}

// willigeSichten gibt auf allen vier Wegen ein Ergebnis.
type willigeSichten struct{}

func (willigeSichten) Besitz(context.Context, string) (*app.BesitzDTO, error) {
	return &app.BesitzDTO{Printing: "abc", Stueck: 1, Besitzt: true}, nil
}

func (willigeSichten) Lagerorte(context.Context) ([]app.LagerortDTO, error) {
	return []app.LagerortDTO{}, nil
}

func (willigeSichten) Dubletten(context.Context, int) ([]app.DublettenDTO, error) {
	return []app.DublettenDTO{}, nil
}

func (willigeSichten) Wuensche(context.Context) ([]app.ExemplarDTO, error) {
	return []app.ExemplarDTO{}, nil
}
