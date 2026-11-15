package scryfall_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Eine Scryfall-Karte wird auf Kartenfelder abgebildet**
// (#1005).
//
// # Woher die Zielnamen stammen
//
// Aus `card_fields` des Pakets `magic.mse-game` — einer **Datendatei**, nicht
// aus MSE2-Quelltext (§11). Sie sind dieselben, die `internal/mtgjson`
// benutzt; die Fundstellen stehen in `docs/FORMATE.md`.
func TestEineKarteWirdAufFelderAbgebildet(t *testing.T) {
	t.Parallel()

	karte, _, fehler := scryfall.LiesKarte(beispiel(t, "einfach.json"))
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	werte, befunde, ok := scryfall.ZuordnungFuer("magic")
	if !ok {
		t.Fatal("es gibt keine Zuordnung für magic")
	}

	if len(befunde) != 0 {
		t.Errorf("die Zuordnung meldet Befunde: %v", befunde)
	}

	gesetzt := werte.Werte(karte)

	erwartet := map[string]string{
		"name":         "Prüfstein",
		"casting cost": "2W",
		"super type":   "Kreatur",
		"sub type":     "Prüfling",
		"rule text":    "Fliegend.\nWenn Prüfstein ins Spiel kommt, ziehe eine Karte.",
		"flavor text":  "Ein Stein, an dem sich alles prüft.",
		"power":        "2",
		"toughness":    "3",
		"rarity":       "rare",
		"illustrator":  "Selbst Gemalt",
	}

	for feld, soll := range erwartet {
		if ist := gesetzt[feld]; ist != soll {
			t.Errorf("%s = %q, erwartet war %q", feld, ist, soll)
		}
	}
}

// **Die Manakosten kommen ohne Klammern an.**
//
// Scryfall schreibt `{2}{W}`, MSE schreibt `2W` — belegt aus
// `add_cards_scripts` in `magic.mse-game`. Die Regel steht in
// `internal/domain`, weil auch MTG JSON sie braucht.
func TestDieManakostenKommenOhneKlammernAn(t *testing.T) {
	t.Parallel()

	karte, _, fehler := scryfall.LiesKarte(beispiel(t, "einfach.json"))
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	zuordnung, _, _ := scryfall.ZuordnungFuer("magic")

	if kosten := zuordnung.Werte(karte)["casting cost"]; kosten != "2W" {
		t.Errorf("Manakosten %q, erwartet war %q", kosten, "2W")
	}
}

// **`mythic` heißt bei MSE `mythic rare`.**
//
// # Warum das die einzige Übersetzung ist
//
// Die anderen fünf Werte — `common`, `uncommon`, `rare`, `special`, `bonus` —
// heißen bei MSE genauso. Für `mythic` gilt es nicht: `card_fields` führt dort
// `mythic rare`. Dieselbe Abbildung wie bei MTG JSON.
func TestMythicHeisstMythicRare(t *testing.T) {
	t.Parallel()

	fuer := map[string]string{
		"mythic":   "mythic rare",
		"rare":     "rare",
		"common":   "common",
		"uncommon": "uncommon",
		"special":  "special",
		"bonus":    "bonus",
	}

	zuordnung, _, _ := scryfall.ZuordnungFuer("magic")

	for scryfallwert, msewert := range fuer {
		t.Run(scryfallwert, func(t *testing.T) {
			t.Parallel()

			karte := &scryfall.Karte{Name: "x", Seltenheit: scryfallwert}
			if ist := zuordnung.Werte(karte)["rarity"]; ist != msewert {
				t.Errorf("%q wurde zu %q, erwartet war %q", scryfallwert, ist, msewert)
			}
		})
	}
}

