// **Nach einem Schreiben bleibt die Lasche stehen** — Befund aus UXR-12
// (#1920).
//
// Nach jedem geglückten Schreiben lädt die Seite das offene Turnier nach.
// `useTurniere.waehle` leerte dabei den offenen Kontext, als wäre ein anderes
// Turnier gewählt: die Werkbank wurde abgebaut und neu aufgebaut, und wer bei
// den Teilnehmenden jemanden wieder anmeldete, stand danach bei den Runden.
// **Regel 1 aus §7 gilt für einen Wechsel des Bezugs** — dasselbe Turnier
// nachzuladen ist keiner.
//
// **Das Nachladen braucht eine Weile**, wie in Wirklichkeit: ein Doppel, das
// sofort antwortet, zeigte den Abbau nie (`turnierlauf.test.tsx`, Befund 1 des
// Nachlaufs über Phase 4.7).
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { aufgeschlagen } from './turnierwerkbank'
import { raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { rundenprobe, teilnehmerprobe, turnierprobe } from './turnierprobe'

import type { TurnierDTO } from '../apiturniere'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

// gewaehlt gibt den Namen der gewählten Lasche.
function gewaehlt(werkbank: HTMLElement) {
  return within(werkbank)
    .getAllByRole('tab')
    .find((lasche) => lasche.getAttribute('aria-selected') === 'true')
    ?.textContent
}

describe('die Werkbank nach einem Schreiben', () => {
  it('bleibt bei der Lasche, auf der geschrieben wurde', async () => {
    const stand = () =>
      turnierprobe({
        teilnehmer: [teilnehmerprobe({ ausgestiegen: true })],
        runden: [rundenprobe()],
      })

    let loeseAuf: (() => void) | undefined

    const Turnier = vi.fn(() => {
      if (Turnier.mock.calls.length === 1) {
        return Promise.resolve({ turnier: stand(), fehler: null })
      }

      return new Promise<{ turnier: TurnierDTO; fehler: null }>((fertig) => {
        loeseAuf = () => {
          fertig({ turnier: stand(), fehler: null })
        }
      })
    })
    const TurnierAussteigen = vi.fn(() => Promise.resolve({ fehler: null }))

    const werkbank = await aufgeschlagen(
      stand(),
      { Turnier, TurnierAussteigen },
      'teilnehmer',
    )

    expect(gewaehlt(werkbank)).toBe(texte.gebietTeilnehmer)

    await userEvent.click(
      screen.getByRole('button', { name: `${texte.einsteigen}: Ada` }),
    )

    // **Gewartet wird auf das Nachladen**, nicht auf den Ruf: erst wenn es
    // läuft, stand der offene Kontext früher auf `null`.
    await waitFor(() => {
      expect(Turnier).toHaveBeenCalledTimes(2)
    })

    expect(werkbank).toBeInTheDocument()
    expect(gewaehlt(werkbank)).toBe(texte.gebietTeilnehmer)

    loeseAuf?.()

    await waitFor(() => {
      expect(TurnierAussteigen).toHaveBeenCalledWith('7', '3', false)
    })
    expect(werkbank).toBeInTheDocument()
    expect(gewaehlt(werkbank)).toBe(texte.gebietTeilnehmer)
  })
})
