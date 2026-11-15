// **Ein gesunder Bestand wird nicht zum Ersetzen angeboten** — F05
// (#2344),
// Befunde A-05 und A-21.
//
// Belegt, voll, gesperrt, nur lesbar und ein Fehler des Datenträgers: in
// allen fünf ist die Datei vermutlich in Ordnung. Die Seite rät dann zum
// erneuten Versuch und bietet weder das Zurückspielen noch den Weg zurück
// an — wer ihn nähme, verlöre, was seit der Sicherung dazukam. Ein Lauf über
// `App`, aus demselben Grund wie `wiederherstellung.test.tsx`.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { kataloge } from './sprache/kataloge'
import { kurzform } from './sprache/kurzform'

afterEach(() => {
  raeumeApi()
})

// **Englisch**: ohne offenen Bestand spricht die Seite die Sprache des
// Systems, und die ist unter jsdom Englisch (`wiederherstellung.test.tsx`).
const worte = kataloge.en.seiten.bestandspflege
const fehler = kataloge.en.fehler

const rueckweg = {
  datei: 'bestand-vor-0.9.0.db',
  schemastand: '0090_probe.sql',
  sets: 2,
  karten: 30,
}

function belege(teile: Record<string, unknown>) {
  belegeApi({
    ...belegeFenster('de'),
    Einstellungen: () => new Promise<never>(() => undefined),
    Datenstand: vi.fn().mockResolvedValue({ stand: 0, fehler: null }),
    Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
    ...teile,
  })
}

async function titelSteht() {
  expect(
    await screen.findByRole('heading', { name: worte.wiederherstellungTitel }),
  ).toBeInTheDocument()
}

describe('Die Lagen des Starts', () => {
  it.each([
    ['belegt', fehler.belegt],
    ['voll', fehler.voll],
    ['gesperrt', fehler.gesperrt],
    ['bestandNurLesbar', fehler.bestandNurLesbar],
    ['bestandEA', fehler.bestandEA],
  ])(
    'rät bei %s zum erneuten Versuch und nicht zum Ersetzen',
    async (schluessel, satz) => {
      belege({
        Startbefund: vi.fn().mockResolvedValue({
          offen: false,
          grund: 'roh',
          schluessel,
          rueckweg,
        }),
      })

      render(<App />)
      await titelSteht()

      expect(screen.getByText(satz)).toBeInTheDocument()
      expect(
        screen.getByText(kurzform(worte.wiederherstellungHinweisHeil)),
      ).toBeInTheDocument()
      expect(
        screen.getByRole('button', { name: worte.erneutVersuchen }),
      ).toBeInTheDocument()
      expect(
        screen.queryByRole('button', { name: worte.waehlen }),
      ).not.toBeInTheDocument()
      expect(screen.queryByText(worte.wegZurueckTitel)).not.toBeInTheDocument()
    },
  )

  it('bietet bei einer beschädigten Datei das Zurückspielen an und keinen Versuch', async () => {
    belege({
      Startbefund: vi.fn().mockResolvedValue({
        offen: false,
        grund: 'roh',
        schluessel: 'bestandUnbrauchbar',
        rueckweg,
      }),
    })

    render(<App />)
    await titelSteht()

    expect(
      screen.getByRole('button', { name: worte.waehlen }),
    ).toBeInTheDocument()
    expect(screen.getByText(worte.wegZurueckTitel)).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: worte.erneutVersuchen }),
    ).not.toBeInTheDocument()
  })

  it('bietet bei einer ungeklärten Lage beides an', async () => {
    belege({
      Startbefund: vi.fn().mockResolvedValue({
        offen: false,
        grund: 'roh',
        schluessel: 'startUngeklaert',
      }),
    })

    render(<App />)
    await titelSteht()

    expect(screen.getByText(fehler.startUngeklaert)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: worte.erneutVersuchen }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: worte.waehlen }),
    ).toBeInTheDocument()
  })

  it('öffnet nach einem geglückten Versuch die Anwendung', async () => {
    const StartWiederholen = vi
      .fn()
      .mockResolvedValue({ offen: true, grund: '' })
    belege({
      Startbefund: vi
        .fn()
        .mockResolvedValueOnce({
          offen: false,
          grund: 'roh',
          schluessel: 'belegt',
        })
        .mockResolvedValue({ offen: true, grund: '' }),
      StartWiederholen,
    })

    render(<App />)
    await userEvent.click(
      await screen.findByRole('button', { name: worte.erneutVersuchen }),
    )

    await waitFor(() => {
      expect(
        screen.queryByRole('heading', { name: worte.wiederherstellungTitel }),
      ).not.toBeInTheDocument()
    })
    expect(StartWiederholen).toHaveBeenCalledTimes(1)
  })

  it('sagt es, wenn auch der Versuch scheitert', async () => {
    belege({
      Startbefund: vi.fn().mockResolvedValue({
        offen: false,
        grund: 'roh',
        schluessel: 'belegt',
      }),
      StartWiederholen: vi.fn().mockResolvedValue({
        offen: false,
        grund: 'roh',
        schluessel: 'belegt',
      }),
    })

    render(<App />)
    await userEvent.click(
      await screen.findByRole('button', { name: worte.erneutVersuchen }),
    )

    expect(await screen.findByText(worte.weiterGescheitert)).toBeInTheDocument()
  })

  it('zeigt einen Fehler der Bindung am Knopf', async () => {
    belege({
      Startbefund: vi.fn().mockResolvedValue({
        offen: false,
        grund: 'roh',
        schluessel: 'belegt',
      }),
      StartWiederholen: vi.fn().mockRejectedValue(new Error('Probe')),
    })

    render(<App />)
    await userEvent.click(
      await screen.findByRole('button', { name: worte.erneutVersuchen }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})
