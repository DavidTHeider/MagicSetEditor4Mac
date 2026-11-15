// **Eine ganze Zahl wird gelesen, nicht erraten** — F03
// (#2245).
//
// `Number.parseInt` liest, so weit es kommt, und schweigt über den Rest:
// „12a" ist 12, „1.5" ist 1, „8080x" ist 8080. Vier Stellen im Fenster gaben
// so eine Zahl weiter, die niemand getippt hat.
import { describe, expect, it } from 'vitest'

import { ganzzahlAus } from './ganzzahl'

describe('ganzzahlAus', () => {
  it.each([
    ['0', 0],
    ['7', 7],
    ['  12 ', 12],
    ['0042', 42],
    ['9007199254740991', 9007199254740991],
  ])('liest %j als %d', (roh, zahl) => {
    expect(ganzzahlAus(roh)).toBe(zahl)
  })

  it.each([
    ['', 'leer'],
    ['   ', 'nur Leerraum'],
    ['12a', 'mit einem Buchstaben dahinter'],
    ['1.5', 'mit Nachkommastellen'],
    ['1,5', 'mit Komma'],
    ['-3', 'negativ'],
    ['+3', 'mit Vorzeichen'],
    ['1e3', 'in Exponentschreibweise'],
    ['0x10', 'hexadezimal'],
    ['viele', 'ein Wort'],
    ['9007199254740992', 'jenseits der sicheren Ganzzahlen'],
    ['١٢', 'in arabisch-indischen Ziffern'],
  ])('weist %j ab (%s)', (roh) => {
    expect(ganzzahlAus(roh)).toBeUndefined()
  })
})
