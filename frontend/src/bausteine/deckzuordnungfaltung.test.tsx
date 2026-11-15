// **Keine Formularwand im Inspector** — UXR-11
// (#1919).
//
// Die Lasche *Zuordnung* zeichnete zu jedem Druck des Decks ein Formular aus
// Exemplar, Menge und Knopf — bei einem Commanderdeck 76 Zeilen, 152 Felder
// und 76 Knöpfe in einer Spalte von 320 Punkten (R3, R4 und R6 in
// `docs/UIREVIEW.md`). Das Artboard *Decks* zeigt an dieser Stelle einen
// Satz — „53 von 60 liegen bei dir. 7 fehlen." — und den Weg zum Tausch.
// Jetzt steht der Satz oben, die Zeilen dahinter eingeklappt, und das
// Formular einer Zeile erst, wenn jemand an ihr zuordnen will.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Deckzuordnung } from './Deckzuordnung'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import {
  commanderdeck,
  deckexemplare,
  zuordnungszeilen,
} from '../test/deckbestand'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { zaehle } from '../test/bedienzaehlung'

import type { Api } from '../api'

const texte = kataloge.de.seiten.decks

function zeige(teile: Partial<Api> = {}) {
  belegeApi({
    DeckZuordnungen: () =>
      Promise.resolve({
        zeilen: zuordnungszeilen(commanderdeck),
        fehlend: 7,
        unbekannt: 0,
        fehler: null,
      }),
    Sammlung: () =>
      Promise.resolve({
        exemplare: deckexemplare(commanderdeck),
        fehler: null,
      }),
    ...teile,
  })

  const { container } = render(
    <Sprachraum sprache="de">
      <Deckzuordnung
        stand={
          {
            fuehre: async (_schluessel: string, was: () => Promise<void>) => {
              await was()
            },
            vorgang: { laeuft: () => false, starte: vi.fn() },
          } as never
        }
        deck={commanderdeck}
      />
    </Sprachraum>,
  )

  return container
}

// zeile gibt den Namen der Zeile an einer Stelle.
function zeile(stelle: number): string {
  return commanderdeck.plaetze[stelle]?.bezeichnung ?? ''
}

afterEach(() => {
  raeumeApi()
})

describe('die Zuordnung eines Commanderdecks', () => {
  it('zeigt oben, was fehlt, und kein einziges Feld', async () => {
    const kasten = zeige()

    expect(
      await screen.findByText(einsetzen(texte.bestandFehlt, { zahl: '7' })),
    ).toBeInTheDocument()
    expect(zaehle(kasten).felder).toBe(0)
    expect(
      screen.getByRole('button', { name: texte.zuordnungen }),
    ).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByRole('list', { name: texte.zuordnungen })).toBeNull()
  })

  it('klappt die Zeilen auf — und noch immer ohne Feld', async () => {
    const kasten = zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.zuordnungen }),
    )

    const liste = screen.getByRole('list', { name: texte.zuordnungen })
    expect(within(liste).getAllByRole('listitem')).toHaveLength(76)
    expect(zaehle(kasten).felder).toBe(0)
  })

  it('zeichnet das Formular einer Zeile erst auf Wunsch, und immer nur eines', async () => {
    const kasten = zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.zuordnungen }),
    )

    const vierzigste = screen.getByRole('button', {
      name: `${texte.zuordnungExemplarWaehlen}: ${zeile(40)}`,
    })
    await userEvent.click(vierzigste)

    expect(vierzigste).toHaveAttribute('aria-expanded', 'true')
    expect(zaehle(kasten).felder).toBe(2)

    await userEvent.click(
      screen.getByRole('button', {
        name: `${texte.zuordnungExemplarWaehlen}: ${zeile(41)}`,
      }),
    )

    expect(vierzigste).toHaveAttribute('aria-expanded', 'false')
    expect(zaehle(kasten).felder).toBe(2)

    await userEvent.click(
      screen.getByRole('button', {
        name: `${texte.zuordnungExemplarWaehlen}: ${zeile(41)}`,
      }),
    )
    expect(zaehle(kasten).felder).toBe(0)
  })

  it('ordnet über das aufgeklappte Formular zu', async () => {
    const DeckZuordnen = vi.fn().mockResolvedValue({
      zuordnung: { id: '1', platzID: '', exemplarID: '', menge: 1 },
      fehler: null,
    })

    zeige({ DeckZuordnen })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.zuordnungen }),
    )
    await userEvent.click(
      screen.getByRole('button', {
        name: `${texte.zuordnungExemplarWaehlen}: ${zeile(40)}`,
      }),
    )
    await userEvent.selectOptions(
      screen.getByLabelText(texte.zuordnungExemplar),
      String(1040),
    )
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.zuordnen}: ${zeile(40)}` }),
    )

    await waitFor(() => {
      expect(DeckZuordnen).toHaveBeenCalledWith(
        commanderdeck.id,
        commanderdeck.plaetze[40]?.id,
        '1040',
        1,
      )
    })
  })

  it('bietet an einer Zeile ohne passendes Exemplar nichts zum Aufklappen an', async () => {
    zeige({
      Sammlung: () => Promise.resolve({ exemplare: [], fehler: null }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.zuordnungen }),
    )

    expect(
      screen.queryByRole('button', {
        name: `${texte.zuordnungExemplarWaehlen}: ${zeile(40)}`,
      }),
    ).toBeNull()
  })
})
