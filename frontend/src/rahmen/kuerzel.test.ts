// **Die Kürzelübersicht ist keine Kopie** —
// #245.
//
// `globaleKuerzel` ist eine Liste neben einer Funktion, und Listen neben
// Funktionen laufen auseinander. Dieser Test schickt jede Zeile durch
// [deuten] und prüft, dass dabei die genannte Art herauskommt.
//
// **Was er nicht prüft:** ob die Liste vollständig ist. Ein Kürzel, das
// `deuten` kennt und das hier fehlt, fällt nicht auf — es stünde dann nur
// nicht in der Tafel. Die Gegenrichtung ist die teurere: ein Kürzel in der
// Tafel, das es nicht gibt, ist eine Anleitung ins Leere.
import { describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { deuten, globaleKuerzel, kuerzelschluessel } from './tastatur'

describe('globaleKuerzel', () => {
  it.each(globaleKuerzel)('$taste führt zu $art', (angabe) => {
    expect(
      deuten({
        code: angabe.code ?? '',
        alt: angabe.alt ?? false,
        taste: angabe.taste,
        meta: true,
        umschalt: angabe.umschalt,
        imEingabefeld: false,
      }),
    ).toEqual({ art: angabe.art })
  })

  // **Kein Kürzel zweimal.** Zwei Zeilen mit derselben Taste und denselben
  // Umschaltern wären zwei Zeilen in der Tafel, von denen eine nie gilt.
  //
  // **⌥ gehört dazu**, seit ⌘[ den Seitengang geht und ⌥⌘[ die Regionen
  // (UX-03.AK-02, #1675): dieselbe Taste, zwei Kürzel.
  it('nennt kein Kürzel zweimal', () => {
    const gesehen = globaleKuerzel.map(kuerzelschluessel)

    expect(new Set(gesehen).size).toBe(globaleKuerzel.length)
  })
})

// **Jedes Kürzel der Tafel trägt eine Beschriftung — in beiden Sprachen.**
//
// # Warum das ein eigener Wächter ist
//
// Die Tafel schlägt die Beschriftung über die **Art** nach. Fehlt der Eintrag
// im Katalog, steht in der Zeile nichts — kein Fehler, keine Meldung, eine
// leere Zelle. Ein neues Kürzel ohne Text sähe im Lauf über die
// Barrierefreiheit genauso aus wie eines mit.
//
// Angestrichen beim Bau von #891: drei Kürzel kamen dazu, und ihre Texte
// hätten in einem der beiden Kataloge fehlen können.
describe('die Beschriftungen der Kürzeltafel', () => {
  it.each(Object.entries(kataloge))(
    'sind in %s vollständig',
    (_sprache, katalog) => {
      const beschriftungen: Record<string, string> =
        katalog.seiten.einstellungen.kuerzel

      for (const angabe of globaleKuerzel) {
        expect(beschriftungen[angabe.art]).toBeTruthy()
      }
    },
  )
})
