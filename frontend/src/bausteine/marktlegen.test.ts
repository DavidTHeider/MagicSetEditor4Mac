// **Was der Markttisch mit einer gezogenen Karte tut** — P12
// (#2380),
// Entwurf `Markttisch09`: aus dem eigenen Fach nach „Ich gebe", aus dem Fach
// des Gegenübers und aus der Wunschliste nach „Ich bekomme".
import { describe, expect, it } from 'vitest'

import {
  ausExemplar,
  ausSichtkarte,
  legen,
  zieleFuer,
  type Fachkarte,
} from './marktlegen'

import type { ExemplarDTO } from '../api'
import type { Handelsposition } from '../apihandelsleitung'
import type { Sichtkarte } from '../apisichtfreigabe'

const exemplar = (teil: Partial<ExemplarDTO>): ExemplarDTO =>
  ({
    id: 7,
    fassung: 1,
    printing: 'p-glut',
    druckID: '12',
    vorlaeufig: false,
    menge: 2,
    zustand: 'nearMint',
    sprache: 'de',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: 'Glutläufer',
    setcode: 'p11',
    nummer: '4',
    lagerort: '',
    wunsch: false,
    bezeichnung: 'Glutläufer · P11 #4',
    herkunft: { art: '', quelle: '', quellProfil: 0, uebernommen: '' },
    ...teil,
  }) as ExemplarDTO

const sicht: Sichtkarte = {
  referenz: 'r1',
  printing: 'p-sturm',
  menge: 2,
  zustand: 'excellent',
  sprache: 'en',
  finish: 'foil',
  bezeichnung: 'Sturmrufer · P11 #3',
}

