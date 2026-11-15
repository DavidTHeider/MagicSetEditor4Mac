// **Zuletzt benutzt über die Naht** —
// #891,
// Werkstück 3 von Phase Y.
package app_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// benutzungsprobe merkt sich, welches Set gemeldet wurde.
type benutzungsprobe struct {
	setID  int64
	fehler error
}

func (b *benutzungsprobe) SetBenutzt(_ context.Context, setID int64) error {
	b.setID = setID

	return b.fehler
}

// **Ein aufgeschlagenes Set wird vermerkt — und die Leiste neu gesetzt.**
//
// Ohne den zweiten Teil zeigte das Untermenü „Zuletzt benutzt" bis zum
// nächsten Start die Reihenfolge von vorhin.
func TestEinBenutztesSetSetztDieLeisteNeu(t *testing.T) {
	t.Parallel()

	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	probe := &benutzungsprobe{}
	anwendung.MitSetbenutzung(probe)

	var gesetzt int

	anwendung.MitMenue(func(context.Context, *app.Menue) { gesetzt++ })

	if antwort := anwendung.SetBenutzt(t.Context(), "7"); antwort.Fehler != nil {
		t.Fatalf("SetBenutzt: %+v", antwort.Fehler)
	}

	if probe.setID != 7 {
		t.Errorf("vermerkt wurde %d", probe.setID)
	}

	if gesetzt != 1 {
		t.Errorf("die Leiste wurde %d-mal gesetzt", gesetzt)
	}
}

// **Eine unbrauchbare Kennung, ein Fehler des Speichers und eine fehlende
// Naht melden sich — und setzen die Leiste nicht.**
func TestSetBenutztMeldetSeineFehler(t *testing.T) {
	t.Parallel()

	t.Run("eine unbrauchbare Kennung", func(t *testing.T) {
		t.Parallel()

		anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))
		anwendung.MitSetbenutzung(&benutzungsprobe{})

		if antwort := anwendung.SetBenutzt(t.Context(), "keine"); antwort.Fehler == nil {
			t.Error("die Kennung kam durch")
		}
	})

	t.Run("ohne Naht", func(t *testing.T) {
		t.Parallel()

		anwendung := &app.App{}

		antwort := anwendung.SetBenutzt(t.Context(), "7")
		if antwort.Fehler == nil {
			t.Fatal("ohne Naht kam es durch")
		}

		if antwort.Fehler.Code != app.CodeValidierung {
			t.Errorf("der Code ist %q", antwort.Fehler.Code)
		}
	})

	t.Run("ein Fehler des Speichers", func(t *testing.T) {
		t.Parallel()

		anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))
		anwendung.MitSetbenutzung(&benutzungsprobe{
			fehler: errors.New("die Platte ist voll"),
		})

		var gesetzt int

		anwendung.MitMenue(func(context.Context, *app.Menue) { gesetzt++ })

		if antwort := anwendung.SetBenutzt(t.Context(), "7"); antwort.Fehler == nil {
			t.Error("der Fehler kam nicht durch")
		}

		if gesetzt != 0 {
			t.Errorf("die Leiste wurde trotzdem gesetzt (%d-mal)", gesetzt)
		}
	})
}

// **Und die Bindung reicht durch** (§16).
func TestDieBindungVermerktEinSet(t *testing.T) {
	t.Parallel()

	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	probe := &benutzungsprobe{}
	anwendung.MitSetbenutzung(probe)
	anwendung.MitMenue(func(context.Context, *app.Menue) {})

	if antwort := app.NeueBindungen(anwendung).SetBenutzt(t.Context(), "9"); antwort.Fehler != nil {
		t.Fatalf("SetBenutzt: %+v", antwort.Fehler)
	}

	if probe.setID != 9 {
		t.Errorf("vermerkt wurde %d", probe.setID)
	}
}

