package service_test

// **Der Dienst, der eine Dimension über ein Set zählt** — Werkstück C von
// Phase 4.11 (#1338).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// dimensionsprobe gibt ein Spiel mit einer Dimension und ein paar Karten.
type dimensionsprobe struct {
	angaben []service.Auszaehlangabe
	karten  []*domain.CardRow
	fehler  error

	// kartenfehler scheitert **nur** am Lesen der Karten — das Angebot
	// steht dann noch.
	kartenfehler error

	// kartenstil ist der Stil, den jede Karte bekommt.
	kartenstil msepaket.Stil
}

func (d *dimensionsprobe) Auszaehldimensionen(_ context.Context, _ int64) (
	[]service.Auszaehlangabe, service.Vorlage, error,
) {
	return d.angaben, service.Vorlage{Kartenfelder: auszaehlfelder}, d.fehler
}

func (d *dimensionsprobe) VorlageFuer(
	_ context.Context, _ *domain.CardRow,
) service.Vorlage {
	return service.Vorlage{Kartenfelder: auszaehlfelder, Stil: d.kartenstil}
}

func (d *dimensionsprobe) Karten(_ context.Context, _ int64) ([]*domain.CardRow, error) {
	if d.kartenfehler != nil {
		return nil, d.kartenfehler
	}

	return d.karten, d.fehler
}

func probedimension() *dimensionsprobe {
	return &dimensionsprobe{
		angaben: []service.Auszaehlangabe{{
			Name:         "farbe",
			Beschreibung: "Wie viele Karten je Farbe",
			Skript:       "card.farbe",
			Gruppen:      []string{"weiss", "rot"},
			Farben:       map[string]string{"rot": "rgb(156,15,15)"},
		}},
		karten: auszaehlkarten("weiss", "rot", "rot"),
	}
}

// **Das Angebot nennt die Dimensionen des Spielpakets.**
func TestDerAuszaehldienstNenntSeineDimensionen(t *testing.T) {
	t.Parallel()

	angaben, fehler := service.NeuerAuszaehldienst(probedimension()).
		Angebot(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Angebot: %v", fehler)
	}

	if len(angaben) != 1 || angaben[0].Name != "farbe" {
		t.Errorf("das Angebot lautet %+v", angaben)
	}
}

// **Eine Auszählung gibt die Gruppen in der erklärten Reihenfolge.**
func TestDerAuszaehldienstZaehltNachDerDimension(t *testing.T) {
	t.Parallel()

	auszaehlung, fehler := service.NeuerAuszaehldienst(probedimension()).
		Auszaehlen(t.Context(), 1, "farbe")
	if fehler != nil {
		t.Fatalf("Auszaehlen: %v", fehler)
	}

	if auszaehlung.Dimension != "farbe" {
		t.Errorf("die Dimension heißt %q", auszaehlung.Dimension)
	}

	if len(auszaehlung.Gruppen) != 2 {
		t.Fatalf("%d Gruppen: %+v", len(auszaehlung.Gruppen), auszaehlung.Gruppen)
	}

	if auszaehlung.Gruppen[0].Name != "weiss" || auszaehlung.Gruppen[0].Karten != 1 {
		t.Errorf("die erste Gruppe ist %+v", auszaehlung.Gruppen[0])
	}

	if auszaehlung.Gruppen[1].Karten != 2 ||
		auszaehlung.Gruppen[1].Farbe != "rgb(156,15,15)" {
		t.Errorf("die zweite Gruppe ist %+v", auszaehlung.Gruppen[1])
	}
}

// **Eine Dimension, die es nicht gibt, ist ein fachliches Nein.**
func TestEineUnbekannteDimensionWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeuerAuszaehldienst(probedimension()).
		Auszaehlen(t.Context(), 1, "gibt es nicht")

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Ein Lesefehler kommt durch — an beiden Wegen.**
func TestEinLesefehlerKommtDurchDenAuszaehldienst(t *testing.T) {
	t.Parallel()

	eigener := errors.New("kein Spielpaket")
	dienst := service.NeuerAuszaehldienst(&dimensionsprobe{fehler: eigener})

	if _, fehler := dienst.Angebot(t.Context(), 1); !errors.Is(fehler, eigener) {
		t.Errorf("Angebot gab %v", fehler)
	}

	if _, fehler := dienst.Auszaehlen(t.Context(), 1, "farbe"); !errors.Is(fehler, eigener) {
		t.Errorf("Auszaehlen gab %v", fehler)
	}
}

