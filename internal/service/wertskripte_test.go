package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Kartenfeld mit `script:` bekommt seinen Wert gerechnet.**
//
// MSE wendet auf jeden Feldwert zwei Skripte an — `default:` **füllt** einen
// leeren Wert, `script:` **rechnet** ihn:
//
//	field().default_script.invokeOnDefault(ctx, value)
//	| field().        script.invokeOn(ctx, value)
//
// `src/data/field/text.cpp`, Zeilen 178–179, Fassung f9d9356 vom 2021-07-06.
// Der bisherige Wert steht dem Skript als `value` zur Verfügung.
//
// **Daran hängt der halbe Kartentext.** Das Feld `text` von `magic.mse-game`
// trägt `save value: false` und entsteht erst so — aus Regeltext und
// Flavortext, getrennt durch `<line>`
// (#345).
func TestEinFeldMitWertskriptWirdGerechnet(t *testing.T) {
	faelle := []struct {
		name     string
		felder   []domain.Field
		werte    []domain.Value
		erwartet string
	}{
		{
			name: "der bisherige Wert steht als `value` bereit",
			felder: []domain.Field{
				{Name: "text", Skript: `"[" + value + "]"`},
			},
			werte:    []domain.Value{{Field: "text", Text: "abc"}},
			erwartet: "[abc]",
		},
		{
			// **Ein Feld, das die Karte gar nicht führt, entsteht hier.** So
			// wie `text` bei MSE: `save value: false`.
			name: "ein Feld ohne gespeicherten Wert entsteht",
			felder: []domain.Field{
				{Name: "regel", Default: `"Fliegend."`},
				{Name: "text", Skript: `combined_editor(field1: card.regel, separator: "-", field2: "x")`},
			},
			erwartet: "Fliegend.<sep>-</sep>x",
		},
		{
			// **Ein gescheitertes Skript lässt den Wert stehen** (§8). Das
			// rettet die 72 Kartenfelder des Packs, deren Skript das noch nicht
			// gebaute `text_filter` ruft.
			name: "ein gescheitertes Skript lässt den Wert stehen",
			felder: []domain.Field{
				{Name: "text", Skript: `gibtsnicht(value)`},
			},
			werte:    []domain.Value{{Field: "text", Text: "bleibt"}},
			erwartet: "bleibt",
		},
		{
			// **Und `default:` läuft davor.** Erst füllen, dann rechnen.
			name: "der Vorgabewert steht dem Skript zur Verfügung",
			felder: []domain.Field{
				{Name: "text", Default: `"aus der Vorgabe"`, Skript: `"[" + value + "]"`},
			},
			erwartet: "[aus der Vorgabe]",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523,
				Felder: []msepaket.Feldstil{{
					Feld: "text", Links: "0", Oben: "0", Breite: "300", Hoehe: "40",
					Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
				}},
			}

			karte := &domain.CardRow{ID: 1, Number: 1, Values: fall.werte}

			plan, _, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), karte, service.Vorlage{
					Stil: stil, Kartenfelder: fall.felder,
				})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for stelle := range plan.Ebenen {
				if plan.Ebenen[stelle].Art != render.ArtText {
					continue
				}

				// Der `<sep>` ist eine Auszeichnung und wird entfernt; geprüft
				// wird der Wert, der ankommt.
				if text := plan.Ebenen[stelle].Klartext(); text != ohneSep(fall.erwartet) {
					t.Errorf("gezeichnet wird %q, erwartet %q", text, ohneSep(fall.erwartet))
				}

				return
			}

			t.Fatal("der Plan führt keine Textebene")
		})
	}
}

// **Ein gescheitertes Wertskript wird benannt.**
//
// Ein leerer Textkasten ohne Meldung sieht aus wie eine Karte ohne Regeltext
// (§8).
func TestEinGescheitertesWertskriptWirdBenannt(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "300", Hoehe: "40",
			Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	_, befunde, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1}, service.Vorlage{
			Stil:         stil,
			Kartenfelder: []domain.Field{{Name: "text", Skript: `gibtsnicht()`}},
		})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	gefunden := false

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "Wertskript") {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("kein Befund über das gescheiterte Wertskript: %v", befunde)
	}
}

