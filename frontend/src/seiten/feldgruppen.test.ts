import { describe, expect, it } from 'vitest'

import { feldform } from '../test/feldproben'
import { gruppiereFelder } from './feldgruppen'

describe('Feldgruppen', () => {
  // **Eine Reihe wird zu einer Gruppe, überschrieben mit dem Grundfeld.**
  it('fasst die Glieder einer Reihe zusammen', () => {
    const gruppen = gruppiereFelder([
      feldform({ name: 'name', aufschrift: 'Name', reihe: 'name', stelle: 1 }),
      feldform({
        name: 'name 2',
        aufschrift: 'Name 2',
        reihe: 'name',
        stelle: 2,
      }),
      feldform({ name: 'rarity', aufschrift: 'Rarity' }),
    ])

    expect(gruppen).toHaveLength(2)
    expect(gruppen[0]?.aufschrift).toBe('Name')
    expect(gruppen[0]?.felder.map((eines) => eines.name)).toEqual([
      'name',
      'name 2',
    ])
    expect(gruppen[1]?.reihe).toBe('')
  })

  // **Ein Feld ohne Geschwister in der Liste bekommt keine Überschrift.**
  //
  // Die Massenänderung bietet nur editierbare und gespeicherte Felder an; von
  // einer Reihe kann deshalb ein einziges Glied übrig bleiben.
  it('lässt ein allein gebliebenes Glied ohne Gruppe', () => {
    const gruppen = gruppiereFelder([
      feldform({ name: 'name 2', reihe: 'name', stelle: 2 }),
    ])

    expect(gruppen).toHaveLength(1)
    expect(gruppen[0]?.reihe).toBe('')
  })

  // **Die Ordnung des Spielpakets bleibt.** Wer über die ganze Liste sammelte,
  // zöge ein spätes Glied nach oben — und änderte eine Reihenfolge, die aus
  // der Datei kommt.
  it('zieht ein getrenntes Glied nicht nach oben', () => {
    const gruppen = gruppiereFelder([
      feldform({ name: 'name', reihe: 'name', stelle: 1 }),
      feldform({ name: 'rarity' }),
      feldform({ name: 'name 2', reihe: 'name', stelle: 2 }),
    ])

    expect(gruppen.map((gruppe) => gruppe.felder.map((f) => f.name))).toEqual([
      ['name'],
      ['rarity'],
      ['name 2'],
    ])
  })

  // **Und die Überschrift fällt auf das Grundfeld zurück**, wenn das erste
  // Glied der Gruppe nicht das Grundfeld ist.
  it('überschreibt mit dem Grundfeld, wenn dieses fehlt', () => {
    const gruppen = gruppiereFelder([
      feldform({ name: 'name 2', reihe: 'name', stelle: 2 }),
      feldform({ name: 'name 3', reihe: 'name', stelle: 3 }),
    ])

    expect(gruppen[0]?.aufschrift).toBe('name')
  })
})