describe('der Markttisch', () => {
  it('nimmt ein eigenes Stück nach „Ich gebe", Wunsch und fremdes Stück nach „Ich bekomme"', () => {
    const mein = ausExemplar(exemplar({}))
    const wunsch = ausExemplar(exemplar({ id: 8, wunsch: true }))
    const sein = ausSichtkarte(sicht)

    expect(zieleFuer([mein])).toEqual(['gebe'])
    expect(zieleFuer([wunsch])).toEqual(['bekomme'])
    expect(zieleFuer([sein])).toEqual(['bekomme'])

    // **Eine gemischte Wahl hat kein gemeinsames Ziel** — und wird nicht
    // halb gelegt.
    expect(zieleFuer([mein, sein])).toEqual([])
    expect(zieleFuer([])).toEqual([])
  })

  it('legt Zeilen an und zählt dieselbe Karte hoch, bis zu ihrer Menge', () => {
    const mein = ausExemplar(exemplar({}))
    const sein = ausSichtkarte(sicht)

    let positionen: Handelsposition[] = legen([], [mein], 'gebe', 'ich', 'du')
    positionen = legen(positionen, [sein], 'bekomme', 'ich', 'du')

    expect(positionen).toEqual([
      expect.objectContaining({
        von: 'ich',
        printing: 'p-glut',
        druckId: 0,
        menge: 1,
        guete: 'nearMint',
        sprache: 'de',
        finish: 'nonfoil',
      }),
      expect.objectContaining({
        von: 'du',
        printing: 'p-sturm',
        menge: 1,
        guete: 'excellent',
        sprache: 'en',
        finish: 'foil',
      }),
    ])

    // **Noch einmal dieselbe Karte**: eine Zeile, Menge zwei — und nicht
    // mehr, als dort liegt.
    positionen = legen(positionen, [mein], 'gebe', 'ich', 'du')
    positionen = legen(positionen, [mein], 'gebe', 'ich', 'du')
    expect(positionen).toHaveLength(2)
    expect(positionen[0]?.menge).toBe(2)
  })

  it('hält Varianten desselben Drucks in eigenen Zeilen', () => {
    // **Befund 3 des Zweitblicks über Phase 8**: ein englisches Foil
    // desselben Drucks zählte die deutsche Zeile hoch, statt eine eigene zu
    // werden — die Vereinbarung hätte etwas anderes gesagt als die Karten.
    const deutsch = ausExemplar(exemplar({}))
    const englisch = ausExemplar(
      exemplar({ id: 8, sprache: 'en', finish: 'foil' }),
    )

    const positionen = legen(
      legen([], [deutsch], 'gebe', 'ich', 'du'),
      [englisch],
      'gebe',
      'ich',
      'du',
    )

    expect(positionen).toEqual([
      expect.objectContaining({ sprache: 'de', finish: 'nonfoil', menge: 1 }),
      expect.objectContaining({ sprache: 'en', finish: 'foil', menge: 1 }),
    ])
  })

  it('zählt dieselbe Variante aus zwei Fachzeilen zusammen — jede bis zu ihrer Menge', () => {
    // **#2497, Befund 3 des Nachlaufs über Phase 8**: ein Wunsch und ein
    // freigegebenes Angebot derselben Variante, je ein Stück, in einem Zug
    // nach „Ich bekomme" — die Grenze war die der gerade gelegten Fachzeile,
    // und aus zwei Stücken wurde eines.
    const wunsch = ausExemplar(
      exemplar({
        id: 8,
        wunsch: true,
        printing: 'p-sturm',
        menge: 1,
        zustand: 'excellent',
        sprache: 'en',
        finish: 'foil',
      }),
    )
    const sein = ausSichtkarte({ ...sicht, menge: 1 })
    const fach = [wunsch, sein]

    const beide = legen([], fach, 'bekomme', 'ich', 'du', fach)
    expect(beide).toEqual([
      expect.objectContaining({ von: 'du', printing: 'p-sturm', menge: 2 }),
    ])

    // **Dieselbe Fachzeile ein zweites Mal** bleibt an der Summe dessen,
    // was im Fach liegt — und nicht darüber.
    expect(legen(beide, [sein], 'bekomme', 'ich', 'du', fach)).toEqual(beide)
    // **Und allein an ihrer eigenen**, wenn keine zweite da ist.
    const allein = legen([], [sein], 'bekomme', 'ich', 'du', [sein])
    expect(legen(allein, [sein], 'bekomme', 'ich', 'du', [sein])).toEqual(
      allein,
    )

    // **Ein Fach, das die gelegte Karte nicht nennt**, hält sie nicht unter
    // ihrer eigenen Menge.
    const zwei = ausSichtkarte(sicht)
    const ohneFach = legen([], [zwei], 'bekomme', 'ich', 'du', [])
    expect(legen(ohneFach, [zwei], 'bekomme', 'ich', 'du', [])[0]?.menge).toBe(
      2,
    )
  })

  it('verkleinert keine Menge, die schon über der Fachzeile steht', () => {
    const mein = ausExemplar(exemplar({ menge: 1 }))
    const schon: Handelsposition = {
      von: 'ich',
      printing: 'p-glut',
      druckId: 0,
      menge: 3,
      guete: 'nearMint',
      sprache: 'de',
      finish: 'nonfoil',
      abbild: '',
      bewertung: 0,
    }

    expect(legen([schon], [mein], 'gebe', 'ich', 'du')).toEqual([schon])
  })

  it('nimmt einen eigenen Druck ohne Scryfall-Kennung über seine Druckzeile', () => {
    const eigen = ausExemplar(exemplar({ printing: '', druckID: '42' }))

    expect(legen([], [eigen], 'gebe', 'ich', 'du')[0]).toEqual(
      expect.objectContaining({ printing: '', druckId: 42 }),
    )
  })

  it('legt nichts an ein Ziel, das die Karte nicht hat', () => {
    const mein = ausExemplar(exemplar({}))
    const vorher: Handelsposition[] = []

    expect(legen(vorher, [mein], 'bekomme', 'ich', 'du')).toBe(vorher)
  })

  it('nennt jede Karte eindeutig und lesbar', () => {
    const karten: Fachkarte[] = [
      ausExemplar(exemplar({})),
      ausExemplar(exemplar({ id: 8, wunsch: true })),
      ausSichtkarte(sicht),
    ]

    expect(karten.map((karte) => karte.kennung)).toEqual([
      'mein:7',
      'wunsch:8',
      'sein:r1',
    ])
    expect(karten[0]?.name).toBe('Glutläufer · P11 #4')
    // **Ohne Druckzeile ist sie null** und nicht `NaN`.
    expect(ausExemplar(exemplar({ druckID: '' })).druckId).toBe(0)
    expect(karten[2]?.name).toBe('Sturmrufer · P11 #3')
  })
})
