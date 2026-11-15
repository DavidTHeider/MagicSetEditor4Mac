package service_test

// **Angebote zu einem Druck** — Werkstück A4 von Phase 5.3
// (#1443).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Druck, den der Markt schon kennt, wird nicht noch einmal aufgelöst.**
func TestEineBekannteFassungWirdNichtAufgeloest(t *testing.T) {
	t.Parallel()

	markt := &marktprobe{angebote: []domain.Marktangebot{{Fremdkennung: "1", Menge: 2}}}
	register := &druckprobe{druck: druckMit("m11", "4711")}

	ergebnis, fehler := service.NeuerMarktabruf(register, markt).
		Angebote(t.Context(), "abc", domain.Angebotswunsch{})
	if fehler != nil {
		t.Fatalf("Angebote: %v", fehler)
	}

	if ergebnis.Fassung != "4711" || ergebnis.Zugeordnet {
		t.Errorf("das Ergebnis ist %+v", ergebnis)
	}

	if markt.fassungsrufe != 0 {
		t.Errorf("die Fassungen wurden %d mal geholt, erwartet keinmal",
			markt.fassungsrufe)
	}

	if len(ergebnis.Angebote) != 1 || ergebnis.Angebote[0].Druck != "abc" {
		t.Errorf("die Angebote sind %+v", ergebnis.Angebote)
	}
}

// **Ein unbekannter Druck wird über seine Scryfall-Kennung gefunden — und die
// Zuordnung wird gemerkt.**
//
// **Gemerkt, weil sie sonst jeden Abruf kostet**: die Blueprints einer
// Erweiterung sind einige Megabyte, und die Zuordnung ändert sich nicht.
func TestEineNeueFassungWirdZugeordnetUndGemerkt(t *testing.T) {
	t.Parallel()

	markt := &marktprobe{
		fassungen: []domain.Fremdfassung{
			{Kennung: "1", Kennungen: []domain.Druckkennung{scryfallkennung("xyz")}},
			{Kennung: "89", Kennungen: []domain.Druckkennung{scryfallkennung("abc")}},
		},
		angebote: []domain.Marktangebot{{Fremdkennung: "1"}},
	}
	register := &druckprobe{druck: druckMit("m11", "")}

	ergebnis, fehler := service.NeuerMarktabruf(register, markt).
		Angebote(t.Context(), "abc", domain.Angebotswunsch{})
	if fehler != nil {
		t.Fatalf("Angebote: %v", fehler)
	}

	if ergebnis.Fassung != "89" || !ergebnis.Zugeordnet {
		t.Errorf("das Ergebnis ist %+v", ergebnis)
	}

	if markt.setcode != "m11" {
		t.Errorf("gefragt wurde nach %q, erwartet m11", markt.setcode)
	}

	if register.gesichert == nil {
		t.Fatal("die Zuordnung wurde nicht gemerkt")
	}

	if wert, _ := register.gesichert.Kennung(
		domain.SystemCardTrader, domain.SemantikDruck); wert != "89" {
		t.Errorf("gemerkt wurde %q", wert)
	}
}

// **Zwei Fassungen zu derselben Kennung sind ein Konflikt, kein Automerge.**
//
// Die Regel kommt aus A2 (#1441): „mehrdeutige oder widersprüchliche
// Cross-IDs als Auflösungskonflikt, kein Namens-Automerge".
func TestZweiFassungenSindEinKonflikt(t *testing.T) {
	t.Parallel()

	markt := &marktprobe{fassungen: []domain.Fremdfassung{
		{Kennung: "88", Kennungen: []domain.Druckkennung{scryfallkennung("abc")}},
		{Kennung: "89", Kennungen: []domain.Druckkennung{scryfallkennung("abc")}},
	}}
	register := &druckprobe{druck: druckMit("m11", "")}

	_, fehler := service.NeuerMarktabruf(register, markt).
		Angebote(t.Context(), "abc", domain.Angebotswunsch{})

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v, erwartet ein Konflikt", fehler)
	}

	if register.gesichert != nil {
		t.Error("bei einem Konflikt wurde etwas gemerkt")
	}

	if markt.angebotsrufe != 0 {
		t.Error("bei einem Konflikt wurde trotzdem nach Angeboten gefragt")
	}
}

// **Keine Fassung ist kein Konflikt, sondern ein »nicht gefunden«.**
func TestKeineFassungIstNichtGefunden(t *testing.T) {
	t.Parallel()

	markt := &marktprobe{fassungen: []domain.Fremdfassung{
		{Kennung: "88", Kennungen: []domain.Druckkennung{scryfallkennung("xyz")}},
	}}

	// **Ein anderer Setcode als in den übrigen Läufen** — sonst prüfte keiner,
	// dass der Setcode des Drucks überhaupt weitergereicht wird.
	_, fehler := service.NeuerMarktabruf(&druckprobe{druck: druckMit("znr", "")}, markt).
		Angebote(t.Context(), "abc", domain.Angebotswunsch{})

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("der Fehler ist %v, erwartet »nicht gefunden«", fehler)
	}
}

