// **Die Vorrichtung der Handlungsläufe** —
// #1623
// und UX-08 (#1680).
//
// **Eine eigene Datei**, weil sechs Läufe sie brauchen — die Leiste, ihr
// Kampf, das Aufklappen, die Befehle, der Tisch und das Gastspiel —, und
// seit die Felder erst auf Zuruf dastehen, klappt jeder davon auf, bevor er
// etwas einträgt. **Unter `src/test/`**, weil sie kein Baustein ist:
// `bausteineohneaufrufer.test.ts` zählte sie sonst mit.
import { render, screen, within } from '@testing-library/react'
import { vi } from 'vitest'
import nutzerereignis from '@testing-library/user-event'

import { Handlungsleiste } from '../bausteine/Handlungsleiste'
import { Sprachraum } from '../sprache/kontext'
import { einObjekt, einePartie } from '../seiten/partieprobe'
import { kataloge } from '../sprache/kataloge'

import type { Partie, Spielobjekt } from '../apipartie'

export const eigene = kataloge.de.seiten.handlungen

export type Hand = ReturnType<typeof nutzerereignis.setup>

// zeige zeichnet die Leiste und gibt die Hand samt dem Spion.
export function zeige({
  erlaubt,
  objekt,
  partie = einePartie(),
}: {
  erlaubt: string[]
  objekt?: Spielobjekt
  partie?: Partie
}) {
  const befehlen = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Handlungsleiste
        partie={partie}
        erlaubt={erlaubt}
        objekt={objekt}
        befehlen={befehlen}
        angehalten={false}
      />
    </Sprachraum>,
  )

  return { hand: nutzerereignis.setup(), befehlen }
}

// eineKarte gibt eine offene Karte auf dem Tisch.
export function eineKarte(teile: Partial<Spielobjekt> = {}): Spielobjekt {
  return einObjekt({
    kennung: 'sitz1/k1/0',
    zone: 'tisch',
    name: 'Wald',
    geheim: false,
    ...teile,
  })
}

// name ist die Aufschrift eines Befehls.
export function name(art: string): string {
  return (eigene.handlungsname as Record<string, string>)[art] ?? art
}

// form gibt die aufgeklappten Felder eines Befehls — oder nichts.
export function form(art: string): HTMLElement | null {
  return screen.queryByRole('group', { name: name(art) })
}

// oeffne klappt die Felder eines Befehls auf und gibt sie.
//
// **Geklickt wird der Knopf mit seinem Namen**, wie ihn jemand am Tisch
// sieht — und nicht eine Kennung, die niemand sieht.
export async function oeffne(hand: Hand, art: string): Promise<HTMLElement> {
  await hand.click(screen.getByRole('button', { name: name(art) }))

  return screen.getByRole('group', { name: name(art) })
}

// stelle stellt einen Befehl mit Feldern — und klappt ihn vorher auf, wo er
// noch zu ist. **Die Knöpfe heißen alle „Stellen"**; gesucht wird deshalb in
// den Feldern, die zum Befehl gehören.
export async function stelle(hand: Hand, art: string) {
  const offen = form(art) ?? (await oeffne(hand, art))

  await hand.click(
    within(offen).getByRole('button', { name: eigene.handlungStellen }),
  )
}
