// **Vertragstests gegen eingecheckte, selbst geschriebene Beispiele** —
// Abnahmekriterium aus #888.
//
// # Warum selbst geschrieben und nicht heruntergeladen
//
// §11: fremde Inhalte kommen nicht ins Repositorium. Die Beispiele haben
// **dieselbe Gestalt** wie eine echte Antwort — Feldnamen, Verschachtelung,
// `null` an den Stellen, die die Dokumentation als nullable führt — und einen
// erfundenen Inhalt.
//
// # Was sie halten und was nicht
//
// Sie halten, dass dieses Programm die **dokumentierte** Gestalt liest. Ob
// Scryfall sie noch so schickt, halten sie nicht — dafür gibt es den Lauf
// hinter dem Build-Tag `netzlauf`.
package scryfall_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

func TestVertragRulings(t *testing.T) {
	t.Parallel()

	hinweise, befunde, fehler := scryfall.LiesRulings(beispiel(t, "rulings.json"))
	if fehler != nil {
		t.Fatalf("LiesRulings: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde = %+v", befunde)
	}

	if len(hinweise) != 2 {
		t.Fatalf("%d Hinweise", len(hinweise))
	}

	if hinweise[0].Quelle != scryfall.QuelleWotc ||
		hinweise[1].Quelle != scryfall.QuelleScryfall {
		t.Errorf("Quellen = %q / %q", hinweise[0].Quelle, hinweise[1].Quelle)
	}

	if hinweise[0].Datum != "2004-10-04" || hinweise[0].OrakelID == "" {
		t.Errorf("erster Hinweis = %+v", hinweise[0])
	}
}

func TestVertragSetangabe(t *testing.T) {
	t.Parallel()

	satz, befunde, fehler := scryfall.LiesSetangabe(beispiel(t, "setangabe.json"))
	if fehler != nil {
		t.Fatalf("LiesSetangabe: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde = %+v", befunde)
	}

	if satz.Code != "prf" || satz.Name != "Prüfsatz" || satz.Art != "expansion" {
		t.Errorf("Satz = %+v", satz)
	}

	if satz.Karten != 3 || satz.Digital {
		t.Errorf("Karten = %d, Digital = %v", satz.Karten, satz.Digital)
	}

	if satz.Erschienen != "2026-08-24" {
		t.Errorf("Erschienen = %q — das Datum wird nicht umgerechnet", satz.Erschienen)
	}
}

func TestVertragSymbologie(t *testing.T) {
	t.Parallel()

	symbole, befunde, fehler := scryfall.LiesSymbole(beispiel(t, "symbologie.json"))
	if fehler != nil {
		t.Fatalf("LiesSymbole: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde = %+v", befunde)
	}

	if len(symbole) != 3 {
		t.Fatalf("%d Symbole", len(symbole))
	}

	// **`{W}` hat einen Manawert von 1.**
	if symbole[0].Manawert == nil || *symbole[0].Manawert != 1 {
		t.Errorf("Manawert von {W} = %v", symbole[0].Manawert)
	}

	// **`{T}` hat gar keinen** — und das ist etwas anderes als null.
	if symbole[1].Manawert != nil {
		t.Errorf("Manawert von {T} = %v, erwartet keinen", *symbole[1].Manawert)
	}

	if symbole[1].IstMana || symbole[1].Vektoradresse != "" {
		t.Errorf("{T} = %+v", symbole[1])
	}

	// **`{20}` kommt in keinem Manakosten vor**, obwohl es ein Manasymbol ist.
	// Genau daran hängt jede Prüfung eines geschriebenen Manakostens.
	if !symbole[2].IstMana || symbole[2].InManakosten {
		t.Errorf("{20} = %+v", symbole[2])
	}
}

// **Ein neues Feld bricht nichts, ein fehlendes Pflichtfeld bricht laut.**
//
// Das ist die Driftregel aus §16, und sie gilt für jede der drei Gestalten —
// deshalb steht sie hier einmal für alle drei und nicht dreimal verstreut.
func TestDriftregelGiltFuerAlleDreiGestalten(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		mitNeuem    string
		ohnePflicht string
		lies        func([]byte) error
	}{
		"Regelhinweise": {
			mitNeuem: `{"object":"list","data":[{"object":"ruling","source":"wotc",
			           "comment":"Da.","voellig_neu":{"tief":1}}]}`,
			ohnePflicht: `{"object":"list","data":[{"object":"ruling","source":"wotc"}]}`,
			lies: func(roh []byte) error {
				_, _, fehler := scryfall.LiesRulings(roh)

				return fehler
			},
		},
		"Setangabe": {
			mitNeuem:    `{"object":"set","code":"prf","name":"Prüfsatz","voellig_neu":[1,2]}`,
			ohnePflicht: `{"object":"set","name":"Prüfsatz"}`,
			lies: func(roh []byte) error {
				_, _, fehler := scryfall.LiesSetangabe(roh)

				return fehler
			},
		},
		"Symbologie": {
			mitNeuem:    `{"object":"list","data":[{"object":"card_symbol","symbol":"{W}","neu":true}]}`,
			ohnePflicht: `{"object":"list","data":[{"object":"card_symbol","english":"weiß"}]}`,
			lies: func(roh []byte) error {
				_, _, fehler := scryfall.LiesSymbole(roh)

				return fehler
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name+" — ein neues Feld bricht nichts", func(t *testing.T) {
			t.Parallel()

			if fehler := fall.lies([]byte(fall.mitNeuem)); fehler != nil {
				t.Errorf("das neue Feld hat gebrochen: %v", fehler)
			}
		})

		t.Run(name+" — ein fehlendes Pflichtfeld bricht laut", func(t *testing.T) {
			t.Parallel()

			if fehler := fall.lies([]byte(fall.ohnePflicht)); fehler == nil {
				t.Error("das fehlende Pflichtfeld kam durch")
			}
		})
	}
}
