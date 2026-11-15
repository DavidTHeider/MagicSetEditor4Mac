// **Ein Turnier als Datei und als Bericht** — DOM-04.AK-03
// (#1699),
// Vermerk 0140: die Werkbank gibt aus und sichert den Bericht, über der
// Liste liest ein Knopf ein und schlägt das neue Turnier auf. Ein
// weggeklickter Dialog sagt nichts; ein Nein steht über der Seite.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Turniere } from './Turniere'
import { aufgeschlagen } from './turnierwerkbank'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { turnierhandlung, turniermenue } from '../test/turnierhandlung'
import { turnierprobe } from './turnierprobe'
import type { Api } from '../api'
import type { TurnierdateiAntwort } from '../apiturniere'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

// datei ist eine Antwort der Dateiwege.
function datei(teil: Partial<TurnierdateiAntwort> = {}): TurnierdateiAntwort {
  return {
    kennung: '7',
    pfad: '/tmp/freitag.mseturnier',
    inhalt: '',
    abgebrochen: false,
    befunde: [],
    fehler: null,
    ...teil,
  }
}

describe('die Ausgabe in der Werkbank', () => {
  it('gibt ein Turnier aus und sagt, wohin', async () => {
    const ausgeben = vi.fn<Api['TurnierAusgabeSpeichern']>(() =>
      Promise.resolve(datei()),
    )
    const werkbank = await aufgeschlagen(turnierprobe(), {
      TurnierAusgabeSpeichern: ausgeben,
    })

    await turnierhandlung('Küchentisch', texte.ausgeben)

    expect(
      await within(werkbank).findByText(
        einsetzen(texte.gesichertUnter, { pfad: '/tmp/freitag.mseturnier' }),
      ),
    ).toBeInTheDocument()
    expect(ausgeben).toHaveBeenCalledWith('7')
  })

  it('sichert den Bericht', async () => {
    const berichten = vi.fn<Api['TurnierBerichtSpeichern']>(() =>
      Promise.resolve(datei({ pfad: '/tmp/freitag.md' })),
    )
    const werkbank = await aufgeschlagen(turnierprobe(), {
      TurnierBerichtSpeichern: berichten,
    })

    await turnierhandlung('Küchentisch', texte.berichtSichern)

    expect(
      await within(werkbank).findByText(
        einsetzen(texte.gesichertUnter, { pfad: '/tmp/freitag.md' }),
      ),
    ).toBeInTheDocument()
    expect(berichten).toHaveBeenCalledWith('7')
  })

  it('sagt nichts, wenn der Dialog weggeklickt wird', async () => {
    const ausgeben = vi.fn<Api['TurnierAusgabeSpeichern']>(() =>
      Promise.resolve(datei({ pfad: '', abgebrochen: true })),
    )
    const werkbank = await aufgeschlagen(turnierprobe(), {
      TurnierAusgabeSpeichern: ausgeben,
    })

    await turnierhandlung('Küchentisch', texte.ausgeben)

    await waitFor(() => {
      expect(ausgeben).toHaveBeenCalled()
    })
    // **Der Eintrag ist wieder frei** — erst dann ist die Antwort angekommen.
    await userEvent.click(await turniermenue('Küchentisch'))
    await waitFor(() => {
      expect(
        screen.getByRole('menuitem', { name: texte.ausgeben }),
      ).not.toHaveAttribute('aria-disabled', 'true')
    })
    await userEvent.keyboard('{Escape}')
    expect(within(werkbank).queryByRole('status')).toBeNull()
  })

  it('meldet ein Nein über der Seite', async () => {
    await aufgeschlagen(turnierprobe(), {
      TurnierAusgabeSpeichern: () =>
        Promise.resolve(
          datei({
            fehler: {
              code: 'unbekannt' as const,
              schluessel: 'unbekannt',
              meldung: 'Die Platte ist voll.',
            },
          }),
        ),
    })

    await turnierhandlung('Küchentisch', texte.ausgeben)

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})

// seite stellt die Turnierseite hin, ohne ein Turnier aufzuschlagen.
function seite(weitere: Record<string, unknown>) {
  belegeApi({
    Turniere: () => Promise.resolve({ turniere: [], fehler: null }),
    Turnier: () =>
      Promise.resolve({
        turnier: turnierprobe({ id: '8', name: 'Samstag' }),
        fehler: null,
      }),
    TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
    ...weitere,
  })

  render(
    <Sprachraum sprache="de">
      <Turniere />
    </Sprachraum>,
  )
}

describe('das Einlesen über der Liste', () => {
  it('liest ein Turnier ein, nennt die Befunde und schlägt es auf', async () => {
    const turnier = vi.fn<Api['Turnier']>(() =>
      Promise.resolve({
        turnier: turnierprobe({ id: '8', name: 'Samstag' }),
        fehler: null,
      }),
    )
    seite({
      Turnier: turnier,
      TurnierEinlesenOeffnen: () =>
        Promise.resolve(datei({ kennung: '8', befunde: ['ohne Format'] })),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.einlesen }),
    )

    const status = await screen.findByRole('status')
    expect(status).toHaveTextContent(texte.eingelesen)
    expect(within(status).getByText('ohne Format')).toBeInTheDocument()
    expect(
      await screen.findByRole('region', { name: 'Samstag' }),
    ).toBeInTheDocument()
    expect(turnier).toHaveBeenCalledWith('8')
  })

  it('sagt nichts und schlägt nichts auf, wenn der Dialog weggeklickt wird', async () => {
    const einlesen = vi.fn<Api['TurnierEinlesenOeffnen']>(() =>
      Promise.resolve(datei({ kennung: '', pfad: '', abgebrochen: true })),
    )
    const turnier = vi.fn<Api['Turnier']>()
    seite({ TurnierEinlesenOeffnen: einlesen, Turnier: turnier })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.einlesen }),
    )

    await waitFor(() => {
      expect(einlesen).toHaveBeenCalled()
    })
    await waitFor(() => {
      expect(screen.getByRole('button', { name: texte.einlesen })).toBeEnabled()
    })
    expect(screen.queryByRole('status')).toBeNull()
    expect(turnier).not.toHaveBeenCalled()
  })

  it('meldet ein Nein über der Seite', async () => {
    seite({
      TurnierEinlesenOeffnen: () =>
        Promise.resolve(
          datei({
            kennung: '',
            fehler: {
              code: 'konflikt' as const,
              schluessel: 'konflikt',
              meldung: 'Ein Turnier dieses Namens gibt es schon.',
            },
          }),
        ),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.einlesen }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    expect(screen.queryByRole('status')).toBeNull()
  })
})
