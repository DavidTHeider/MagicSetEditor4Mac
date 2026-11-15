package domain_test

// **Marken an jedem Fachobjekt** — UX-12
// (#1684),
// Vermerk 0135.

import (
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJedeMarkenartLaesstSichLesen(t *testing.T) {
	t.Parallel()

	for _, art := range domain.Markenarten() {
		gelesen, fehler := domain.PruefeMarkenart(string(art))
		if fehler != nil || gelesen != art {
			t.Errorf("%s: %v", art, fehler)
		}
	}

	// **Neun Arten, seit das Symbol eine Ressource ist** — Vermerk 0135,
	// Punkt 9 hatte es mit DOM-02 angekündigt (#1697).
	if len(domain.Markenarten()) != 9 || !slices.Contains(domain.Markenarten(), domain.MarkenartSymbol) {
		t.Errorf("%v", domain.Markenarten())
	}

	if _, fehler := domain.PruefeMarkenart("schrift"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("schrift: %v", fehler)
	}
}

func TestEinFilterHeisstUndOderOder(t *testing.T) {
	t.Parallel()

	getragen := []string{"tausch", "binder 2"}

	for _, fall := range []struct {
		name         string
		marken       []string
		verknuepfung string
		trifft       bool
	}{
		{"und mit beiden", []string{"Tausch", "binder  2"}, "und", true},
		{"und mit einer fehlenden", []string{"tausch", "verliehen"}, "und", false},
		{"oder mit einer", []string{"verliehen", "TAUSCH"}, "oder", true},
		{"oder mit keiner", []string{"verliehen"}, "oder", false},
		{"ohne Marke ist kein Filter", nil, "und", true},
		{"ohne Verknüpfung heißt und", []string{"tausch", "verliehen"}, "", false},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			filter, fehler := domain.NeuerMarkenfilter(fall.marken, fall.verknuepfung)
			if fehler != nil {
				t.Fatal(fehler)
			}

			if filter.Trifft(getragen) != fall.trifft {
				t.Errorf("%+v trifft %v nicht wie erwartet", filter, getragen)
			}
		})
	}

	for _, falsch := range []struct {
		name         string
		marken       []string
		verknuepfung string
	}{
		{"eine unbekannte Verknüpfung", []string{"tausch"}, "xor"},
		{"eine leere Marke", []string{"  "}, "und"},
	} {
		if _, fehler := domain.NeuerMarkenfilter(falsch.marken, falsch.verknuepfung); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s: %v", falsch.name, fehler)
		}
	}
}

func TestGespeicherteFilterSindEineListe(t *testing.T) {
	t.Parallel()

	wert := `[{"name":"Zum Tausch","sicht":"exemplar","marken":["Tausch"],"verknuepfung":"oder"}]`

	filter, fehler := domain.GespeicherteFilterAus(wert)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if len(filter) != 1 || filter[0].Name != "Zum Tausch" || filter[0].Sicht != domain.MarkenartExemplar ||
		!slices.Equal(filter[0].Filter.Marken, []string{"tausch"}) || filter[0].Filter.Verknuepfung != domain.VerknuepfungOder {
		t.Errorf("%+v", filter)
	}

	// **Leer heißt keine** — die Vorgabe der Einstellung.
	if leer, fehler := domain.GespeicherteFilterAus(""); fehler != nil || len(leer) != 0 {
		t.Errorf("leer: %v %v", leer, fehler)
	}

	for _, falsch := range []struct{ name, wert string }{
		{"kein JSON", `[{`},
		{"ohne Namen", `[{"name":" ","sicht":"deck","marken":["a"]}]`},
		{"eine unbekannte Sicht", `[{"name":"a","sicht":"schrift","marken":["a"]}]`},
		{"ohne Marke", `[{"name":"a","sicht":"deck","marken":[]}]`},
		{"zweimal derselbe Name in einer Sicht", `[{"name":"a","sicht":"deck","marken":["x"]},{"name":"A","sicht":"deck","marken":["y"]}]`},
		{"eine falsche Verknüpfung", `[{"name":"a","sicht":"deck","marken":["x"],"verknuepfung":"xor"}]`},
	} {
		if fehler := domain.PruefeGespeicherteFilter(falsch.wert); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s: %v", falsch.name, fehler)
		}
	}

	// **Derselbe Name in zwei Sichten ist zweimal gemeint.**
	if fehler := domain.PruefeGespeicherteFilter(`[{"name":"a","sicht":"deck","marken":["x"]},{"name":"a","sicht":"set","marken":["x"]}]`); fehler != nil {
		t.Errorf("zwei Sichten: %v", fehler)
	}
}

// **Eine Zahlkennung hat genau eine Form** — sonst bestünde `011` die Frage
// nach Deck 11 und läge danach als `011` in einer Textspalte, die weder die
// Abfrage mit `11` noch der Löschauslöser findet (Befund 3 des Zweitblicks
// über UX-12).
func TestEineZahlkennungHatEineForm(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		art        domain.Markenart
		roh, soll  string
		abgewiesen bool
	}{
		{art: domain.MarkenartDeck, roh: "11", soll: "11"},
		{art: domain.MarkenartDeck, roh: "011", soll: "11"},
		{art: domain.MarkenartSet, roh: " +7 ", soll: "7"},
		{art: domain.MarkenartKarte, roh: "42", soll: "42"},
		{art: domain.MarkenartExemplar, roh: "11.0", abgewiesen: true},
		{art: domain.MarkenartTurnier, roh: "1e1", abgewiesen: true},
		{art: domain.MarkenartExemplar, roh: "0", abgewiesen: true},
		{art: domain.MarkenartDeck, roh: "-3", abgewiesen: true},
		{art: domain.MarkenartDeck, roh: "zwei", abgewiesen: true},
		// **Text bleibt Text**: eine Partie und ein Handel heißen nach ihrer
		// Kennung, ein Paket nach Art und Name.
		{art: domain.MarkenartPartie, roh: "011", soll: "011"},
		{art: domain.MarkenartHandel, roh: "h-011", soll: "h-011"},
		{art: domain.MarkenartPaket, roh: "game/magic", soll: "game/magic"},
	} {
		kennung, fehler := fall.art.Kennung(fall.roh)
		if fall.abgewiesen {
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("%s %q: %q, %v", fall.art, fall.roh, kennung, fehler)
			}

			continue
		}

		if fehler != nil || kennung != fall.soll {
			t.Errorf("%s %q: %q, %v — erwartet %q", fall.art, fall.roh, kennung, fehler, fall.soll)
		}
	}
}