// **Ein geklammertes Wertskript wird ebenso gelesen.**
//
// Beide Formen kommen vor: `magic.mse-game` schreibt seine Wertskripte als
// eingerückten Block ohne Klammern, andere Vorlagen klammern.
func TestEinGeklammertesWertskriptWirdGelesen(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "300", Hoehe: "40",
			Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(),
			&domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
				{Field: "text", Text: "abc"},
			}},
			service.Vorlage{
				Stil:         stil,
				Kartenfelder: []domain.Field{{Name: "text", Skript: `{ "[" + value + "]" }`}},
			})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for stelle := range plan.Ebenen {
		if plan.Ebenen[stelle].Art == render.ArtText {
			if text := plan.Ebenen[stelle].Klartext(); text != "[abc]" {
				t.Errorf("gezeichnet wird %q, erwartet %q", text, "[abc]")
			}

			return
		}
	}

	t.Fatal("der Plan führt keine Textebene")
}

// ohneSep nimmt die `<sep>`-Klammern heraus — sie sind Auszeichnung.
func ohneSep(text string) string {
	return strings.NewReplacer("<sep>", "", "</sep>", "").Replace(text)
}

// **Ein abgebrochener Lauf rechnet keine Wertskripte mehr.**
//
// `magic.mse-game` führt 249 Kartenfelder, und die Wertskripte sind die
// längste Strecke im Vorschaulauf — wer während des Tippens die Karte
// wechselt, wartet sonst darauf, dass die alte fertig gerechnet wird (§7).
func TestEinAbgebrochenerLaufRechnetKeineWertskripte(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "300", Hoehe: "40",
			Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	_, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(ctx,
			&domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
				{Field: "text", Text: "abc"},
			}},
			service.Vorlage{
				Stil:         stil,
				Kartenfelder: []domain.Field{{Name: "text", Skript: `"[" + value + "]"`}},
			})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("der Fehler ist %v, erwartet context.Canceled", err)
	}
}

// **Polster und Zeilenhöhen dürfen Skripte sein — und was nicht rechnet, ist null.**
//
// Ein Polster ist eine Zugabe; wo keines steht, steht keines. Ein Befund je
// fehlender Angabe stünde auf jeder Karte und sagte nichts.
func TestPolsterUndZeilenhoehenWerdenAusgerechnet(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "300", Hoehe: "100",
			Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			Polster: msepaket.Polsterangabe{
				Links: "6", Rechts: "{ 2 + 2 }", Oben: "3", Unten: "{ gibtsnicht() }",
			},
			Zeilenhoehen: msepaket.Zeilenhoehenangabe{Weich: "0.9", Hart: "{ 1.2 }"},
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: "abc"},
	}}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for stelle := range plan.Ebenen {
		ebene := &plan.Ebenen[stelle]
		if ebene.Art != render.ArtText {
			continue
		}

		if ebene.Satz.Polster != (render.Polster{Links: 6, Rechts: 4, Oben: 3}) {
			t.Errorf("das Polster ist %+v", ebene.Satz.Polster)
		}

		if ebene.Satz.Zeilenfaktoren != (render.Zeilenfaktoren{Weich: 0.9, Hart: 1.2}) {
			t.Errorf("die Zeilenfaktoren sind %+v", ebene.Satz.Zeilenfaktoren)
		}

		return
	}

	t.Fatal("der Plan führt keine Textebene")
}

// **Ein Wertskript, das einen Fehlerwert ergibt, wird gemeldet — nicht
// eingesetzt.**
//
// Bei MSE endet ein Wertskript in einer Zuweisung, und die ruft `toString()`;
// bei einem `ScriptDelayedError` wirft die (`src/script/value.cpp`, Zeile 148).
// `Value::update` fängt den Fehler je Feld, und der alte Wert bleibt stehen.
//
// **Was das gekostet hat:** ohne diese Prüfung landete der Wortlaut `Fehler` —
// der Name der Wertart — als Kartenwert im Plan. 28-mal über die 190 Stile von
// `M15-Magic-Pack` stand damit an einem Wahlfeld „der Wert Fehler hat kein
// Wahlbild" (#345).
//
// Fassung f9d9356 vom 2021-07-06.
func TestEinWertskriptMitFehlerwertWirdGemeldet(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Felder: []msepaket.Feldstil{{
				Feld: "name", Links: "0", Oben: "0", Breite: "50", Hoehe: "10",
				Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
			}},
		},
		Kartenfelder: []domain.Field{
			{Name: "name", Kind: domain.KindText, Skript: "card.gibtsnicht"},
		},
	}

	karte := &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}}}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "gibtsnicht") {
		t.Fatalf("Befunde: %v", befunde)
	}

	// **Der alte Wert bleibt stehen** — nicht der Name der Wertart.
	if len(plan.Ebenen) < 2 || plan.Ebenen[1].Klartext() != "Feuerball" {
		t.Errorf("gezeichnet wird %q, erwartet den alten Wert",
			plan.Ebenen[len(plan.Ebenen)-1].Klartext())
	}
}
