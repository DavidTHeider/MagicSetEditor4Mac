// **Die Karten in die berechnete Reihenfolge bringen** — Befund 1 des
// Zweitblicks über Phase 3.7
// (#1168).
import { describe, expect, it } from 'vitest'

import type { KartenzeileDTO } from '../api'
import { geordnet } from './listenordnung'

// zeile baut eine Kartenzeile mit einer Kennung.
function zeile(id: string): KartenzeileDTO {
  return { id, nummer: 0, name: id, felder: 0 }
}

// kennungen liest die Kennungen einer Liste.
function kennungen(karten: readonly KartenzeileDTO[]): string[] {
  return karten.map((karte) => karte.id)
}

describe('Die Karten in der berechneten Reihenfolge', () => {
  it('stellt sie um', () => {
    const karten = [zeile('1'), zeile('2'), zeile('3')]

    expect(kennungen(geordnet(karten, ['3', '1', '2']))).toEqual([
      '3',
      '1',
      '2',
    ])
  })

  // **Ohne Ordnung bleibt es, wie es kam.**
  //
  // Solange der Aufbau lädt, ist sie leer — und eine leere Ordnung darf die
  // Liste nicht leeren.
  it('lässt die Liste stehen, solange keine Ordnung dasteht', () => {
    const karten = [zeile('1'), zeile('2')]

    expect(geordnet(karten, [])).toBe(karten)
  })

  // **Was die Ordnung nicht kennt, steht hinten.**
  //
  // Eine Karte, die nach dem letzten Aufbau entstanden ist, verschwände sonst
  // aus der Liste, bis jemand neu lädt.
  it('hängt an, was die Ordnung nicht kennt', () => {
    const karten = [zeile('1'), zeile('2'), zeile('3')]

    expect(kennungen(geordnet(karten, ['3']))).toEqual(['3', '1', '2'])
  })

  // **Was es nicht gibt, wird übersprungen** — eine gelöschte Karte steht
  // noch in der Ordnung von vorhin.
  it('überspringt eine Kennung ohne Karte', () => {
    const karten = [zeile('1'), zeile('2')]

    expect(kennungen(geordnet(karten, ['9', '2', '1']))).toEqual(['2', '1'])
  })
})
