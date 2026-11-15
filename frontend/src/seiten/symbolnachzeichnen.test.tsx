// **Ein Bild nachzeichnen — im Fenster** —
// #1340,
// Werkstück F von Phase 4.11.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { leererSetkopf } from '../test/setkopfprobe'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Symbol } from './Symbol'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.symbol

function punkt(x: number, y: number) {
  return {
    x,
    y,
    griffDavorX: 0,
    griffDavorY: 0,
    griffDanachX: 0,
    griffDanachY: 0,
    kurveDanach: true,
    sperre: 'direction',
  }
}

function nachgezeichnet() {
  return {
    teile: [
      {
        art: 'Form',
        name: 'Form 1',
        kombination: 'merge',
        punkte: [punkt(0.5, 0.15), punkt(0.85, 0.5), punkt(0.5, 0.85)],
        spiegelart: '',
        kopien: 0,
        mitteX: 0,
        mitteY: 0,
        griffX: 0,
        griffY: 0,
        teile: [],
      },
    ],
    name: '',
    fehler: null,
  }
}

function zeige(belegung: Record<string, unknown>) {
  belegeApi({
    Setangaben: vi.fn().mockResolvedValue({
      felder: [],
      werte: [],
      kopf: leererSetkopf,
      fehler: null,
    }),
    ...belegung,
  })

  render(
    <Sprachraum sprache="de">
      <Symbol setID="7" />
    </Sprachraum>,
  )
}

describe('Bild nachzeichnen', () => {
  it('hängt die nachgezeichneten Formen an die Zeichnung', async () => {
    const SymbolNachzeichnen = vi.fn().mockResolvedValue(nachgezeichnet())

    zeige({ SymbolNachzeichnen })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.nachzeichnen }),
    )

    expect(SymbolNachzeichnen).toHaveBeenCalledTimes(1)

    expect(
      await screen.findByRole('button', { name: 'Punkt 1 von Form 1' }),
    ).toBeInTheDocument()
  })

  // **Und der Handgriff landet im Verlauf.**
  //
  // Ohne das wäre ein versehentlich gewähltes Bild eine Zeichnung, die man von
  // Hand wieder auseinandernehmen müsste.
  it('lässt sich zurücknehmen', async () => {
    zeige({ SymbolNachzeichnen: vi.fn().mockResolvedValue(nachgezeichnet()) })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.nachzeichnen }),
    )

    expect(
      await screen.findByRole('button', { name: 'Punkt 1 von Form 1' }),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.zuruecknehmen }),
    )

    expect(
      screen.queryByRole('button', { name: 'Punkt 1 von Form 1' }),
    ).not.toBeInTheDocument()
  })

  // **Ein Abbruch im Dateidialog ist kein Handgriff.**
  //
  // Er gibt eine leere Liste ohne Fehler. Ihn einzutragen hieße, einen leeren
  // Schritt in den Verlauf zu legen — und ⌘Z täte dann nichts Sichtbares.
  it('trägt einen Abbruch nicht in den Verlauf ein', async () => {
    zeige({
      SymbolNachzeichnen: vi.fn().mockResolvedValue({
        teile: [],
        name: '',
        fehler: null,
      }),
    })

    const knopf = await screen.findByRole('button', {
      name: texte.nachzeichnen,
    })

    await userEvent.click(knopf)

    expect(
      screen.getByRole('button', { name: texte.zuruecknehmen }),
    ).toBeDisabled()
  })

  // **Ein Fehler steht am Editor und nicht in der Konsole** (§6).
  it('zeigt einen Fehler an', async () => {
    zeige({
      SymbolNachzeichnen: vi.fn().mockResolvedValue({
        teile: [],
        name: '',
        fehler: {
          code: 'validierung',
          schluessel: 'fehler.unbekannt',
          meldung: 'kein lesbares Bild',
        },
      }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.nachzeichnen }),
    )

    expect(await screen.findByText(/kein lesbares Bild/)).toBeInTheDocument()
  })
})
