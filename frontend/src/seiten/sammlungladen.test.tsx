// **Die Sammlung sagt, dass sie lädt, und bietet nach einem Lesefehler den
// nächsten Versuch an** — G11
// (#2277).
//
// B03 verzögerte `Sammlungsfenster` um vier Sekunden: die Mitte der Seite
// blieb leer — kein Leerzustand, denn noch lud sie, und kein Hinweis, dass
// sie lud. Scheiterte das Lesen, stand die Meldung da, und um es noch einmal
// zu versuchen, blieb nur, die Seite zu verlassen.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Sammlungsfensterantwort } from '../antwortensammlung'
import type { MarkenJeObjektAntwort } from '../apiobjektmarken'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Sammlung } from './Sammlung'

const zustaende = kataloge.de.bausteine.zustaende
const eigene = kataloge.de.seiten.sammlung

const leeresFenster: Sammlungsfensterantwort = {
  exemplare: [],
  gesamt: 0,
  insgesamt: 0,
  stueck: 0,
  fehler: null,
}

const fehlschlag = {
  code: 'unbekannt' as const,
  schluessel: '',
  meldung: 'die Sammlung schweigt',
}

function zeige(
  fenster: () => Promise<Sammlungsfensterantwort>,
  marken: () => Promise<MarkenJeObjektAntwort> = () =>
    Promise.resolve({ marken: {}, fehler: null }),
) {
  belegeApi({
    Sammlung: () => Promise.resolve({ exemplare: [], fehler: null }),
    Sammlungsfenster: fenster,
    Portfolio: () =>
      Promise.resolve({ zeilen: [], ohnePreis: 0, fehler: null }),
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
    MarkenJeObjekt: marken,
  })

  render(
    <Sprachraum sprache="de">
      <Sammlung bestandGeaendert={vi.fn()} />
    </Sprachraum>,
  )
}

afterEach(raeumeApi)

describe('Die Sammlung beim Öffnen', () => {
  it('sagt, dass sie lädt, und behauptet solange keine leere Sammlung', async () => {
    zeige(() => new Promise(() => undefined))

    expect(await screen.findByRole('status')).toHaveTextContent(zustaende.laedt)
    expect(screen.queryByText(eigene.leerTitel)).toBeNull()
  })

  // **Ein Lesefehler belegt keine leere Sammlung** — Befund 5 des
  // Zweitblicks über Phase G: neben der Meldung stand „Noch nichts
  // eingetragen" samt Handlungen, eine Aussage über Daten, die nie kamen.
  it('behauptet nach einem ersten Lesefehler keine leere Sammlung', async () => {
    zeige(() => Promise.resolve({ ...leeresFenster, fehler: fehlschlag }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'die Sammlung schweigt',
    )
    expect(screen.queryByText(eigene.leerTitel)).toBeNull()
  })

  it('bietet nach einem Lesefehler den nächsten Versuch an', async () => {
    let gefragt = 0
    zeige(() => {
      gefragt++

      return Promise.resolve(
        gefragt === 1
          ? { ...leeresFenster, fehler: fehlschlag }
          : leeresFenster,
      )
    })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'die Sammlung schweigt',
    )

    await userEvent.click(
      screen.getByRole('button', { name: zustaende.wiederholen }),
    )

    expect(await screen.findByText(eigene.leerTitel)).toBeVisible()
    expect(screen.queryByRole('alert')).toBeNull()
  })

  // **Und die Marken** — ihr Lesefehler steht hinter dem der Sammlung, und
  // nachgeladen wird, was scheiterte.
  it('lädt nach einem Lesefehler der Marken die Marken erneut', async () => {
    let gefragt = 0
    zeige(
      () => Promise.resolve(leeresFenster),
      () => {
        gefragt++

        return Promise.resolve(
          gefragt === 1
            ? {
                marken: {},
                fehler: { ...fehlschlag, meldung: 'die Marken schweigen' },
              }
            : { marken: {}, fehler: null },
        )
      },
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'die Marken schweigen',
    )

    await userEvent.click(
      screen.getByRole('button', { name: zustaende.wiederholen }),
    )

    await waitFor(() => {
      expect(screen.queryByRole('alert')).toBeNull()
    })
    expect(gefragt).toBe(2)
  })
})
