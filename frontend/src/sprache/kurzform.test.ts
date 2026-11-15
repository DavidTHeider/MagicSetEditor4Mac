// **Ein kurzer Satz sichtbar, der Rest hinter dem Fragezeichen** — UX-15.AK-02
// und AK-03
// (#1771).
import { describe, expect, it } from 'vitest'

import { kurzform, langform, ohneAuszeichnung } from './kurzform'

describe('kurzform', () => {
  it('nimmt den ersten Satz samt seinem Punkt', () => {
    expect(kurzform('Das gilt sofort. Und danach noch etwas.')).toBe(
      'Das gilt sofort.',
    )
  })

  it('nimmt auch ein Frage- oder Ausrufezeichen', () => {
    expect(kurzform('Wirklich? Ja, wirklich.')).toBe('Wirklich?')
  })

  // **Ein Text ohne zweiten Satz ist ganz kurz** — dann gibt es nichts zu
  // verstecken.
  it('gibt einen einzelnen Satz unverändert', () => {
    expect(kurzform('Nur dieser eine Satz.')).toBe('Nur dieser eine Satz.')
  })

  // **Ein Punkt in einer Zahl oder Abkürzung trennt nicht** — getrennt wird
  // an einem Punkt, dem ein Leerzeichen folgt, und nicht an jedem.
  it('trennt nicht an einem Punkt ohne Abstand', () => {
    expect(kurzform('Ab Fassung 1.2.3 gilt das.')).toBe(
      'Ab Fassung 1.2.3 gilt das.',
    )
  })

  // **Und Auszeichnung bleibt stehen**: die Kataloge schreiben `**…**`, und
  // ein abgeschnittenes Sternchenpaar zeigte die Sternchen.
  it('lässt Auszeichnung ganz', () => {
    expect(kurzform('**Wichtig** ist das hier. Der Rest nicht.')).toBe(
      '**Wichtig** ist das hier.',
    )
  })
})

describe('langform', () => {
  it('gibt, was nach dem ersten Satz kommt', () => {
    expect(langform('Das gilt sofort. Und danach noch etwas.')).toBe(
      'Und danach noch etwas.',
    )
  })

  // **Nichts dahinter heißt: kein Fragezeichen.** Ein Hilfesymbol, das eine
  // leere Erklärung öffnet, ist eine Zusage ohne Einlösung (§10).
  it('gibt nichts, wo nichts mehr kommt', () => {
    expect(langform('Nur dieser eine Satz.')).toBeUndefined()
  })

  it('gibt nichts für einen leeren Text', () => {
    expect(langform('')).toBeUndefined()
  })

  // **Und nichts, wo hinter dem Punkt nur Leerraum steht.** Ein Hinweis mit
  // einem Leerzeichen am Ende bekäme sonst ein Fragezeichen, das eine leere
  // Erklärung öffnet.
  it('gibt nichts für einen Rest aus Leerraum', () => {
    expect(langform('Nur dieser eine Satz. ')).toBeUndefined()
  })
})

describe('ohneAuszeichnung', () => {
  it('nimmt Fettung und wörtliche Auszeichnung weg', () => {
    expect(ohneAuszeichnung('**fett** und `wörtlich` und roh')).toBe(
      'fett und wörtlich und roh',
    )
  })

  it('lässt einen Text ohne Auszeichnung, wie er ist', () => {
    expect(ohneAuszeichnung('ganz gewöhnlich')).toBe('ganz gewöhnlich')
  })
})

// **Und beides zusammen ist der ganze Text** — ohne diesen Lauf könnte die
// Teilung etwas verschlucken, und niemand sähe es (§10, „keine Information
// geht verloren").
describe('kurzform und langform zusammen', () => {
  const faelle = [
    'Das gilt sofort. Und danach noch etwas. Und ein drittes.',
    'Nur dieser eine Satz.',
    '**Wichtig** ist das hier. Der Rest nicht.',
    'Ab Fassung 1.2.3 gilt das.',
  ]

  for (const text of faelle) {
    it(`verliert nichts an „${text.slice(0, 20)}…"`, () => {
      const rest = langform(text)

      expect([kurzform(text), rest].filter(Boolean).join(' ')).toBe(text)
    })
  }
})
