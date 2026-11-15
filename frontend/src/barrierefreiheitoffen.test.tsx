// **Barrierefreiheit aufgeklappt** — Z4 des Zweitblicks über Phase 6.13
// (#1906).
//
// `barrierefreiheit.test.tsx` prüft jede Seite **im Ausgangszustand**. Die
// Entschlackung hat vieles hinter Falten gelegt — ein `Aufklappkasten` hängt
// seinen Inhalt erst beim Öffnen ein —, und ein Rollen-, Namens- oder
// Beschriftungsfehler darin stand für axe nie da.
//
// # Was hier geschieht
//
// **Je Seite und Sprache wird jede eingeklappte Falte im Inhalt geöffnet**,
// bis keine mehr zu ist (Falten in Falten eingeschlossen), und dann geprüft.
// **Bleibt nach fünf Runden eine zu, ist der Lauf rot** und nennt sie.
// Gesucht wird nach `aria-expanded="false"` an einem Knopf **ohne**
// `aria-haspopup`: das sind die Falten; Menüs, Popover und Auswahllisten
// öffnen etwas über der Seite und gehören nicht hierher.
//
// # Was er nicht prüft
//
// **Den Fokusweg** in eine Falte und zurück — den halten die Läufe der
// Bausteine (`aufklappkasten.test.tsx`, `schublade.test.tsx`). Und **die
// Schubladen der Setseiten**: die stehen schon im Ausgangszustand offen da,
// `geladenesDokumentDer` gibt das ganze Dokument samt Portal zurück.
import { fireEvent } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { seiten } from './navigation'
import { sprachen } from './sprache/kataloge'
import { verstoesse } from './test/axe'
import { antwortAbwarten } from './test/api'
import { belegeProbeset } from './test/seitenvorrichtung'
import { geladenesDokumentDer, zurRuhe } from './test/geladeneseite'

// falten gibt die eingeklappten Falten im Inhalt.
function falten(dokument: HTMLElement): HTMLElement[] {
  return [
    ...dokument.querySelectorAll<HTMLElement>(
      'main button[aria-expanded="false"]:not([aria-haspopup]), [role=dialog] button[aria-expanded="false"]:not([aria-haspopup])',
    ),
  ]
}

// benennen nennt eine Falte, wie ein Mensch sie liest.
function benennen(knopf: HTMLElement): string {
  return knopf.getAttribute('aria-label') ?? knopf.textContent
}

describe('Barrierefreiheit aufgeklappt', () => {
  for (const sprache of sprachen) {
    it.each(seiten)(
      `hält $schluessel aufgeklappt in ${sprache} ein`,
      async (seite) => {
        belegeProbeset()

        const dokument = await geladenesDokumentDer(seite.schluessel, sprache)

        // **Höchstens fünf Runden**: eine Falte, die beim Öffnen eine neue
        // eingeklappte zeigt, braucht eine zweite; eine, die sich nie öffnet,
        // hielte den Lauf sonst fest.
        for (let runde = 0; runde < 5; runde++) {
          const zu = falten(dokument)
          if (zu.length === 0) {
            break
          }

          for (const knopf of zu) {
            fireEvent.click(knopf)
          }

          await zurRuhe(dokument, seite.schluessel)
        }

        // **Was nach fünf Runden noch zu ist, ist ein Fehler und kein Rest**
        // — N2 des Nachlaufs über 6.13 (#1906): ohne diese Zeile lief axe
        // auch dann, wenn eine Falte jenseits der fünften Ebene nie aufging,
        // und ihr Inhalt blieb ungeprüft bei grünem Lauf.
        expect(falten(dokument).map(benennen)).toEqual([])

        expect(await verstoesse(dokument)).toEqual([])
      },
    )
  }

  // **Die Gegenprobe, dass überhaupt etwas aufgeht** — ein Lauf, der keine
  // Falte fände, wäre grün und prüfte nichts (§10).
  it('öffnet auf den Setseiten Falten', async () => {
    belegeProbeset()

    const dokument = await geladenesDokumentDer('setangaben', 'de')

    const vorher = falten(dokument).length
    expect(vorher).toBeGreaterThan(0)

    for (const knopf of falten(dokument)) {
      fireEvent.click(knopf)
    }
    await antwortAbwarten()

    expect(falten(dokument).length).toBeLessThan(vorher)
  })
})
