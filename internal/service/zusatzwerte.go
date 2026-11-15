package service

import (
	"path"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// zusatzdatenfeld ist der Name, unter dem eine Karte die Werte der
// Zusatzfelder trägt — `extra_data:` in der Datei, `extra data` im Bestand.
const zusatzdatenfeld = "extra_data"

// gespeicherteZusatzwerte gibt, was die Karte unter `extra_data:` für **ihre**
// Stilvorlage trägt — Feldname in kanonischer Form auf Wert
// (#2298).
//
// # Wie MSE2 es liest
//
// `extra_data` ist ein `DelayedIndexMaps` je Stilvorlage: der Block bleibt
// ungelesen, bis jemand die Werte **einer** Vorlage verlangt
// (`Card::extraDataFor`, `src/data/card.cpp`, Zeilen 63–65;
// `DelayedIndexMaps::get`, `src/util/delayed_index_maps.hpp`, Zeilen 19–33).
// Gefragt wird mit `stylesheet.name()` — dem Paketnamen ohne Endung
// (`Package::name`, `src/util/io/package.cpp`, Zeilen 43–53). Die Schlüssel
// liest `Reader` in kanonischer Form, Leerzeichen als Unterstrich
// (`src/util/io/reader.cpp`, Zeile 217). Fassung f9d9356 vom 2021-07-06.
//
// **Die Blöcke anderer Vorlagen bleiben unberührt** — MSE2 schreibt sie
// ungelesen zurück; hier steht der ganze Wert ohnehin unverändert im
// Bestand.
//
// **Ein Block, der sich nicht zerteilen lässt, wird benannt** (§8): dann
// gelten die Vorgaben, wie bei einer Karte ohne `extra_data:`, und der Befund
// sagt, warum.
func gespeicherteZusatzwerte(
	felder []msescript.Eintrag, stilpaket string, zusatzfelder []domain.Field,
) (map[string]string, []domain.Befund) {
	var text string

	for _, feld := range felder {
		if feld.Name == zusatzdatenfeld {
			text = feld.Wert.AlsText()
		}
	}

	return zusatzwerteAus(text, stilpaket, zusatzfelder)
}

// zusatzwerteAus liest die Werte einer Stilvorlage aus dem Text von
// `extra_data:` — der Kern von [gespeicherteZusatzwerte], auch für die
// Pflege ([Zusatzwertpflege]).
func zusatzwerteAus(text, stilpaket string, zusatzfelder []domain.Field) (map[string]string, []domain.Befund) {
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}

	baum, err := msepaket.Zerteile(zusatzdatenfeld, []byte(text))
	if err != nil {
		return nil, []domain.Befund{{
			Ort:     msepaket.Entkanonisch(zusatzdatenfeld),
			Meldung: "die Werte der Zusatzfelder ließen sich nicht lesen, es gelten die Vorgaben: " + err.Error(),
		}}
	}

	altformEinordnen(baum, zusatzfelder)

	vorlage := stilvorlagenname(stilpaket)

	for _, block := range baum.Wurzel {
		if block.Zeilenart != msepaket.ArtEintrag || msepaket.Kanonisch(block.Schluessel) != vorlage {
			continue
		}

		werte := make(map[string]string, len(block.Kinder))

		for _, kind := range block.Kinder {
			if kind.Zeilenart == msepaket.ArtEintrag {
				werte[msepaket.Kanonisch(kind.Schluessel)] = kind.Blocktext()
			}
		}

		return werte, nil
	}

	return nil, nil
}

// stilvorlagenname ist der Name, unter dem eine Karte die Werte einer
// Stilvorlage führt — der Paketname ohne Endung, in kanonischer Form
// (`Package::name`, `src/util/io/package.cpp`, Zeilen 43–53, Fassung f9d9356
// vom 2021-07-06).
func stilvorlagenname(stilpaket string) string {
	return msepaket.Kanonisch(strings.TrimSuffix(path.Base(stilpaket), path.Ext(stilpaket)))
}

// altformEinordnen bringt die flache Altform in ihre Gestalt — Befund 6 des
// Zweitblicks über E/F.
//
// Ein Bestand, der vor #2298 eingelesen wurde, trägt die Zusatzdaten ohne
// ihre zweite Ebene:
//
//	magic-nouveau:
//	keyword_icons: flying, haste
//
// **Ein Blockanfang ist eine Zeile ohne Wert und ohne Kinder, deren Name kein
// Feld dieses Stils ist** — eine Stilvorlage heißt nie wie ein Feld. Was
// danach auf derselben Ebene folgt, gehört zu ihm, bis der nächste beginnt.
// Die Zeilen verlieren dabei ihre Rohfassung: geschrieben werden sie mit der
// Einrückung ihrer neuen Tiefe. Eine Form mit zweiter Ebene bleibt, wie sie
// ist.
func altformEinordnen(baum *msepaket.Baum, zusatzfelder []domain.Field) {
	feldnamen := make(map[string]bool, len(zusatzfelder))
	for i := range zusatzfelder {
		feldnamen[msepaket.Kanonisch(zusatzfelder[i].Name)] = true
	}

	geordnet := make([]*msepaket.Knoten, 0, len(baum.Wurzel))

	var block *msepaket.Knoten

	for _, knoten := range baum.Wurzel {
		switch {
		case knoten.Zeilenart != msepaket.ArtEintrag:
			geordnet = append(geordnet, knoten)

		case len(knoten.Kinder) == 0 && strings.TrimSpace(knoten.Wert) == "" &&
			!feldnamen[msepaket.Kanonisch(knoten.Schluessel)]:
			block = knoten
			geordnet = append(geordnet, knoten)

		case block != nil && len(knoten.Kinder) == 0:
			knoten.Roh = ""
			block.Kinder = append(block.Kinder, knoten)

		default:
			block = nil
			geordnet = append(geordnet, knoten)
		}
	}

	baum.Wurzel = geordnet
}
