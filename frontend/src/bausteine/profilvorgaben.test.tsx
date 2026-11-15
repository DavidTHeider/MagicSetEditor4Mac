// **Vorgaben, die auf Zeilen zeigen — im Fenster** — Werkstück 15 von
// Phase 4.12
// (#1351),
// Punkte 4 bis 7.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Profilvorgaben } from './Profilvorgaben'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { VorgabenstandDTO } from '../api'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.nutzer

// vorgabename gibt die Beschriftung einer Art — wie die Komponente.
function vorgabename(art: string): string {
  return eigene.vorgabenamen[art] ?? art
}

// einStand ist die Vorlage, aus der die Fälle ihre Abweichung bauen.
function einStand(teile: Partial<VorgabenstandDTO> = {}): VorgabenstandDTO {
  return {
    art: 'spiel',
    mehrwertig: false,
    werte: [],
    moeglich: [{ wert: 'magic', beschriftung: 'Magic' }],
    verwaist: [],
    ...teile,
  }
}

function zeige(
  staende: VorgabenstandDTO[],
  teile: Record<string, unknown> = {},
) {
  belegeApi({
    Profilvorgaben: vi.fn(() => Promise.resolve({ vorgaben: staende })),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Profilvorgaben profilID="10" />
    </Sprachraum>,
  )
}

describe('Die Vorgaben eines Profils', () => {
  it('zeigt eine einwertige Art als Auswahl', async () => {
    zeige([einStand({ werte: ['magic'] })])

    const wahl = await screen.findByRole('combobox', {
      name: vorgabename('spiel'),
    })

    expect(wahl).toHaveValue('magic')

    // **„ohne Vorgabe" steht mit darin** — sonst ließe sich eine gesetzte
    // Vorgabe nicht wieder wegnehmen.
    expect(
      screen.getByRole('option', { name: eigene.vorgabenOhne }),
    ).toBeInTheDocument()
  })

  it('zeigt eine mehrwertige Art als Kästchen', async () => {
    // **Kästchen und keine Mehrfachauswahl**: ein `<select multiple>`
    // verlangt Strg-Klick und zeigt nicht, was gewählt ist.
    zeige([
      einStand({
        art: 'deck',
        mehrwertig: true,
        werte: ['1'],
        moeglich: [
          { wert: '1', beschriftung: 'Erstes' },
          { wert: '2', beschriftung: 'Zweites' },
        ],
      }),
    ])

    expect(
      await screen.findByRole('checkbox', { name: 'Erstes' }),
    ).toBeChecked()
    expect(screen.getByRole('checkbox', { name: 'Zweites' })).not.toBeChecked()
  })

  it('setzt eine einwertige Art', async () => {
    const setzen = vi.fn(() => Promise.resolve({ vorgaben: [einStand()] }))

    zeige([einStand()], { ProfilvorgabeSetzen: setzen })

    await userEvent.selectOptions(
      await screen.findByRole('combobox', {
        name: vorgabename('spiel'),
      }),
      'magic',
    )

    expect(setzen).toHaveBeenCalledWith({
      profilID: '10',
      art: 'spiel',
      werte: ['magic'],
    })
  })

  // **Gesetzt ist gesetzt** (§7, Regel 3; #1970): kam der Stand danach
  // nicht, liest die Seite selbst neu — und was sie dann nicht lesen kann,
  // steht da, ohne dass das Setzen als gescheitert gilt.
  it('nimmt ein Setzen mit Nachladefehler als geschehen', async () => {
    const lesen = vi
      .fn()
      .mockResolvedValueOnce({ vorgaben: [einStand()] })
      .mockResolvedValue({
        vorgaben: [],
        fehler: { code: 'validierung', schluessel: '', meldung: 'klemmt' },
      })
    const setzen = vi.fn(() =>
      Promise.resolve({
        vorgaben: [],
        nachladefehler: {
          code: 'validierung' as const,
          schluessel: '',
          meldung: 'klemmt',
        },
        fehler: null,
      }),
    )

    zeige([], { Profilvorgaben: lesen, ProfilvorgabeSetzen: setzen })

    await userEvent.selectOptions(
      await screen.findByRole('combobox', { name: vorgabename('spiel') }),
      'magic',
    )

    expect(await screen.findByRole('alert')).toHaveTextContent('klemmt')
    expect(lesen).toHaveBeenCalledTimes(2)
  })

  it('räumt eine Art ab, wenn „ohne Vorgabe" gewählt wird', async () => {
    const setzen = vi.fn(() => Promise.resolve({ vorgaben: [einStand()] }))

    zeige([einStand({ werte: ['magic'] })], { ProfilvorgabeSetzen: setzen })

    await userEvent.selectOptions(
      await screen.findByRole('combobox', {
        name: vorgabename('spiel'),
      }),
      '',
    )

    expect(setzen).toHaveBeenCalledWith({
      profilID: '10',
      art: 'spiel',
      werte: [],
    })
  })

  it('hängt eine mehrwertige Wahl hinten an', async () => {
    // **Hinten und nicht vorn**: die Reihenfolge ist bei den Präferenzdecks
    // die Aussage, und wer eines dazunimmt, meint es als das nächste.
    const setzen = vi.fn(() => Promise.resolve({ vorgaben: [einStand()] }))

    zeige(
      [
        einStand({
          art: 'deck',
          mehrwertig: true,
          werte: ['1'],
          moeglich: [
            { wert: '1', beschriftung: 'Erstes' },
            { wert: '2', beschriftung: 'Zweites' },
          ],
        }),
      ],
      { ProfilvorgabeSetzen: setzen },
    )

    await userEvent.click(
      await screen.findByRole('checkbox', { name: 'Zweites' }),
    )

    expect(setzen).toHaveBeenCalledWith({
      profilID: '10',
      art: 'deck',
      werte: ['1', '2'],
    })
  })

  it('nimmt eine mehrwertige Wahl wieder heraus', async () => {
    const setzen = vi.fn(() => Promise.resolve({ vorgaben: [einStand()] }))

    zeige(
      [
        einStand({
          art: 'deck',
          mehrwertig: true,
          werte: ['1', '2'],
          moeglich: [
            { wert: '1', beschriftung: 'Erstes' },
            { wert: '2', beschriftung: 'Zweites' },
          ],
        }),
      ],
      { ProfilvorgabeSetzen: setzen },
    )

    await userEvent.click(
      await screen.findByRole('checkbox', { name: 'Erstes' }),
    )

    expect(setzen).toHaveBeenCalledWith({
      profilID: '10',
      art: 'deck',
      werte: ['2'],
    })
  })

  it('erklärt eine Art, für die es nichts gibt', async () => {
    // **§6: ein leerer Zustand erklärt den nächsten Schritt** — eine leere
    // Auswahl sagte nur, dass etwas fehlt, nicht was.
    zeige([einStand({ moeglich: [] })])

    expect(await screen.findByText(eigene.vorgabenLeer)).toBeInTheDocument()

    expect(
      screen.queryByRole('combobox', { name: vorgabename('spiel') }),
    ).not.toBeInTheDocument()
  })

  it('sagt, wenn eine Vorgabe ins Leere zeigt', async () => {
    // **Sie steht weiter da und wird gemeldet** — wegzuräumen nähme jemandem
    // seine Wahl, während er nicht hinsieht (§8).
    zeige([
      einStand({
        art: 'deck',
        mehrwertig: true,
        werte: ['9'],
        moeglich: [],
        verwaist: ['9'],
      }),
    ])

    const hinweis = await screen.findByRole('status')

    expect(hinweis).toHaveTextContent(eigene.vorgabenVerwaist)
    expect(hinweis).toHaveTextContent('9')
  })

  it('nennt eine Art, für die es keinen Namen gibt, bei ihrem Schlüssel', async () => {
    // **Der Rückfall, den es geben muss.** Käme aus Go eine sechste Art,
    // stünde hier sonst nichts — und niemand sähe, dass es sie gibt.
    zeige([einStand({ art: 'kartenrücken' })])

    expect(
      await screen.findByRole('combobox', { name: 'kartenrücken' }),
    ).toBeInTheDocument()
  })

  it('meldet, was schiefgeht', async () => {
    zeige([einStand()], {
      ProfilvorgabeSetzen: vi.fn(() =>
        Promise.resolve({
          fehler: { code: 'konflikt', schluessel: 'konflikt' },
        }),
      ),
    })

    await userEvent.selectOptions(
      await screen.findByRole('combobox', {
        name: vorgabename('spiel'),
      }),
      'magic',
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.konflikt,
    )
  })

  it('meldet auch, was beim Lesen schiefgeht', async () => {
    belegeApi({
      Profilvorgaben: vi.fn(() =>
        Promise.resolve({
          vorgaben: [],
          fehler: {
            code: 'konflikt' as const,
            schluessel: 'konflikt',
            meldung: '',
          },
        }),
      ),
    })

    render(
      <Sprachraum sprache="de">
        <Profilvorgaben profilID="10" />
      </Sprachraum>,
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.konflikt,
    )
  })
})
