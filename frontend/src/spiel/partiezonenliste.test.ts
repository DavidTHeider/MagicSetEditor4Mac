// **Welche Zonen der Tisch zeigt** — PLAY-03.AK-01
// (#1711).
import { describe, expect, it } from 'vitest'

import { einObjekt, einePartie } from '../seiten/partieprobe'
import { zonenDerPartie } from './partiezonenliste'

const sandkasten = einePartie()

const commander = einePartie({
  regelwerk: {
    kennung: 'commander',
    fassung: 1,
    strenge: 'begleitet',
    abgewaehlt: [],
  },
})

describe('zonenDerPartie', () => {
  // **Sechs Zonen hat jede Partie** — CR 400.1 zählt sie auf, und keine
  // Variante kommt ohne sie aus.
  it('zeigt die sechs, die jede Partie hat', () => {
    expect(zonenDerPartie(sandkasten)).toEqual([
      'hand',
      'tisch',
      'stapel',
      'bibliothek',
      'friedhof',
      'exil',
    ])
  })

  // **Die Kommandozone gehört zur Variante** — CR 903.3: dort steht der
  // Commander. Ein leeres Feld an jedem Sitz einer gewöhnlichen Partie ist
  // Platz, den niemand braucht.
  it('nimmt die Kommandozone zur Commanderpartie dazu', () => {
    expect(zonenDerPartie(commander)).toContain('kommando')
  })

  // **Und sie kommt auch sonst, sobald etwas darin liegt** — ein Emblem
  // entsteht dort (CR 114.2), auch im Sandkasten. Eine Zone auszublenden,
  // in der etwas liegt, hieße eine Karte zu verstecken (§8).
  it('nimmt sie dazu, wo etwas darin liegt', () => {
    expect(
      zonenDerPartie(
        einePartie({
          objekte: [einObjekt({ kennung: 'e1', zone: 'kommando' })],
        }),
      ),
    ).toContain('kommando')
  })

  // **Die Reihenfolge ist die des Blicks** — von der Hand zum Tisch und dann
  // in die Ablagen; das Kommando zuletzt, weil es das seltenste ist.
  it('stellt die Kommandozone ans Ende', () => {
    expect(zonenDerPartie(commander).at(-1)).toBe('kommando')
  })
})
