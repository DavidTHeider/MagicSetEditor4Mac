// **Bedarf und Zuordnung im Fenster** — Werkstück U9 von Phase 5.1
// (#1420).
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Deckzuordnung } from './Deckzuordnung'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { deckprobe, platzprobe } from '../seiten/deckprobe'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { keineHerkunft } from '../apiherkunft'

import type { ExemplarDTO } from '../api'
import type { ZuordnungszeileDTO } from '../apidecks'

const texte = kataloge.de.seiten.decks

function exemplar(teil: Partial<ExemplarDTO> = {}): ExemplarDTO {
  return {
    id: 1,
    fassung: 1,
    printing: 'a1b2c3d4-0000-4000-8000-00000000000a',
    druckID: '4',
    vorlaeufig: false,
    menge: 2,
    zustand: 'nearMint',
    sprache: 'en',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: 'Blitz',
    setcode: 'm11',
    nummer: '149',
    lagerort: '',
    wunsch: false,
    bezeichnung: 'Blitz · M11 #149',
    herkunft: keineHerkunft,
    ...teil,
  }
}

function zeile(teil: Partial<ZuordnungszeileDTO> = {}): ZuordnungszeileDTO {
  return {
    printing: 'a1b2c3d4-0000-4000-8000-00000000000a',
    druckID: '4',
    vorlaeufig: false,
    bezeichnung: 'Blitz · M11 #149',
    gebraucht: 4,
    vorhanden: 2,
    fehlend: 2,
    grund: '',
    reserviert: 0,
    zugeordnet: 0,
    frei: 2,
    zuordnungen: [],
    ...teil,
  }
}

const deck = deckprobe({ plaetze: [platzprobe({ druckID: '4' })] })

function zeige(teile: Parameters<typeof belegeApi>[0] = {}) {
  const fuehre = vi.fn(
    async (_schluessel: string, was: () => Promise<void>) => {
      await was()
    },
  )

  belegeApi({
    DeckZuordnungen: () =>
      Promise.resolve({
        zeilen: [zeile()],
        fehlend: 2,
        unbekannt: 0,
        fehler: null,
      }),
    Sammlung: () => Promise.resolve({ exemplare: [exemplar()], fehler: null }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Deckzuordnung
        stand={
          {
            fuehre,
            vorgang: { laeuft: () => false, starte: vi.fn() },
          } as never
        }
        deck={deck}
      />
    </Sprachraum>,
  )

  return fuehre
}

afterEach(() => {
  raeumeApi()
})

// aufklappen öffnet die Zeilen von Bedarf und Zuordnung.
//
// **Seit UXR-11** (#1919)
// steht oben der Satz, was fehlt, und die Zeilen dahinter eingeklappt — das
// Formular einer Zeile erst auf Wunsch (`deckzuordnungfaltung.test.tsx`).
async function aufklappen() {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.zuordnungen }),
  )
}

