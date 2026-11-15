package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Kontext eines Stils beginnt, wie MSE ihn beginnt** — Befund 1 des
// Nachlaufs über Phase G, zu VK-01b
// (#2182).
//
// `SetScriptContext::getContext(stylesheet)` bindet vor beiden Initskripten
// `stylesheet`, `card_style`, als `card` die **erste Karte des Sets** und als
// `styling` die **setweiten** Stiloptionen; erst `getContext(card)` setzt
// danach die Optionen der einzelnen Karte (`src/script/script_manager.cpp`,
// Zeilen 39–44 und 64–73, Fassung f9d9356 vom 2021-07-06). Hier stand der
// Kontext mit den Optionen der ersten Karte **dieses Stils** da, mit einem
// leeren `card` und ohne `card_style` — und ein Initskript, das eines davon
// las, gab je nach Reihenfolge der Karten anderes.

func einStil(stil msepaket.Stil) service.Ausfuhrstile {
	return func(string) (msepaket.Stil, bool) { return stil, true }
}

func fuehreMitStil(t *testing.T, stil msepaket.Stil, vorlage msepaket.Exportvorlage, karten []*domain.CardRow) string {
	t.Helper()

	vorlage.KurzerName = "probe"
	vorlage.Skript = `to_string(for each card in cards do card.gestalt + ";")`

	ausgabe, befunde, err := probelauf(service.Ausfuhrmittel{Stile: einStil(stil)}).Fuehre(
		t.Context(), vorlage, karten, nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v — %v", err, befunde)
	}

	return ausgabe
}

func TestDasInitskriptSiehtDieSetweitenStiloptionen(t *testing.T) {
	t.Parallel()

	stil := msepaket.Stil{
		Paket:        "probe-normal.mse-style",
		Stiloptionen: []domain.StyleField{{Name: "modus", Kind: domain.KindText, Initial: "setweit"}},
		Initskript:   `anfang := styling.modus`,
	}
	vorlage := msepaket.Exportvorlage{
		Kartenfelder: []domain.Field{{Name: "gestalt", Kind: domain.KindText, Skript: `anfang`}},
	}

	eigen := &domain.CardRow{ID: 1, Values: []domain.Value{
		{Field: service.Stilschalterfeld, Text: "true"},
		{Field: service.Stilwertefeld, Text: "modus: karte"},
	}}
	schlicht := &domain.CardRow{ID: 2}

	for name, karten := range map[string][]*domain.CardRow{
		"die eigene Karte zuerst":  {eigen, schlicht},
		"die eigene Karte zuletzt": {schlicht, eigen},
	} {
		if ausgabe := fuehreMitStil(t, stil, vorlage, karten); ausgabe != "setweit;setweit;" {
			t.Errorf("%s: %q, erwartet setweit;setweit;", name, ausgabe)
		}
	}
}

func TestDasInitskriptSiehtDieErsteKarteDesSets(t *testing.T) {
	t.Parallel()

	stil := msepaket.Stil{Paket: "probe-normal.mse-style", Initskript: `anfang := card.name`}
	vorlage := msepaket.Exportvorlage{Kartenfelder: []domain.Field{
		{Name: "name", Kind: domain.KindText},
		{Name: "gestalt", Kind: domain.KindText, Skript: `anfang`},
	}}
	karten := []*domain.CardRow{
		{ID: 1, Values: []domain.Value{{Field: "name", Text: "Erste"}}},
		{ID: 2, Values: []domain.Value{{Field: "name", Text: "Zweite"}}},
	}

	if ausgabe := fuehreMitStil(t, stil, vorlage, karten); ausgabe != "Erste;Erste;" {
		t.Errorf("%q, erwartet Erste;Erste;", ausgabe)
	}
}

func TestDasInitskriptSiehtCardStyle(t *testing.T) {
	t.Parallel()

	stil := msepaket.Stil{Paket: "probe-normal.mse-style", Initskript: `anfang := card_style`}
	vorlage := msepaket.Exportvorlage{Kartenfelder: []domain.Field{
		{Name: "gestalt", Kind: domain.KindText, Skript: `if anfang == nil then "fehlt" else "da"`},
	}}

	if ausgabe := fuehreMitStil(t, stil, vorlage, []*domain.CardRow{{ID: 1}}); ausgabe != "da;" {
		t.Errorf("%q, erwartet da;", ausgabe)
	}
}

// **Ein Stil wird einmal gefragt, nicht je Karte** — Befund 3 desselben
// Nachlaufs: die Stilquelle öffnet und liest ihr Paket bei jedem Ruf
// (`Stilwaehler.Stilpaket`), und gefragt wurde zweimal je Karte.
func TestEinStilWirdNichtJeKarteGefragt(t *testing.T) {
	t.Parallel()

	for _, anzahl := range []int{1, 86, 1000} {
		gefragt := 0
		stile := func(string) (msepaket.Stil, bool) {
			gefragt++

			return msepaket.Stil{Paket: "probe-normal.mse-style", Initskript: `form := { "normal" }`}, true
		}

		karten := make([]*domain.CardRow, anzahl)
		for i := range karten {
			karten[i] = &domain.CardRow{ID: int64(i + 1)}
		}

		vorlage := msepaket.Exportvorlage{
			KurzerName: "probe", Skript: `"fertig"`,
			Kartenfelder: []domain.Field{{Name: "gestalt", Kind: domain.KindText, Skript: `form()`}},
		}

		if _, befunde, err := probelauf(service.Ausfuhrmittel{Stile: stile}).Fuehre(
			t.Context(), vorlage, karten, nil, nil); err != nil {
			t.Fatalf("ausführen: %v — %v", err, befunde)
		}

		if gefragt > 3 {
			t.Errorf("%d Karten, ein Stil: %d Fragen an die Stilquelle", anzahl, gefragt)
		}
	}
}
