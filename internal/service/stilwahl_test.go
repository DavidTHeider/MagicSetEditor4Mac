package service_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// stilablage ist die Naht zu `storage` — eine Attrappe.
type stilablage struct {
	werte  []domain.Value
	spiel  string
	stil   string
	fehler error

	// setzfehler scheitert **nur** beim Schreiben, nicht beim Lesen.
	//
	// `fehler` trifft beides — und damit endete `Wechseln` schon am Lesen, und
	// sein Schreibfehlerweg blieb unbetreten
	// (#693).
	setzfehler error
}

func (s *stilablage) Stilwerte(context.Context, int64) ([]domain.Value, error) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	return s.werte, nil
}

// **Und die beiden Einzeloperationen** — seit dem Zweitblick über Phase L
// schreibt der Dienst eine Option für sich, nicht die ganze Liste.
func (s *stilablage) StiloptionSetzen(_ context.Context, _ int64, feld, wert string) error {
	if s.setzfehler != nil {
		return s.setzfehler
	}

	for i := range s.werte {
		if s.werte[i].Field == feld {
			s.werte[i].Text = wert

			return nil
		}
	}

	s.werte = append(s.werte, domain.Value{Field: feld, Text: wert})

	return nil
}

func (s *stilablage) StiloptionEntfernen(_ context.Context, _ int64, feld string) error {
	if s.setzfehler != nil {
		return s.setzfehler
	}

	uebrig := make([]domain.Value, 0, len(s.werte))

	for _, wert := range s.werte {
		if wert.Field != feld {
			uebrig = append(uebrig, wert)
		}
	}

	s.werte = uebrig

	return nil
}

func (s *stilablage) StilSetzen(
	_ context.Context, _ int64, spiel, stil string, werte []domain.Value,
) error {
	if s.fehler != nil {
		return s.fehler
	}

	if s.setzfehler != nil {
		return s.setzfehler
	}

	s.spiel, s.stil, s.werte = spiel, stil, werte

	return nil
}

// feldliste ist die Naht zu den Feldnamen eines Sets.
type feldliste struct {
	felder []string
	fehler error
}

func (f *feldliste) Felder(context.Context, int64) ([]string, error) {
	return f.felder, f.fehler
}

// saatkatalog gibt die Pakete der Saat, wie der Katalog sie führt.
func saatkatalog() *ablage {
	return &ablage{pakete: []domain.Paket{
		{Art: "mse-game", Name: "probe", Pfad: "/p/probe.mse-game"},
		{
			Art: "mse-style", Name: "probe-schlicht", Version: "2026-08-15",
			Kurzname: "schlicht und klar",
			Pfad:     "/p/probe-schlicht.mse-style",
			Braucht:  []domain.Bedarf{{Paket: "probe.mse-game", Version: "2026-08-15"}},
		},
		{
			Art: "mse-style", Name: "fremd", Version: "2026-01-01", Pfad: "/p/fremd.mse-style",
			Braucht: []domain.Bedarf{{Paket: "anderes.mse-game"}},
		},
	}}
}

// waehler baut einen Stilwähler über der Saat des Repositoriums.
func waehler(t *testing.T, kopf domain.SetHead, ablageWerte *stilablage, felder *feldliste) *service.Stilwaehler {
	t.Helper()

	return service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{kopf}},
		saatkatalog(), ablageWerte, felder, festeSaatwurzel(t),
	)
}

// **Die Auswahl zeigt die installierten Stile — den eigenen zuerst.**
//
// Eine Liste von 194 Stilen, in der die brauchbaren irgendwo stehen, ist keine
// Auswahl. Ein fremder wird trotzdem **gezeigt**: er ist eine Möglichkeit mit
// Preis, und den nennt `Folgen`.
func TestDieAuswahlZeigtDenEigenenStilZuerst(t *testing.T) {
	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-schlicht"},
		&stilablage{}, &feldliste{})

	angebote, err := waehle.Stile(t.Context(), 1)
	if err != nil {
		t.Fatalf("Stile: %v", err)
	}

	if len(angebote) != 2 {
		t.Fatalf("Angebote: %+v", angebote)
	}

	if angebote[0].Paket != "probe-schlicht" || !angebote[0].EigenesSpiel || !angebote[0].Gewaehlt {
		t.Errorf("erstes Angebot: %+v", angebote[0])
	}

	if angebote[1].Paket != "fremd" || angebote[1].EigenesSpiel {
		t.Errorf("zweites Angebot: %+v", angebote[1])
	}

	// **Und jedes trägt seinen Anzeigenamen** — UX-05, S07.NAT-02 (#1677):
	// der Kurzname großgeschrieben, ohne ihn die Kennung. **„Und" wird groß**:
	// MSE2 nimmt nur englische kurze Wörter aus (`capitalize_in_place`).
	if angebote[0].Anzeigename != "Schlicht Und Klar" || angebote[1].Anzeigename != "Fremd" {
		t.Errorf("Anzeigenamen: %q, %q", angebote[0].Anzeigename, angebote[1].Anzeigename)
	}
}

