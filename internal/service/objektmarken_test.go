package service_test

// **Marken an jedem Fachobjekt** — UX-12
// (#1684),
// Vermerk 0135.

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// markendoppel hält Objekte und ihre Marken im Arbeitsspeicher.
type markendoppel struct {
	vorhanden map[string]bool
	marken    map[string][]string
	fehler    error

	vergeben  []string
	umbenannt []string
}

func neuesMarkendoppel() *markendoppel {
	return &markendoppel{
		vorhanden: map[string]bool{"deck/11": true, "deck/12": true, "deck/13": true, "partie/p1": true},
		marken:    map[string][]string{},
	}
}

func (d *markendoppel) Vorhanden(_ context.Context, art domain.Markenart, kennung string) (bool, error) {
	return d.vorhanden[string(art)+"/"+kennung], d.fehler
}

func (d *markendoppel) Setzen(_ context.Context, art domain.Markenart, kennung, marke string) error {
	schluessel := string(art) + "/" + kennung
	if !slices.Contains(d.marken[schluessel], marke) {
		d.marken[schluessel] = append(d.marken[schluessel], marke)
	}

	return nil
}

func (d *markendoppel) SetzenAlle(_ context.Context, art domain.Markenart, kennungen []string, marke string) error {
	d.vergeben = append(append(d.vergeben, string(art), marke), kennungen...)

	return d.fehler
}

func (d *markendoppel) Entfernen(_ context.Context, art domain.Markenart, kennung, marke string) error {
	schluessel := string(art) + "/" + kennung
	d.marken[schluessel] = slices.DeleteFunc(d.marken[schluessel], func(m string) bool { return m == marke })

	return nil
}

func (d *markendoppel) An(_ context.Context, art domain.Markenart, kennung string) ([]string, error) {
	return d.marken[string(art)+"/"+kennung], nil
}

func (d *markendoppel) JeObjekt(_ context.Context, art domain.Markenart) (map[string][]string, error) {
	je := map[string][]string{}

	for schluessel, marken := range d.marken {
		if kennung, ist := cutPrefix(schluessel, string(art)+"/"); ist {
			je[kennung] = marken
		}
	}

	return je, d.fehler
}

func (d *markendoppel) Katalog(context.Context) ([]domain.Markenposten, error) {
	return []domain.Markenposten{{Marke: "tausch", Anzahl: 2}}, d.fehler
}

func (d *markendoppel) Umbenennen(_ context.Context, alt, neu string, zusammen bool) (int, error) {
	d.umbenannt = append(d.umbenannt, alt, neu, map[bool]string{true: "zusammen", false: "allein"}[zusammen])

	return len(d.marken), d.fehler
}

func (d *markendoppel) UeberallEntfernen(context.Context, string) (int, error) {
	return len(d.marken), d.fehler
}

func cutPrefix(text, vorn string) (string, bool) {
	if len(text) < len(vorn) || text[:len(vorn)] != vorn {
		return "", false
	}

	return text[len(vorn):], true
}

func TestEineMarkeGehtNurAnEinObjektDesProfils(t *testing.T) {
	t.Parallel()

	doppel := neuesMarkendoppel()
	marken := service.NeueObjektmarken(doppel)

	if fehler := marken.Setzen(t.Context(), "deck", "11", "  Zum  TAUSCH "); fehler != nil {
		t.Fatal(fehler)
	}

	if an, fehler := marken.An(t.Context(), "deck", "11"); fehler != nil || !slices.Equal(an, []string{"zum tausch"}) {
		t.Errorf("an: %v %v", an, fehler)
	}

	for _, fall := range []struct {
		name, art, kennung, marke string
		art2                      error
	}{
		{"ein fremdes Deck", "deck", "99", "tausch", domain.ErrNotFound},
		{"eine unbekannte Art", "schrift", "11", "tausch", domain.ErrValidation},
		{"eine leere Marke", "deck", "11", "  ", domain.ErrValidation},
	} {
		if fehler := marken.Setzen(t.Context(), fall.art, fall.kennung, fall.marke); !errors.Is(fehler, fall.art2) {
			t.Errorf("setzen, %s: %v", fall.name, fehler)
		}

		if fehler := marken.Entfernen(t.Context(), fall.art, fall.kennung, fall.marke); !errors.Is(fehler, fall.art2) {
			t.Errorf("entfernen, %s: %v", fall.name, fehler)
		}
	}

	if _, fehler := marken.An(t.Context(), "partie", "fehlt"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("an einer fehlenden Partie: %v", fehler)
	}

	if fehler := marken.Entfernen(t.Context(), "deck", "11", "Zum Tausch"); fehler != nil || len(doppel.marken["deck/11"]) != 0 {
		t.Errorf("entfernen: %v %v", fehler, doppel.marken)
	}

	// **Ein Fehler der Ablage kommt an.**
	doppel.fehler = errors.New("die Platte klemmt")
	if fehler := marken.Setzen(t.Context(), "deck", "11", "x"); fehler == nil {
		t.Error("der Fehler der Ablage ging verloren")
	}
}

