package service

// **Die setweiten Stiloptionen gehen in den Block `styling:`** — F05
// (#2247).
//
// MSE2 legt sie je Stilpaket ab:
//
//	IndexMap<FieldP, ValueP>& Set::stylingDataFor(const StyleSheet& stylesheet) {
//	  return styling_data.get(stylesheet.name(), stylesheet.styling_fields);
//	}
//
// `src/data/set.cpp`, Zeilen 99–101, Fassung f9d9356 vom 2021-07-06 — und so
// steht es in jedem kanonischen Set:
//
//	styling:
//		magic-m15-altered:
//			text_box_mana_symbols: magic-mana-small.mse-symbol-font
//
// **Bis zum 2026-10-10 kam eine hier gesetzte Option gar nicht in die Datei**,
// und der eingelesene Block stand zusätzlich unter `set info:`, wo MSE2 ihn
// mit „Unexpected key: 'styling'" abwies (Setdatei-Rundlauf gegen MSE2,
// `internal/flaechen/setrundlauf_fremd_test.go`).
//
// # Was hier **nicht** geschieht
//
// **Eine zurückgesetzte Option verschwindet nicht aus dem Block.** Was das Set
// aus seiner Datei mitbrachte, liegt als Setwert [Setstilfeld] und nicht unter
// den Optionen, die hier gesetzt wurden; zurückgesetzt wird nur diese. Dieselbe
// Regel wie bei den Setangaben (`setwertebloecke.go`): es wird keine Zeile
// entfernt, die niemand angeordnet hat.

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// stilblockAngleichen schreibt die hier gesetzten Stiloptionen unter das
// Stilpaket des Sets.
//
// **Ohne Option bleibt der Baum, wie er ist** — die Bytezusage gilt für alles,
// was niemand angefasst hat
// ([0002](../../docs/entscheidungen/0002-sqlite-als-arbeitsformat.md)), und
// eine Option, deren Wert schon dasteht, ändert keine Zeile.
func stilblockAngleichen(baum *msepaket.Baum, kopf domain.SetHead, stilwerte []domain.Value, aufgeloest string) {
	if len(stilwerte) == 0 {
		return
	}

	// **Ein Zeilenende für jede neue Zeile** — das der Datei. Eine neue Ebene
	// hat noch keine Zeile, an der sich ihr eigenes ablesen ließe.
	ende := zeilenendeVon(baum.Wurzel)

	block := baum.Ersten(Setstilfeld)
	if block == nil {
		block = &msepaket.Knoten{Zeilenart: msepaket.ArtEintrag, Schluessel: Setstilfeld}
		baum.Wurzel = angehaengt(baum.Wurzel, block, ende)
	}

	paket := stilpaketImBlock(block, kopf, ende, aufgeloest)

	for _, wert := range stilwerte {
		zeile := paket.Ersten(wert.Field)

		switch {
		case zeile == nil:
			paket.Kinder = angehaengt(paket.Kinder, &msepaket.Knoten{
				Zeilenart:  msepaket.ArtEintrag,
				Schluessel: msepaket.Kanonisch(wert.Field),
				Wert:       wert.Text,
			}, ende)

		case wertText(zeile) != wert.Text:
			// **Der alte Block geht mit, wenn er der Wert war** — derselbe
			// Grund wie an der Karte (#704).
			if zeile.Wert == "" {
				zeile.Kinder = nil
			}

			zeile.Roh = ""
			zeile.Wert = wert.Text
		}
	}
}

// stilpaketImBlock gibt den Knoten des Stilpakets — unter dem aufgelösten
// Paketnamen, sonst unter dem Namen mit Spiel davor oder dem rohen, in dieser
// Reihenfolge ([stilpaketnamen]); fehlt er, wird er unter dem aufgelösten
// angelegt, ohne Auflösung mit Spiel davor, so wie `magic-set-generator` und
// jedes kanonische Set ihn schreiben.
//
// **Ist er aufgelöst, gilt nur sein Name** (Befund 2 des Zweitblicks über
// Phase G, zweiter Teil, und Befund 1 des Nachlaufs): stand im Bestand der
// volle Paketname, wurde aus `magic-m15` sonst `magic-magic-m15`, und ein
// solcher Block aus einer älteren Sicherung gewann noch gegen den richtigen.
// Gelesen wird die Option unter dem Namen des Pakets ([setstilwerte]); ein
// Block unter einem anderen bleibt, wie er steht.
func stilpaketImBlock(block *msepaket.Knoten, kopf domain.SetHead, ende, aufgeloest string) *msepaket.Knoten {
	namen := stilpaketnamen(kopf.Style, kopf.Game)
	if aufgeloest != "" {
		namen = []string{aufgeloest}
	}

	for _, name := range namen {
		if paket := block.Ersten(name); paket != nil {
			return paket
		}
	}

	paket := &msepaket.Knoten{Zeilenart: msepaket.ArtEintrag, Schluessel: namen[0]}
	block.Kinder = angehaengt(block.Kinder, paket, ende)

	return paket
}

// angehaengt hängt einen selbst gebauten Knoten an eine Ebene.
//
// **Die Zeile davor bekommt ihr Ende** — die letzte Zeile einer Datei hat oft
// keins, und ohne es klebte die neue an ihr. Gemeint ist die letzte Zeile des
// ganzen Teilbaums, nicht das letzte Kind: unter ihm kann eingerückt noch
// etwas stehen.
func angehaengt(ebene []*msepaket.Knoten, neu *msepaket.Knoten, ende string) []*msepaket.Knoten {
	if letzte := letzteZeile(ebene); letzte != nil && letzte.Zeilenende == "" {
		letzte.Zeilenende = ende
	}

	neu.Zeilenende = ende

	return append(ebene, neu)
}

func letzteZeile(ebene []*msepaket.Knoten) *msepaket.Knoten {
	if len(ebene) == 0 {
		return nil
	}

	letzte := ebene[len(ebene)-1]
	if tiefer := letzteZeile(letzte.Kinder); tiefer != nil {
		return tiefer
	}

	return letzte
}