// **Die Typzeile wird an ihrem Gedankenstrich getrennt.**
//
// # Warum, und woran das belegt ist
//
// Scryfall liefert **eine** Zeile: „Legendary Creature — Human Wizard". MSE
// führt zwei Felder, `super type` und `sub type`, und `internal/mtgjson`
// füllt sie aus `supertypes`+`types` und `subtypes`. Der Gedankenstrich ist
// genau die Stelle, an der MTG JSON die beiden trennt.
//
// **Der Strich ist ein Halbgeviertstrich (U+2014), kein Bindestrich.** Ein
// Kartentyp kann einen Bindestrich enthalten („Time Lord Doctor" nicht, aber
// „Jump-Start" als Schlüsselwort); wer am Bindestrich träfe, zerschnitte ihn.
func TestDieTypzeileWirdAmGedankenstrichGetrennt(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct{ zeile, ober, unter string }{
		"mit Untertyp":       {"Kreatur — Prüfling", "Kreatur", "Prüfling"},
		"ohne Untertyp":      {"Spontanzauber", "Spontanzauber", ""},
		"mehrere Untertypen": {"Legendäre Kreatur — Mensch Zauberer", "Legendäre Kreatur", "Mensch Zauberer"},
		"leer":               {"", "", ""},
		"mit Bindestrich":    {"Artefakt — Ausrüstungs-Teil", "Artefakt", "Ausrüstungs-Teil"},
		"zwei Striche":       {"Kreatur — Held — Mensch", "Kreatur", "Held — Mensch"},
	}

	zuordnung, _, _ := scryfall.ZuordnungFuer("magic")

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gesetzt := zuordnung.Werte(&scryfall.Karte{Name: "x", Typzeile: fall.zeile})

			if ist := gesetzt["super type"]; ist != fall.ober {
				t.Errorf("super type = %q, erwartet war %q", ist, fall.ober)
			}

			if ist := gesetzt["sub type"]; ist != fall.unter {
				t.Errorf("sub type = %q, erwartet war %q", ist, fall.unter)
			}
		})
	}
}

// **Jedes übergangene Feld nennt seinen Grund.**
//
// # Warum ein Wächter und kein Vertrauen
//
// Ohne diese Prüfung wäre ein Feld, das jemand vergessen hat, von einem
// unterschieden, das absichtlich fehlt — und niemand wüsste, welches.
// „Auch eine Ablehnung braucht Deckung" (§10). Denselben Wächter hat
// `internal/mtgjson`.
func TestJedesUebergangeneFeldNenntSeinenGrund(t *testing.T) {
	t.Parallel()

	zuordnung, _, _ := scryfall.ZuordnungFuer("magic")

	for _, feld := range zuordnung.Felder {
		if feld.Feld == "" && strings.TrimSpace(feld.Grund) == "" {
			t.Errorf("das Scryfall-Feld %q wird übergangen und nennt keinen Grund", feld.Quelle)
		}

		if feld.Feld != "" && feld.Grund != "" {
			t.Errorf("das Scryfall-Feld %q hat ein Ziel **und** einen Grund", feld.Quelle)
		}
	}
}

// **Und jedes Feld des DTO steht in der Zuordnung — oder wird begründet
// übergangen.**
//
// # Warum das die eigentliche Zusage ist
//
// Ein Feld, das gelesen und dann nirgends erwähnt wird, ist unsichtbar: es
// steht im DTO, niemand ordnet es zu, und niemandem fällt es auf. Genau das
// war die Lage vor dem Fähigkeitskatalog aus Phase T, nur eine Ebene tiefer.
func TestJedesGeleseneFeldStehtInDerZuordnung(t *testing.T) {
	t.Parallel()

	zuordnung, _, _ := scryfall.ZuordnungFuer("magic")

	genannt := map[string]bool{}
	for _, feld := range zuordnung.Felder {
		genannt[feld.Quelle] = true
	}

	for _, feld := range scryfall.GeleseneFelder() {
		if !genannt[feld] {
			t.Errorf("das DTO liest %q, und die Zuordnung für magic erwähnt es nicht — "+
				"weder mit Ziel noch mit Grund", feld)
		}
	}
}

// **Ein unbekanntes Spiel bekommt keine erfundene Zuordnung.**
func TestEinUnbekanntesSpielBekommtKeineZuordnung(t *testing.T) {
	t.Parallel()

	if _, _, ok := scryfall.ZuordnungFuer("gibtsnicht"); ok {
		t.Error("für ein unbekanntes Spiel kam eine Zuordnung")
	}
}
