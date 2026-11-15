// **Die nächste und die vorige Karte** —
// #1290.
//
// # Portiert, nicht erfunden (§15)
//
// MSE2 hat das, und zwar seit jeher: `ID_CARD_PREV` und `ID_CARD_NEXT` stehen
// in `src/util/window_id.hpp`, Zeilen 98–99, und hängen im Kartenmenü
// (`src/gui/set/cards_panel.cpp`, Zeilen 68–69). Was sie tun, steht in
// `src/gui/control/item_list.cpp`, Zeilen 37–55, Fassung f9d9356 vom
// 2021-07-06:
//
//	bool ItemList::canSelectPrevious() const {
//	  return selected_item_pos - 1 >= 0;
//	}
//	bool ItemList::canSelectNext() const {
//	  return selected_item_pos >= 0 && (size_t)(selected_item_pos + 1) < sorted_list.size();
//	}
//
// Drei Dinge sind daran abzulesen, und alle drei stehen unten als Fall:
//
//   - **Eine Stelle in der sortierten Liste**, nicht der Nachbar im Bestand.
//     `sorted_list` ist ausdrücklich „pos->item"; hier ist es die Ordnung, die
//     `Kartenlistenaufbau` bringt.
//   - **Kein Umlauf.** Am Anfang gibt es kein Vorher, am Ende kein Nachher —
//     `cards_panel.cpp` prüft es sogar zweimal, mit dem Vermerk „Forwarded
//     events may cause this to occur even at the top."
//   - **Ohne Auswahl geht nichts.** `selected_item_pos` ist dann `-1`, und
//     `canSelectNext` gibt `false`; `canSelectPrevious` ebenfalls, weil
//     `-1 - 1 >= 0` nicht zutrifft.

import { describe, expect, it } from 'vitest'

import { nachbarkarte } from './kartennachbar'
import type { KartenzeileDTO } from '../api'

const zeile = (id: string): KartenzeileDTO =>
  ({ id, nummer: Number(id), aufschriften: [] }) as unknown as KartenzeileDTO

const drei = [zeile('1'), zeile('2'), zeile('3')]

describe('nachbarkarte', () => {
  it('gibt die nächste Karte in der Liste', () => {
    expect(nachbarkarte(drei, '2', 'vor')).toBe('3')
  })

  it('gibt die vorige Karte in der Liste', () => {
    expect(nachbarkarte(drei, '2', 'zurueck')).toBe('1')
  })

  it('gibt am Ende nichts — es gibt keinen Umlauf', () => {
    expect(nachbarkarte(drei, '3', 'vor')).toBeUndefined()
  })

  it('gibt am Anfang nichts — es gibt keinen Umlauf', () => {
    expect(nachbarkarte(drei, '1', 'zurueck')).toBeUndefined()
  })

  it('gibt ohne Auswahl nichts — `selected_item_pos` ist dann -1', () => {
    expect(nachbarkarte(drei, undefined, 'vor')).toBeUndefined()
    expect(nachbarkarte(drei, undefined, 'zurueck')).toBeUndefined()
  })

  it('gibt für eine Karte, die nicht in der Liste steht, nichts', () => {
    expect(nachbarkarte(drei, 'gibtsnicht', 'vor')).toBeUndefined()
  })

  it('gibt in einer leeren Liste nichts', () => {
    expect(nachbarkarte([], '1', 'vor')).toBeUndefined()
  })

  it('folgt der Reihenfolge der Liste und nicht der Kennung', () => {
    const umgedreht = [zeile('3'), zeile('1'), zeile('2')]

    expect(nachbarkarte(umgedreht, '3', 'vor')).toBe('1')
    expect(nachbarkarte(umgedreht, '1', 'zurueck')).toBe('3')
  })
})
