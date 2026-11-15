// **Eine geladene Zeichnung und die Punkttasten** —
// #890,
// Werkstück 3 von Phase X.
//
// # Was hier geprüft wird, und warum es diese Reihenfolge ist
//
// **Beides kam vom Zweitblick über Phase X.** Eine geladene Zeichnung ließ
// sich nicht bearbeiten, und die Tastaturhilfe versprach zwei Tasten, die die
// Seite nicht kannte. Der Schnitt folgt der Frage: in `symbol.test.tsx` steht,
// **wie gezeichnet wird**, hier, **was mit einer vorhandenen Zeichnung
// geschieht**.
import { render, screen, waitFor } from '@testing-library/react'
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

// dreieck ist die Antwort, die die Go-Seite auf einen Grundformknopf gibt.
//
// **Drei Punkte, und die Zahlen sind die echten** — `service.Grundform` gibt
// sie für ein Dreieck über der Mitte (0,5|0,5) mit halber Größe 0,4.
function dreieck() {
  return {
    teile: [
      {
        art: 'Form',
        name: 'Dreieck',
        kombination: 'overlap',
        punkte: [punkt(0.5, 0.1), punkt(0.9, 0.8), punkt(0.1, 0.8)],
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

function punkt(x: number, y: number) {
  return {
    x,
    y,
    griffDavorX: 0,
    griffDavorY: 0,
    griffDanachX: 0,
    griffDanachY: 0,
    kurveDanach: false,
    sperre: 'free',
  }
}

// bei liest eine Koordinate eines Punktes als Zahl.
function bei(name: string, achse: 'cx' | 'cy') {
  return Number(screen.getByRole('button', { name }).getAttribute(achse))
}

// ohneSymbol ist die Antwort eines Sets, das noch keines hat.
function ohneSymbol() {
  return { felder: [], werte: [], fehler: null }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Symbol setID="7" />
    </Sprachraum>,
  )
}

// **Eine geladene Zeichnung lässt sich bearbeiten** —
// #890.
//
// # Der Befund
//
// Die geladene Zeichnung lag in `vorhandenes.daten`, jede Änderung griff auf
// den weiterhin **leeren** Zustand `teile`. Wer ein vorhandenes Symbol öffnete
// und einen Punkt bewegte, sah nichts geschehen; ein Grundformknopf ersetzte
// die Zeichnung, statt sie zu ergänzen; „Teil löschen" löschte nichts.
//
// Angestrichen vom Zweitblick über Phase X. **Der Editor war für genau den
// Fall unbrauchbar, für den es ihn gibt** — ein Symbol, das schon da ist.
describe('Eine geladene Zeichnung', () => {
  function mitGeladenem() {
    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [{ feld: 'symbol', wert: 'symbol1.mse-symbol' }],
        kopf: leererSetkopf,
        fehler: null,
      }),
      Symbol: vi.fn().mockResolvedValue({
        teile: dreieck().teile,
        name: 'symbol1.mse-symbol',
        fehler: null,
      }),
      Grundform: vi.fn().mockResolvedValue({
        teile: dreieck().teile.map((teil) => ({ ...teil, name: 'Rechteck' })),
        name: '',
        fehler: null,
      }),
    })

    zeige()
  }

  it('lässt sich an ihren Punkten verschieben', async () => {
    mitGeladenem()

    const erster = await screen.findByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    erster.focus()
    await userEvent.keyboard('{Shift>}{ArrowRight}{/Shift}')

    await waitFor(() => {
      expect(bei('Punkt 1 von Dreieck', 'cx')).toBeCloseTo(192, 6)
    })
  })

  it('wird von einer Grundform ergänzt und nicht ersetzt', async () => {
    mitGeladenem()

    await screen.findByRole('button', { name: 'Punkt 1 von Dreieck' })

    await userEvent.click(screen.getByRole('button', { name: texte.rechteck }))

    // **Beide stehen da.**
    expect(
      await screen.findByRole('button', { name: 'Punkt 1 von Rechteck' }),
    ).toBeInTheDocument()

    expect(
      screen.getByRole('button', { name: 'Punkt 1 von Dreieck' }),
    ).toBeInTheDocument()
  })

  // **Ein Ladefehler steht als Meldung da und nicht als leere Zeichenfläche.**
  //
  // Eine beschädigte `.mse-symbol` sah aus wie „noch nichts gezeichnet" — und
  // wer das liest, zeichnet ein zweites Symbol über ein vorhandenes. Fremde
  // Daten sind der Normalfall (§8). Angestrichen vom Zweitblick über Phase X.
  it('meldet einen Ladefehler, statt leer auszusehen', async () => {
    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [{ feld: 'symbol', wert: 'symbol1.mse-symbol' }],
        kopf: leererSetkopf,
        fehler: null,
      }),
      Symbol: vi
        .fn()
        .mockRejectedValue(new Error('Die Zeichnung ist beschädigt.')),
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Die Zeichnung ist beschädigt.',
    )
  })

  it('lässt sich löschen', async () => {
    mitGeladenem()

    await screen.findByRole('button', { name: 'Punkt 1 von Dreieck' })

    await userEvent.click(
      screen.getByRole('button', { name: 'Dreieck löschen' }),
    )

    expect(await screen.findByText(texte.keineTeile)).toBeInTheDocument()
  })
})

