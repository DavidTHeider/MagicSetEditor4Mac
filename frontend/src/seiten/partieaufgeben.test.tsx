// **Aufgeben fragt nach** — P1b
// (#2370),
// Befund C-06: ein Einzelklick beendete das Spiel, sofort und endgültig
// (CR 104.3a, Fassung 2026-09-25). Abbrechen erzeugt kein Ereignis,
// Bestätigen genau eines; an einem beendeten Spiel gibt es nichts mehr
// aufzugeben.
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Partietisch } from './Partietisch'
import { Sprachraum } from '../sprache/kontext'
import { einLauf, einePartie } from './partieprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.spielen

function zeige(stand = 'laeuft') {
  const befohlen = vi.fn()
  const partie = einePartie({ fuer: 'sitz1', stand })
  render(
    <Sprachraum sprache="de">
      <Partietisch
        lauf={einLauf({ partie, befehlen: befohlen })}
        partie={partie}
      />
    </Sprachraum>,
  )

  return befohlen
}

it('fragt vor dem Aufgeben nach, und Abbrechen befiehlt nichts', async () => {
  const befohlen = zeige()
  const hand = nutzerereignis.setup()

  await hand.click(screen.getByRole('button', { name: eigene.aufgeben }))

  expect(befohlen).not.toHaveBeenCalled()
  expect(screen.getByText(eigene.aufgebenFrage)).toBeInTheDocument()

  await hand.click(screen.getByRole('button', { name: eigene.aufgebenNein }))

  expect(befohlen).not.toHaveBeenCalled()
  expect(screen.queryByText(eigene.aufgebenFrage)).toBeNull()
})

it('befiehlt nach dem Bestätigen genau einmal', async () => {
  const befohlen = zeige()
  const hand = nutzerereignis.setup()

  await hand.click(screen.getByRole('button', { name: eigene.aufgeben }))
  await hand.click(screen.getByRole('button', { name: eigene.aufgebenJa }))

  expect(befohlen).toHaveBeenCalledTimes(1)
  expect(befohlen).toHaveBeenCalledWith('zug.aufgeben', {})
})

it('bietet an einem beendeten Spiel kein Aufgeben an', () => {
  zeige('beendet')

  expect(screen.getByRole('button', { name: eigene.aufgeben })).toBeDisabled()
})

// **Endet die Partie, während die Frage steht, geht die Frage** — sonst
// schickte „Ja, aufgeben“ den Befehl an ein beendetes Spiel.
it('nimmt die Rückfrage zurück, wenn die Partie inzwischen endet', async () => {
  const befohlen = vi.fn()
  const zeichne = (stand: string) => {
    const partie = einePartie({ fuer: 'sitz1', stand })

    return (
      <Sprachraum sprache="de">
        <Partietisch
          lauf={einLauf({ partie, befehlen: befohlen })}
          partie={partie}
        />
      </Sprachraum>
    )
  }
  const { rerender } = render(zeichne('laeuft'))
  const hand = nutzerereignis.setup()
  await hand.click(screen.getByRole('button', { name: eigene.aufgeben }))
  expect(screen.getByText(eigene.aufgebenFrage)).toBeInTheDocument()

  rerender(zeichne('beendet'))

  expect(screen.queryByText(eigene.aufgebenFrage)).toBeNull()
  expect(befohlen).not.toHaveBeenCalled()
})
