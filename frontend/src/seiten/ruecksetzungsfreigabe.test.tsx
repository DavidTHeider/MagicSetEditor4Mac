// **Die Freigabe für den Neuanfang wird am Fenster erteilt** — G12
// (#2282).
//
// Der Weg von der Vorschau über den Knopf bis zum beiseitegelegten Bestand —
// und das Nein, und was danach bleibt.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Ruecksetzung } from './Ruecksetzung'
import { Schreibrechtraum } from '../haken/schreibrecht'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const worte = kataloge.de.seiten.ruecksetzung

afterEach(() => {
  raeumeApi()
})

const anwendung = {
  art: 'anwendung',
  betroffen: [],
  hindernisse: [],
  bereit: true,
  nameNoetig: false,
  nachweisNoetig: true,
  sicherungEmpfohlen: true,
  revision: 'marke-der-vorschau',
  fehler: null,
}

describe('die Freigabe für den Neuanfang', () => {
  it('verlangt für die ganze Anwendung eine Freigabe statt eines Namens', async () => {
    const tun = vi.fn().mockResolvedValue({
      art: 'anwendung',
      zurueckgesetzt: 0,
      beiseite: '/alt/bestand.db',
      fehler: null,
    })

    belegeApi({
      Ruecksetzungsarten: () => Promise.resolve(['anwendung']),
      Ruecksetzungsvorschau: () => Promise.resolve(anwendung),
      Zuruecksetzen: tun,
      RuecksetzungFreigabeErteilen: () =>
        Promise.resolve({
          kennung: 'freigabe-1',
          laeuftab: '2026-10-10T12:30:00Z',
          fehler: null,
        }),
    })

    const nutzer = userEvent.setup()

    render(
      <Sprachraum sprache="de">
        <Ruecksetzung gewechselt={vi.fn()} veraendert={vi.fn()} />
      </Sprachraum>,
    )

    await nutzer.click(
      await screen.findByRole('button', { name: worte.nachsehen }),
    )

    expect(await screen.findByText(worte.nichtsBetroffen)).toBeInTheDocument()
    expect(screen.queryByRole('textbox', { name: worte.nameTippen })).toBeNull()

    // **Keine Kennung zum Abtippen** (G12, #2282): sie stand nirgends, und
    // ausgestellt hat sie bis dahin auch niemand.
    expect(screen.queryByRole('textbox')).toBeNull()

    const knopf = screen.getByRole('button', { name: worte.tun })
    expect(knopf).toBeDisabled()

    await nutzer.click(screen.getByRole('button', { name: worte.freigeben }))

    expect(
      await screen.findByText(worte.freigegeben.split('{')[0] ?? '', {
        exact: false,
      }),
    ).toBeInTheDocument()
    // **Die Kennung steht da** — für die Kommandozeile, die sie einlöst.
    expect(screen.getByText('freigabe-1')).toBeInTheDocument()
    await waitFor(() => {
      expect(knopf).toBeEnabled()
    })

    await nutzer.click(knopf)

    await waitFor(() => {
      expect(tun).toHaveBeenCalledWith(
        'anwendung',
        '',
        '',
        '',
        'freigabe-1',
        'marke-der-vorschau',
      )
    })

    expect(
      await screen.findByText(/\/alt\/bestand\.db/, { exact: false }),
    ).toBeInTheDocument()
  })

  it('meldet eine verweigerte Freigabe am Ort und gibt den Knopf nicht frei', async () => {
    const tun = vi.fn()

    belegeApi({
      Ruecksetzungsarten: () => Promise.resolve(['anwendung']),
      Ruecksetzungsvorschau: () => Promise.resolve(anwendung),
      Zuruecksetzen: tun,
      RuecksetzungFreigabeErteilen: () =>
        Promise.resolve({
          kennung: '',
          laeuftab: '',
          fehler: {
            code: 'erlaubnis',
            schluessel: '',
            meldung: 'Einen Nachweis stellt aus, wer mindestens mitarbeitet.',
            feld: '',
          },
        }),
    })

    const nutzer = userEvent.setup()

    render(
      <Sprachraum sprache="de">
        <Ruecksetzung gewechselt={vi.fn()} veraendert={vi.fn()} />
      </Sprachraum>,
    )

    await nutzer.click(
      await screen.findByRole('button', { name: worte.nachsehen }),
    )
    await nutzer.click(
      await screen.findByRole('button', { name: worte.freigeben }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'wer mindestens mitarbeitet',
    )
    expect(screen.getByRole('button', { name: worte.tun })).toBeDisabled()
    expect(
      screen.queryByText(worte.freigegeben.split('{')[0] ?? '', {
        exact: false,
      }),
    ).toBeNull()
    expect(tun).not.toHaveBeenCalled()
  })

  // **Ein Leser erteilt keine Freigabe** — Befund 4 des Zweitblicks über
  // Phase G: der Knopf ruhte nur während des Ladens, die Absage kam erst vom
  // Dienst.
  it('sperrt die Freigabe für einen Leser', async () => {
    const erteilen = vi.fn()

    belegeApi({
      Ruecksetzungsarten: () => Promise.resolve(['anwendung']),
      Ruecksetzungsvorschau: () => Promise.resolve(anwendung),
      RuecksetzungFreigabeErteilen: erteilen,
    })

    const nutzer = userEvent.setup()

    render(
      <Sprachraum sprache="de">
        <Schreibrechtraum nurLesend>
          <Ruecksetzung gewechselt={vi.fn()} veraendert={vi.fn()} />
        </Schreibrechtraum>
      </Sprachraum>,
    )

    await nutzer.click(
      await screen.findByRole('button', { name: worte.nachsehen }),
    )

    const knopf = await screen.findByRole('button', { name: worte.freigeben })
    expect(knopf).toHaveAttribute('aria-disabled', 'true')
    expect(knopf).toHaveAccessibleDescription(
      kataloge.de.bausteine.nurlesen.knopf,
    )

    await nutzer.click(knopf)
    expect(erteilen).not.toHaveBeenCalled()
  })
})
