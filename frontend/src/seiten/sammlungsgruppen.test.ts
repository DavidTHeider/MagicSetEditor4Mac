// **Gruppen und Ansichten der Sammlung** — Werkstück U8 von Phase 5.1 (#1419):
// die Sicht kennt eine Ansicht (Tabelle, Galerie, Gruppen) und eine
// Gruppierung (Set, Lagerort, Zustand, Sprache); `gruppiert` liefert
// Abschnitte mit Kopf. Farbe und Typ warten auf A2 — ein Exemplar kennt heute
// keine Farbe (Vermerk 0085).
import { describe, expect, it } from 'vitest'

import { keineHerkunft } from '../apiherkunft'
import type { ExemplarDTO } from '../api'
import { gruppiert, leereSicht } from './sammlungssicht'

function exemplar(teile: Partial<ExemplarDTO> = {}): ExemplarDTO {
  return {
    id: 1,
    fassung: 1,
    printing: 'abc',
    druckID: '',
    vorlaeufig: false,
    menge: 1,
    zustand: 'nearMint',
    sprache: 'en',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: '',
    setcode: '',
    nummer: '',
    lagerort: '',
    wunsch: false,
    bezeichnung: 'abc',
    herkunft: keineHerkunft,
    ...teile,
  }
}

const bestand = [
  exemplar({ id: 1, setcode: 'm15', lagerort: 'Kiste A', menge: 2 }),
  exemplar({ id: 2, setcode: 'khm', lagerort: 'Kiste A', zustand: 'good' }),
  exemplar({ id: 3, setcode: 'm15', lagerort: '', sprache: 'de', menge: 4 }),
]

describe('Die Sicht', () => {
  it('beginnt als Tabelle ohne Gruppierung', () => {
    expect(leereSicht.ansicht).toBe('tabelle')
    expect(leereSicht.gruppierung).toBe('keine')
  })
})

describe('gruppiert', () => {
  it('gibt ohne Gruppierung einen einzigen Abschnitt ohne Schlüssel', () => {
    const abschnitte = gruppiert(bestand, 'keine')

    expect(abschnitte).toHaveLength(1)
    expect(abschnitte[0]?.schluessel).toBe('')
    expect(abschnitte[0]?.exemplare.map((e) => e.id)).toEqual([1, 2, 3])
    expect(abschnitte[0]?.menge).toBe(7)
  })

  it('gruppiert nach Set in der Reihenfolge der Liste und zählt die Mengen', () => {
    const abschnitte = gruppiert(bestand, 'set')

    expect(abschnitte.map((a) => a.schluessel)).toEqual(['m15', 'khm'])
    expect(abschnitte[0]?.exemplare.map((e) => e.id)).toEqual([1, 3])
    expect(abschnitte[0]?.menge).toBe(6)
    expect(abschnitte[1]?.menge).toBe(1)
  })

  // **Ein leerer Schlüssel ist eine eigene Gruppe** — „ohne Lagerort" ist
  // eine Antwort, kein Loch.
  it('führt Exemplare ohne Wert als eigene Gruppe am Ende', () => {
    const abschnitte = gruppiert(bestand, 'lagerort')

    expect(abschnitte.map((a) => a.schluessel)).toEqual(['Kiste A', ''])
    expect(abschnitte[1]?.exemplare.map((e) => e.id)).toEqual([3])
  })

  it.each([
    ['zustand', ['nearMint', 'good']],
    ['sprache', ['en', 'de']],
  ] as const)('gruppiert nach %s', (gruppierung, erwartet) => {
    expect(gruppiert(bestand, gruppierung).map((a) => a.schluessel)).toEqual(
      erwartet,
    )
  })

  it('gibt für eine leere Liste keine Abschnitte', () => {
    expect(gruppiert([], 'set')).toEqual([])
    expect(gruppiert([], 'keine')).toEqual([])
  })
})
