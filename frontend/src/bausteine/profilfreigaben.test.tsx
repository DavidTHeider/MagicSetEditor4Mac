// **Wer an diesem Profil was darf, im Fenster** — Werkstück 11 von Phase 4.12
// (#1351),
// Punkte 8 bis 10.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Profilfreigaben } from './Profilfreigaben'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { FreigabeDTO, NutzerDTO } from '../api'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.nutzer

// dieNutzer sind der Eigner und zwei andere.
const dieNutzer = [
  { id: '1', name: 'Ada' },
  { id: '2', name: 'Grace' },
  { id: '3', name: 'Alan' },
] as unknown as readonly NutzerDTO[]

// eineFreigabe ist die Vorlage, aus der die Fälle ihre Abweichung bauen.
function eineFreigabe(teile: Partial<FreigabeDTO> = {}): FreigabeDTO {
  return {
    profilID: '10',
    nutzerID: '2',
    name: 'Grace',
    rolle: 'mitarbeiter',
    angelegt: '2026-09-06T10:00:00Z',
    ...teile,
  }
}

function zeige(teile: Record<string, unknown> = {}) {
  belegeApi({
    Freigaben: vi.fn(() =>
      Promise.resolve({
        freigaben: [eineFreigabe()],
        sichtbarkeit: 'privat',
        eigeneRolle: 'verwalter',
      }),
    ),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Profilfreigaben profilID="10" nutzer={dieNutzer} eigner="1" />
    </Sprachraum>,
  )
}

describe('Die Freigaben eines Profils', () => {
  it('zeigt, wer welche Rolle hat', async () => {
    zeige()

    const zeile = (await screen.findByText('Grace')).closest('li')

    // **In der Zeile gesucht und nicht auf der Seite**: derselbe Rollenname
    // steht auch im Auswahlfeld darüber.
    expect(zeile).toHaveTextContent(eigene.rollennamen.mitarbeiter ?? '')
  })

  it('sagt es, wenn niemand etwas darf', async () => {
    zeige({
      Freigaben: vi.fn(() =>
        Promise.resolve({
          freigaben: [],
          sichtbarkeit: 'privat',
          eigeneRolle: 'verwalter',
        }),
      ),
    })

    expect(await screen.findByText(eigene.freigabenLeer)).toBeInTheDocument()
  })

  it('stellt weder den Eigner noch einen schon Freigegebenen zur Wahl', async () => {
    zeige()

    await screen.findByText('Grace')

    const wahl = screen.getByLabelText(eigene.freigabeNutzer)

    // **Alan bleibt übrig** — Ada ist der Eigner, Grace hat schon eine.
    const namen = [...wahl.querySelectorAll('option')].map(
      (eine) => eine.textContent,
    )

    expect(namen).toEqual([eigene.freigabeNiemand, 'Alan'])
  })

  it('vergibt eine Rolle und liest danach neu', async () => {
    const setzen = vi.fn(() =>
      Promise.resolve({
        freigaben: [],
        sichtbarkeit: 'privat',
        eigeneRolle: 'verwalter',
      }),
    )
    const lesen = vi.fn(() =>
      Promise.resolve({
        freigaben: [eineFreigabe()],
        sichtbarkeit: 'privat',
        eigeneRolle: 'verwalter',
      }),
    )

    zeige({ Freigaben: lesen, FreigabeSetzen: setzen })

    await screen.findByText('Grace')

    await userEvent.selectOptions(
      screen.getByLabelText(eigene.freigabeNutzer),
      '3',
    )
    await userEvent.selectOptions(
      screen.getByLabelText(eigene.freigabeRolle),
      'verwalter',
    )
    await userEvent.click(
      screen.getByRole('button', { name: eigene.freigabeVergeben }),
    )
    await antwortAbwarten()

    expect(setzen).toHaveBeenCalledWith({
      profilID: '10',
      nutzerID: '3',
      rolle: 'verwalter',
    })

    // **Regel 3: nach dem geglückten Schreiben wird nachgeladen.**
    expect(lesen).toHaveBeenCalledTimes(2)
  })

  it('zieht eine Freigabe zurück', async () => {
    const loeschen = vi.fn(() =>
      Promise.resolve({
        freigaben: [],
        sichtbarkeit: 'privat',
        eigeneRolle: 'verwalter',
      }),
    )

    zeige({ FreigabeLoeschen: loeschen })

    await screen.findByText('Grace')

    await userEvent.click(
      screen.getByRole('button', { name: eigene.freigabeZurueckziehen }),
    )
    await antwortAbwarten()

    expect(loeschen).toHaveBeenCalledWith('10', '2')
  })

  it('meldet, woran das Vergeben scheitert', async () => {
    zeige({
      FreigabeSetzen: vi.fn(() =>
        Promise.resolve({
          freigaben: [],
          sichtbarkeit: '',
          eigeneRolle: '',
          fehler: { code: 'erlaubnis', schluessel: 'erlaubnis' },
        }),
      ),
    })

    await screen.findByText('Grace')

    await userEvent.selectOptions(
      screen.getByLabelText(eigene.freigabeNutzer),
      '3',
    )
    await userEvent.click(
      screen.getByRole('button', { name: eigene.freigabeVergeben }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.erlaubnis,
    )
  })

  it('zeigt einem Nichtverwalter kein Formular, sondern den Grund', async () => {
    zeige({
      Freigaben: vi.fn(() =>
        Promise.resolve({
          freigaben: [eineFreigabe()],
          sichtbarkeit: 'bestand',
          eigeneRolle: 'mitarbeiter',
        }),
      ),
    })

    expect(
      await screen.findByText(eigene.freigabenNurLesen),
    ).toBeInTheDocument()

    expect(
      screen.queryByRole('button', { name: eigene.freigabeVergeben }),
    ).not.toBeInTheDocument()

    // **Und auch keinen Knopf zum Zurückziehen** — die Liste bleibt lesbar.
    expect(
      screen.queryByRole('button', { name: eigene.freigabeZurueckziehen }),
    ).not.toBeInTheDocument()
  })

  it('bietet kein Formular an, wenn niemand mehr übrig ist', async () => {
    zeige({
      Freigaben: vi.fn(() =>
        Promise.resolve({
          freigaben: [
            eineFreigabe(),
            eineFreigabe({ nutzerID: '3', name: 'Alan', rolle: 'leser' }),
          ],
          sichtbarkeit: 'privat',
          eigeneRolle: 'verwalter',
        }),
      ),
    })

    expect(await screen.findByText('Alan')).toBeInTheDocument()

    expect(
      screen.queryByLabelText(eigene.freigabeNutzer),
    ).not.toBeInTheDocument()
  })

  it('zeigt eine Rolle, die der Katalog nicht kennt, im Klartext', async () => {
    // **§16: Drift wird erkannt, nicht erlitten.** Ein neuerer Bestand mit
    // einer vierten Rolle darf hier keine leere Zelle erzeugen.
    zeige({
      Freigaben: vi.fn(() =>
        Promise.resolve({
          freigaben: [eineFreigabe({ rolle: 'kurator' })],
          sichtbarkeit: 'privat',
          eigeneRolle: 'verwalter',
        }),
      ),
    })

    const zeile = (await screen.findByText('Grace')).closest('li')

    expect(zeile).toHaveTextContent('kurator')
  })

  it('meldet, woran das Lesen scheitert', async () => {
    zeige({
      Freigaben: vi.fn(() =>
        Promise.resolve({
          freigaben: [],
          sichtbarkeit: '',
          eigeneRolle: '',
          fehler: { code: 'erlaubnis', schluessel: 'erlaubnis' },
        }),
      ),
    })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.erlaubnis,
    )
  })

  // **Ein Nachladefehler ist kein gescheiterter Schritt** (§7, Regel 3;
  // #1970): wer seine eigene Freigabe zurückzieht, darf die Liste danach
  // nicht mehr lesen. Die Seite nimmt den Schritt als geschehen, liest
  // selbst neu — und was sie dann nicht lesen darf, steht an der Liste.
  it('nimmt ein Zurückziehen mit Nachladefehler als geschehen', async () => {
    const erlaubnis = { code: 'erlaubnis' as const, schluessel: 'erlaubnis' }
    const lesen = vi
      .fn()
      .mockResolvedValueOnce({
        freigaben: [eineFreigabe()],
        sichtbarkeit: 'privat',
        eigeneRolle: 'verwalter',
      })
      .mockResolvedValue({
        freigaben: [],
        sichtbarkeit: '',
        eigeneRolle: '',
        fehler: erlaubnis,
      })
    const loeschen = vi.fn(() =>
      Promise.resolve({
        freigaben: [],
        sichtbarkeit: '',
        eigeneRolle: '',
        nachladefehler: { ...erlaubnis, meldung: 'nicht mehr lesbar' },
        fehler: null,
      }),
    )

    zeige({ Freigaben: lesen, FreigabeLoeschen: loeschen })

    await userEvent.click(
      await screen.findByRole('button', { name: eigene.freigabeZurueckziehen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.erlaubnis,
    )
    expect(loeschen).toHaveBeenCalledWith('10', '2')
    expect(lesen).toHaveBeenCalledTimes(2)
  })
})
