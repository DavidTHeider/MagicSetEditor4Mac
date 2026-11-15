// **Ist ein Turnier offen, faltet die Liste** — UXR-12
// (#1920).
//
// Gemessen am nativen Fenster (`uxr12`, 2026-10-03): mit acht Turnieren und
// einem Abend zu sechzehn Teilnehmenden rollte die Seite in den Laschen
// *Runden* und *Teilnehmende* 2,7 und 2,8 Bildschirmhöhen tief — über R4
// (`docs/UIREVIEW.md`, höchstens 2,5). Unter dem offenen Turnier stand die
// ganze Liste; wer an einem Abend arbeitet, braucht sie nicht. Sie steht
// jetzt hinter „Alle Turniere (N)" und ohne offenes Turnier wie bisher.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { aufgeschlagen } from './turnierwerkbank'
import { raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { turnierprobe } from './turnierprobe'
import { einsetzen } from '../sprache/einsetzen'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

const freitag = turnierprobe({ id: '7', name: 'Freitag' })
const montag = turnierprobe({ id: '8', name: 'Montag' })

describe('die Liste unter einem offenen Turnier', () => {
  it('steht hinter „Alle Turniere“ und klappt auf Wunsch auf', async () => {
    await aufgeschlagen(freitag, {
      Turniere: () =>
        Promise.resolve({ turniere: [freitag, montag], fehler: null }),
    })

    const falte = screen.getByRole('button', {
      name: einsetzen(texte.alleTurniere, { anzahl: '2' }),
    })

    expect(falte).toHaveAttribute('aria-expanded', 'false')
    expect(
      screen.queryByRole('button', { name: `${texte.waehlen}: Montag` }),
    ).toBeNull()

    await userEvent.click(falte)

    const montagOeffnen = screen.getByRole('button', {
      name: `${texte.waehlen}: Montag`,
    })
    expect(montagOeffnen).toBeVisible()
  })
})