describe('Bedarf und Zuordnung', () => {
  it('zeigt die vier Zahlen nebeneinander — und dass etwas fehlt', async () => {
    zeige()

    expect(
      await screen.findByText(einsetzen(texte.bestandFehlt, { zahl: '2' })),
    ).toBeInTheDocument()

    await aufklappen()
    expect(
      screen.getByText(
        einsetzen(texte.zuordnungszahlen, {
          vorhanden: '2',
          gebraucht: '4',
          frei: '2',
          zugeordnet: '0',
          reserviert: '0',
        }),
      ),
    ).toBeInTheDocument()
  })

  it('ordnet ein gewähltes Exemplar zu und lädt danach neu', async () => {
    const zuordnen = vi.fn().mockResolvedValue({
      zuordnung: { id: '1', platzID: '3', exemplarID: '1', menge: 1 },
      fehler: null,
    })

    zeige({ DeckZuordnen: zuordnen })
    await aufklappen()
    await userEvent.click(
      screen.getByRole('button', {
        name: `${texte.zuordnungExemplarWaehlen}: Blitz · M11 #149`,
      }),
    )

    const wahl = await screen.findByLabelText(texte.zuordnungExemplar)
    await userEvent.selectOptions(wahl, '1')

    await userEvent.click(
      screen.getByRole('button', {
        name: `${texte.zuordnen}: Blitz · M11 #149`,
      }),
    )

    await waitFor(() => {
      expect(zuordnen).toHaveBeenCalledWith(deck.id, '3', '1', 1)
    })
  })

  it('löst eine bestehende Zuordnung', async () => {
    const loesen = vi.fn().mockResolvedValue({ fehler: null })

    zeige({
      DeckZuordnungen: () =>
        Promise.resolve({
          zeilen: [
            zeile({
              zugeordnet: 2,
              frei: 0,
              zuordnungen: [
                {
                  id: '9',
                  platzID: '3',
                  exemplarID: '1',
                  menge: 2,
                  bezeichnung: 'Blitz · M11 #149',
                  abweichend: false,
                },
              ],
            }),
          ],
          fehlend: 2,
          unbekannt: 0,
          fehler: null,
        }),
      DeckZuordnungLoesen: loesen,
    })
    await aufklappen()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.zuordnungLoesen}: Blitz · M11 #149`,
      }),
    )

    await waitFor(() => {
      expect(loesen).toHaveBeenCalledWith(deck.id, '9')
    })
  })

  it('markiert eine Zuordnung auf einen anderen Druck', async () => {
    zeige({
      DeckZuordnungen: () =>
        Promise.resolve({
          zeilen: [
            zeile({
              zuordnungen: [
                {
                  id: '9',
                  platzID: '3',
                  exemplarID: '1',
                  menge: 1,
                  bezeichnung: 'Blitz · LEA #161',
                  abweichend: true,
                },
              ],
            }),
          ],
          fehlend: 2,
          unbekannt: 0,
          fehler: null,
        }),
    })

    await aufklappen()

    expect(await screen.findByText(texte.abweichend)).toHaveAttribute(
      'title',
      texte.abweichendErklaerung,
    )
  })

  it('holt die Fehlkarten als Liste — und sagt, wenn nichts fehlt', async () => {
    zeige({
      DeckFehlkarten: vi
        .fn()
        .mockResolvedValueOnce({ liste: '2 Blitz (M11) 149\n', fehler: null })
        .mockResolvedValueOnce({ liste: '', fehler: null }),
    })

    const knopf = await screen.findByRole('button', { name: texte.fehlkarten })

    await userEvent.click(knopf)

    const liste = await screen.findByLabelText(texte.fehlkartenliste)
    expect(liste).toHaveTextContent('2 Blitz (M11) 149')
    // **Sie rollt, also erreicht sie die Tastatur** (G10).
    expect(liste).toHaveAttribute('tabindex', '0')

    await userEvent.click(knopf)

    await waitFor(() => {
      expect(
        within(screen.getByLabelText(texte.fehlkartenliste)).getByText(
          texte.fehlkartenKeine,
        ),
      ).toBeInTheDocument()
    })
  })

  it('sagt es, wenn das Deck leer ist', async () => {
    zeige({
      DeckZuordnungen: () =>
        Promise.resolve({ zeilen: [], fehlend: 0, unbekannt: 0, fehler: null }),
    })

    expect(await screen.findByText(texte.zuordnungenLeer)).toBeInTheDocument()
    expect(screen.getByText(texte.bestandVollstaendig)).toBeInTheDocument()
  })

  it('meldet, wenn der Stand nicht zu lesen ist', async () => {
    zeige({
      DeckZuordnungen: () =>
        Promise.resolve({
          zeilen: [],
          fehlend: 0,
          unbekannt: 0,
          fehler: {
            code: 'unbekannt' as const,
            schluessel: '',
            meldung: 'kein Weg',
          },
        }),
    })

    expect(await screen.findByRole('alert')).toHaveTextContent('kein Weg')
  })
})
