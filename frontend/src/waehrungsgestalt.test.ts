// **Dieselbe Gestalt wie in Go** — UX-06.AK-02 und AK-04
// (#1678).
//
// Die Fälle sind dieselben wie in `internal/domain/waehrungsgestalt_test.go`:
// wer die Regel an einer Seite ändert, sieht die andere rot.
import { afterEach, describe, expect, it, vi } from 'vitest'

import { waehrungsbefund, waehrungsvorschlaege } from './waehrungsgestalt'

afterEach(() => {
  vi.restoreAllMocks()
})

describe('waehrungsbefund', () => {
  it.each([
    ['EUR', true],
    ['EURO', false],
    ['EU', false],
    ['€', false],
    ['978', false],
    ['ÄUR', false],
    ['E R', false],
  ])('%j hat die Gestalt eines Codes: %s', (eingabe, gestalt) => {
    expect(waehrungsbefund(eingabe, true) === undefined).toBe(gestalt)
  })

  it('nennt leer keinen Befund — ob es fehlen darf, weiß der Aufrufer', () => {
    expect(waehrungsbefund('   ', false)).toBeUndefined()
  })

  it('hebt klein Geschriebenes an, wo die Stelle dahinter es auch tut', () => {
    expect(waehrungsbefund(' eur ', true)).toBeUndefined()
  })

  it('sagt, wie der Code richtig heißt, wo klein nicht angenommen wird', () => {
    expect(waehrungsbefund('eur', false)).toEqual({
      art: 'gross',
      gross: 'EUR',
    })
  })

  // **Erst die Gestalt, dann die Schreibweise**: „schreib groß" wäre für
  // `euro` ein falscher Rat.
  it('rät bei falscher Gestalt nicht zur Großschreibung', () => {
    expect(waehrungsbefund('euro', false)).toEqual({ art: 'gestalt' })
  })
})

describe('waehrungsvorschlaege', () => {
  it('nennt die Währungen der Plattform mit ihrem Namen', () => {
    const vorschlaege = waehrungsvorschlaege('de')

    expect(vorschlaege.find((einer) => einer.wert === 'EUR')).toEqual({
      wert: 'EUR',
      bezeichnung: 'Euro',
    })
    expect(vorschlaege.length).toBeGreaterThan(100)
  })

  it('lässt einen Code ohne Namen für sich stehen', () => {
    vi.spyOn(Intl, 'supportedValuesOf').mockReturnValue(['XTS'])
    vi.spyOn(Intl.DisplayNames.prototype, 'of').mockReturnValue(undefined)

    expect(waehrungsvorschlaege('de')).toEqual([
      { wert: 'XTS', bezeichnung: 'XTS' },
    ])
  })

  // **Ohne `Intl` bleibt das Feld ein Freitext mit Prüfung** — nur ohne
  // Vorschläge, und ohne dass der Baum zerbricht.
  it('gibt ohne Plattformliste keine Vorschläge', () => {
    vi.spyOn(Intl, 'supportedValuesOf').mockImplementation(() => {
      throw new RangeError('unbekannt')
    })

    expect(waehrungsvorschlaege('de')).toEqual([])
  })
})
