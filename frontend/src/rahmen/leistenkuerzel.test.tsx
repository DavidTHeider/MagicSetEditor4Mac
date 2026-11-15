// **Was die Seitenleiste über Tasten sagt.**
//
// Aus `Rahmen.test.tsx` herausgelöst, als die Datei über 400 Zeilen wuchs
// (§2). Der Schnitt folgt der Frage: dort steht, **was** die Leiste zeigt und
// wohin sie führt; hier, **welche Taste sie dabei ausweist**.
import { render } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { leisteneintraege } from '../leistenordnung'
import { leisteneintrag } from '../test/leistenweg'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Rahmen } from './Rahmen'

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Rahmen aktuell="start" wechseln={vi.fn()} suchen={vi.fn()}>
        <p>{kataloge.de.seiten.start.was}</p>
      </Rahmen>
    </Sprachraum>,
  )
}

// **Eine Seite ohne Kürzel zeigt kein ⌘ ohne Taste** —
// #1164.
//
// # Wie das gefunden wurde
//
// Bei der Sichtprüfung im laufenden Fenster, nachdem „Vorlagen" als erste
// **Anwendungs**seite ohne Kürzel dazukam: in der Leiste stand ein einzelnes
// ⌘ hinter ihrem Namen. Bis dahin hatte jede Seite in dieser Liste eine Taste,
// und der Zweig war nie betreten worden.
//
// **Kein Test hat das gesehen** und keiner konnte es: die Zeichenfolge stand
// da, nur ohne Bedeutung. Das ist die Sorte Befund, für die §3 die
// Sichtprüfung verlangt.
describe('Eine Seite ohne Tastenkürzel', () => {
  // **Der Satz lautet seit dem 2026-09-03 genauer, und der alte ist weg.**
  //
  // Hier stand „eine Seite ohne `kuerzel` zeigt kein ⌘", und das war die
  // Zusage, solange die Leiste nur die **erste** Reihe zeigte. Seit sie auch
  // ⌥⌘ zeigt, trägt eine Seite ohne `kuerzel` sehr wohl ein ⌘.
  //
  // **Ein zweiter Anlauf prüfte „ohne beide Reihen" — und lief über eine
  // leere Liste**: seit #1291 hat jede Seite eine Taste. Ein Test, dessen Fall
  // nicht mehr vorkommt, ist grün und sagt nichts (§10); er steht deshalb
  // nicht hier.
  //
  // Was bleibt, ist der Befund von #1164 selbst: **ein ⌘ ohne Taste
  // dahinter** — unabhängig davon, aus welcher Reihe sie kommt.
  it('lässt kein ⌘ ohne Zeichen dahinter stehen', async () => {
    zeige()

    for (const eintrag of leisteneintraege) {
      const knopf = await leisteneintrag(eintrag.ziel)

      expect(knopf.textContent).not.toMatch(/⌘\s*$/)
    }
  })
})

// **Jede Seite der Leiste zeigt ihre Taste** — Werkstück C von Phase 4.10
// (#1325).
//
// # Warum das ein Befund war
//
// *Vorlagen*, *Sammlung*, *Decks* und *Turniere* standen ohne jede Taste in
// der Leiste — und hatten seit #1291 eine, nur in der **zweiten** Reihe, die
// die Leiste nicht anzeigte.
//
// `rahmen/tastatur.ts` schreibt den Befund selbst auf: *„⌘1…⌘8, ⌘K, ⌘S, ⌘Z,
// ⌘, stehen — **sichtbar** sind sie nirgends."* Für die Tafel in den
// Einstellungen ist das behoben worden, für die Leiste nicht.
//
// **Behoben ist es zweifach**: die Leiste zeigt jetzt auch ⌥⌘ — und die vier
// brauchen es nicht mehr, weil die erste Reihe seither ganz ihr gehört.
describe('Die Seitenleiste', () => {
  // **Seit UX-02 trägt jeder Eintrag eine**
  // (#1674),
  // und das ist keine Aufweichung, sondern die Folge davon, dass die Leiste
  // **Bereiche** führt: die Ziffer gehört dem Bereich, und *Karten*,
  // *Vorlagen*, *Turniere* und *Spielen* stehen nicht mehr einzeln darin.
  //
  // **Was dieser Lauf nicht hält**: dass es zu jedem Eintrag überhaupt eine
  // Taste **gibt**. Das ist eine Aussage über die Ordnung und steht in
  // `bereichsordnung.test.ts`; hier steht die andere Hälfte — dass sie auch
  // **dasteht**.
  it('zeigt an jedem Eintrag seine Taste', async () => {
    zeige()

    for (const eintrag of leisteneintraege) {
      const knopf = await leisteneintrag(eintrag.ziel)

      expect(knopf.textContent, `${eintrag.ziel} zeigt keine Taste`).toContain(
        `⌘${eintrag.kuerzel}`,
      )
    }
  })
})
