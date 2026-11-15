// Vorlagen aus einer MSE2-Installation übernehmen (#55).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { type Api } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Uebernahme } from './Uebernahme'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function belegeUebernahme(teile: Partial<Api>) {
  return belegeApi({
    MSE2Vorschlaege: () => Promise.resolve({ orte: ['/heim/.wine/mse/data'] }),
    ...teile,
  })
}

function zeige(uebernommen = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Uebernahme uebernommen={uebernommen} />
    </Sprachraum>,
  )

  return uebernommen
}

describe('Übernahme aus MSE2', () => {
  // **Vorgeschlagen, nicht ungefragt gelesen** — die Zusage aus #55.
  it('schlägt die üblichen Orte vor, ohne sie zu lesen', async () => {
    const VorlagenUebernehmen = vi.fn()
    belegeUebernahme({ VorlagenUebernehmen })
    zeige()

    expect(await screen.findByText('/heim/.wine/mse/data')).toBeInTheDocument()

    // **Noch ist nichts geschehen.** Ein Vorschlag ist kein Auftrag.
    expect(VorlagenUebernehmen).not.toHaveBeenCalled()
  })

  it('erklärt, was zu tun ist, wenn kein Ort gefunden wurde', async () => {
    belegeUebernahme({ MSE2Vorschlaege: () => Promise.resolve({ orte: [] }) })
    zeige()

    expect(
      await screen.findByText(texte.seiten.pakete.mse2KeinVorschlag),
    ).toBeInTheDocument()
  })

  it('übernimmt aus einem Vorschlag und meldet, was dabei herauskam', async () => {
    const VorlagenUebernehmen = vi.fn(() =>
      Promise.resolve({
        uebernommen: [
          {
            art: 'mse-game',
            name: 'magic',
            anzeigename: 'magic',
            version: '2009-07-23',
            pfad: '/p/magic.mse-game',
            herkunft: '/heim/.wine/mse/data/magic.mse-game',
            pruefsumme: '',
            braucht: [],
            gruppe: '',
            rang: 1000000,
            artrang: 20,
            herkunftsart: 'mitgeliefert' as const,
            hauptdatei: 'style',
            geaendert: false,
            gebrauchtVon: [],
            benutzt: 0,
          },
        ],
        uebergangen: ['magic-m15.mse-style'],
        befunde: ['kaputt.mse-style: nicht lesbar'],
        abgebrochen: false,
      }),
    )

    belegeUebernahme({ VorlagenUebernehmen })
    const uebernommen = zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: 'Von hier übernehmen — /heim/.wine/mse/data',
      }),
    )

    expect(VorlagenUebernehmen).toHaveBeenCalledWith('/heim/.wine/mse/data')
    expect(
      await screen.findByText('1 Vorlagen übernommen.'),
    ).toBeInTheDocument()

    // **Übergangen und benannt**, nicht überschrieben.
    expect(screen.getByText('magic-m15.mse-style')).toBeInTheDocument()
    expect(
      screen.getByText('kaputt.mse-style: nicht lesbar'),
    ).toBeInTheDocument()

    await waitFor(() => {
      expect(uebernommen).toHaveBeenCalled()
    })
  })

  it('fragt nach einem Verzeichnis, wenn keiner der Vorschläge passt', async () => {
    const VorlagenUebernehmen = vi.fn(() =>
      Promise.resolve({
        uebernommen: [],
        uebergangen: [],
        befunde: [],
        abgebrochen: false,
      }),
    )

    belegeUebernahme({ VorlagenUebernehmen })
    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.pakete.mse2Waehlen,
      }),
    )

    // **Ein leerer Vorschlag heißt „frag nach".** Der Dialog gehört nach Go;
    // ein Fenster bekäme einen Namen und keinen Pfad.
    expect(VorlagenUebernehmen).toHaveBeenCalledWith('')
  })

  it('zeigt nichts, wenn jemand den Dialog abbricht', async () => {
    belegeUebernahme({
      VorlagenUebernehmen: () =>
        Promise.resolve({
          uebernommen: [],
          uebergangen: [],
          befunde: [],
          abgebrochen: true,
        }),
    })
    const uebernommen = zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.pakete.mse2Waehlen,
      }),
    )

    await waitFor(() => {
      expect(uebernommen).not.toHaveBeenCalled()
    })

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('meldet einen Fehler am Ort', async () => {
    belegeUebernahme({
      VorlagenUebernehmen: () =>
        Promise.resolve({
          uebernommen: [],
          uebergangen: [],
          befunde: [],
          abgebrochen: false,
          fehler: {
            code: 'validierung' as const,
            schluessel: 'fehler.validierung',
            meldung: 'Das ist kein Verzeichnis mit Vorlagen.',
          },
        }),
    })
    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.pakete.mse2Waehlen,
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Das ist kein Verzeichnis mit Vorlagen.',
    )
  })
})
