// **Die Rechnung hinter dem Markenfilter** —
// #1208.
import { describe, expect, it } from 'vitest'

import type { KartenzeileDTO } from '../api'
import { gefiltert, imSet, wirksam } from './markenauswahl'

// zeile gibt eine Kartenzeile, wie die Liste sie bekommt.
function zeile(id: string): KartenzeileDTO {
  return { id, nummer: 0, name: `Karte ${id}`, felder: 1 }
}

describe('imSet', () => {
  // **Jede Marke einmal** — auch wenn zehn Karten sie tragen.
  it('nennt jede Marke einmal', () => {
    expect(
      imSet({ '1': ['fertig'], '2': ['fertig'], '3': ['prüfen'] }, 'de'),
    ).toEqual(['fertig', 'prüfen'])
  })

  // **Alphabetisch, und über `localeCompare`** — `sort()` ohne Vergleich
  // ordnet nach Zeichencodes, und dort steht „ä" hinter „z".
  it('ordnet alphabetisch, auch mit Umlaut', () => {
    expect(
      imSet({ '1': ['zuletzt'], '2': ['änderung'], '3': ['anfang'] }, 'de'),
    ).toEqual(['änderung', 'anfang', 'zuletzt'])
  })

  // **Ohne Marken nichts** — dann zeichnet der Filter gar nicht erst.
  it('gibt ohne Marken eine leere Liste', () => {
    expect(imSet({}, 'de')).toEqual([])
    expect(imSet({ '1': [] }, 'de')).toEqual([])
  })
})

describe('gefiltert', () => {
  const karten = [zeile('1'), zeile('2')]
  const jeKarte = { '1': ['fertig'] }

  // **Leer heißt: alle.**
  it('lässt ohne Marke alles stehen', () => {
    expect(gefiltert(karten, jeKarte, '')).toBe(karten)
  })

  it('nimmt weg, was die Marke nicht trägt', () => {
    expect(gefiltert(karten, jeKarte, 'fertig').map((eine) => eine.id)).toEqual(
      ['1'],
    )
  })

  // **Eine Marke, die es nicht gibt, ergibt nichts** — und der Filter
  // verschwindet dann von selbst, weil `imSet` sie nicht mehr anbietet.
  it('gibt für eine unbekannte Marke nichts', () => {
    expect(gefiltert(karten, jeKarte, 'gibtesnicht')).toEqual([])
  })
})

describe('wirksam', () => {
  // **Eine Marke, die es noch gibt, filtert.**
  it('lässt eine angebotene Marke stehen', () => {
    expect(wirksam(['fertig', 'prüfen'], 'fertig')).toBe('fertig')
  })

  // **Eine, die verschwunden ist, filtert nicht mehr.**
  //
  // Wer die letzte Karte entmarkiert, nimmt die Marke aus dem Set — die
  // Kachelreihe zeichnet sie nicht mehr. Bliebe die Wahl stehen, wäre die
  // Liste **leer und ohne Weg zurück**: der Knopf für »Alle« ist mit ihr
  // verschwunden. Im laufenden Fenster gesehen (§3).
  it('gibt eine verschwundene Marke frei', () => {
    expect(wirksam([], 'fertig')).toBe('')
    expect(wirksam(['prüfen'], 'fertig')).toBe('')
  })

  // **Und »alle« bleibt »alle«.**
  it('lässt die leere Wahl leer', () => {
    expect(wirksam(['fertig'], '')).toBe('')
  })
})
