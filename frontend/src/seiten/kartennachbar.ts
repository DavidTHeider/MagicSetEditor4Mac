// Die nächste und die vorige Karte in der Liste.
//
// **ported** aus `src/gui/control/item_list.cpp`, Zeilen 37–55, Fassung
// f9d9356 vom 2021-07-06
// (#1290).
//
// **Eine eigene Datei neben `kartenauswahl.ts`**, weil sie eine eigene Frage
// beantwortet: dort steht, was nach einem **Löschen** aufgeschlagen wird — die
// nachrückende Karte —, hier, wohin ein Tastendruck führt. Beides sieht in der
// Mitte gleich aus und ist an den Rändern verschieden: das Nachrücken greift am
// Ende auf die letzte verbliebene zurück, das Weiterblättern hört auf.

import type { KartenzeileDTO } from '../api'

// Richtung sagt, wohin geblättert wird.
export type Richtung = 'vor' | 'zurueck'

// nachbarkarte gibt die Kennung der nächsten oder vorigen Karte.
//
// `undefined`, wenn es keine gibt — am Anfang, am Ende, ohne Auswahl und bei
// einer Karte, die nicht in der Liste steht.
//
// # Kein Umlauf, und das ist portiert
//
//	bool ItemList::canSelectPrevious() const {
//	  return selected_item_pos - 1 >= 0;
//	}
//
// MSE2 prüft die Grenze **zweimal** — einmal beim Aktivieren des Menüeintrags
// und noch einmal beim Ausführen, mit dem Vermerk „Forwarded events may cause
// this to occur even at the top" (`src/gui/set/cards_panel.cpp`, Zeilen
// 295–302). Ein Umlauf wäre bequem und stünde in keiner Quelle.
//
// # Die Liste ist die **sortierte**
//
// `sorted_list` ist bei MSE2 ausdrücklich „a map: pos->item"
// (`item_list.hpp`, Zeile 109). Hier ist das die Liste, die [geordnet] aus dem
// Kartenlistenaufbau gemacht hat — nicht die Reihenfolge im Bestand und nicht
// die der Kennungen.
export function nachbarkarte(
  karten: readonly KartenzeileDTO[],
  gewaehlt: string | undefined,
  richtung: Richtung,
): string | undefined {
  if (gewaehlt === undefined) {
    return undefined
  }

  const stelle = karten.findIndex((karte) => karte.id === gewaehlt)
  if (stelle < 0) {
    return undefined
  }

  return karten[stelle + (richtung === 'vor' ? 1 : -1)]?.id
}
