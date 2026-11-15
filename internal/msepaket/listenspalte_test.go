package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Was ein Kartenfeld über seine Listenspalte sagt** — Werkstück 3 von
// Phase 3.7
// (#1168).
//
//	int       card_list_column; ///< What column to use in the card list?
//	UInt      card_list_width;  ///< Width of the card list column (pixels).
//	bool      card_list_visible;///< Is this field shown in the card list?
//	bool      card_list_allow;  ///< Is this field allowed to appear in the card list?
//	Alignment card_list_align;  ///< Alignment of the card list colummn.
//
// `src/data/field.hpp`, Zeilen 55–60, Fassung f9d9356 vom 2021-07-06.
//
// Über `Full-Magic-Pack` gezählt: 284-mal `card list column:`, 277-mal
// `card list width:`, 56-mal `card list visible:`, 16-mal `card list allow:`
// und 14-mal `card list alignment:`.
func TestEinKartenfeldTraegtSeineListenspalte(t *testing.T) {
	t.Parallel()

	pfad := probespielpaket(t, `mse version: 2.0.0
card field:
	type: text
	name: name
	card list name: Name
	card list visible: true
	card list column: 100
	card list width: 150
	card list alignment: left
card field:
	type: text
	name: nummer
	card list column: 0200
	card list width: 60
	card list alignment: right
card field:
	type: text
	name: geheim
	card list allow: false
card field:
	type: text
	name: schlicht
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	spiel, befunde := paket.Spielmodell()
	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	nach := func(name string) domain.Field {
		t.Helper()

		for _, feld := range spiel.Fields {
			if feld.Name == name {
				return feld
			}
		}

		t.Fatalf("das Feld %q fehlt", name)

		return domain.Field{}
	}

	t.Run("das sichtbare trägt alles", func(t *testing.T) {
		feld := nach("name")

		if !feld.Listensichtbar || feld.Listenstelle != 100 ||
			feld.Listenbreite != 150 || feld.Listenausrichtung != domain.ListeLinks {
			t.Errorf("%+v", feld)
		}

		// **Ohne Angabe erlaubt** — MSE setzt `card_list_allow` im
		// Konstruktor auf `true` (`src/data/field.cpp`, `Field::Field`, Fassung
		// f9d9356 vom 2021-07-06).
		if !feld.Listenerlaubt {
			t.Error("das Feld gilt als nicht erlaubt")
		}
	})

	t.Run("eine Stelle mit führender Null ist eine Zahl", func(t *testing.T) {
		feld := nach("nummer")

		if feld.Listenstelle != 200 {
			t.Errorf("Stelle %d, erwartet 200", feld.Listenstelle)
		}

		if feld.Listenausrichtung != domain.ListeRechts {
			t.Errorf("Ausrichtung %q", feld.Listenausrichtung)
		}

		// **Ohne Angabe unsichtbar** — MSE setzt `card_list_visible` auf
		// `false`; nur wer es schreibt, will die Spalte von Anfang an sehen.
		if feld.Listensichtbar {
			t.Error("das Feld gilt als sichtbar")
		}
	})

	t.Run("ein verbotenes Feld sagt es", func(t *testing.T) {
		if nach("geheim").Listenerlaubt {
			t.Error("das verbotene Feld gilt als erlaubt")
		}
	})

	t.Run("ohne jede Angabe gelten die Vorgaben", func(t *testing.T) {
		feld := nach("schlicht")

		if !feld.Listenerlaubt || feld.Listensichtbar || feld.Listenstelle != 0 ||
			feld.Listenbreite != 0 || feld.Listenausrichtung != domain.ListeLinks {
			t.Errorf("%+v", feld)
		}
	})
}

// **Eine unlesbare Zahl ist eine Null und kein Fehler.**
//
// §8: fremde Daten sind der Normalfall. Ein Paket von 2007 mit `card list
// column: viele` soll sich öffnen lassen — die Spalte steht dann eben vorn.
func TestEineUnlesbareListenangabeIstEineNull(t *testing.T) {
	t.Parallel()

	pfad := probespielpaket(t, `mse version: 2.0.0
card field:
	type: text
	name: name
	card list column: viele
	card list width: breit
	card list alignment: schräg
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	spiel, befunde := paket.Spielmodell()
	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	feld := spiel.Fields[0]
	if feld.Listenstelle != 0 || feld.Listenbreite != 0 ||
		feld.Listenausrichtung != domain.ListeLinks {
		t.Errorf("%+v", feld)
	}
}