// **Ein Wunsch filtert auch, was der Anbieter durchgelassen hat.**
//
// **Der Filter steht in der Abfrage *und* hier**, und das ist kein doppelter
// Weg: ein Anbieter, der ihn übergeht, schickt sonst Posten, nach denen
// niemand gefragt hat — und niemand sähe es (§16, Drift).
func TestDerWunschFiltertAuchDieAntwort(t *testing.T) {
	t.Parallel()

	markt := &marktprobe{angebote: []domain.Marktangebot{
		{Fremdkennung: "1", Sprache: "en"},
		{Fremdkennung: "2", Sprache: "de"},
	}}

	ergebnis, fehler := service.NeuerMarktabruf(
		&druckprobe{druck: druckMit("m11", "4711")}, markt).
		Angebote(t.Context(), "abc", domain.Angebotswunsch{Sprache: "de"})
	if fehler != nil {
		t.Fatalf("Angebote: %v", fehler)
	}

	if len(ergebnis.Angebote) != 1 || ergebnis.Angebote[0].Fremdkennung != "2" {
		t.Errorf("durchgelassen wurde %+v", ergebnis.Angebote)
	}

	if markt.wunsch.Sprache != "de" {
		t.Errorf("der Anbieter bekam %+v — der Wunsch ging nicht mit hinaus", markt.wunsch)
	}

	if ergebnis.Wunsch.Sprache != "de" {
		t.Errorf("das Ergebnis nennt den Wunsch %+v", ergebnis.Wunsch)
	}
}

// **Ein Druck, den es nicht gibt, ist »nicht gefunden«.**
func TestEinUnbekannterDruckIstNichtGefunden(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeuerMarktabruf(&druckprobe{}, &marktprobe{}).
		Angebote(t.Context(), "abc", domain.Angebotswunsch{})

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("der Fehler ist %v, erwartet »nicht gefunden«", fehler)
	}
}

// **Und ein Fehler des Anbieters bleibt einer.**
func TestEinFehlerDesAnbietersBleibtEiner(t *testing.T) {
	t.Parallel()

	eigen := errors.New("die Gegenstelle schweigt")

	for _, fall := range []struct {
		name  string
		markt *marktprobe
		druck *domain.Druck
	}{
		{"beim Auflösen", &marktprobe{fassungsfehler: eigen}, druckMit("m11", "")},
		{"beim Abrufen", &marktprobe{angebotsfehler: eigen}, druckMit("m11", "4711")},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, fehler := service.NeuerMarktabruf(
				&druckprobe{druck: fall.druck}, fall.markt).
				Angebote(t.Context(), "abc", domain.Angebotswunsch{})

			if !errors.Is(fehler, eigen) {
				t.Fatalf("der Fehler ist %v", fehler)
			}
		})
	}
}

// **Die Befunde beider Schritte reisen mit.**
func TestDieBefundeBeiderSchritteReisenMit(t *testing.T) {
	t.Parallel()

	markt := &marktprobe{
		fassungen: []domain.Fremdfassung{
			{Kennung: "89", Kennungen: []domain.Druckkennung{scryfallkennung("abc")}},
		},
		fassungsbefunde: []domain.Befund{{Meldung: "eine Fassung ohne Kennung"}},
		angebotsbefunde: []domain.Befund{{Meldung: "ein Posten ohne Währung"}},
	}

	ergebnis, fehler := service.NeuerMarktabruf(
		&druckprobe{druck: druckMit("m11", "")}, markt).
		Angebote(t.Context(), "abc", domain.Angebotswunsch{})
	if fehler != nil {
		t.Fatalf("Angebote: %v", fehler)
	}

	if len(ergebnis.Befunde) != 2 {
		t.Errorf("die Befunde sind %+v", ergebnis.Befunde)
	}
}

