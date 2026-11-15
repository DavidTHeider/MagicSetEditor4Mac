// **Das Symbol unter dem Schreibzeiger** — Werkstück U7 von Phase 5.1
// (#1418).
//
// # Warum über die Symbole aus Go und nicht über das Wort
//
// Ein Feld heißt `casting cost` — mit Leerzeichen — und im Skript
// `casting_cost`. Wer nur das Wort unter dem Zeiger nähme, fände in der
// Stildatei `cost` und im Skript `casting_cost`, und keines von beiden stünde
// im Index. Deshalb wird umgekehrt gefragt: **welches belegte Symbol deckt
// den Zeiger** — in beiden Schreibweisen, als ganzes Wort, und bei mehreren
// das längste. Erst wenn keines passt, bleibt das Wort selbst: dann ist es
// ein Name aus der Bibliothek oder einem eingebundenen Paket, und der
// Inspector sagt das.
import { istWortzeichen } from './wortstelle'

import type { PaketsymbolDTO } from '../api'

// Symbolstelle ist der Name unter dem Zeiger — mit Definition, wenn das Paket
// eine hat.
export interface Symbolstelle {
  name: string
  symbol: PaketsymbolDTO | undefined
}

// symbolAnStelle sagt, welchen Namen der Zeiger an `stelle` berührt.
export function symbolAnStelle(
  text: string,
  stelle: number,
  symbole: readonly PaketsymbolDTO[],
): Symbolstelle | null {
  if (stelle < 0 || stelle > text.length) {
    return null
  }

  const von = text.lastIndexOf('\n', stelle - 1) + 1
  const ende = text.indexOf('\n', stelle)
  const zeile = text.slice(von, ende === -1 ? text.length : ende)
  const position = stelle - von

  let bestes: { symbol: PaketsymbolDTO; laenge: number } | null = null

  for (const symbol of symbole) {
    for (const form of schreibweisen(symbol.name)) {
      if (
        deckt(zeile, form, position) &&
        (bestes === null || form.length > bestes.laenge)
      ) {
        bestes = { symbol, laenge: form.length }
      }
    }
  }

  if (bestes !== null) {
    return { name: bestes.symbol.name, symbol: bestes.symbol }
  }

  const wort = wortUm(zeile, position)

  return wort === '' ? null : { name: wort, symbol: undefined }
}

// schreibweisen gibt die Form der Datei und die des Skripts — einmal, wenn
// sie gleich sind.
function schreibweisen(name: string): string[] {
  const skript = name.replaceAll(' ', '_')

  return skript === name ? [name] : [name, skript]
}

// deckt sagt, ob `form` als ganzes Wort in der Zeile steht und die Position
// einschließt.
function deckt(zeile: string, form: string, position: number): boolean {
  for (
    let fund = zeile.indexOf(form);
    fund !== -1;
    fund = zeile.indexOf(form, fund + 1)
  ) {
    const bis = fund + form.length
    const ganz =
      (fund === 0 || !istWortzeichen(zeile.charAt(fund - 1))) &&
      (bis === zeile.length || !istWortzeichen(zeile.charAt(bis)))

    if (ganz && fund <= position && position <= bis) {
      return true
    }
  }

  return false
}

// wortUm gibt das Wort, in dem oder an dessen Rand die Position steht.
function wortUm(zeile: string, position: number): string {
  let von = position
  let bis = position

  while (von > 0 && istWortzeichen(zeile.charAt(von - 1))) {
    von -= 1
  }

  while (bis < zeile.length && istWortzeichen(zeile.charAt(bis))) {
    bis += 1
  }

  return zeile.slice(von, bis)
}
