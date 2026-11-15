// **Eigene Übersetzungen in den Einstellungen** — LOC-02
// (#1688).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Eigenuebersetzungen } from './Eigenuebersetzungen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { sprachantwort, sprachpaketprobe } from '../test/sprachprobe'

import type { Api } from '../api'

const texte = kataloge.de.seiten.einstellungen.eigenuebersetzung

afterEach(() => {
  raeumeApi()
})

function zeige(teile: Partial<Api> = {}, wert = '') {
  const setze = vi.fn()

  belegeApi({
    Sprachpakete: () =>
      Promise.resolve({ pakete: [sprachpaketprobe()], fehler: null }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Eigenuebersetzungen
        eintrag={{
          schluessel: 'sprache.eigene',
          wert,
          wirksam: wert,
          werte: [],
          bereich: 'anwendung',
        }}
        angehalten={false}
        setze={setze}
      />
    </Sprachraum>,
  )

  return setze
}

describe('Die eigenen Übersetzungen', () => {
  it('nennt jede mit ihren Zahlen und setzt die verwendete', async () => {
    const setze = zeige()

    expect(await screen.findByText('Português')).toBeInTheDocument()
    expect(
      screen.getByText('1 von 3 übersetzt · 0 veraltet'),
    ).toBeInTheDocument()

    await userEvent.selectOptions(
      screen.getByRole('combobox', { name: texte.verwenden }),
      'pt-BR',
    )

    expect(setze).toHaveBeenCalledWith('pt-BR')
  })

  it('sagt, was zu tun ist, wenn es noch keine gibt', async () => {
    zeige({ Sprachpakete: () => Promise.resolve({ pakete: [], fehler: null }) })

    expect(await screen.findByText(texte.leer)).toBeInTheDocument()
  })

  // **Angelegt heißt: der Editor geht auf** — über der Basis, mit allem als
  // fehlend.
  it('legt eine an und öffnet ihren Editor', async () => {
    const SprachpaketAnlegen = vi.fn(() =>
      Promise.resolve(
        sprachantwort({
          paket: sprachpaketprobe({ kennung: 'da', name: 'Dansk' }),
        }),
      ),
    )
    zeige({
      SprachpaketAnlegen,
      Sprachpaket: () =>
        Promise.resolve(
          sprachantwort({
            paket: sprachpaketprobe({ kennung: 'da', name: 'Dansk' }),
          }),
        ),
    })

    await userEvent.type(
      screen.getByRole('textbox', { name: texte.kennung }),
      'da',
    )
    await userEvent.type(
      screen.getByRole('textbox', { name: texte.name }),
      'Dansk',
    )
    await userEvent.click(
      screen.getByRole('button', { name: new RegExp(texte.anlegen) }),
    )

    expect(SprachpaketAnlegen).toHaveBeenCalledWith('da', 'Dansk', 'en')
    expect(
      await screen.findByRole('heading', { name: 'Dansk (da)' }),
    ).toBeInTheDocument()

    // **Zurück zur Liste** — und sie lädt nach.
    await userEvent.click(screen.getByRole('button', { name: texte.zurueck }))
    expect(
      await screen.findByRole('heading', { name: texte.titel }),
    ).toBeInTheDocument()
  })

  // **Der Befund steht am Feld** (§6).
  it('zeigt eine falsche Kennung an der Kennung', async () => {
    zeige({
      SprachpaketAnlegen: () =>
        Promise.resolve(
          sprachantwort({
            fehler: {
              code: 'validierung',
              schluessel: 'sprachkennungEingebaut',
              meldung: 'eingebaut',
              feld: 'kennung',
              werte: { kennung: 'ja' },
            },
          }),
        ),
    })

    await userEvent.type(
      screen.getByRole('textbox', { name: texte.kennung }),
      'ja',
    )
    await userEvent.type(screen.getByRole('textbox', { name: texte.name }), 'J')
    await userEvent.click(
      screen.getByRole('button', { name: new RegExp(texte.anlegen) }),
    )

    expect(
      await screen.findByText(/ist eine eingebaute Sprache/),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('textbox', { name: texte.kennung }),
    ).toHaveAttribute('aria-invalid', 'true')
  })

  it('liest ein, und ein weggeklickter Dialog ist kein Fehler', async () => {
    const SprachpaketEinlesenOeffnen = vi
      .fn()
      .mockResolvedValueOnce(sprachantwort({ abgebrochen: true }))
      .mockResolvedValueOnce(sprachantwort())
    zeige({ SprachpaketEinlesenOeffnen })

    await screen.findByText('Português')
    await userEvent.click(
      screen.getByRole('checkbox', { name: texte.ersetzen }),
    )
    await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))
    expect(SprachpaketEinlesenOeffnen).toHaveBeenCalledWith(true)
    expect(screen.queryByRole('status')).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))
    expect(await screen.findByRole('status')).toHaveTextContent(
      'Português eingelesen.',
    )
  })

  it('gibt aus und löscht erst nach der Rückfrage', async () => {
    const SprachpaketLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))
    zeige({
      SprachpaketAusgabeSpeichern: () =>
        Promise.resolve(sprachantwort({ pfad: '/tmp/pt.json' })),
      SprachpaketLoeschen,
    })

    const zeile = (await screen.findByText('Português')).closest(
      'li',
    ) as HTMLElement

    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.ausgeben }),
    )
    expect(await screen.findByRole('status')).toHaveTextContent('/tmp/pt.json')

    await userEvent.click(
      within(zeile).getByRole('button', { name: new RegExp(texte.loeschen) }),
    )
    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.abbrechen }),
    )
    expect(SprachpaketLoeschen).not.toHaveBeenCalled()

    await userEvent.click(
      within(zeile).getByRole('button', { name: new RegExp(texte.loeschen) }),
    )
    await userEvent.click(
      within(zeile).getByRole('button', { name: new RegExp(texte.loeschenJa) }),
    )
    expect(SprachpaketLoeschen).toHaveBeenCalledWith('pt-BR')
  })

  it('zeigt einen Fehler ohne Feld unter der Liste', async () => {
    zeige({
      SprachpaketAusgabeSpeichern: () =>
        Promise.resolve(
          sprachantwort({
            fehler: {
              code: 'validierung',
              schluessel: 'sprachpaketOhneLizenz',
              meldung: 'ohne',
              feld: 'lizenz',
              werte: {},
            },
          }),
        ),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.ausgeben }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(/Ohne Lizenz/)
  })

  it('öffnet eine vorhandene zum Bearbeiten und zeigt einen fremden Stand, wie er heißt', async () => {
    zeige({
      Sprachpakete: () =>
        Promise.resolve({
          pakete: [sprachpaketprobe({ stand: 'eigen' })],
          fehler: null,
        }),
      Sprachpaket: () => Promise.resolve(sprachantwort()),
    })

    const zeile = (await screen.findByText('Português')).closest(
      'li',
    ) as HTMLElement
    expect(within(zeile).getByText(/eigen$/)).toBeInTheDocument()

    await userEvent.click(
      within(zeile).getByRole('button', { name: new RegExp(texte.bearbeiten) }),
    )
    expect(
      await screen.findByRole('heading', { name: 'Português (pt-BR)' }),
    ).toBeInTheDocument()
  })

  it('legt über der gewählten Basis an und zeigt deren Befund', async () => {
    const SprachpaketAnlegen = vi.fn(() =>
      Promise.resolve(
        sprachantwort({
          fehler: {
            code: 'validierung',
            schluessel: 'validierung',
            meldung: 'Keine Sprache des Fensters.',
            feld: 'basis',
            werte: {},
          },
        }),
      ),
    )
    zeige({ SprachpaketAnlegen })

    await userEvent.type(
      screen.getByRole('textbox', { name: texte.kennung }),
      'da',
    )
    await userEvent.type(
      screen.getByRole('textbox', { name: texte.name }),
      'Dansk',
    )
    await userEvent.selectOptions(
      screen.getByRole('combobox', { name: texte.basis }),
      'fr',
    )
    await userEvent.click(
      screen.getByRole('button', { name: new RegExp(texte.anlegen) }),
    )

    expect(SprachpaketAnlegen).toHaveBeenCalledWith('da', 'Dansk', 'fr')
    expect(
      await screen.findByText(/Die Eingabe ist nicht gültig/),
    ).toBeInTheDocument()
  })

  // **Ein Fehler ohne Feld und ein abgewiesener Ruf** stehen unter der
  // Liste — nicht verschluckt.
  it.each([
    [
      'ohne Feld',
      () =>
        Promise.resolve(
          sprachantwort({
            fehler: {
              code: 'unbekannt',
              schluessel: 'unbekannt',
              meldung: 'x',
              werte: {},
            },
          }),
        ),
    ],
    ['abgewiesen', () => Promise.reject(new Error('die Brücke schweigt'))],
  ])('zeigt einen Fehler %s', async (_, SprachpaketAusgabeSpeichern) => {
    zeige({ SprachpaketAusgabeSpeichern })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.ausgeben }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})