// **Die Optionen kommen aus dem Paket** — und ihr Wert aus dem Bestand.
func TestDieOptionenKommenAusDemPaket(t *testing.T) {
	ablageWerte := &stilablage{werte: []domain.Value{{Field: "hintergrund", Text: "dunkel"}}}
	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-schlicht"},
		ablageWerte, &feldliste{})

	optionen, err := waehle.Optionen(t.Context(), 1)
	if err != nil {
		t.Fatalf("Optionen: %v", err)
	}

	// **Drei seit #764**: die Saat trägt eine Farboption mit `default:`, damit
	// der Vorgabezustand einen Fall hat.
	if len(optionen) != 3 {
		t.Fatalf("Optionen: %+v", optionen)
	}

	if optionen[0].Name != "hintergrund" || optionen[0].Wert != "dunkel" {
		t.Errorf("erste Option: %+v", optionen[0])
	}

	// **Ohne eigenen Wert gilt der des Pakets.** Ein leeres Feld wäre eine
	// dritte Möglichkeit, die keine Vorlage kennt.
	if optionen[1].Wert != optionen[1].Initial || optionen[1].Wert != "no" {
		t.Errorf("zweite Option: %+v", optionen[1])
	}
}

// **Ein Set ohne Stil hat keine Optionen — und keinen Fehler.**
func TestOhneStilKeineOptionen(t *testing.T) {
	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe"}, &stilablage{}, &feldliste{})

	optionen, err := waehle.Optionen(t.Context(), 1)
	if err != nil {
		t.Fatalf("Optionen: %v", err)
	}

	if len(optionen) != 0 {
		t.Errorf("Optionen: %+v", optionen)
	}
}

// **Was ein Wechsel kostet, steht vorher da** — die erste Zusage aus #41.
func TestDieFolgenStehenVorDemWechsel(t *testing.T) {
	ablageWerte := &stilablage{werte: []domain.Value{
		{Field: "hintergrund", Text: "dunkel"},
		{Field: "nur im alten Stil", Text: "ja"},
	}}
	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-schlicht"},
		ablageWerte, &feldliste{})

	folgen, err := waehle.Folgen(t.Context(), 1, "probe-bild")
	if err != nil {
		t.Fatalf("Folgen: %v", err)
	}

	if folgen.Spielwechsel {
		t.Error("derselbe Stil desselben Spiels gilt als Spielwechsel")
	}

	// `probe-bild` führt gar keine Stiloptionen — beide Werte gingen verloren.
	if !slices.Equal(folgen.VerloreneOptionen, []string{"hintergrund", "nur im alten Stil"}) {
		t.Errorf("verlorene Optionen: %v", folgen.VerloreneOptionen)
	}

	if len(folgen.UngedeuteteFelder) != 0 {
		t.Errorf("ungedeutete Felder ohne Spielwechsel: %v", folgen.UngedeuteteFelder)
	}
}

// **Bei einem Spielwechsel werden die Felder benannt, die das neue Spiel nicht
// kennt.**
//
// Sie gehen nicht verloren — der Bestand führt jeden Feldwert mit (#48) —, aber
// gezeichnet werden sie nicht mehr. Der Unterschied gehört in die Meldung.
func TestEinSpielwechselBenenntDieUngedeutetenFelder(t *testing.T) {
	felder := &feldliste{felder: []string{"name", "text", "farbe", "flavor"}}
	waehle := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "anderes", Style: "fremd"}}},
		saatkatalog(), &stilablage{}, felder, festeSaatwurzel(t),
	)

	folgen, err := waehle.Folgen(t.Context(), 1, "probe-schlicht")
	if err != nil {
		t.Fatalf("Folgen: %v", err)
	}

	if !folgen.Spielwechsel {
		t.Fatal("kein Spielwechsel erkannt")
	}

	// Das Spiel der Saat führt `name`, `text` und `farbe` — `flavor` nicht.
	if !slices.Equal(folgen.UngedeuteteFelder, []string{"flavor"}) {
		t.Errorf("ungedeutete Felder: %v", folgen.UngedeuteteFelder)
	}
}

