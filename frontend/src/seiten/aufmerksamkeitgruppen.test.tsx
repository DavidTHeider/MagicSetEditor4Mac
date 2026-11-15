// **Paketbefunde als eine Zeile** — AP-8
// (#2365),
// Befund B-14: 274 Paketbefunde beherrschten die Aufmerksamkeit, jeder mit
// seinem rohen Skriptsatz. Jetzt steht für alle Pakete eine Zeile da; sie
// öffnet die Paketliste gefiltert auf **genau diese** Pakete — die Namen
// kommen aus derselben Antwort wie die Zahl, damit beide zum selben Stand
// gehören (`PLAN-0.9.0.md`: Wert und Filterziel tragen dieselbe Epoche).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import type { Anschluss } from '../anschluss'
import type { Anschlussstelle } from '../anschlussraum'
import type { AufmerksamkeitAntwort, HinweisDTO } from '../apiaufmerksamkeit'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from '../test/api'
import { AufmerksamkeitProbe } from '../test/aufmerksamkeitprobe'

const texte = kataloge.de.seiten.start.aufmerksamkeit

const uebergeben: Anschluss[] = []
vi.mock('../anschlussraum', async (echt) => ({
  ...(await echt<typeof import('../anschlussraum')>()),
  useAnschluss: (): Anschlussstelle => ({
    offen: undefined,
    uebergib: (anschluss) => {
      uebergeben.push(anschluss)
    },
    verbrauche: () => undefined,
  }),
}))

const paket = (stelle: number): HinweisDTO => ({
  art: 'paketBefund',
  objekt: {
    art: 'paket',
    kennung: `paket-${String(stelle)}.mse-style`,
    name: `Paket ${String(stelle)}`,
    angeheftet: false,
  },
  anzahl: 2,
  grund: `style: Zeile ${String(stelle)}`,
  ort: '',
  schluessel: '',
  werte: [],
})

const deck: HinweisDTO = {
  art: 'deckFehlend',
  objekt: { art: 'deck', kennung: '3', name: 'Mono Rot', angeheftet: false },
  anzahl: 4,
  grund: '',
  ort: '',
  schluessel: '',
  werte: [],
}

function hinstellen(hinweise: HinweisDTO[]) {
  const antwort: AufmerksamkeitAntwort = {
    hinweise,
    durchgesehen: { sets: 0, decks: 1, pakete: 12 },
    befunde: [],
    fehler: null,
  }
  belegeApi({ Aufmerksamkeit: vi.fn().mockResolvedValue(antwort) })
  render(
    <Sprachraum sprache="de">
      <AufmerksamkeitProbe waehleSet={vi.fn()} stand={0} />
    </Sprachraum>,
  )
}

describe('die Pakete in der Aufmerksamkeit', () => {
  it('stehen als eine Zeile, die genau diese Pakete gefiltert öffnet', async () => {
    hinstellen([
      paket(0),
      deck,
      ...Array.from({ length: 11 }, (_, i) => paket(i + 1)),
    ])
    const bereich = await screen.findByRole('region', { name: texte.titel })

    expect(
      within(bereich).queryByRole('button', { name: /^Paket 3/ }),
    ).toBeNull()
    const sammel = within(bereich).getByRole('button', {
      name: /12 eigene Pakete mit Befund im Skript/,
    })
    // **Die Zahl zählt Objekte, nicht Zeilen** — dreizehn brauchen
    // Aufmerksamkeit, auch wenn zwölf davon eine Zeile teilen.
    expect(
      within(bereich).getByRole('heading', { name: `${texte.titel} 13` }),
    ).toBeInTheDocument()

    await userEvent.click(sammel)

    expect(uebergeben.at(-1)).toEqual({
      art: 'paketauswahl',
      name: '12 eigene Pakete mit Befund im Skript',
      namen: Array.from(
        { length: 12 },
        (_, i) => `paket-${String(i)}.mse-style`,
      ),
    })
  })

  // **Ihre Stelle ist die des ersten** — die Reihenfolge der Antwort bleibt.
  it('steht, wo das erste Paket stand', async () => {
    hinstellen([deck, paket(0), paket(1)])
    const bereich = await screen.findByRole('region', { name: texte.titel })

    const zeilen = within(bereich)
      .getAllByRole('listitem')
      .map((zeile) => zeile.textContent)

    expect(zeilen).toHaveLength(2)
    expect(zeilen[0]).toContain('Mono Rot')
    expect(zeilen[1]).toContain('2 eigene Pakete mit Befund im Skript')
  })
})
