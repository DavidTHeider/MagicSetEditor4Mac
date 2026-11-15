// **Das Fach *Symbol*** — DOM-02
// (#1697):
// ein eingesetztes Symbol steht sofort im Editor, und beim Symbolfeld einer
// Karte gibt es keine Bibliothek.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { leererSetkopf } from '../test/setkopfprobe'
import {
  bibliothekAufklappen,
  krone,
  mitSymbolen,
  ohneMarken,
} from '../test/symbolbibliothek'
import { Symbolfach } from './Symbolfach'

const texte = kataloge.de.seiten.karten.symbolbibliothek

afterEach(() => {
  raeumeApi()
})

// setangaben ist die Antwort mit einem Symbolwert — oder ohne.
const setangaben = (symbol?: string) => ({
  felder: [],
  werte: symbol === undefined ? [] : [{ feld: 'symbol', wert: symbol }],
  kopf: leererSetkopf,
  fehler: null,
})

describe('Das Fach Symbol', () => {
  it('liest nach dem Einsetzen das Symbol des Sets neu', async () => {
    const angaben = vi
      .fn()
      .mockResolvedValueOnce(setangaben())
      .mockResolvedValue(setangaben('symbol1.mse-symbol'))
    const symbol = vi.fn().mockResolvedValue({
      teile: [],
      name: 'symbol1.mse-symbol',
      fehler: null,
    })
    belegeApi({
      Setangaben: angaben,
      Symbol: symbol,
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(krone)),
      BibliothekssymbolEinsetzen: vi
        .fn()
        .mockResolvedValue({ datei: 'symbol1.mse-symbol', fehler: null }),
      MarkenJeObjekt: ohneMarken,
    })

    render(
      <Sprachraum sprache="de">
        <Symbolfach setID="7" ziel={undefined} />
      </Sprachraum>,
    )
    // **Die Bibliothek steht eingeklappt** — seit UXR-07 (#1915).
    await bibliothekAufklappen()

    await userEvent.click(
      within(await screen.findByRole('listitem', { name: 'Krone' })).getByRole(
        'button',
        { name: texte.einsetzen },
      ),
    )

    await waitFor(() => {
      expect(symbol).toHaveBeenCalledWith('7', 'symbol1.mse-symbol')
    })
  })

  // **Der Editor führt, die Bibliothek steht eingeklappt darunter** — UXR-07
  // (#1915). Die Lasche maß 4,1 Bildschirme (R4 aus `docs/UIREVIEW.md`), und
  // die Bibliothek ist die seltenere Handlung. **Eingeklappt, nicht
  // entfernt**: ein Klick öffnet sie, und erst dann fragt sie den Bestand.
  it('legt die Bibliothek eingeklappt unter den Editor', async () => {
    const bibliothek = vi.fn().mockResolvedValue(mitSymbolen(krone))
    belegeApi({
      Setangaben: vi.fn().mockResolvedValue(setangaben()),
      Symbolbibliothek: bibliothek,
      MarkenJeObjekt: ohneMarken,
    })

    render(
      <Sprachraum sprache="de">
        <Symbolfach setID="7" ziel={undefined} />
      </Sprachraum>,
    )

    const knopf = await screen.findByRole('button', {
      name: texte.ueberschrift,
    })
    expect(knopf).toHaveAttribute('aria-expanded', 'false')
    expect(
      screen.queryByRole('button', { name: texte.aufnehmen }),
    ).not.toBeInTheDocument()
    expect(bibliothek).not.toHaveBeenCalled()

    await userEvent.click(knopf)

    expect(
      await screen.findByRole('listitem', { name: 'Krone' }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('region', { name: texte.ueberschrift }),
    ).toBeInTheDocument()

    // **Und der Bereich ohne eigene Überschrift hält WCAG 2.2 AA** — sein
    // Name kommt aus `aria-label`, der Knopf darüber trägt denselben.
    expect(await verstoesse(document.body)).toEqual([])
  })

  it('zeigt beim Symbolfeld einer Karte keine Bibliothek', async () => {
    const bibliothek = vi.fn()
    belegeApi({
      Setangaben: vi.fn().mockResolvedValue(setangaben()),
      Karte: vi.fn().mockResolvedValue({
        karte: { id: '3', werte: [] },
        fehler: null,
      }),
      Symbolbibliothek: bibliothek,
    })

    render(
      <Sprachraum sprache="de">
        <Symbolfach setID="7" ziel={{ karteID: '3', feld: 'symbol' }} />
      </Sprachraum>,
    )

    await screen.findByRole('heading', {
      name: kataloge.de.seiten.karten.symbol.ueberschrift,
    })
    expect(
      screen.queryByRole('heading', { name: texte.ueberschrift }),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: texte.ueberschrift }),
    ).not.toBeInTheDocument()
    expect(bibliothek).not.toHaveBeenCalled()
  })
})
