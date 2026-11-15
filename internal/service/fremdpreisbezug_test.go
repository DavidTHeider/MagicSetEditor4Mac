package service_test

// **Preisverlauf von einer fremden Quelle** — Werkstück A5 von Phase 5.3
// (#1444).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine bekannte Fassung wird nicht noch einmal gesucht.**
func TestEineBekannteFremdfassungWirdNichtGesucht(t *testing.T) {
	t.Parallel()

	quelle := &fremdquelle{staende: []domain.Preisstand{{Wert: 417, Waehrung: "USD"}}}
	register := &druckprobe{druck: fremddruck("mtg-m11-149")}

	bilanz, fehler := service.NeuerFremdpreisbezug(register, quelle, &ablageprobe{}).
		Holen(t.Context(), "abc", domain.Angebotswunsch{})
	if fehler != nil {
		t.Fatalf("Holen: %v", fehler)
	}

	if bilanz.Fassung != "mtg-m11-149" || bilanz.Zugeordnet {
		t.Errorf("die Bilanz ist %+v", bilanz)
	}

	if quelle.suchen != 0 {
		t.Errorf("es wurde %d mal gesucht, erwartet keinmal", quelle.suchen)
	}

	if bilanz.Staende != 1 {
		t.Errorf("%d Stände abgelegt, erwartet 1", bilanz.Staende)
	}
}

// **Eine unbekannte wird gesucht, zugeordnet und gemerkt.**
func TestEineNeueFremdfassungWirdGemerkt(t *testing.T) {
	t.Parallel()

	quelle := &fremdquelle{
		fassungen: []domain.Fremdfassung{{Kennung: "mtg-m11-149"}},
		staende:   []domain.Preisstand{{Wert: 417, Waehrung: "USD"}},
	}
	register := &druckprobe{druck: fremddruck("")}

	bilanz, fehler := service.NeuerFremdpreisbezug(register, quelle, &ablageprobe{}).
		Holen(t.Context(), "abc", domain.Angebotswunsch{})
	if fehler != nil {
		t.Fatalf("Holen: %v", fehler)
	}

	if bilanz.Fassung != "mtg-m11-149" || !bilanz.Zugeordnet {
		t.Errorf("die Bilanz ist %+v", bilanz)
	}

	if register.gesichert == nil {
		t.Fatal("die Zuordnung wurde nicht gemerkt")
	}

	if wert, _ := register.gesichert.Kennung(
		domain.SystemScrydex, domain.SemantikDruck); wert != "mtg-m11-149" {
		t.Errorf("gemerkt wurde %q", wert)
	}
}

// **Zwei Fassungen sind ein Konflikt, keine ist »nicht gefunden«.**
//
// Dieselbe Regel wie bei CardTrader und aus derselben Quelle (A2, #1441):
// „mehrdeutige oder widersprüchliche Cross-IDs als Auflösungskonflikt, kein
// Namens-Automerge".
func TestDieAufloesungKenntKonfliktUndLeere(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name      string
		fassungen []domain.Fremdfassung
		erwartet  error
	}{
		{"keine", nil, domain.ErrNotFound},
		{
			"zwei",
			[]domain.Fremdfassung{{Kennung: "eins"}, {Kennung: "zwei"}},
			domain.ErrConflict,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			quelle := &fremdquelle{fassungen: fall.fassungen}

			_, fehler := service.NeuerFremdpreisbezug(
				&druckprobe{druck: fremddruck("")}, quelle, &ablageprobe{}).
				Holen(t.Context(), "abc", domain.Angebotswunsch{})

			if !errors.Is(fehler, fall.erwartet) {
				t.Fatalf("der Fehler ist %v, erwartet %v", fehler, fall.erwartet)
			}

			if quelle.verlaeufe != 0 {
				t.Error("trotz ungeklärter Zuordnung wurde gefragt")
			}
		})
	}
}

