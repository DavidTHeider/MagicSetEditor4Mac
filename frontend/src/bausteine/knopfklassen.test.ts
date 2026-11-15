// **Haupt- und Nebenknopf einer Rückfrage sind gleich groß** — #2412.
//
// Der Hauptknopf von „Diese Sicherung zurückspielen?" trug nur Farbe und
// Fokusring: sein Text stieß an den Rand, stand größer als „Abbrechen" daneben,
// und der Knopf war niedriger als sein Nachbar — gefunden in PERF01 am
// nativen Fenster. Derselbe Knopf in `Bestandschiffre` trug die Größe.
//
// **Was er nicht hält**: ob ein Knopf, der nicht von hier liest, dieselbe
// Größe trägt — die Bauteilbibliothek (`Knopf`) hält ihre eigene.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

import { describe, expect, it } from 'vitest'

import { hauptknopf, nebenknopf } from './knopfklassen'
import { dateienUnter } from '../test/dateien'

const groesse = ['px-3', 'py-1.5', 'text-sm', 'font-medium']

describe('Die Klassen der Rückfrageknöpfe', () => {
  it('geben Haupt- und Nebenknopf dieselbe Größe', () => {
    for (const klasse of groesse) {
      expect(hauptknopf.split(' ')).toContain(klasse)
      expect(nebenknopf.split(' ')).toContain(klasse)
    }
  })

  // **Und dieselbe Rückmeldung wie `Knopf`** — AP-4 (#2361), B-11.
  it('antworten auf Zeiger und Druck wie jeder Knopf', () => {
    for (const klassen of [hauptknopf, nebenknopf]) {
      expect(klassen.split(' ')).toEqual(
        expect.arrayContaining([
          'transition-[color,background-color,border-color,opacity,translate]',
          'duration-(--dauer-kurz)',
          'enabled:not-aria-disabled:active:translate-y-px',
        ]),
      )
    }
    expect(hauptknopf.split(' ')).toContain(
      'enabled:not-aria-disabled:hover:bg-marke/90',
    )
    expect(nebenknopf.split(' ')).toContain(
      'enabled:not-aria-disabled:hover:bg-betont',
    )
  })

  it('stehen an einer Stelle — keine Datei führt einen eigenen Hauptknopf', () => {
    const quellen = join(import.meta.dirname, '..')
    const eigene = dateienUnter(quellen, '**/*.tsx', {
      ohne: ['**/*.test.tsx', 'test/**'],
    }).filter((datei) =>
      /^const hauptknopf =/m.test(readFileSync(join(quellen, datei), 'utf8')),
    )

    expect(eigene).toEqual([])
  })
})
