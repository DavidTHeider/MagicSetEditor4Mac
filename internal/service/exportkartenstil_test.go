package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Jede Karte der Ausfuhr rechnet im Kontext ihres Stils** — VK-01b
// (#2182).
//
// MSE hält je Stilbogen einen Kontext, lässt darin erst das `init script` des
// Spiels und dann das des Stils laufen und bindet vorher `stylesheet` und
// `styling`; eine Karte bekommt den Kontext ihres eigenen Stils, sonst den des
// Sets (`SetScriptContext::getContext`, `src/script/script_manager.cpp`,
// Zeilen 29–74, Fassung f9d9356 vom 2021-07-06). Gemessen in VK-01
// (`docs/belege/vk01-kartenstil.md`): `shape` rechnet `stylesheet;
// card_shape()`, und ohne den Stil gab die Ausfuhr den Wert des Spiels.
func TestDieAusfuhrRechnetJedeKarteImKontextIhresStils(t *testing.T) {
	t.Parallel()

	stile := func(paket string) (msepaket.Stil, bool) {
		switch paket {
		case "":
			return msepaket.Stil{
				Paket: "probe-normal.mse-style", Breite: 375,
				Initskript: `form := { "normal" }`,
			}, true
		case "quer":
			return msepaket.Stil{
				Paket: "probe-quer.mse-style", Breite: 523,
				Initskript: `form := { "quer" }`,
			}, true
		}

		return msepaket.Stil{}, false
	}

	vorlage := msepaket.Exportvorlage{
		KurzerName:  "probe",
		Spielskript: `form := { "spiel" }`,
		Kartenfelder: []domain.Field{
			{Name: "name", Kind: domain.KindText},
			{Name: "gestalt", Kind: domain.KindText, Skript: `stylesheet; form()`},
			{Name: "breite", Kind: domain.KindText, Skript: `to_string(stylesheet.card_width)`},
		},
	}

	karten := []*domain.CardRow{
		{ID: 1, Values: []domain.Value{{Field: "name", Text: "Anton"}}},
		{ID: 2, Values: []domain.Value{{Field: "name", Text: "Berta"}, {Field: "stylesheet", Text: "quer"}}},
		{ID: 3, Values: []domain.Value{{Field: "name", Text: "Cäsar"}}},
	}

	faelle := []struct {
		name, skript, erwartet string
	}{
		{
			"das Initskript des Stils schlägt das des Spiels",
			`to_string(for each card in cards do card.gestalt + ";")`,
			"normal;quer;normal;",
		},
		{
			"stylesheet ist der Stil der Karte",
			`to_string(for each card in cards do card.breite + ";")`,
			"375;523;375;",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			vorlage := vorlage
			vorlage.Skript = fall.skript

			ausgabe, befunde, err := probelauf(service.Ausfuhrmittel{Stile: stile}).Fuehre(
				t.Context(), vorlage, karten, nil, nil)
			if err != nil {
				t.Fatalf("ausführen: %v — %v", err, befunde)
			}

			if ausgabe != fall.erwartet {
				t.Errorf("%q, erwartet %q — Befunde %v", ausgabe, fall.erwartet, befunde)
			}
		})
	}
}

// **Ohne Stilquelle rechnet die Ausfuhr wie bisher** — mit dem Spiel allein.
// Ein Lauf ohne Bestand setzt keine ([service.Ausfuhrstile]), und dann darf
// ein Feld nicht an einem fehlenden Stil scheitern, das ohne ihn rechnet.
func TestOhneStilquelleRechnetDieAusfuhrMitDemSpiel(t *testing.T) {
	t.Parallel()

	vorlage := msepaket.Exportvorlage{
		KurzerName:  "probe",
		Skript:      `to_string(for each card in cards do card.gestalt + ";")`,
		Spielskript: `form := { "spiel" }`,
		Kartenfelder: []domain.Field{
			{Name: "gestalt", Kind: domain.KindText, Skript: `form()`},
		},
	}

	karten := []*domain.CardRow{
		{ID: 1},
		{ID: 2, Values: []domain.Value{{Field: "stylesheet", Text: "quer"}}},
	}

	ausgabe, befunde, err := probelauf(service.Ausfuhrmittel{}).Fuehre(t.Context(), vorlage, karten, nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v — %v", err, befunde)
	}

	if ausgabe != "spiel;spiel;" {
		t.Errorf("%q, erwartet spiel;spiel; — Befunde %v", ausgabe, befunde)
	}
}

// **Ein Stil, den die Quelle nicht kennt, fällt auf den des Sets zurück** — und
// kennt sie auch den nicht, rechnet die Karte mit dem Spiel allein. So wählt
// auch der Zeichenpfad ([service.Stilspeicher.StilFuerKarte]); dass der Stil
// fehlt, meldet die Ausfuhr einmal an `card.stylesheet`.
func TestEinUnbekannterStilFaelltAufDenDesSetsZurueck(t *testing.T) {
	t.Parallel()

	vorlage := msepaket.Exportvorlage{
		KurzerName:  "probe",
		Skript:      `to_string(for each card in cards do card.gestalt + ";")`,
		Spielskript: `form := { "spiel" }`,
		Kartenfelder: []domain.Field{
			{Name: "gestalt", Kind: domain.KindText, Skript: `form()`},
		},
	}

	karten := []*domain.CardRow{
		{ID: 1, Values: []domain.Value{{Field: "stylesheet", Text: "fehlt"}}},
	}

	for name, fall := range map[string]struct {
		setstil  bool
		erwartet string
	}{
		"auf den Stil des Sets": {true, "normal;"},
		"ohne den aufs Spiel":   {false, "spiel;"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stile := func(paket string) (msepaket.Stil, bool) {
				if paket == "" && fall.setstil {
					return msepaket.Stil{Paket: "probe-normal.mse-style", Initskript: `form := { "normal" }`}, true
				}

				return msepaket.Stil{}, false
			}

			ausgabe, befunde, err := probelauf(service.Ausfuhrmittel{Stile: stile}).Fuehre(
				t.Context(), vorlage, karten, nil, nil)
			if err != nil {
				t.Fatalf("ausführen: %v — %v", err, befunde)
			}

			if ausgabe != fall.erwartet {
				t.Errorf("%q, erwartet %q", ausgabe, fall.erwartet)
			}
		})
	}
}