// **Ein Druck, den es nicht gibt, ist »nicht gefunden«.**
func TestEinUnbekannterDruckHatKeinenFremdverlauf(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeuerFremdpreisbezug(
		&druckprobe{}, &fremdquelle{}, &ablageprobe{}).
		Holen(t.Context(), "abc", domain.Angebotswunsch{})

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Die Stände werden abgelegt — alle oder keiner.**
//
// Dieselbe Zusage wie beim Bulkbezug (A3): ein halber Verlauf wäre ein Stand,
// den es so nie gab, und niemand könnte ihn als halb erkennen.
func TestEinFehlerDerAblageHaeltDenBezugAuf(t *testing.T) {
	t.Parallel()

	eigen := errors.New("die Ablage klemmt")

	_, fehler := service.NeuerFremdpreisbezug(
		&druckprobe{druck: fremddruck("mtg-m11-149")},
		&fremdquelle{staende: []domain.Preisstand{{Wert: 1}}},
		&ablageprobe{fehler: eigen}).
		Holen(t.Context(), "abc", domain.Angebotswunsch{})

	if !errors.Is(fehler, eigen) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Und der Abrufzeitpunkt kommt von hier.**
//
// Der Beobachtungstag kommt von der Quelle (0093); **wann gefragt wurde**,
// weiß nur dieser Dienst — und ohne ihn ließe sich ein alter Verlauf nicht
// von einem frisch geholten unterscheiden.
func TestDerAbrufzeitpunktKommtVomDienst(t *testing.T) {
	t.Parallel()

	ablage := &ablageprobe{}

	_, fehler := service.NeuerFremdpreisbezug(
		&druckprobe{druck: fremddruck("mtg-m11-149")},
		&fremdquelle{staende: []domain.Preisstand{{Wert: 1}, {Wert: 2}}},
		ablage).Holen(t.Context(), "abc", domain.Angebotswunsch{})
	if fehler != nil {
		t.Fatalf("Holen: %v", fehler)
	}

	for _, stand := range ablage.eingetragen {
		if stand.Erhoben.IsZero() {
			t.Errorf("ein Stand kam ohne Abrufzeitpunkt: %+v", stand)
		}
	}
}

// **Ein Fehler der Quelle reist weiter.**
func TestEinFehlerDerFremdquelleReistWeiter(t *testing.T) {
	t.Parallel()

	eigen := errors.New("die Gegenstelle schweigt")

	for _, fall := range []struct {
		name   string
		quelle *fremdquelle
		druck  *domain.Druck
	}{
		{"beim Suchen", &fremdquelle{suchfehler: eigen}, fremddruck("")},
		{
			"beim Holen",
			&fremdquelle{verlaufsfehler: eigen},
			fremddruck("mtg-m11-149"),
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, fehler := service.NeuerFremdpreisbezug(
				&druckprobe{druck: fall.druck}, fall.quelle, &ablageprobe{}).
				Holen(t.Context(), "abc", domain.Angebotswunsch{})

			if !errors.Is(fehler, eigen) {
				t.Fatalf("der Fehler ist %v", fehler)
			}
		})
	}
}

// fremddruck baut einen Druck mit Alias und wahlweise Scrydex-Kennung.
func fremddruck(fassung string) *domain.Druck {
	druck := &domain.Druck{
		ID: 7, Name: "Lightning Bolt", Setcode: "m11", Nummer: "149", Sprache: "en",
		Kennungen: []domain.Druckkennung{scryfallkennung(derAlias)},
	}

	if fassung != "" {
		druck.Kennungen = append(druck.Kennungen, domain.Druckkennung{
			System: domain.SystemScrydex, Semantik: domain.SemantikDruck, Wert: fassung,
		})
	}

	return druck
}

// fremdquelle antwortet, was ihr gesagt wurde.
type fremdquelle struct {
	fassungen  []domain.Fremdfassung
	suchfehler error
	suchen     int

	staende        []domain.Preisstand
	befunde        []domain.Befund
	verlaufsfehler error
	verlaeufe      int
}

func (f *fremdquelle) Familie() domain.Preisquelle   { return "scrydex" }
func (f *fremdquelle) System() domain.Kennungssystem { return domain.SystemScrydex }

func (f *fremdquelle) Fassungen(_ context.Context, _ domain.Druck) (
	[]domain.Fremdfassung, error,
) {
	f.suchen++

	return f.fassungen, f.suchfehler
}

func (f *fremdquelle) Verlauf(
	_ context.Context, _, _ string, _ domain.Angebotswunsch,
) ([]domain.Preisstand, []domain.Befund, error) {
	f.verlaeufe++

	return f.staende, f.befunde, f.verlaufsfehler
}

// ablageprobe merkt sich, was eingetragen wurde.
type ablageprobe struct {
	eingetragen []domain.Preisstand
	fehler      error
}

func (a *ablageprobe) Eintragen(_ context.Context, staende ...domain.Preisstand) error {
	if a.fehler != nil {
		return a.fehler
	}

	a.eingetragen = append(a.eingetragen, staende...)

	return nil
}

// **Ein Fehler der Druckablage reist weiter.**
func TestEinFehlerDerAblageBeimFremdbezugReistWeiter(t *testing.T) {
	t.Parallel()

	eigen := errors.New("die Ablage klemmt")

	_, fehler := service.NeuerFremdpreisbezug(
		&druckprobe{lesefehler: eigen}, &fremdquelle{}, &ablageprobe{}).
		Holen(t.Context(), "abc", domain.Angebotswunsch{})

	if !errors.Is(fehler, eigen) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Ein gescheitertes Merken hält den Bezug nicht auf** (§7, Regel 3).
//
// Die Zuordnung spart den nächsten Abruf. Wer sie zur Bedingung machte, ließe
// einen Schreibfehler eine Auskunft verhindern, die schon dasteht.
func TestEinGescheitertesMerkenHaeltDenFremdbezugNichtAuf(t *testing.T) {
	t.Parallel()

	register := &druckprobe{
		druck:        fremddruck(""),
		sicherfehler: errors.New("die Ablage klemmt"),
	}

	bilanz, fehler := service.NeuerFremdpreisbezug(register, &fremdquelle{
		fassungen: []domain.Fremdfassung{{Kennung: "mtg-m11-149"}},
		staende:   []domain.Preisstand{{Wert: 417, Waehrung: "USD"}},
	}, &ablageprobe{}).Holen(t.Context(), "abc", domain.Angebotswunsch{})
	if fehler != nil {
		t.Fatalf("Holen: %v", fehler)
	}

	if bilanz.Staende != 1 {
		t.Errorf("%d Stände abgelegt, erwartet 1", bilanz.Staende)
	}

	if len(bilanz.Befunde) != 1 {
		t.Errorf("das gescheiterte Merken steht nicht als Befund: %+v", bilanz.Befunde)
	}
}
