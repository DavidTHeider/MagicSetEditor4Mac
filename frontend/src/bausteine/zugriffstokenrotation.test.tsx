// **Rotation, Ablaufwahlen und Sammelwiderruf am Fenster** — Werkstück O2
// von Phase 5.2 (#1426).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Zugriffstoken } from './Zugriffstoken'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { TokenDTO } from '../api'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.nutzer

function einToken(teile: Partial<TokenDTO> = {}): TokenDTO {
  return {
    id: '1',
    profilID: '10',
    bezeichnung: 'Nachtlauf',
    principal: 'Rechner',
    scope: 'lesen',
    ablauf: '',
    angelegt: '2026-09-05T12:00:00Z',
    abgelaufen: false,
    letzteNutzung: '',
    vorgaengerBis: '',
    partie: '',
    sitz: '',
    fassung: '1',
    ...teile,
  }
}

function zeige(teile: Record<string, unknown> = {}) {
  belegeApi({
    Token: vi.fn(() => Promise.resolve({ token: [einToken()] })),
    Tokenablaufwahlen: vi.fn(() =>
      Promise.resolve({
        wahlen: [
          { name: 'tag', ablauf: '2026-09-09T12:00:00Z' },
          { name: 'monat', ablauf: '2026-10-08T12:00:00Z' },
          { name: 'jahr', ablauf: '2027-09-08T12:00:00Z' },
          { name: 'nie', ablauf: '' },
        ],
        fehler: null,
      }),
    ),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Zugriffstoken profilID="10" />
    </Sprachraum>,
  )
}

describe('Rotation und Sammelwiderruf', () => {
  it('fragt vor der Rotation nach und zeigt dann den neuen Klartext', async () => {
    const rotieren = vi.fn(() =>
      Promise.resolve({
        token: einToken({
          vorgaengerBis: '2026-09-08T13:00:00Z',
          fassung: '2',
        }),
        klartext: 'mse_1_NEU',
      }),
    )

    zeige({ TokenRotieren: rotieren })

    await userEvent.click(
      await screen.findByRole('button', { name: eigene.tokenRotieren }),
    )

    // **Noch nichts geschehen** — erst die Rückfrage.
    expect(rotieren).not.toHaveBeenCalled()

    const rueckfrage = screen.getByRole('group', {
      name: eigene.tokenRotierenFrage,
    })

    await userEvent.click(
      within(rueckfrage).getByRole('button', { name: eigene.tokenRotieren }),
    )

    expect(await screen.findByText('mse_1_NEU')).toBeInTheDocument()
    expect(rotieren).toHaveBeenCalledWith('1')
  })

  it('lässt die Rotation abbrechen', async () => {
    const rotieren = vi.fn()

    zeige({ TokenRotieren: rotieren })

    await userEvent.click(
      await screen.findByRole('button', { name: eigene.tokenRotieren }),
    )
    await userEvent.click(
      within(
        screen.getByRole('group', { name: eigene.tokenRotierenFrage }),
      ).getByRole('button', { name: kataloge.de.bausteine.aktionen.abbrechen }),
    )

    expect(
      screen.queryByRole('group', { name: eigene.tokenRotierenFrage }),
    ).not.toBeInTheDocument()
    expect(rotieren).not.toHaveBeenCalled()
  })

  it('zieht alle Token nach Rückfrage zurück und sagt, wie viele', async () => {
    const widerrufen = vi.fn(() =>
      Promise.resolve({ gezogen: 2, fehler: null }),
    )

    zeige({ TokenAlleWiderrufen: widerrufen })

    await userEvent.click(
      await screen.findByRole('button', { name: eigene.tokenAlleWiderrufen }),
    )

    expect(widerrufen).not.toHaveBeenCalled()

    await userEvent.click(
      within(
        screen.getByRole('group', { name: eigene.tokenAlleWiderrufenFrage }),
      ).getByRole('button', { name: eigene.tokenAlleWiderrufen }),
    )

    expect(
      await screen.findByText('2 Token zurückgezogen.'),
    ).toBeInTheDocument()
    expect(widerrufen).toHaveBeenCalledWith('10')
  })

  it('lässt den Sammelwiderruf abbrechen', async () => {
    const widerrufen = vi.fn()

    zeige({ TokenAlleWiderrufen: widerrufen })

    await userEvent.click(
      await screen.findByRole('button', { name: eigene.tokenAlleWiderrufen }),
    )
    await userEvent.click(
      within(
        screen.getByRole('group', { name: eigene.tokenAlleWiderrufenFrage }),
      ).getByRole('button', { name: kataloge.de.bausteine.aktionen.abbrechen }),
    )

    expect(
      screen.queryByRole('group', { name: eigene.tokenAlleWiderrufenFrage }),
    ).not.toBeInTheDocument()
    expect(widerrufen).not.toHaveBeenCalled()
  })

  // **Die Ablaufwahlen kommen aus Go und setzen den Zeitpunkt.**
  it('nimmt den Zeitpunkt einer Ablaufwahl', async () => {
    const anlegen = vi.fn(() =>
      Promise.resolve({ token: einToken({ id: '2' }), klartext: 'mse_2_X' }),
    )

    zeige({ TokenAnlegen: anlegen })

    await userEvent.type(
      await screen.findByLabelText(eigene.tokenBezeichnung),
      'Setzer',
    )
    await userEvent.type(screen.getByLabelText(eigene.tokenPrincipal), 'Cron')
    await userEvent.selectOptions(
      screen.getByLabelText(eigene.tokenAblaufwahl),
      'monat',
    )
    await userEvent.click(
      screen.getByRole('button', { name: eigene.tokenAnlegen }),
    )

    expect(anlegen).toHaveBeenCalledWith(
      expect.objectContaining({ ablauf: '2026-10-08T12:00:00Z' }),
    )
  })

  // **Eine Wahl, die der Katalog nicht kennt, bleibt beim Namen** — sie
  // kommt aus Go, und ein leerer Eintrag verschwiege sie.
  it('nennt eine unbekannte Ablaufwahl beim Namen', async () => {
    zeige({
      Tokenablaufwahlen: vi.fn(() =>
        Promise.resolve({
          wahlen: [{ name: 'sonderfrist', ablauf: '2026-09-10T12:00:00Z' }],
          fehler: null,
        }),
      ),
    })

    expect(
      await screen.findByRole('option', { name: 'sonderfrist' }),
    ).toBeInTheDocument()
  })

  it('meldet, wenn die Ablaufwahlen nicht zu lesen sind', async () => {
    zeige({
      Tokenablaufwahlen: vi.fn(() =>
        Promise.resolve({
          wahlen: [],
          fehler: { code: 'unbekannt', meldung: 'keine Uhr', feld: '' },
        }),
      ),
    })

    expect(await screen.findByRole('alert')).toHaveTextContent('keine Uhr')
  })
})
