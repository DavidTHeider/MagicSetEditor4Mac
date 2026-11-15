// **Die Vorrichtungen der Paketseite** — geteilt von `paketseite.test.tsx`,
// `pakete.test.tsx`, `paketlauf.test.tsx` und `bezugsquellen.test.tsx`, seit
// der Bezug in einer Schublade hinter *Paket beziehen …* steht (UXR-17,
// #1925).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'

import { type Api } from '../api'
import { Pakete } from '../seiten/Pakete'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from './api'

export const texte = kataloge.de

// belegeSeite belegt, was die Seite beim Aufschlagen liest.
//
// **Weil die Seite es zeigt.** Ohne Beleg meldete die Fassade „die Methode
// ist nicht belegt" — mit `role="alert"`, und ein Lauf, der genau eine
// Meldung sucht, fände zwei. Genau dafür ist die Vorrichtung streng: eine,
// die stillschweigend nichts liefert, machte aus dem vergessenen Beleg eine
// leere Seite.
export function belegeSeite(teile: Partial<Api>) {
  return belegeApi({
    PaketArten: () =>
      Promise.resolve([{ art: 'mse-game', brauchtSpiel: false }]),
    Pakete: () => Promise.resolve({ pakete: [], befunde: [] }),
    MSE2Vorschlaege: () => Promise.resolve({ orte: [] }),
    Abhaengigkeiten: () => Promise.resolve({ befunde: [] }),
    Bezugsquellen: () => Promise.resolve({ quellen: [] }),
    PaketBezuege: () => Promise.resolve({ bezuege: [] }),
    ...teile,
  })
}

export function zeige() {
  render(
    <Sprachraum sprache="de">
      <Pakete wechseln={vi.fn()} />
    </Sprachraum>,
  )
}

// oeffneBezug drückt den führenden Knopf im Kopf und gibt die Schublade.
export async function oeffneBezug(): Promise<HTMLElement> {
  await userEvent.click(
    await screen.findByRole('button', {
      name: texte.seiten.pakete.plan.beziehen,
    }),
  )

  return screen.findByRole('dialog', {
    name: texte.seiten.pakete.plan.beziehenTitel,
  })
}

// oeffneKasten klappt einen der Kästen auf — auf der Seite oder in der
// Schublade.
export async function oeffneKasten(titel: string): Promise<void> {
  await userEvent.click(screen.getByRole('button', { name: titel }))
}
