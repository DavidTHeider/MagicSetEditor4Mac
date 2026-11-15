// **Was dem Menschen gehört, auf der Einstellungsseite** — Werkstück 8 von
// Phase 4.12
// (#1351),
// Punkte 9, 10 und 11.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Nutzereinstellungsraum } from '../nutzereinstellungsraum'
import { Persoenliches } from './Persoenliches'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'

import type { NutzereinstellungDTO } from '../api'
import { probeprofil } from '../test/profil'

afterEach(() => {
  raeumeApi()
  document.documentElement.removeAttribute('data-bewegung')
  document.documentElement.removeAttribute('data-kontrast')
})

const eintraege: NutzereinstellungDTO[] = [
  {
    schluessel: 'barrierefreiheit.bewegung',
    wert: 'system',
    werte: ['system', 'reduziert'],
  },
  {
    schluessel: 'barrierefreiheit.kontrast',
    wert: 'system',
    werte: ['system', 'hoch'],
  },
  {
    schluessel: 'sicherheit.faktorbeimwechsel',
    wert: 'verlangen',
    werte: ['verlangen', 'nurAnmelden'],
  },
]

const gezeigtesProfil = () =>
  Promise.resolve({
    profil: probeprofil({
      id: '10',
      nutzerID: '1',
      name: 'Mein Profil',
      fassung: '1',
    }),
  })

function zeige(teile: Parameters<typeof belegeApi>[0] = {}) {
  belegeApi({
    GezeigtesProfil: vi.fn(gezeigtesProfil),
    Nutzereinstellungen: vi.fn(() =>
      Promise.resolve({ einstellungen: eintraege }),
    ),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Nutzereinstellungsraum>
        <Persoenliches />
      </Nutzereinstellungsraum>
    </Sprachraum>,
  )
}

describe('Mein Zugang', () => {
  it('zeigt jede Einstellung mit ihren Werten', async () => {
    zeige()

    const wahl = await screen.findByLabelText(
      kataloge.de.seiten.einstellungen.namen['barrierefreiheit.bewegung'],
    )

    // **Zwei Werte, die kein Paar bilden, sind zwei Segmente** — UXR-F2
    // (#2008).
    expect(within(wahl).getAllByRole('radio')).toHaveLength(2)
    expect(
      within(wahl).getByRole('radio', {
        name: kataloge.de.seiten.einstellungen.werte[
          'barrierefreiheit.bewegung.system'
        ],
      }),
    ).toBeChecked()
  })

  it('schreibt eine Änderung und liest neu', async () => {
    let gesetzt = false

    // **Der zweite Ruf gibt den neuen Stand** — so wie der Bestand es täte.
    const stand = () => ({
      einstellungen: gesetzt
        ? eintraege.map((eine) =>
            eine.schluessel === 'barrierefreiheit.kontrast'
              ? { ...eine, wert: 'hoch' }
              : eine,
          )
        : eintraege,
    })

    zeige({
      NutzereinstellungSetzen: vi.fn(() => {
        gesetzt = true

        return Promise.resolve(stand())
      }),
      Nutzereinstellungen: vi.fn(() => Promise.resolve(stand())),
    })

    const wahl = await screen.findByLabelText(
      kataloge.de.seiten.einstellungen.namen['barrierefreiheit.kontrast'],
    )

    const hoch =
      kataloge.de.seiten.einstellungen.werte['barrierefreiheit.kontrast.hoch']

    await userEvent.click(within(wahl).getByRole('radio', { name: hoch }))

    // **Gewartet wird auf die Wirkung** (§3): der neue Stand steht erst da,
    // wenn nachgeladen ist.
    await waitFor(() => {
      expect(within(wahl).getByRole('radio', { name: hoch })).toBeChecked()
    })
  })

  it('sichert die Auskunft über den Dialog', async () => {
    const sichern = vi.fn(() => Promise.resolve({}))

    zeige({ PersonendatenSichern: sichern })

    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.seiten.nutzer.auskunftSichern,
      }),
    )

    expect(sichern).toHaveBeenCalledWith('1')
  })

  it('meldet, was beim Sichern schiefgeht', async () => {
    zeige({
      PersonendatenSichern: vi.fn(() =>
        Promise.resolve({
          fehler: {
            code: 'validierung' as const,
            schluessel: 'validierung',
            meldung: 'Kein Platz mehr.',
          },
        }),
      ),
    })

    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.seiten.nutzer.auskunftSichern,
      }),
    )

    // **Der Schlüssel gewinnt über den Satz aus Go** — die Meldung kommt aus
    // dem Katalog, und genau das ist die Zusage der Fehlerübersetzung.
    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.validierung,
    )
  })

  // **Ein Ladefehler steht über der Liste und macht die Seite nicht leer.**
  it('meldet einen Ladefehler', async () => {
    zeige({
      Nutzereinstellungen: vi.fn(() =>
        Promise.resolve({
          einstellungen: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'unbekannt',
            meldung: 'Die Platte schweigt.',
          },
        }),
      ),
    })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.unbekannt,
    )
  })

  // **Ohne Nutzer gibt es nichts zu sichern** — und der Knopf sagt es.
  //
  // **Der Beleg, dass jemand zuhört, steht daneben**: die Erklärung ist da,
  // die Seite ist also gezeichnet. Ohne ihn prüfte diese Verneinung nur, dass
  // noch gar nichts dasteht.
  it('hält den Knopf an, solange kein Nutzer feststeht', async () => {
    belegeApi({
      GezeigtesProfil: vi.fn(() => Promise.resolve({ profil: null })),
      Nutzereinstellungen: vi.fn(() =>
        Promise.resolve({ einstellungen: eintraege }),
      ),
    })

    render(
      <Sprachraum sprache="de">
        <Nutzereinstellungsraum>
          <Persoenliches />
        </Nutzereinstellungsraum>
      </Sprachraum>,
    )

    await screen.findByText(kataloge.de.seiten.nutzer.auskunft)

    expect(
      screen.getByRole('button', {
        name: kataloge.de.seiten.nutzer.auskunftSichern,
      }),
    ).toBeDisabled()
  })
})
