// **Die Kennzahlen des offenen Sets** — AP-9, Teil 2
// (#2366),
// Entwurf `Bereiche09`: am Setkopf die Karten und die Karten mit Befund —
// aus dem Prüfstand des Sets, gelesen und nicht geprüft (PERF02).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Anschluss } from '../anschluss'
import type { Anschlussstelle } from '../anschlussraum'
import type { SetproblemstandDTO } from '../apiprobleme'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Setleiste } from './Setleiste'

const texte = kataloge.de.allgemein.arbeitsplatz.leiste

const uebergeben: Anschluss[] = []
vi.mock('../anschlussraum', async (echt) => ({
  ...(await echt<typeof import('../anschlussraum')>()),
  useAnschluss: (): Anschlussstelle => ({
    offen: undefined,
    uebergib: (anschluss) => {
      uebergeben.push(anschluss)
    },
    verbrauche: () => undefined,
  }),
}))

afterEach(() => {
  raeumeApi()
  uebergeben.length = 0
})

function stand(teil: Partial<SetproblemstandDTO> = {}): SetproblemstandDTO {
  return {
    setID: '7',
    name: 'Tischprobe',
    urteil: 'aktuell',
    zahl: 2,
    sicher: true,
    geprueft: 5,
    gesamt: 5,
    beendet: '2026-10-17T03:00:00Z',
    alter: 240,
    karten: [
      { karteID: '71', befunde: ['a'] },
      { karteID: '72', befunde: ['b'] },
    ],
    ...teil,
  }
}

function zeige(staende: SetproblemstandDTO[]) {
  belegeApi({
    Setprobleme: () =>
      Promise.resolve({
        staende: [stand({ setID: '8', name: 'Anderes', zahl: 9 }), ...staende],
        fehler: null,
      }),
  })
  const ziele = { wechseln: vi.fn(), oeffneKarte: vi.fn() }
  render(
    <Sprachraum sprache="de">
      <Setleiste setID="7" karten={5} verlaufsstand={0} {...ziele} />
    </Sprachraum>,
  )

  return ziele
}

const leiste = async () =>
  within(await screen.findByRole('toolbar', { name: texte.bezeichnung }))

describe('die Leiste des Sets', () => {
  it('führt zu den Karten und zu den Karten mit Befund', async () => {
    const ziele = zeige([stand()])

    await userEvent.click(
      await (await leiste()).findByRole('button', { name: /^Karten 5/ }),
    )
    expect(ziele.wechseln).toHaveBeenCalledWith('karten')

    await userEvent.click(
      await (
        await leiste()
      ).findByRole('button', {
        name: `${texte.befund} 2 geprüft vor 4 min`,
      }),
    )
    expect(ziele.oeffneKarte).toHaveBeenCalledWith('7', '71')
    expect(uebergeben.at(-1)).toEqual({
      art: 'kartenauswahl',
      name: '2 Karten mit Befund',
      setID: '7',
      karteIDs: ['71', '72'],
    })
  })

  // **Keine Null über Ungeprüftes, kein Ziel ins Leere** — wie auf der
  // Startseite (AP-8).
  it('sagt „nicht geprüft“, solange der Stand nicht gilt', async () => {
    zeige([stand({ sicher: false, urteil: 'veraltet' })])

    expect(
      await (await leiste()).findByText(texte.nichtGeprueft),
    ).toBeInTheDocument()
    expect((await leiste()).getAllByRole('button')).toHaveLength(1)
  })

  it('führt mit null Befunden nirgends hin', async () => {
    zeige([stand({ zahl: 0, karten: [] })])

    expect(await (await leiste()).findByText('0')).toBeInTheDocument()
    expect((await leiste()).getAllByRole('button')).toHaveLength(1)
  })

  it('sagt „nicht geprüft“ für ein Set ohne Stand', async () => {
    zeige([])

    expect(
      await (await leiste()).findByText(texte.nichtGeprueft),
    ).toBeInTheDocument()
  })

  it('sagt „nicht gelesen“, wenn der Stand nicht zu lesen war', async () => {
    belegeApi({
      Setprobleme: () =>
        Promise.resolve({
          staende: [],
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            meldung: 'kaputt',
            feld: '',
            werte: {},
          },
        }),
    })
    render(
      <Sprachraum sprache="de">
        <Setleiste
          setID="7"
          karten={5}
          verlaufsstand={0}
          wechseln={vi.fn()}
          oeffneKarte={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(
      await (await leiste()).findByText(texte.nichtGelesen),
    ).toBeInTheDocument()
  })
})