// **Der Wechsel behält, was der neue Stil kennt.**
func TestDerWechselBehaeltWasBleibenKann(t *testing.T) {
	ablageWerte := &stilablage{werte: []domain.Value{
		{Field: "hintergrund", Text: "dunkel"},
		{Field: "gibt es dort nicht", Text: "ja"},
	}}
	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-bild"},
		ablageWerte, &feldliste{})

	if err := waehle.Wechseln(t.Context(), 1, "probe-schlicht"); err != nil {
		t.Fatalf("wechseln: %v", err)
	}

	if ablageWerte.stil != "probe-schlicht" || ablageWerte.spiel != "probe" {
		t.Errorf("gesetzt: %q / %q", ablageWerte.spiel, ablageWerte.stil)
	}

	if len(ablageWerte.werte) != 1 || ablageWerte.werte[0].Field != "hintergrund" {
		t.Errorf("Werte: %+v", ablageWerte.werte)
	}
}

// **Eine Option, die der Stil nicht führt, wird abgewiesen.**
//
// Sonst stünde im Bestand ein Wert, den kein Skript liest und den niemand mehr
// ändern kann.
func TestEineFremdeOptionWirdAbgewiesen(t *testing.T) {
	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-schlicht"},
		&stilablage{}, &feldliste{})

	err := waehle.OptionSetzen(t.Context(), 1, "gibt es nicht", "egal")
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("kein Validierungsfehler: %v", err)
	}
}

// **Eine bekannte Option wird gesetzt und ersetzt ihren Vorwert.**
func TestEineOptionWirdGesetztUndErsetzt(t *testing.T) {
	ablageWerte := &stilablage{werte: []domain.Value{{Field: "hintergrund", Text: "hell"}}}
	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-schlicht"},
		ablageWerte, &feldliste{})

	if err := waehle.OptionSetzen(t.Context(), 1, "hintergrund", "dunkel"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if len(ablageWerte.werte) != 1 || ablageWerte.werte[0].Text != "dunkel" {
		t.Fatalf("Werte: %+v", ablageWerte.werte)
	}

	if err := waehle.OptionSetzen(t.Context(), 1, "rahmen zeigen", "yes"); err != nil {
		t.Fatalf("zweites Setzen: %v", err)
	}

	if len(ablageWerte.werte) != 2 {
		t.Errorf("Werte: %+v", ablageWerte.werte)
	}
}

// **Ein Set, das der Bestand nicht führt, ist ein `ErrNotFound`** — auf jedem
// der vier Wege.
func TestEinUnbekanntesSetIstNichtGefunden(t *testing.T) {
	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe"}, &stilablage{}, &feldliste{})

	if _, err := waehle.Stile(t.Context(), 99); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Stile: %v", err)
	}

	if _, err := waehle.Optionen(t.Context(), 99); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Optionen: %v", err)
	}

	if _, err := waehle.Folgen(t.Context(), 99, "probe-schlicht"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Folgen: %v", err)
	}

	if err := waehle.OptionSetzen(t.Context(), 99, "hintergrund", "hell"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("OptionSetzen: %v", err)
	}
}

// **Ein Stil, den es nicht gibt, ebenso — und ein leerer Name ist ein
// Pflichtfeld.**
func TestEinUnbekannterStilLaesstSichNichtWaehlen(t *testing.T) {
	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-schlicht"},
		&stilablage{}, &feldliste{})

	if err := waehle.Wechseln(t.Context(), 1, "niegesehen"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Wechseln: %v", err)
	}

	if err := waehle.Wechseln(t.Context(), 1, ""); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("leerer Stil: %v", err)
	}
}

// **Innerhalb eines Spiels in MSE2s Ordnung** — erst `position hint:`, dann
// der Name ([domain.VergleicheNachRang]; UX-05, #1677). Alphabetisch stand
// „magic-DayNight" vor „M15 Mainframe", dem Stil mit dem kleinsten Hinweis.
func TestDieAuswahlOrdnetNachRang(t *testing.T) {
	probe := []domain.Bedarf{{Paket: "probe.mse-game"}}
	katalog := &ablage{pakete: []domain.Paket{
		{Art: "mse-style", Name: "probe-a", Rang: 20, Braucht: probe},
		{Art: "mse-style", Name: "fremd", Rang: 1, Braucht: []domain.Bedarf{{Paket: "anderes.mse-game"}}},
		{Art: "mse-style", Name: "probe-b", Rang: 5, Braucht: probe},
		{Art: "mse-style", Name: "probe-Z", Rang: 5, Braucht: probe},
	}}
	waehle := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "probe", Style: "probe-a"}}},
		katalog, &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	angebote, err := waehle.Stile(t.Context(), 1)
	if err != nil {
		t.Fatalf("Stile: %v", err)
	}

	reihe := make([]string, 0, len(angebote))
	for _, angebot := range angebote {
		reihe = append(reihe, angebot.Paket)
	}

	// **Der eigene bleibt vorn**, auch gegen einen kleineren fremden Rang.
	if want := []string{"probe-Z", "probe-b", "probe-a", "fremd"}; !slices.Equal(reihe, want) {
		t.Errorf("Reihe %v, erwartet %v", reihe, want)
	}
}
