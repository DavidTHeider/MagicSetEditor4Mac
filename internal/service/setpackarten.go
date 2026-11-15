// **Die Packarten eines Sets** — Lesen aus der Datei, Schreiben zurück
// (#1081).
//
// Sie fielen bis zum 2026-08-25 in den Rohweg: byte-genau erhalten und im
// Bestand unsichtbar. `Packbereich` kannte nur `pack type:` aus dem
// Spielpaket, und ein Set, das eine eigene Art nennt, zog aus der falschen.

package service

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// packartAus liest einen `pack type:`-Block eines Sets.
//
// **Über [msepaket.PackAus] und nicht neu geschrieben**: dieselbe Auflösung
// von `auto`, dieselben Vorgaben für `amount` und `weight`. Zwei Fassungen
// derselben Regel liefen auseinander, und dann zöge dasselbe Pack aus einem
// Spielpaket anders als aus einem Set.
func packartAus(knoten *msepaket.Knoten) *domain.PackArt {
	gelesen := msepaket.PackAus(knoten)

	art := &domain.PackArt{
		Name:       gelesen.Name,
		Select:     gelesen.Auswahl,
		Filter:     gelesen.Filter,
		Selectable: gelesen.Waehlbar,
		Items:      make([]domain.PackItem, 0, len(gelesen.Posten)),
	}

	for _, posten := range gelesen.Posten {
		art.Items = append(art.Items, domain.PackItem{
			Name: posten.Name, Amount: posten.Anzahl, Weight: posten.Gewicht,
		})
	}

	return art
}

// AlsPaketpacks übersetzt die Packarten eines Sets in die Gestalt, die der
// Packzieher kennt.
//
// **Damit `Packbereich` eine Liste sieht und nicht zwei Gestalten.** Die
// Vorrangregel steht dort; hier steht nur die Übersetzung.
func AlsPaketpacks(arten []*domain.PackArt) []msepaket.Pack {
	packs := make([]msepaket.Pack, 0, len(arten))

	for _, art := range arten {
		posten := make([]msepaket.Packposten, 0, len(art.Items))
		for _, item := range art.Items {
			posten = append(posten, msepaket.Packposten{
				Name: item.Name, Anzahl: item.Amount, Gewicht: item.Weight,
			})
		}

		packs = append(packs, msepaket.Pack{
			Name:     art.Name,
			Auswahl:  art.Select,
			Filter:   art.Filter,
			Posten:   posten,
			Waehlbar: art.Selectable,
		})
	}

	return packs
}

// PackartenMitVorrang legt die Packarten eines Sets vor die des Spiels.
//
// # Die Regel
//
// MSE sucht eine Packart **zuerst im Set**, dann im Spiel:
//
//	FOR_EACH_CONST(type, set->pack_types)       { if (type->name == name) return &*type; }
//	FOR_EACH_CONST(type, set->game->pack_types) { if (type->name == name) return &*type; }
//
// `src/data/pack.cpp`, Zeilen 373–392, Fassung f9d9356 vom 2021-07-06. Das
// Feld heißt dort *„Additional/replacement pack types"* (`src/data/set.hpp`,
// Zeile 57) — **beides**: eine neue Art kommt hinzu, eine gleichnamige
// ersetzt.
//
// # Warum die eigenen vorn stehen
//
// Weil MSE sie zuerst durchläuft, und weil die Reihenfolge dieser Liste die
// ist, in der die Oberfläche sie anbietet. Eine Art, die eine des Spiels
// ersetzt, an deren Stelle zu setzen wäre eine zweite Regel — und MSE hat
// keine.
//
// # Was diese Funktion nicht tut
//
// Sie zieht nicht und sie filtert nicht. Sie stellt die Liste auf; alles
// Weitere macht [Packzieher] mit derselben Liste, gleich woher eine Art kommt.
func PackartenMitVorrang(eigene []*domain.PackArt, spiel []msepaket.Pack) []msepaket.Pack {
	umgesetzt := AlsPaketpacks(eigene)

	// **Die erste gewinnt**, auch unter den eigenen: MSE bricht die Schleife
	// beim ersten Treffer ab (`return &*type;`, Zeile 377).
	gesehen := make(map[string]bool, len(umgesetzt))

	arten := make([]msepaket.Pack, 0, len(umgesetzt)+len(spiel))

	for _, art := range umgesetzt {
		if gesehen[art.Name] {
			continue
		}

		gesehen[art.Name] = true

		arten = append(arten, art)
	}

	for _, art := range spiel {
		if gesehen[art.Name] {
			continue
		}

		gesehen[art.Name] = true

		arten = append(arten, art)
	}

	return arten
}
