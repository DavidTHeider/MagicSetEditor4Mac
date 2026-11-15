// **Was eine Partie als Ganzes angeht** — Anhalten, Fortsetzen, Beenden.
//
// # Wogegen hier geprüft wird
//
// **Dass nur der Standwechsel angeboten wird, der gerade geht**, und dass ein
// Beenden gefragt wird: es ist endgültig, und §6 verlangt die Frage dort, wo
// geklickt wird.
//
// Die Läufe standen bis zum 2026-09-24 in `partiebefehle.test.tsx`; die
// Knöpfe stehen seitdem unter dem Tisch (UX-08,
// #1680).
import { render, screen, within } from '@testing-library/react'
import { expect, it, vi } from 'vitest'
import nutzerereignis from '@testing-library/user-event'

import { Partiefuehrung } from './Partiefuehrung'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { einLauf, einePartie, element } from './partieprobe'

import type { Partie } from '../apipartie'
import type { Partielauf } from '../haken/usePartie'

const eigene = kataloge.de.seiten.spielen

// zeige zeichnet die Führung der Partie und gibt die Hand.
function zeige(lauf: Partielauf, partie: Partie = einePartie()) {
  render(
    <Sprachraum sprache="de">
      <Partiefuehrung lauf={lauf} partie={partie} />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// **Anhalten geht bei einer laufenden Partie, Fortsetzen nicht.**
it('bietet nur den Standwechsel an, der gerade geht', async () => {
  const gewechselt = vi.fn()
  const hand = zeige(einLauf({ wechseln: gewechselt }))

  expect(screen.getByRole('button', { name: eigene.fortsetzen })).toBeDisabled()

  await hand.click(screen.getByRole('button', { name: eigene.pausieren }))

  expect(gewechselt).toHaveBeenCalledWith('pausieren')
})

// **Bei einer pausierten Partie umgekehrt.**
it('setzt eine pausierte Partie fort', async () => {
  const gewechselt = vi.fn()
  const hand = zeige(
    einLauf({ wechseln: gewechselt }),
    einePartie({ stand: 'pausiert' }),
  )

  expect(screen.getByRole('button', { name: eigene.pausieren })).toBeDisabled()

  await hand.click(screen.getByRole('button', { name: eigene.fortsetzen }))

  expect(gewechselt).toHaveBeenCalledWith('fortsetzen')
})

// **Eine beendete Partie lässt sich nicht noch einmal beenden.**
it('bietet einer beendeten Partie nichts mehr an', () => {
  zeige(einLauf(), einePartie({ stand: 'beendet' }))

  for (const name of [eigene.pausieren, eigene.fortsetzen, eigene.beenden]) {
    expect(screen.getByRole('button', { name })).toBeDisabled()
  }
})

// **Während ein Wechsel läuft, ruht der nächste** (§7, Regel 4).
it('sperrt die Knöpfe, solange ein Wechsel läuft', () => {
  zeige(einLauf({ laeuft: (was) => was === 'partie-p1' }))

  for (const name of [eigene.pausieren, eigene.fortsetzen, eigene.beenden]) {
    expect(screen.getByRole('button', { name })).toBeDisabled()
  }
})

// **Das Beenden fragt, und ein Nein tut nichts.**
it('fragt vor dem Beenden und tut bei Nein nichts', async () => {
  const gewechselt = vi.fn()
  const hand = zeige(einLauf({ wechseln: gewechselt }))

  await hand.click(screen.getByRole('button', { name: eigene.beenden }))

  expect(screen.getByText(eigene.beendenFrage)).toBeInTheDocument()

  await hand.click(screen.getByRole('button', { name: eigene.beendenNein }))

  expect(gewechselt).not.toHaveBeenCalled()
  expect(screen.queryByText(eigene.beendenFrage)).not.toBeInTheDocument()
})

// **Und ein Ja beendet.**
it('beendet nach der Rückfrage', async () => {
  const gewechselt = vi.fn()
  const hand = zeige(einLauf({ wechseln: gewechselt }))

  await hand.click(screen.getByRole('button', { name: eigene.beenden }))

  // **Die Rückfrage trägt dieselbe Aufschrift wie der Knopf**, und das ist
  // richtig so: sie sagt, was geschieht. Gesucht wird deshalb in ihr.
  const frage = screen.getByText(eigene.beendenFrage).closest('form, div')

  await hand.click(
    within(element(frage)).getByRole('button', { name: eigene.beendenJa }),
  )

  expect(gewechselt).toHaveBeenCalledWith('beenden')
  expect(screen.queryByText(eigene.beendenFrage)).not.toBeInTheDocument()
})