// **Eingabetaste fügt einen Punkt ein, Rückschritt löscht ihn** —
// #890.
//
// # Warum dieser Lauf spät kam und trotzdem zählt
//
// Die Tastaturhilfe der Seite nannte beide Tasten, **bevor die Seite sie
// konnte** — und `service.PunktEinfuegen` stand gerechnet und geprüft da, ohne
// dass ein Aufrufer sie erreichte. Beides zusammen ist die Klasse, die §10 an
// erster Stelle nennt: eine Zusage ohne Deckung, und ein Paket ohne Aufrufer.
//
// **Gerechnet wird auf der Go-Seite**, auch hier: das Teilen einer Kurve
// braucht de Casteljau, und eine zweite Rechnung im Frontend liefe mit jener
// auseinander.
describe('Die Punkttasten', () => {
  it('fügt mit der Eingabetaste einen Punkt ein', async () => {
    const PunktEinfuegen = vi.fn().mockResolvedValue({
      teile: [
        {
          ...dreieck().teile[0],
          name: 'Dreieck',
          punkte: [
            punkt(0.5, 0.1),
            punkt(0.7, 0.45),
            punkt(0.9, 0.8),
            punkt(0.1, 0.8),
          ],
        },
      ],
      name: '',
      fehler: null,
    })

    belegeApi({
      Grundform: vi.fn().mockResolvedValue(dreieck()),
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
      PunktEinfuegen,
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))

    const erster = await screen.findByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    erster.focus()
    await userEvent.keyboard('{Enter}')

    await waitFor(() => {
      expect(PunktEinfuegen).toHaveBeenCalledWith(expect.anything(), {
        teil: 0,
        punkt: 0,
      })
    })

    // **Vier Punkte statt drei** — und der neue steht an zweiter Stelle.
    expect(
      await screen.findByRole('button', { name: 'Punkt 4 von Dreieck' }),
    ).toBeInTheDocument()
  })

  it('löscht mit dem Rückschritt', async () => {
    const PunktLoeschen = vi.fn().mockResolvedValue({
      teile: [
        {
          ...dreieck().teile[0],
          name: 'Dreieck',
          punkte: [punkt(0.9, 0.8), punkt(0.1, 0.8)],
        },
      ],
      name: '',
      fehler: null,
    })

    belegeApi({
      Grundform: vi.fn().mockResolvedValue(dreieck()),
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
      PunktLoeschen,
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))

    const erster = await screen.findByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    erster.focus()
    await userEvent.keyboard('{Backspace}')

    await waitFor(() => {
      expect(PunktLoeschen).toHaveBeenCalled()
    })

    await waitFor(() => {
      expect(
        screen.queryByRole('button', { name: 'Punkt 3 von Dreieck' }),
      ).not.toBeInTheDocument()
    })
  })

  // **Und ein Löschen, das zu wenige Punkte übrig ließe, räumt nichts.**
  //
  // §7, Regel 3: was dasteht, bleibt stehen — die Meldung sagt, warum.
  it('behält die Zeichnung, wenn das Löschen abgewiesen wird', async () => {
    belegeApi({
      Grundform: vi.fn().mockResolvedValue(dreieck()),
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
      PunktLoeschen: vi
        .fn()
        .mockRejectedValue(new Error('Eine Form braucht mindestens zwei.')),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))

    const erster = await screen.findByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    erster.focus()
    await userEvent.keyboard('{Backspace}')

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Eine Form braucht mindestens zwei.',
    )

    // **Die drei Punkte stehen noch.**
    expect(
      screen.getByRole('button', { name: 'Punkt 3 von Dreieck' }),
    ).toBeInTheDocument()
  })
})