// **Und ein gescheitertes Merken hält den Abruf nicht auf** (§7, Regel 3).
//
// Die Zuordnung ist Bestenfalls-Arbeit: sie spart den nächsten Abruf. Wer sie
// zur Bedingung machte, ließe einen Lesefehler eine Auskunft verhindern, die
// schon dasteht.
func TestEinGescheitertesMerkenHaeltDenAbrufNichtAuf(t *testing.T) {
	t.Parallel()

	markt := &marktprobe{
		fassungen: []domain.Fremdfassung{
			{Kennung: "89", Kennungen: []domain.Druckkennung{scryfallkennung("abc")}},
		},
		angebote: []domain.Marktangebot{{Fremdkennung: "1"}},
	}
	register := &druckprobe{
		druck:        druckMit("m11", ""),
		sicherfehler: errors.New("die Ablage klemmt"),
	}

	ergebnis, fehler := service.NeuerMarktabruf(register, markt).
		Angebote(t.Context(), "abc", domain.Angebotswunsch{})
	if fehler != nil {
		t.Fatalf("Angebote: %v", fehler)
	}

	if len(ergebnis.Angebote) != 1 {
		t.Fatalf("die Angebote sind %+v", ergebnis.Angebote)
	}

	if len(ergebnis.Befunde) != 1 {
		t.Errorf("das gescheiterte Merken steht nicht als Befund: %+v", ergebnis.Befunde)
	}
}

// derAlias ist die Scryfall-Kennung, unter der jeder Lauf hier fragt.
//
// **Als Konstante und nicht als Stelle**: sie wechselt in keinem Lauf, und
// eine Stelle, die immer denselben Wert bekommt, sieht wie eine Wahl aus.
const derAlias = "abc"

// druckMit baut einen Druck mit Setcode, Scryfall-Alias und Marktkennung.
//
// **Der Alias steht immer** — ein Druck ohne ihn erreicht [service.Marktabruf]
// nicht: gefunden wird er über ihn ([service.Marktregister]).
func druckMit(setcode, fassung string) *domain.Druck {
	druck := &domain.Druck{
		ID: 7, Name: "Probe", Setcode: setcode, Nummer: "1", Sprache: "en",
		Kennungen: []domain.Druckkennung{scryfallkennung(derAlias)},
	}

	if fassung != "" {
		druck.Kennungen = append(druck.Kennungen, domain.Druckkennung{
			System: domain.SystemCardTrader, Semantik: domain.SemantikDruck, Wert: fassung,
		})
	}

	return druck
}

func scryfallkennung(wert string) domain.Druckkennung {
	return domain.Druckkennung{
		System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: wert,
	}
}

// druckprobe ist ein Register mit einem Druck.
type druckprobe struct {
	druck        *domain.Druck
	gesichert    *domain.Druck
	sicherfehler error
	lesefehler   error
}

func (d *druckprobe) NachScryfall(_ context.Context, _ string) (*domain.Druck, error) {
	if d.lesefehler != nil {
		return nil, d.lesefehler
	}

	if d.druck == nil {
		return nil, nil
	}

	return d.druck, nil
}

func (d *druckprobe) Sichern(_ context.Context, druck *domain.Druck) error {
	if d.sicherfehler != nil {
		return d.sicherfehler
	}

	d.gesichert = druck

	return nil
}

// marktprobe ist ein Marktplatz, der antwortet, was ihm gesagt wurde.
type marktprobe struct {
	fassungen       []domain.Fremdfassung
	fassungsbefunde []domain.Befund
	fassungsfehler  error
	fassungsrufe    int
	setcode         string

	angebote        []domain.Marktangebot
	angebotsbefunde []domain.Befund
	angebotsfehler  error
	angebotsrufe    int
	wunsch          domain.Angebotswunsch
}

func (m *marktprobe) Quelle() domain.Preisquelle { return "probe" }

func (m *marktprobe) System() domain.Kennungssystem { return domain.SystemCardTrader }

func (m *marktprobe) Fassungen(_ context.Context, setcode string) (
	[]domain.Fremdfassung, []domain.Befund, error,
) {
	m.fassungsrufe++
	m.setcode = setcode

	return m.fassungen, m.fassungsbefunde, m.fassungsfehler
}

func (m *marktprobe) Angebote(
	_ context.Context, _ string, wunsch domain.Angebotswunsch,
) ([]domain.Marktangebot, []domain.Befund, error) {
	m.angebotsrufe++
	m.wunsch = wunsch

	return m.angebote, m.angebotsbefunde, m.angebotsfehler
}

// **Ein Fehler der Druckablage reist weiter.**
//
// **Nicht dasselbe wie »den Druck gibt es nicht«**: dort ist die Antwort eine
// Auskunft, hier kam gar keine. Wer beides gleich behandelte, meldete bei
// einer klemmenden Ablage „diesen Druck gibt es nicht" — und das ist falsch.
func TestEinFehlerDerDruckablageReistWeiter(t *testing.T) {
	t.Parallel()

	eigen := errors.New("die Ablage klemmt")

	_, fehler := service.NeuerMarktabruf(
		&druckprobe{lesefehler: eigen}, &marktprobe{}).
		Angebote(t.Context(), "abc", domain.Angebotswunsch{})

	if !errors.Is(fehler, eigen) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}