// **Ein Spiel ohne Dimensionen gibt eine leere Liste und keinen Fehler.**
//
// §6: ein leerer Zustand erklärt den nächsten Schritt. Ein Set ohne
// Spielpaket hat keine erklärten Dimensionen, und das ist kein Fehler.
func TestOhneDimensionenKommtEineLeereListe(t *testing.T) {
	t.Parallel()

	angaben, fehler := service.NeuerAuszaehldienst(&dimensionsprobe{}).
		Angebot(t.Context(), 1)

	if fehler != nil || angaben == nil || len(angaben) != 0 {
		t.Errorf("Angebot gab %+v / %v", angaben, fehler)
	}
}

// **Die Befunde kommen mit — und als Liste, nie als `nil`.**
func TestEineAuszaehlungTraegtIhreBefunde(t *testing.T) {
	t.Parallel()

	probe := probedimension()
	probe.angaben[0].Skript = "gibtesnicht()"

	auszaehlung, fehler := service.NeuerAuszaehldienst(probe).
		Auszaehlen(t.Context(), 1, "farbe")
	if fehler != nil {
		t.Fatalf("Auszaehlen: %v", fehler)
	}

	if len(auszaehlung.Befunde) != 1 {
		t.Errorf("die Befunde lauten %+v", auszaehlung.Befunde)
	}

	// **Die Gruppen stehen trotzdem da** — die erklärten, mit null Karten.
	// Ein Diagramm ohne Achse sagte nicht, dass es die Achse gibt.
	if len(auszaehlung.Gruppen) != 2 {
		t.Errorf("die Gruppen lauten %+v", auszaehlung.Gruppen)
	}
}

// **Ein Lesefehler an den Karten hält die Auszählung an.**
//
// **Und zwar mit dem Fehler und nicht mit einer leeren Zählung.** Null
// Gruppen sähen aus wie „das Set ist leer" — und der Unterschied ist der
// zwischen „nichts da" und „nicht gelesen" (§8).
func TestEinKartenfehlerHaeltDieAuszaehlungAn(t *testing.T) {
	t.Parallel()

	eigener := errors.New("der Bestand antwortet nicht")

	probe := probedimension()
	probe.kartenfehler = eigener

	if _, fehler := service.NeuerAuszaehldienst(probe).
		Auszaehlen(t.Context(), 1, "farbe"); !errors.Is(fehler, eigener) {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Ein Abbruch bricht ab — mitten in den Karten.**
//
// Eine Auszählung über 3.761 Karten kostet gemessen 25 Sekunden. Wer währenddessen
// das Set wechselt, wartet sonst darauf, dass die alte fertig gerechnet wird
// (§7, Regel 4).
func TestEinAbbruchHaeltDieAuszaehlungAn(t *testing.T) {
	t.Parallel()

	abgebrochen, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	werte, _ := service.Auszaehlwerte(abgebrochen,
		auszaehlkarten("weiss", "rot"), auszaehlprobe(""), nil,
		service.Auszaehlangabe{Name: "farbe", Skript: "card.farbe"})

	if len(werte) != 0 {
		t.Errorf("nach dem Abbruch kamen %d Werte: %+v", len(werte), werte)
	}
}

// **Das `init script:` eines Stils läuft — einmal je Stilpaket.**
//
// `magic-m15-mainframe-dfc` schreibt `card_shape` um; ohne diesen Schritt gab
// die Dimension „color category" für jede zweiseitige Karte einen Wert statt
// zwei. Gemerkt wird je Paket: `magic-m15-altered/style` hat allein 700
// Zeilen, und ein Set aus 3.761 Karten benutzt eine Handvoll Stile.
func TestDasInitskriptDesStilsLaeuftUndWirdGemerkt(t *testing.T) {
	t.Parallel()

	probe := probedimension()
	probe.angaben[0].Skript = "stilwort()"

	// **Ohne erklärte Gruppen**: sonst gälte die Erklärung, und der Wert aus
	// dem Stil stünde in keiner von ihnen.
	probe.angaben[0].Gruppen = nil
	probe.angaben[0].Farben = nil

	probe.kartenstil = msepaket.Stil{
		Paket:      "probe.mse-style",
		Initskript: `stilwort := { "aus dem Stil" }`,
	}

	auszaehlung, fehler := service.NeuerAuszaehldienst(probe).
		Auszaehlen(t.Context(), 1, "farbe")
	if fehler != nil {
		t.Fatalf("Auszaehlen: %v", fehler)
	}

	if len(auszaehlung.Befunde) != 0 {
		t.Fatalf("Befunde: %+v", auszaehlung.Befunde)
	}

	// **Drei Karten, eine Gruppe** — alle drei bekommen denselben Wert aus
	// dem Stilskript. Ohne den Lauf des Initskripts wären es null Gruppen und
	// ein Befund.
	if len(auszaehlung.Gruppen) != 1 ||
		auszaehlung.Gruppen[0].Name != "aus dem Stil" ||
		auszaehlung.Gruppen[0].Karten != 3 {
		t.Errorf("die Gruppen lauten %+v", auszaehlung.Gruppen)
	}
}
