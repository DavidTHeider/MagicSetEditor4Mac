package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Paket nennt seinen kurzen und seinen vollen Namen.**
//
//	String short_name;      ///< Short name of this package
//	…
//	if (short_name.empty()) {
//	  if (!full_name.empty()) short_name = full_name;
//	  short_name = name();
//	}
//
// `Packaged`, `src/util/io/package.hpp`, Zeile 266, und `package.cpp`, Zeilen
// 539 und 623–625, Fassung f9d9356 vom 2021-07-06.
//
// **Die Vorlagen lesen `stylesheet.short_name`** — 190 Befunde über die 190
// Stile von `M15-Magic-Pack`, also einer je Stil
// (#270).
//
// **Und das Zurückfallen ist Teil der Regel** — mit einer Eigenheit, die man
// nachlesen muss: die mittlere Zeile ist **wirkungslos**. Sie wird von der
// nächsten sofort überschrieben, und ein `full name:` ohne `short name:`
// erreicht `short_name` damit **nie**. Es gilt `name()`, der Dateiname **ohne
// Endung**.
//
// **Hier stand bis #636
// eine Kette mit `full name:` in der Mitte** — mit der Begründung, MSEs
// Fassung sehe wie ein Versehen aus. Sie ist trotzdem die, gegen die ein
// Feldskript vergleicht, und `full_name` steht in derselben Umgebung ohnehin
// unter eigenem Namen. Die Abweichung kaufte nichts und kostete Treue.
//
// **Folgenlos über dieses Pack**: alle 376 Stilpakete von `full-magic-pack`
// erklären ein `short name:`, der Rückfall greift also nirgends (§10 — die
// Reichweite der Messung ist die Reichweite der Aussage).
func TestEinPaketNenntSeinenNamen(t *testing.T) {
	faelle := []struct {
		name, kurz, voll, paket, erwartet string
	}{
		{name: "der kurze Name", kurz: "M15", voll: "Magic M15", erwartet: "M15"},
		{
			name: "ein voller Name allein zählt nicht", voll: "Magic M15",
			paket: "magic-m15.mse-style", erwartet: "magic-m15",
		},
		{name: "sonst der Dateiname ohne Endung", paket: "magic-m15.mse-style", erwartet: "magic-m15"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			vorlage := service.Vorlage{
				Stil: msepaket.Stil{
					Spiel: "probe", Breite: 100, Hoehe: 100,
					KurzerName: fall.kurz, VollerName: fall.voll, Paket: fall.paket,
					Felder: []msepaket.Feldstil{{
						Feld: "name", Links: "0", Oben: "0", Breite: "80", Hoehe: "10",
						Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
					}},
				},
				Kartenfelder: []domain.Field{{
					Name: "name", Kind: domain.KindText, Skript: "stylesheet.short_name",
				}},
			}

			plan, befunde, err := service.NeuerPlanbauer().
				Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			if len(befunde) > 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			if len(plan.Ebenen) < 2 || plan.Ebenen[1].Klartext() != fall.erwartet {
				t.Errorf("gezeichnet wird %q, erwartet %q",
					plan.Ebenen[len(plan.Ebenen)-1].Klartext(), fall.erwartet)
			}
		})
	}
}

// **Eine Karte trägt Notizen — `card.notes`.**
//
//	class Card … {
//	  IndexMap<FieldP, ValueP> data;
//	  /// Notes for this card
//	  String notes;
//	  …
//	}
//
// `src/data/card.hpp`, Zeilen 36–40, Fassung f9d9356 vom 2021-07-06; reflektiert
// in `card.cpp`, Zeile 91. **Es ist kein Kartenfeld des Spiels**, sondern eine
// Eigenschaft jeder Karte — deshalb steht es in keiner `card field:`-Liste und
// fehlte hier.
//
// `magic.mse-game` liest es an drei Stellen, darunter die Kartennummer:
// `exporter_num_grabber(exporter_num_filter(card.notes))`. **192 Befunde** über
// die 190 Stile (#270).
func TestEineKarteTraegtNotizen(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Felder: []msepaket.Feldstil{{
				Feld: "name", Links: "0", Oben: "0", Breite: "80", Hoehe: "10",
				Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
			}},
		},
		Kartenfelder: []domain.Field{{
			Name: "name", Kind: domain.KindText, Skript: `"[" + card.notes + "]"`,
		}},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "notes", Text: "gemerkt"}}}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(plan.Ebenen) < 2 || plan.Ebenen[1].Klartext() != "[gemerkt]" {
		t.Errorf("gezeichnet wird %q, erwartet %q",
			plan.Ebenen[len(plan.Ebenen)-1].Klartext(), "[gemerkt]")
	}
}

// **Ein Vorgabewert, der einen Fehlerwert ergibt, wird gemeldet.**
//
// Dieselbe Regel wie beim Wertskript: MSEs Zuweisung ruft `toString()`, und
// bei einem `ScriptDelayedError` wirft die (`src/script/value.cpp`, Zeile 148,
// Fassung f9d9356 vom 2021-07-06). Der Wert bleibt leer — erfunden wird keiner
// (#345).
func TestEinVorgabewertMitFehlerwertWirdGemeldet(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Felder: []msepaket.Feldstil{{
				Feld: "name", Links: "0", Oben: "0", Breite: "80", Hoehe: "10",
				Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
			}},
		},
		Kartenfelder: []domain.Field{
			{Name: "name", Kind: domain.KindText, Default: "card.gibtsnicht"},
		},
	}

	_, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "gibtsnicht") {
		t.Errorf("Befunde: %v", befunde)
	}
}

// **Ein Punkt im Namen bleibt stehen** — abgeschnitten wird nur die Endung.
//
// `Package::name()` sucht `find_last_of('.')`; ein Paket, das
// `magic.m15.mse-style` heißt, behält sein erstes Stück
// (`src/util/io/package.cpp`, Zeilen 43–53, Fassung f9d9356 vom 2021-07-06).
func TestNurDieEndungFaelltWeg(t *testing.T) {
	for name, fall := range map[string][2]string{
		"mit Endung":       {"magic-m15.mse-style", "magic-m15"},
		"mit zwei Punkten": {"magic.m15.mse-style", "magic.m15"},
		"ohne Endung":      {"magic-m15", "magic-m15"},
		"gar nichts":       {"", ""},
	} {
		t.Run(name, func(t *testing.T) {
			if ist := service.Kurzname("", fall[0]); ist != fall[1] {
				t.Errorf("Kurzname(\"\", %q) = %q, erwartet %q", fall[0], ist, fall[1])
			}
		})
	}
}
