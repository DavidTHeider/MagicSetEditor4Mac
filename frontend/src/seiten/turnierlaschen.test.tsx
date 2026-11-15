// **Die Laschen der Werkbank in der Ordnung des Artboards** — UXR-12
// (#1920).
//
// Das Artboard *Turniere* zeichnet *Runde · Rangliste · Teilnehmende ·
// Decklisten* — was während eines Abends gebraucht wird, vorn. Gebaut war
// *Teilnehmer · Runden · Decklisten · Rangliste · Protokoll*: die Rangliste,
// nach der jede Runde gefragt wird, stand an vierter Stelle, und die Lasche
// hieß anders als der Schritt des Assistenten, der dieselben Menschen
// anmeldet. Ein Turnier ohne Runde schlägt weiter bei den Teilnehmenden auf
// — dort fängt es an.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { Turniere } from './Turniere'
import { aufgeschlagen } from './turnierwerkbank'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { abend, rangliste } from '../test/turnierbestand'
import { kataloge } from '../sprache/kataloge'
import { turnierprobe } from './turnierprobe'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

// laschen gibt die Namen der Laschen und die gewählte.
function laschen(werkbank: HTMLElement) {
  const alle = within(werkbank).getAllByRole('tab')

  return {
    namen: alle.map((lasche) => lasche.textContent),
    gewaehlt: alle.find(
      (lasche) => lasche.getAttribute('aria-selected') === 'true',
    )?.textContent,
  }
}

describe('die Laschen der Werkbank', () => {
  it('stehen in der Ordnung des Artboards und öffnen bei den Runden', async () => {
    const werkbank = await aufgeschlagen(
      abend,
      {
        TurnierRangliste: () => Promise.resolve({ rangliste, fehler: null }),
      },
      'runden',
    )

    expect(laschen(werkbank)).toEqual({
      namen: [
        texte.gebietRunden,
        texte.gebietRangliste,
        texte.gebietTeilnehmer,
        texte.gebietDecklisten,
        texte.gebietProtokoll,
      ],
      gewaehlt: texte.gebietRunden,
    })
  })

  it('heißen wie der Schritt des Assistenten', () => {
    expect(texte.gebietTeilnehmer).toBe(texte.schrittTeilnehmende)
  })

  it('öffnen ein Turnier ohne Runde bei den Teilnehmenden', async () => {
    // **Ohne `aufgeschlagen`** — es wählte die Lasche selbst.
    belegeApi({
      Turniere: () =>
        Promise.resolve({ turniere: [turnierprobe()], fehler: null }),
      Turnier: () => Promise.resolve({ turnier: turnierprobe(), fehler: null }),
      TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
    })
    render(
      <Sprachraum sprache="de">
        <Turniere />
      </Sprachraum>,
    )
    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Küchentisch`,
      }),
    )
    const werkbank = await screen.findByRole('region', { name: 'Küchentisch' })

    expect(laschen(werkbank).gewaehlt).toBe(texte.gebietTeilnehmer)
  })
})
