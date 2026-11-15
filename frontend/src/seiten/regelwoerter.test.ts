// **Vom Schlüssel zum Satz** — PLAY-02.AK-01
// (#1710).

import { expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import {
  regelbereichwort,
  regelgradwort,
  regelgruppenwort,
} from './regelwoerter'

const eigene = kataloge.de.seiten.spielen

// **Jeder Bereich, jeder Grad und jede Gruppe hat ein Wort** — und keines
// davon ist der Schlüssel.
//
// **Das ist die Zusage, für die es diese Datei gibt** (§10): ein fehlendes
// Wort fiele sonst nur dem auf, der genau diese Zeile im Fenster sieht.
it('übersetzt jeden Bereich, jeden Grad und jede Gruppe', () => {
  for (const bereich of [
    'zonen',
    'zugstruktur',
    'prioritaet',
    'stapel',
    'combat',
    'zustandsbasiert',
    'typen',
    'marken',
    'anhaengsel',
    'spielsteine',
    'mana',
    'wechselwirkungen',
    'commander',
  ]) {
    expect(regelbereichwort(eigene, bereich)).not.toBe(bereich)
  }

  for (const grad of ['automatisch', 'begleitet', 'manuell', 'unbekannt']) {
    expect(regelgradwort(eigene, grad)).not.toBe(grad)
  }

  for (const gruppe of ['ablauf', 'mechanik']) {
    expect(regelgruppenwort(eigene, gruppe)).not.toBe(gruppe)
  }
})

// **Ein unbekannter Schlüssel bleibt stehen** — hässlich und ehrlich.
it('gibt einen unbekannten Schlüssel unverändert zurück', () => {
  expect(regelbereichwort(eigene, 'zonnen')).toBe('zonnen')
  expect(regelgradwort(eigene, 'fast')).toBe('fast')
  expect(regelgruppenwort(eigene, 'sonstiges')).toBe('sonstiges')
})

// **Und ein Eintrag, der kein Text ist, gilt nicht als Wort.**
//
// **Der Katalog ist verschachtelt**: unter `seiten` stehen Objekte, und ein
// Griff, der eines davon als Satz ausgäbe, schriebe `[object Object]` ins
// Fenster.
it('nimmt nur Zeichenketten als Wort', () => {
  const verschachtelt = { regelbereichZonen: { tief: 'ja' } }

  expect(regelbereichwort(verschachtelt, 'zonen')).toBe('zonen')
})