func TestDieMassenvergabeNenntIhreObjekte(t *testing.T) {
	t.Parallel()

	doppel := neuesMarkendoppel()
	marken := service.NeueObjektmarken(doppel)

	if fehler := marken.Vergeben(t.Context(), "deck", []string{"13", "11", "13"}, "Turnier"); fehler != nil {
		t.Fatal(fehler)
	}

	if !slices.Equal(doppel.vergeben, []string{"deck", "turnier", "11", "13"}) {
		t.Errorf("vergeben: %v", doppel.vergeben)
	}

	for _, fall := range []struct {
		name, art string
		kennungen []string
		marke     string
	}{
		{"ohne Objekt", "deck", nil, "x"},
		{"eine unbekannte Art", "schrift", []string{"1"}, "x"},
		{"eine leere Marke", "deck", []string{"11"}, ""},
	} {
		if fehler := marken.Vergeben(t.Context(), fall.art, fall.kennungen, fall.marke); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s: %v", fall.name, fehler)
		}
	}
}

func TestFilternHeisstUndOderOder(t *testing.T) {
	t.Parallel()

	doppel := neuesMarkendoppel()
	doppel.marken = map[string][]string{
		"deck/11": {"rot", "turnier"}, "deck/12": {"rot"}, "deck/13": {"blau"}, "partie/p1": {"rot"},
	}
	marken := service.NeueObjektmarken(doppel)

	for _, fall := range []struct {
		name         string
		gesucht      []string
		verknuepfung string
		erwartet     []string
	}{
		{"und", []string{"Rot", "turnier"}, "und", []string{"11"}},
		{"oder", []string{"turnier", "blau"}, "oder", []string{"11", "13"}},
		{"nur die eigene Art", []string{"rot"}, "und", []string{"11", "12"}},
	} {
		getroffen, fehler := marken.Filtern(t.Context(), "deck", fall.gesucht, fall.verknuepfung)
		if fehler != nil || !slices.Equal(getroffen, fall.erwartet) {
			t.Errorf("%s: %v %v", fall.name, getroffen, fehler)
		}
	}

	for _, fall := range []struct {
		name, art    string
		gesucht      []string
		verknuepfung string
	}{
		{"ohne Marke", "deck", nil, "und"},
		{"eine unbekannte Art", "schrift", []string{"rot"}, "und"},
		{"eine falsche Verknüpfung", "deck", []string{"rot"}, "xor"},
	} {
		if _, fehler := marken.Filtern(t.Context(), fall.art, fall.gesucht, fall.verknuepfung); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s: %v", fall.name, fehler)
		}
	}

	doppel.fehler = errors.New("die Platte klemmt")
	if _, fehler := marken.Filtern(t.Context(), "deck", []string{"rot"}, "und"); fehler == nil {
		t.Error("der Fehler der Ablage ging verloren")
	}
}

