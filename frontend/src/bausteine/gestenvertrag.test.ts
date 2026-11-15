import { describe, expect, it } from 'vitest'

import {
  absichtAusZug,
  gestentaste,
  greifen,
  type Gestenbezug,
  type Gestenzustand,
} from './gestenvertrag'

const bezug: Gestenbezug = { profil: 'ada', epoche: 3 }
const exil = { kennung: 'exil', name: 'Exil', sitz: 'sitz-1' }
const ziele = [
  { kennung: 'friedhof', name: 'Friedhof', sitz: 'sitz-1' },
  exil,
  { kennung: 'hand', name: 'Hand', sitz: 'sitz-1' },
]

// schritt ist ein Schritt, der etwas bewirkt — sonst ist der Fall falsch
// gebaut und soll laut scheitern.
function schritt(zustand: Gestenzustand, taste: string) {
  const ergebnis = gestentaste(zustand, taste)
  if (ergebnis === undefined) {
    throw new Error(`»${taste}« bewirkt nichts`)
  }

  return ergebnis
}

const gegriffen = (): Gestenzustand =>
  greifen({ objekt: 'karte-7', quelle: 'schlachtfeld', ziele }, bezug).zustand

describe('der Gestenvertrag', () => {
  it('greift ein Objekt und sagt das erste Ziel an', () => {
    const { zustand, ansage } = greifen(
      { objekt: 'karte-7', quelle: 'schlachtfeld', ziele },
      bezug,
    )

    expect(zustand).toEqual({
      art: 'gegriffen',
      objekt: 'karte-7',
      quelle: 'schlachtfeld',
      ziele,
      stelle: 0,
      bezug,
    })
    expect(ansage).toEqual({
      schluessel: 'gegriffen',
      ziel: 'Friedhof',
      stelle: 1,
      von: 3,
    })
  })

  it('greift nicht, wenn es kein Ziel gibt', () => {
    expect(
      greifen({ objekt: 'karte-7', quelle: 'hand', ziele: [] }, bezug),
    ).toEqual({
      zustand: { art: 'ruhend' },
      ansage: { schluessel: 'keinZiel' },
    })
  })

  it.each([
    ['ArrowRight', 1],
    ['ArrowDown', 1],
    ['ArrowLeft', 2],
    ['ArrowUp', 2],
  ])('wandert mit %s zyklisch durch die Ziele', (taste, stelle) => {
    const { zustand, ansage } = schritt(gegriffen(), taste)

    expect(zustand).toMatchObject({ art: 'gegriffen', stelle })
    expect(ansage).toEqual({
      schluessel: 'ziel',
      ziel: ziele[stelle]?.name,
      stelle: stelle + 1,
      von: 3,
    })
  })

  it.each(['Enter', ' '])(
    'legt mit »%s« ab und gibt die Absicht mit Ziel-Sitz, Profil und Epoche',
    (taste) => {
      const nach = schritt(schritt(gegriffen(), 'ArrowRight').zustand, taste)

      expect(nach.zustand).toEqual({ art: 'ruhend' })
      expect(nach.absicht).toEqual({
        objekt: 'karte-7',
        quelle: 'schlachtfeld',
        ziel: 'exil',
        zielSitz: 'sitz-1',
        profil: 'ada',
        epoche: 3,
        weg: 'tastatur',
      })
      expect(nach.ansage).toEqual({ schluessel: 'abgelegt', ziel: 'Exil' })
    },
  )

  it('bricht mit Escape ab, ohne Absicht', () => {
    expect(gestentaste(gegriffen(), 'Escape')).toEqual({
      zustand: { art: 'ruhend' },
      ansage: { schluessel: 'abgebrochen' },
    })
  })

  // **Ein Griff ohne Ziele** entsteht über `greifen` nicht; wer ihn von Hand
  // baut, bekommt ihn unverändert zurück statt einer Absicht ins Leere.
  it('lässt einen Griff ohne Ziele, was er ist', () => {
    const leer: Gestenzustand = {
      art: 'gegriffen',
      objekt: 'karte-7',
      quelle: 'hand',
      ziele: [],
      stelle: 0,
      bezug,
    }

    expect(gestentaste(leer, 'ArrowRight')).toBeUndefined()
    expect(gestentaste(leer, 'Enter')).toBeUndefined()
  })

  it('lässt andere Tasten und den ruhenden Zustand unberührt', () => {
    expect(gestentaste(gegriffen(), 'a')).toBeUndefined()
    expect(gestentaste({ art: 'ruhend' }, 'Enter')).toBeUndefined()
  })

  it('baut aus einem Zug mit Maus oder Finger dieselbe Absicht', () => {
    expect(
      absichtAusZug(
        { objekt: 'karte-7', quelle: 'schlachtfeld', ziel: exil },
        bezug,
        'maus',
      ),
    ).toEqual({
      objekt: 'karte-7',
      quelle: 'schlachtfeld',
      ziel: 'exil',
      zielSitz: 'sitz-1',
      profil: 'ada',
      epoche: 3,
      weg: 'maus',
    })
  })

  it('lässt den Ziel-Sitz weg, wo das Ziel keinen hat', () => {
    expect(
      absichtAusZug(
        {
          objekt: 'exemplar-4',
          quelle: 'sammlung',
          ziel: { kennung: 'fach:eigen', name: 'Von dir' },
        },
        bezug,
        'beruehrung',
      ),
    ).not.toHaveProperty('zielSitz')
  })
})
