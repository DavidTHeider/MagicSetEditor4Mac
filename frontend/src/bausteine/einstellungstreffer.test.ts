// **Wonach die Einstellungssuche vergleicht** — UX-04.AK-03
// (#1676).
//
// **Herausgelöst am 2026-09-21**, als der Trefferkasten wegfiel: die Regel
// stand in ihm, und seitdem brauchen sie **zwei** — die Gruppenliste für
// ihre Zahlen und die Fläche für ihre Marken. Eine Regel an zwei Stellen
// wäre die Falle aus §10.
import { describe, expect, it } from 'vitest'

import { passtZurFrage } from './einstellungstreffer'
import { kataloge } from '../sprache/kataloge'

import type { EinstellungDTO } from '../api'

const texte = kataloge.de

function eine(teile: Partial<EinstellungDTO>): EinstellungDTO {
  return {
    schluessel: 'sprache',
    wert: 'de',
    wirksam: 'de',
    werte: ['de', 'en'],
    bereich: 'anwendung',
    ...teile,
  }
}

describe('passtZurFrage', () => {
  it('findet über den Schlüssel', () => {
    expect(passtZurFrage(eine({}), 'sprach', texte, 'de')).toBe(true)
  })

  it('findet über den Namen aus dem Katalog', () => {
    expect(
      passtZurFrage(
        eine({ schluessel: 'schriftgroesse' }),
        'schriftgröße',
        texte,
        'de',
      ),
    ).toBe(true)
  })

  // **Und über den Eigennamen eines Dienstes** — er steht nicht im
  // Katalog der Einstellungen, sondern in `dienstnamen.ts`.
  it('findet über den Eigennamen eines Dienstes', () => {
    expect(
      passtZurFrage(
        eine({ schluessel: 'dienst.scryfall' }),
        'scryfall',
        texte,
        'de',
      ),
    ).toBe(true)
  })

  it('findet über den Hinweis', () => {
    expect(
      passtZurFrage(
        eine({ schluessel: 'netz.proxy.http' }),
        'HTTP_PROXY',
        texte,
        'de',
      ),
    ).toBe(true)
  })

  it('findet über einen erlaubten Wert', () => {
    expect(
      passtZurFrage(
        eine({ schluessel: 'erscheinungsbild', werte: ['hell', 'dunkel'] }),
        'dunkel',
        texte,
        'de',
      ),
    ).toBe(true)
  })

  // **Groß und klein ist dasselbe** — wer „Proxy" tippt, meint `proxy`.
  it('unterscheidet nicht nach Schreibweise', () => {
    expect(passtZurFrage(eine({}), 'SPRACHE', texte, 'de')).toBe(true)
  })

  // **Auch ein Schlüssel ohne Namen und ohne Hinweis** — er kann aus einer
  // neueren Go-Seite kommen, und dann bleibt der Schlüssel selbst.
  it('findet einen unbenannten Schlüssel über sich selbst', () => {
    const fremd = eine({ schluessel: 'gibtesnochnicht', werte: [] })

    expect(passtZurFrage(fremd, 'gibtesnochnicht', texte, 'de')).toBe(true)
    expect(passtZurFrage(fremd, 'sprache', texte, 'de')).toBe(false)
  })

  it('lehnt ab, was nirgends vorkommt', () => {
    expect(passtZurFrage(eine({}), 'nilpferd', texte, 'de')).toBe(false)
  })

  // **Eine leere Frage passt auf nichts**, und das ist der Unterschied
  // zwischen „ich suche nichts" und „alles passt".
  it('passt bei leerer Frage auf nichts', () => {
    expect(passtZurFrage(eine({}), '', texte, 'de')).toBe(false)
  })
})
