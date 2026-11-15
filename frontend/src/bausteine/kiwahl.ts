// **Was in einer Modellwahl steckt** — herausgezogen für
// #1539.
//
// # Warum eine eigene Datei
//
// **Weil `anbieter/alias` an drei Stellen aufgetrennt wurde** — Textvorschlag,
// Bildvorschlag und seit dieser Änderung die Kostenfrage. Dreimal dasselbe
// `indexOf('/')` ist dreimal dieselbe Gelegenheit, den Sonderfall zu
// vergessen: der **Alias** darf Schrägstriche tragen, der Anbieter nicht.
import type { KIModellDTO } from '../antworten'

// anbieterAus gibt den Anbieterteil einer Wahl.
export function anbieterAus(gewaehlt: string): string {
  const trennung = gewaehlt.indexOf('/')

  return trennung < 0 ? gewaehlt : gewaehlt.slice(0, trennung)
}

// aliasAus gibt den Rest — **mit** seinen Schrägstrichen.
export function aliasAus(gewaehlt: string): string {
  const trennung = gewaehlt.indexOf('/')

  return trennung < 0 ? '' : gewaehlt.slice(trennung + 1)
}

// kostetGeld sagt, ob das gewählte Modell anderswo rechnet.
//
// **Aus dem Katalog und nicht aus dem Namen.** Wo ein Anbieter rechnet, sagt
// er selbst (`ort`); eine Liste von Anbieternamen hier wäre eine zweite
// Wahrheit, die beim nächsten Anbieter ausbliebe.
//
// **Und eine unbekannte Wahl kostet nichts** — dann gibt es nichts zu fragen,
// und der Knopf steht ohnehin still.
export function kostetGeld(modelle: KIModellDTO[], gewaehlt: string): boolean {
  return modelle.some(
    (modell) =>
      `${modell.anbieter}/${modell.alias}` === gewaehlt &&
      modell.ort !== derOrtLokal,
  )
}

// derOrtLokal ist der Ort, an dem nichts kostet.
export const derOrtLokal = 'lokal'
