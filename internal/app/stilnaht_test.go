// **Die Stilnaht und die Paketnaht für die Flächen ohne Fenster**
// (#1107).
package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Naht kommt zusammen, listet, prüft und wechselt.**
//
// **Gegen die Saat**: ohne Pakete wäre die Liste leer, und ein Lauf über eine
// leere Liste prüfte die Umsetzung nicht.
func TestDieStilnahtListetUndWechselt(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatInsZuhause(t, zuhause)

	db := frischerBestand(t)
	setID, _ := saatsetImBestand(t, db)

	weg, nahtfehler := app.StilnahtUeber(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	stile, lesefehler := weg.Stile(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("Stile: %v", lesefehler)
	}

	if len(stile) < 2 {
		t.Fatalf("%d Stile — die Saat trägt mehrere: %+v", len(stile), stile)
	}

	ziel := andererStil(t, stile)

	// **Die Folgen ändern nichts** — das ist ihre Zusage.
	if _, fehler := weg.Folgen(probekontext(t), setID, ziel); fehler != nil {
		t.Fatalf("Folgen: %v", fehler)
	}

	if gewaehlt := gewaehlterStil(t, weg, setID); gewaehlt == ziel {
		t.Fatal("»Folgen« hat gewechselt")
	}

	if fehler := weg.Wechseln(probekontext(t), setID, ziel); fehler != nil {
		t.Fatalf("Wechseln: %v", fehler)
	}

	if gewaehlt := gewaehlterStil(t, weg, setID); gewaehlt != ziel {
		t.Errorf("gewechselt zu %q statt zu %q", gewaehlt, ziel)
	}
}

// andererStil gibt einen Stil, der nicht der gewählte ist.
func andererStil(t *testing.T, stile []app.StilDTO) string {
	t.Helper()

	for _, stil := range stile {
		if !stil.Gewaehlt {
			return stil.Paket
		}
	}

	t.Fatal("alle Stile sind gewählt — dann prüft der Wechsel nichts")

	return ""
}

// gewaehlterStil liest den markierten Stil.
func gewaehlterStil(t *testing.T, weg *app.Stilweg, setID int64) string {
	t.Helper()

	stile, fehler := weg.Stile(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("Stile: %v", fehler)
	}

	for _, stil := range stile {
		if stil.Gewaehlt {
			return stil.Paket
		}
	}

	return ""
}

// **Die Optionen lassen sich setzen und zurückstellen.**
func TestDieStilnahtSetztUndStelltZurueck(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatInsZuhause(t, zuhause)

	db := frischerBestand(t)
	setID, _ := saatsetImBestand(t, db)

	weg, nahtfehler := app.StilnahtUeber(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	optionen, lesefehler := weg.Optionen(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("Optionen: %v", lesefehler)
	}

	if len(optionen) == 0 {
		t.Fatalf("keine Option — der Saatstil müsste welche haben")
	}

	name := optionen[0].Name

	if fehler := weg.OptionSetzen(probekontext(t), setID, name, "probe"); fehler != nil {
		t.Fatalf("OptionSetzen: %v", fehler)
	}

	if wert, gesetzt := option(t, weg, setID, name); wert != "probe" || !gesetzt {
		t.Fatalf("nach dem Setzen: %q (gesetzt: %v)", wert, gesetzt)
	}

	if fehler := weg.OptionZuruecksetzen(probekontext(t), setID, name); fehler != nil {
		t.Fatalf("OptionZuruecksetzen: %v", fehler)
	}

	if _, gesetzt := option(t, weg, setID, name); gesetzt {
		t.Error("die Option gilt nach dem Zurücksetzen noch als gesetzt")
	}
}

// option liest eine Stiloption.
func option(t *testing.T, weg *app.Stilweg, setID int64, name string) (string, bool) {
	t.Helper()

	optionen, fehler := weg.Optionen(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("Optionen: %v", fehler)
	}

	for _, option := range optionen {
		if option.Name == name {
			return option.Wert, option.Gesetzt
		}
	}

	t.Fatalf("die Option %q steht nicht in der Liste", name)

	return "", false
}

// **Ohne Zuhause kommen beide Nähte nicht zusammen.**
func TestDieStilnahtMeldetIhreNahtfehler(t *testing.T) {
	t.Setenv("HOME", "")

	if _, fehler := app.StilnahtUeber(frischerBestand(t)); fehler == nil {
		t.Error("kein Fehler ohne Zuhause")
	}

	if _, fehler := app.PaketnahtUeber(frischerBestand(t)); fehler == nil {
		t.Error("kein Fehler ohne Zuhause bei der Paketnaht")
	}
}

// **Ein Lesefehler kommt durch jede der sechs Methoden.**
func TestDieStilnahtReichtLesefehlerDurch(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatInsZuhause(t, zuhause)

	db := frischerBestand(t)

	weg, nahtfehler := app.StilnahtUeber(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	// **Ohne die Tabelle scheitert jedes Lesen** — der Weg, auf dem der Fehler
	// durch die Naht kommt.
	if _, err := db.ExecContext(probekontext(t), `DROP TABLE sets`); err != nil {
		t.Fatalf("Tabelle entfernen: %v", err)
	}

	if _, err := weg.Stile(probekontext(t), 1); err == nil {
		t.Error("Stile ohne Tabelle kam durch")
	}

	if _, err := weg.Folgen(probekontext(t), 1, "probe-schlicht"); err == nil {
		t.Error("Folgen ohne Tabelle kam durch")
	}

	if err := weg.Wechseln(probekontext(t), 1, "probe-schlicht"); err == nil {
		t.Error("Wechseln ohne Tabelle kam durch")
	}

	if _, err := weg.Optionen(probekontext(t), 1); err == nil {
		t.Error("Optionen ohne Tabelle kam durch")
	}

	if err := weg.OptionSetzen(probekontext(t), 1, "rand", "x"); err == nil {
		t.Error("OptionSetzen ohne Tabelle kam durch")
	}

	if err := weg.OptionZuruecksetzen(probekontext(t), 1, "rand"); err == nil {
		t.Error("OptionZuruecksetzen ohne Tabelle kam durch")
	}
}

// stilprobe antwortet mit einer Option ohne Wahlmöglichkeiten.
type stilprobe struct {
	auffrischfehler error
}

// **Ohne Spielskript** — siehe `stilwahl_test.go`.
func (stilprobe) Spielskript(context.Context, int64) (string, error) { return "", nil }

func (stilprobe) Stile(context.Context, int64) ([]service.Stilangebot, error) {
	return []service.Stilangebot{{Paket: "probe-schlicht"}}, nil
}

func (stilprobe) Folgen(context.Context, int64, string) (*service.Folgen, error) {
	return &service.Folgen{}, nil
}

func (stilprobe) Wechseln(context.Context, int64, string) error { return nil }

// Optionen gibt eine **Textoption ohne Wahlmöglichkeiten**.
//
// **Der Fall, der über die Saat nicht zu erreichen ist**: ihre Optionen
// tragen alle eine Auswahl, und der Zweig, der `nil` in eine leere Liste
// verwandelt, bliebe unbetreten — bis eine Vorlage ihn beträte und `null` im
// JSON stünde (§3, „Zwischen den Prüfebenen").
func (stilprobe) Optionen(context.Context, int64) ([]service.Stiloption, error) {
	return []service.Stiloption{{
		StyleField: domain.StyleField{Name: "titel", Kind: domain.KindText},
	}}, nil
}

func (stilprobe) Kartenfelder(context.Context, int64) ([]domain.Field, error) {
	return nil, nil
}

func (stilprobe) Setfelder(context.Context, int64) ([]domain.Field, error) { return nil, nil }

func (stilprobe) OptionSetzen(context.Context, int64, string, string) error { return nil }

func (stilprobe) OptionZuruecksetzen(context.Context, int64, string) error { return nil }

func (stilprobe) Stilpaket(context.Context, int64, string) (msepaket.Stil, error) {
	return msepaket.Stil{}, domain.ErrNotFound
}

// Auffrischen scheitert, wenn die Probe es soll.
func (s stilprobe) Auffrischen(context.Context) ([]domain.Befund, error) {
	return nil, s.auffrischfehler
}

// **Eine Option ohne Wahlmöglichkeiten kommt als leere Liste, nicht als nil.**
func TestEineOptionOhneWahlKommtAlsLeereListe(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	weg, nahtfehler := app.StilnahtUeber(frischerBestand(t))
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	weg.MitWaehler(stilprobe{})

	optionen, fehler := weg.Optionen(probekontext(t), 1)
	if fehler != nil {
		t.Fatalf("Optionen: %v", fehler)
	}

	if len(optionen) != 1 {
		t.Fatalf("%d Optionen statt einer", len(optionen))
	}

	if optionen[0].Moeglichkeiten == nil {
		t.Error("die Wahlmöglichkeiten kamen als nil — eine leere Liste wäre richtig")
	}
}

// **Ein Auffrischfehler hält die Stilliste auf.**
//
// **Das ist Absicht**: eine Liste, die einen unlesbaren Katalog verschweigt,
// zeigt zu wenige Stile und sagt nicht, warum.
func TestEinAuffrischfehlerHaeltDieStilelisteAuf(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg, nahtfehler := app.StilnahtUeber(frischerBestand(t))
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	weg.MitWaehler(stilprobe{}).
		MitKatalog(stilprobe{auffrischfehler: errors.New("der Katalog ist zu")})

	if _, err := weg.Stile(probekontext(t), 1); err == nil {
		t.Error("der Auffrischfehler kam nicht durch")
	}
}
