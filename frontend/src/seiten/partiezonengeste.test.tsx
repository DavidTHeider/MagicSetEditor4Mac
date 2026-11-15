// **Der Tastenweg aus `Tisch09` an der Liste der Zonen** — GESTE (#2358):
// ⇥ zur Karte, Leertaste greift, Pfeile wählen die Zone, ↩ legt ab, Escape
// bricht ab. Gestellt wird **derselbe Befehl** wie beim Ziehen mit der Maus
// (`objekt.schieben`, `tischbefehle.ts`) — und nur, wenn der Kern ihn erlaubt.
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import type { Partie } from '../apipartie'
import { Sprachraum } from '../sprache/kontext'
import { Partiezonen } from './Partiezonen'
import { einObjekt, einePartie, ruhenderLauf } from './partieprobe'

function zeige(
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void,
  erlaubt: string[] = ['objekt.ziehen'],
  partie: Partie = einePartie({
    objekte: [
      einObjekt({ kennung: 'k1', zone: 'tisch', name: 'Wald', geheim: false }),
    ],
  }),
) {
  const baum = (welche: Partie) => (
    <Sprachraum sprache="de">
      <Partiezonen
        partie={welche}
        texturen={ruhenderLauf()}
        gewaehlt={undefined}
        waehle={() => undefined}
        wirkungen={new Map()}
        erlaubt={erlaubt}
        befehlen={befehlen}
        handlungen={[]}
        greife={() => undefined}
        tastenFuer={() => ''}
        taste={() => false}
      />
    </Sprachraum>
  )
  const gezeichnet = render(baum(partie))

  return {
    nutzer: nutzerereignis.setup(),
    neu: (welche: Partie) => {
      gezeichnet.rerender(baum(welche))
    },
  }
}

const karte = () => screen.getByRole('button', { name: /Wald/ })
const ansage = () =>
  document.querySelector('[data-gestenansage]')?.textContent ?? ''

it('legt eine Karte mit Leertaste, Pfeil und Eingabe in eine andere Zone', async () => {
  const befehlen = vi.fn()
  const { nutzer } = zeige(befehlen)

  karte().focus()
  await nutzer.keyboard(' ')
  expect(ansage()).toMatch(/^Gegriffen\. Ziel: /)
  await nutzer.keyboard('{ArrowRight}{Enter}')

  expect(befehlen).toHaveBeenCalledTimes(1)
  const [art, nutzlast] = befehlen.mock.calls[0] as [
    string,
    { objekt: string; zone: string; sitz: string },
  ]
  // **Dieselbe gedeutete Geste wie mit der Maus** — P5 (#2373): der Kern
  // entscheidet, was sie heißt, und der Sitz geht mit.
  expect(art).toBe('objekt.ziehen')
  expect(nutzlast.objekt).toBe('k1')
  expect(nutzlast.zone).not.toBe('tisch')
  expect(nutzlast.sitz).not.toBe('')
  expect(ansage()).toMatch(/^Abgelegt: /)
})

it('bricht mit Escape ab und stellt nichts', async () => {
  const befehlen = vi.fn()
  const { nutzer } = zeige(befehlen)

  karte().focus()
  await nutzer.keyboard(' {Escape}')

  expect(befehlen).not.toHaveBeenCalled()
  expect(ansage()).toBe('Abgebrochen. Nichts wurde bewegt.')
})

it('greift nicht, wenn der Kern keine Geste erlaubt', async () => {
  const befehlen = vi.fn()
  const { nutzer } = zeige(befehlen, ['objekt.tappen'])

  karte().focus()
  await nutzer.keyboard(' ')

  expect(ansage()).toBe('')
})

it('lässt einen Griff fallen, wenn sich die Partie inzwischen geändert hat', async () => {
  const befehlen = vi.fn()
  const partie = einePartie({
    objekte: [
      einObjekt({ kennung: 'k1', zone: 'tisch', name: 'Wald', geheim: false }),
    ],
  })
  const { nutzer, neu } = zeige(befehlen, ['objekt.ziehen'], partie)

  karte().focus()
  await nutzer.keyboard(' ')
  neu({ ...partie, fassung: `${partie.fassung}-danach` })
  karte().focus()
  await nutzer.keyboard('{ArrowRight}{Enter}')

  expect(befehlen).not.toHaveBeenCalled()
})

// **Eine andere Taste lässt den Griff, wie er ist** — sie geht ihren Weg, und
// ein Pfeil danach wählt weiter die Zone.
it('lässt andere Tasten durch, ohne den Griff zu lösen', async () => {
  const befehlen = vi.fn()
  const { nutzer } = zeige(befehlen)

  karte().focus()
  await nutzer.keyboard(' ')
  const gegriffen = ansage()
  await nutzer.keyboard('x')

  expect(ansage()).toBe(gegriffen)
  expect(karte()).toHaveAttribute('data-gegriffen', 'true')
  await nutzer.keyboard('{ArrowRight}')
  expect(ansage()).toMatch(/^Ziel: /)
})
