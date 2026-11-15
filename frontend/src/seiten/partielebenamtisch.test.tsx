// **Leben am Tisch** — P4
// (#2372),
// Befund C-20: der Regler am eigenen Sitz stellt `spieler.leben` mit der
// Änderung — **nur, wenn der Kern den Befehl gerade anbietet**; sonst steht
// die Zahl ohne Knöpfe da.
import { fireEvent, render, screen } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { Partietisch } from './Partietisch'
import { Sprachraum } from '../sprache/kontext'
import { einLauf, einePartie } from './partieprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.spielen

function zeige(erlaubt?: string[]) {
  const befohlen = vi.fn()
  const partie = einePartie({ fuer: 'sitz1' })
  render(
    <Sprachraum sprache="de">
      <Partietisch
        lauf={einLauf({
          partie,
          befehlen: befohlen,
          ...(erlaubt === undefined ? {} : { erlaubt }),
        })}
        partie={partie}
      />
    </Sprachraum>,
  )

  return befohlen
}

it('stellt die Änderung am eigenen Leben als Befehl', () => {
  const befohlen = zeige()

  const [weniger] = screen.getAllByRole('button', {
    name: eigene.lebenWeniger,
  })
  expect(weniger).toBeDefined()
  if (weniger !== undefined) {
    fireEvent.click(weniger)
  }

  expect(befohlen).toHaveBeenCalledWith('spieler.leben', { um: -1 })
})

it('zeigt ohne den Befehl keine Knöpfe', () => {
  zeige(['zug.passen'])

  expect(screen.queryByRole('button', { name: eigene.lebenWeniger })).toBeNull()
})