func TestUmbenennenUndEntfernenSagenWennEsNichtsGibt(t *testing.T) {
	t.Parallel()

	doppel := neuesMarkendoppel()
	marken := service.NeueObjektmarken(doppel)

	// **Ohne Objekt mit der Marke gibt es nichts umzubenennen.**
	if _, fehler := marken.Umbenennen(t.Context(), "tausch", "handel", false); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("umbenennen ohne Objekt: %v", fehler)
	}

	if _, fehler := marken.UeberallEntfernen(t.Context(), "tausch"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("entfernen ohne Objekt: %v", fehler)
	}

	doppel.marken["deck/11"] = []string{"tausch"}

	if zahl, fehler := marken.Umbenennen(t.Context(), " Tausch", "HANDEL", true); fehler != nil || zahl != 1 ||
		!slices.Equal(doppel.umbenannt, []string{"tausch", "handel", "allein", "tausch", "handel", "zusammen"}) {
		t.Errorf("umbenennen: %d %v %v", zahl, fehler, doppel.umbenannt)
	}

	// **Dieselbe Normalform ist kein neuer Name** — und keine Ablehnung.
	if zahl, fehler := marken.Umbenennen(t.Context(), "Handel", "handel", false); fehler != nil || zahl != 0 {
		t.Errorf("gleicher Name: %d %v", zahl, fehler)
	}

	if zahl, fehler := marken.UeberallEntfernen(t.Context(), "Handel"); fehler != nil || zahl != 1 {
		t.Errorf("entfernen: %d %v", zahl, fehler)
	}

	for _, fall := range []struct{ name, alt, neu string }{
		{"alt leer", "", "x"}, {"neu leer", "x", " "},
	} {
		if _, fehler := marken.Umbenennen(t.Context(), fall.alt, fall.neu, false); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s: %v", fall.name, fehler)
		}
	}

	if _, fehler := marken.UeberallEntfernen(t.Context(), ""); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("leer entfernen: %v", fehler)
	}

	if katalog, fehler := marken.Katalog(t.Context()); fehler != nil || len(katalog) != 1 {
		t.Errorf("Katalog: %v %v", katalog, fehler)
	}

	doppel.fehler = errors.New("die Platte klemmt")
	if _, fehler := marken.An(t.Context(), "deck", "11"); fehler == nil {
		t.Error("der Fehler beim Nachsehen ging verloren")
	}
}

// **Die Marken jedes Objekts einer Art** — für die Listen der Sichten.
func TestJeObjektNenntDieMarkenEinerArt(t *testing.T) {
	t.Parallel()

	doppel := neuesMarkendoppel()
	doppel.marken = map[string][]string{"deck/11": {"rot"}, "partie/p1": {"blau"}}
	marken := service.NeueObjektmarken(doppel)

	je, fehler := marken.JeObjekt(t.Context(), "deck")
	if fehler != nil || len(je) != 1 || !slices.Equal(je["11"], []string{"rot"}) {
		t.Errorf("%v %v", je, fehler)
	}

	if _, fehler := marken.JeObjekt(t.Context(), "schrift"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("schrift: %v", fehler)
	}
}

// **Die Kennung geht in ihrer einen Form an die Ablage** — `011` findet Deck
// 11 und liegt danach unter `11` (Befund 3 des Zweitblicks über UX-12).
func TestEineZahlkennungGehtInIhrerFormAnDieAblage(t *testing.T) {
	t.Parallel()

	doppel := neuesMarkendoppel()
	marken := service.NeueObjektmarken(doppel)

	if fehler := marken.Setzen(t.Context(), "deck", "011", "tausch"); fehler != nil {
		t.Fatal(fehler)
	}

	if an, fehler := marken.An(t.Context(), "deck", "+11"); fehler != nil || !slices.Equal(an, []string{"tausch"}) {
		t.Errorf("an: %v %v", an, fehler)
	}

	if fehler := marken.Entfernen(t.Context(), "deck", " 11", "tausch"); fehler != nil || len(doppel.marken["deck/11"]) != 0 {
		t.Errorf("entfernen: %v %v", doppel.marken, fehler)
	}

	if fehler := marken.Vergeben(t.Context(), "deck", []string{"012", "12", "11"}, "rot"); fehler != nil {
		t.Fatal(fehler)
	}

	if !slices.Equal(doppel.vergeben, []string{"deck", "rot", "11", "12"}) {
		t.Errorf("vergeben: %v", doppel.vergeben)
	}

	for _, ruf := range []func() error{
		func() error { return marken.Setzen(t.Context(), "deck", "11.0", "x") },
		func() error { _, fehler := marken.An(t.Context(), "deck", "1e1"); return fehler },
		func() error { return marken.Vergeben(t.Context(), "deck", []string{"11", "zwei"}, "x") },
	} {
		if fehler := ruf(); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("keine Zahl: %v", fehler)
		}
	}
}
