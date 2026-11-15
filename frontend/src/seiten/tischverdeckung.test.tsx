// **Unter einer Schublade schreibt kein wartender Klick** — Nachlauf zum
// Zweitblick über Phase 7, Rest R5. Ein Einzelklick wartet die
// Doppelklickfrist ab; öffnet in der Zeit eine Schublade über dem Tisch,
// bleibt der Tisch stehen — verdeckt und `inert` —, und der Klick stellte
// seinen Befehl dahinter. Die Verdeckung gehört deshalb zur Bedienbarkeit,
// wie Sichtwechsel und Vorhang.
import { act, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'

import { Partietisch } from './Partietisch'
import { Ruecknahmeraum } from '../ruecknahme'
import { Sprachraum } from '../sprache/kontext'
import { Verdeckung } from '../verdeckung'
import { doppelklickfrist } from '../spiel/kartengriff'
import { einLauf, einObjekt, einePartie } from './partieprobe'

import type { Partielauf } from '../haken/usePartie'

beforeEach(() => {
  vi.useFakeTimers()
})

afterEach(() => {
  vi.useRealTimers()
})

const wald = einObjekt({
  kennung: 'sitz1/wald/0',
  zone: 'tisch',
  name: 'Wald',
  geheim: false,
})

function Tisch({ lauf, verdeckt }: { lauf: Partielauf; verdeckt: boolean }) {
  return (
    <Sprachraum sprache="de">
      <Ruecknahmeraum>
        <Verdeckung value={verdeckt}>
          <Partietisch lauf={lauf} partie={lauf.partie ?? einePartie()} />
        </Verdeckung>
      </Ruecknahmeraum>
    </Sprachraum>
  )
}

it('stellt einen wartenden Klick nicht, wenn eine Schublade den Tisch verdeckt', () => {
  const befehlen = vi.fn()
  const lauf = einLauf({
    befehlen,
    partie: einePartie({ objekte: [wald] }),
    handlungen: [
      {
        objekt: 'sitz1/wald/0',
        handlungen: ['objekt.tappen', 'objekt.mana'],
        primaer: 'objekt.tappen',
        zweit: 'objekt.mana',
      },
    ],
  })

  const { rerender } = render(<Tisch lauf={lauf} verdeckt={false} />)

  fireEvent.click(screen.getByRole('button', { name: /Wald/ }), { detail: 1 })
  rerender(<Tisch lauf={lauf} verdeckt />)
  act(() => {
    vi.advanceTimersByTime(doppelklickfrist * 2)
  })

  expect(befehlen).not.toHaveBeenCalled()
})

// **Die Gegenprobe**: ohne Schublade wirkt derselbe Klick nach der Frist.
it('stellt ihn ohne Schublade nach der Frist', () => {
  const befehlen = vi.fn()
  const lauf = einLauf({
    befehlen,
    partie: einePartie({ objekte: [wald] }),
    handlungen: [
      {
        objekt: 'sitz1/wald/0',
        handlungen: ['objekt.tappen', 'objekt.mana'],
        primaer: 'objekt.tappen',
        zweit: 'objekt.mana',
      },
    ],
  })

  render(<Tisch lauf={lauf} verdeckt={false} />)

  fireEvent.click(screen.getByRole('button', { name: /Wald/ }), { detail: 1 })
  act(() => {
    vi.advanceTimersByTime(doppelklickfrist * 2)
  })

  expect(befehlen).toHaveBeenCalledWith('objekt.tappen', {
    objekt: 'sitz1/wald/0',
  })
})