// **Das Untermenü führt höchstens neun Sets.**
//
// Neun, wie MSE2 (`Settings::max_recent_sets`). Ohne die Grenze wüchse die
// Ablage mit jedem Set, das jemand je geöffnet hat.
func TestDasUntermenueFuehrtHoechstensNeun(t *testing.T) {
	t.Parallel()

	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitSetbenutzung(&benutzungsprobe{})

	for nummer := range 12 {
		if antwort := anwendung.SetAnlegen(t.Context(),
			"Set "+string(rune('A'+nummer)), "probe", "probe-schlicht"); antwort.Fehler != nil {
			t.Fatalf("anlegen: %+v", antwort.Fehler)
		}
	}

	var leiste *app.Menue

	anwendung.MitMenue(func(_ context.Context, gesetzt *app.Menue) { leiste = gesetzt })

	if antwort := anwendung.SetBenutzt(t.Context(), "1"); antwort.Fehler != nil {
		t.Fatalf("SetBenutzt: %+v", antwort.Fehler)
	}

	ablage := untermenueMitNamen(leiste, app.MenuetexteFuer("de").Ablage)
	if ablage == nil {
		t.Fatal("die Leiste führt keine Ablage")
	}

	zuletzt := untermenueMitNamen(&app.Menue{Eintraege: ablage.Untermenue.Eintraege},
		app.MenuetexteFuer("de").Zuletzt)
	if zuletzt == nil {
		t.Fatal("die Ablage führt kein Untermenü »Zuletzt benutzt«")
	}

	if len(zuletzt.Untermenue.Eintraege) != 9 {
		t.Errorf("das Untermenü führt %d Einträge", len(zuletzt.Untermenue.Eintraege))
	}
}

// **Ein Bestand, der sich nicht lesen lässt, gibt eine leere Liste.**
//
// Die Menüleiste ist Bestenfalls-Arbeit; ein Programm, das wegen eines
// Lesefehlers gar keine bekäme, wäre unbedienbar. Der Fehler meldet sich an
// jeder anderen Stelle lauter.
func TestEinUnlesbarerBestandGibtEineLeereListe(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitBestand(leseprobe{fehler: errors.New("die Datei ist fort")})

	var leiste *app.Menue

	anwendung.MitMenue(func(_ context.Context, gesetzt *app.Menue) { leiste = gesetzt })

	if antwort := anwendung.SpracheSetzen("de"); antwort.Fehler != nil {
		t.Fatalf("SpracheSetzen: %+v", antwort.Fehler)
	}

	ablage := untermenueMitNamen(leiste, app.MenuetexteFuer("de").Ablage)
	if ablage == nil {
		t.Fatal("die Leiste führt keine Ablage")
	}

	if untermenueMitNamen(&app.Menue{Eintraege: ablage.Untermenue.Eintraege},
		app.MenuetexteFuer("de").Zuletzt) != nil {
		t.Error("das Untermenü steht trotz Lesefehler da")
	}
}

// **Ein Eintrag ohne Handlung stürzt nicht ab.**
//
// Dieselbe Zusage wie bei den übrigen Einträgen: ein `nil`-Rückruf ist ein
// Programmierfehler, aber einer, der sich beim Bauen zeigen soll und nicht
// beim Klicken.
func TestEinZuletzteintragOhneHandlungBleibtKlickbar(t *testing.T) {
	t.Parallel()

	leiste := app.Menueleiste(app.Menuhandlungen{}, app.MenuetexteFuer("de"),
		[]app.Zuletzteintrag{{ID: "7", Name: "Probeset"}})

	ablage := untermenueMitNamen(leiste, app.MenuetexteFuer("de").Ablage)
	zuletzt := untermenueMitNamen(&app.Menue{Eintraege: ablage.Untermenue.Eintraege},
		app.MenuetexteFuer("de").Zuletzt)

	eintrag := zuletzt.Untermenue.Eintraege[0]
	eintrag.Handlung()
}

// **Ohne Bestand gibt es keine Liste** — und keinen Absturz.
//
// Der Fall ist erreichbar: `SpracheSetzen` kommt, bevor der Bestand offen ist,
// wenn der Start scheiterte (#962).
func TestOhneBestandBleibtDieListeLeer(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}

	var leiste *app.Menue

	anwendung.MitMenue(func(_ context.Context, gesetzt *app.Menue) { leiste = gesetzt })

	if antwort := anwendung.SpracheSetzen("de"); antwort.Fehler != nil {
		t.Fatalf("SpracheSetzen: %+v", antwort.Fehler)
	}

	ablage := untermenueMitNamen(leiste, app.MenuetexteFuer("de").Ablage)
	if ablage == nil {
		t.Fatal("die Leiste führt keine Ablage")
	}

	if untermenueMitNamen(&app.Menue{Eintraege: ablage.Untermenue.Eintraege},
		app.MenuetexteFuer("de").Zuletzt) != nil {
		t.Error("das Untermenü steht ohne Bestand da")
	}

	// **Und der Fehlerweg meldet sich mit seinem Grund** — nicht mit einem
	// Nullzeiger.
	if antwort := anwendung.SetBenutzt(t.Context(), "7"); antwort.Fehler == nil {
		t.Error("ohne Naht kam es durch")
	}
}
