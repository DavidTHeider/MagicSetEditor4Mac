// **Die Liste der Sammlung in drei Ansichten** — Werkstück U8 von Phase 5.1
// (#1419): Tabelle, Gruppen mit Köpfen und Summen, Galerie mit Kacheln — alle
// über dieselbe Fensterliste, alle mit dem Haken für die Massenvorgänge und
// der Wahl für den Inspector.
import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { keineHerkunft } from '../apiherkunft'
import type { ExemplarDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { leereSicht, type Sicht } from './sammlungssicht'
import { Sammlungsliste } from './Sammlungsliste'

const texte = kataloge.de.seiten.sammlung

// element gibt einen Knoten als HTMLElement — oder scheitert laut.
function element(knoten: Element | null | undefined): HTMLElement {
  if (!(knoten instanceof HTMLElement)) {
    throw new Error('kein Element')
  }

  return knoten
}

function exemplar(teil: Partial<ExemplarDTO> = {}): ExemplarDTO {
  return {
    id: 1,
    fassung: 1,
    printing: 'probe-druck',
    druckID: '',
    vorlaeufig: false,
    menge: 3,
    zustand: 'nearMint',
    sprache: 'en',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: '',
    setcode: 'm15',
    nummer: '',
    lagerort: 'Kiste A',
    wunsch: false,
    bezeichnung: 'probe-druck',
    herkunft: keineHerkunft,
    ...teil,
  }
}

const bestand = [
  exemplar({ id: 1, bezeichnung: 'Erste', name: 'Erste', menge: 2 }),
  exemplar({
    id: 2,
    bezeichnung: 'Zweite',
    setcode: 'khm',
    lagerort: '',
    finish: 'foil',
  }),
  exemplar({ id: 3, bezeichnung: 'Dritte', menge: 4 }),
  exemplar({ id: 4, bezeichnung: 'Vierte' }),
  exemplar({ id: 5, bezeichnung: 'Fünfte' }),
]

beforeEach(() => {
  Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
    configurable: true,
    get: () => 600,
  })
})

function zeige(teil: Partial<Sicht> = {}, gehalten?: number) {
  const anhaken = vi.fn()
  const waehlen = vi.fn()
  render(
    <Sprachraum sprache="de">
      <Sammlungsliste
        exemplare={bestand}
        sicht={{ ...leereSicht, ...teil }}
        gewaehlt={new Set([2])}
        anhaken={anhaken}
        gehalten={gehalten}
        waehlen={waehlen}
      />
    </Sprachraum>,
  )

  return { anhaken, waehlen }
}

describe('Die Tabelle', () => {
  it('zeigt je Exemplar Bezeichnung, Menge, Zustand, Sprache, Set und Lagerort', () => {
    zeige()

    const gitter = screen.getByRole('grid', { name: texte.listeBezeichnung })
    const zeile = within(gitter).getByText('Erste').closest('[role="row"]')
    expect(zeile).not.toBeNull()
    expect(zeile).toHaveTextContent('2 × Near Mint')
    expect(zeile).toHaveTextContent('en · Nonfoil')
    expect(zeile).toHaveTextContent('m15')
    expect(zeile).toHaveTextContent('Kiste A')
    // Die Kennung steht nicht in der Zeile — sie gehört dem Inspector.
    expect(within(gitter).queryByText('probe-druck')).not.toBeInTheDocument()
  })

  it('hakt an und wählt', async () => {
    const { anhaken, waehlen } = zeige()

    const haken = screen.getByRole('checkbox', { name: 'Dritte' })
    expect(screen.getByRole('checkbox', { name: 'Zweite' })).toBeChecked()
    await userEvent.click(haken)
    expect(anhaken).toHaveBeenCalledWith(3, true)

    await userEvent.click(screen.getByText('Vierte'))
    expect(waehlen).toHaveBeenCalledWith(4)
  })

  // **Schmal stehen Sprache, Set und Lagerort unter der Bezeichnung** —
  // AP-12 (#2369), B-28. Zwischen Navigator und Inspector blieben der
  // Bezeichnung neben fünf festen Spalten keine 60 Punkte: erst „Counters…“,
  // dann, umgebrochen, Zeilen, die in die Nachbarn liefen — eine Zeile ist 40
  // Punkte hoch, damit die Fensterliste rechnen kann. Ob die Liste schmal
  // ist, fragt sie sich selbst (`@container`); jsdom wertet das nicht aus,
  // deshalb stehen hier die Klassen, und das Fenster misst `ap12feinschliff09`.
  it('stellt schmal Sprache, Set und Lagerort unter die Bezeichnung', () => {
    zeige()

    const gitter = screen.getByRole('grid', { name: texte.listeBezeichnung })
    const name = within(gitter).getByText('Erste')
    expect(name).toHaveClass('truncate')

    const zelle = element(name.closest('[role="gridcell"]'))
    expect(zelle).toHaveAttribute('title', 'Erste')
    const darunter = within(zelle).getByText('en · Nonfoil · m15 · Kiste A')
    expect(darunter).toHaveClass('@2xl:hidden', 'truncate', 'text-xs/4')
    expect(darunter).toHaveAttribute('title', 'en · Nonfoil · m15 · Kiste A')

    // **Ohne Lagerort kein Strich am Ende.**
    const zweite = element(
      within(gitter).getByText('Zweite').closest('[role="gridcell"]'),
    )
    expect(within(zweite).getByText('en · Foil · khm')).toBeInTheDocument()

    // **Breit stehen sie in ihren Spalten** — und die Köpfe so breit wie
    // die Zellen darunter.
    const zeile = element(zelle.closest('[role="row"]'))
    const zellen = within(zeile).getAllByRole('gridcell')
    const koepfe = within(gitter).getAllByRole('columnheader')
    expect(koepfe).toHaveLength(zellen.length)
    for (const [stelle, kopf] of koepfe.entries()) {
      for (const klasse of ['w-8', 'w-28', 'w-20', 'w-12', 'w-24', 'flex-1']) {
        expect(kopf.classList.contains(klasse)).toBe(
          zellen[stelle]?.classList.contains(klasse),
        )
      }
    }
    for (const stelle of [3, 4, 5]) {
      expect(zellen[stelle]).toHaveClass('hidden', '@2xl:block')
      expect(koepfe[stelle]).toHaveClass('hidden', '@2xl:block')
    }
  })

  // **Die Zeile wächst mit der Schrift** — Zweitblick über Phase 6 von
  // 0.9.0, zweiter Teil, Befund 5: schmal braucht eine Zeile zwei Textzeilen
  // (20 und 16 Punkte bei 16 Punkten Wurzelschrift); bei „sehr groß“ (125 %)
  // sind es 45, und 40 Punkte Zeilenhöhe schnitten sie ab.
  it('rechnet die Zeilenhöhe mit der Wurzelschrift', () => {
    document.documentElement.style.fontSize = '20px'
    try {
      zeige()
      const zeile = element(
        within(screen.getByRole('grid', { name: texte.listeBezeichnung }))
          .getByText('Erste')
          .closest('[role="row"]'),
      )
      expect(zeile.style.height).toBe('50px')
    } finally {
      document.documentElement.style.fontSize = ''
    }
  })

  it('liest einen unbekannten Zustand als Near Mint', () => {
    render(
      <Sprachraum sprache="de">
        <Sammlungsliste
          exemplare={[exemplar({ bezeichnung: 'Rätsel', zustand: 'kaputt' })]}
          sicht={leereSicht}
          gewaehlt={new Set()}
          anhaken={vi.fn()}
          gehalten={undefined}
          waehlen={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(
      screen.getByText('Rätsel').closest('[role="row"]'),
    ).toHaveTextContent('3 × Near Mint')
  })

  it('zeichnet das gehaltene Exemplar aus', () => {
    zeige({}, 3)

    const zeile = screen.getByText('Dritte').closest('[role="row"]')
    expect(zeile).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByText('Erste').closest('[role="row"]')).toHaveAttribute(
      'aria-selected',
      'false',
    )
  })
})

describe('Die Tastatur in der Tabelle', () => {
  it('wählt beim Wandern die gehaltene Zeile', () => {
    const { waehlen } = zeige()

    const erste = element(screen.getByText('Erste').closest('[role="row"]'))
    erste.focus()
    fireEvent.keyDown(erste, { key: 'ArrowDown' })

    expect(waehlen).toHaveBeenLastCalledWith(2)
  })
})

describe('Die Gruppen', () => {
  it('setzt Köpfe mit Anzahl und Summe, ohne Angabe zuletzt', () => {
    zeige({ ansicht: 'gruppen', gruppierung: 'lagerort' })

    const koepfe = screen.getAllByRole('rowheader')
    expect(koepfe.map((k) => k.textContent)).toEqual([
      `Kiste A4 Exemplare · 12 Stück`,
      `${texte.gruppeOhne}1 Exemplare · 3 Stück`,
    ])
    // Die Zeilen der Gruppe folgen ihrem Kopf.
    const zeilen = screen.getAllByRole('row')
    expect(zeilen[1]).toHaveTextContent('Kiste A')
    expect(zeilen[2]).toHaveTextContent('Erste')
  })

  it('wählt eine Zeile, aber keinen Kopf, und zeichnet die gehaltene aus', async () => {
    const { waehlen } = zeige({ ansicht: 'gruppen', gruppierung: 'set' }, 3)

    await userEvent.click(element(screen.getAllByRole('rowheader')[0]))
    expect(waehlen).not.toHaveBeenCalled()

    await userEvent.click(screen.getByText('Erste'))
    expect(waehlen).toHaveBeenCalledWith(1)

    expect(screen.getByText('Dritte').closest('[role="row"]')).toHaveAttribute(
      'aria-selected',
      'true',
    )
    const kopf = screen.getAllByRole('rowheader')[0]?.closest('[role="row"]')
    expect(kopf).toHaveAttribute('aria-selected', 'false')
  })
})

// **Auch in den Gruppen mit der Tastatur** — angestrichen vom Zweitblick
// über Phase 6.3 (G9): der Halt wanderte, aber gewählt wurde nichts, und ein
// Exemplar ließ sich dort nur mit der Maus aufschlagen.
describe('Die Tastatur in den Gruppen', () => {
  it('wählt beim Wandern die gehaltene Zeile und keinen Kopf', () => {
    const { waehlen } = zeige({ ansicht: 'gruppen', gruppierung: 'lagerort' })

    const erste = element(screen.getByText('Erste').closest('[role="row"]'))
    erste.focus()
    fireEvent.keyDown(erste, { key: 'ArrowDown' })

    // **Die nächste Zeile der Gruppe** — in Kiste A folgt auf »Erste« die
    // Dritte, nicht die Zweite.
    expect(waehlen).toHaveBeenLastCalledWith(3)

    // **Ein Kopf wird gehalten, aber nicht gewählt** — `Home` führt auf den
    // Kopf von Kiste A.
    const bisher = waehlen.mock.calls.length
    fireEvent.keyDown(document.activeElement ?? erste, { key: 'Home' })
    expect(waehlen).toHaveBeenCalledTimes(bisher)
  })
})

describe('Die Galerie', () => {
  it('legt Kacheln zu vieren in Reihen und wählt per Klick', async () => {
    const { waehlen } = zeige({ ansicht: 'galerie' })

    const gitter = screen.getByRole('grid', { name: texte.listeBezeichnung })
    const reihen = within(gitter).getAllByRole('row')
    // Kopfzeile plus zwei Reihen à vier Kacheln für fünf Exemplare.
    expect(reihen).toHaveLength(3)
    expect(within(element(reihen[1])).getAllByRole('button')).toHaveLength(4)

    await userEvent.click(screen.getByRole('button', { name: 'Fünfte wählen' }))
    expect(waehlen).toHaveBeenCalledWith(5)
  })

  it('drückt die gehaltene Kachel', () => {
    zeige({ ansicht: 'galerie' }, 2)

    expect(
      screen.getByRole('button', { name: 'Zweite wählen' }),
    ).toHaveAttribute('aria-pressed', 'true')
    expect(
      screen.getByRole('button', { name: 'Erste wählen' }),
    ).toHaveAttribute('aria-pressed', 'false')
  })
})

const marke = kataloge.de.bausteine.druck.vorlaeufig

// **Ein Druck ohne Alias ist vorläufig** — Werkstück A2 von Phase 5.3 (#1441).
describe('die Liste und die vorläufige Identität', () => {
  it('markiert die Zeile in Tabelle und Galerie', () => {
    const vorlaeufig = [
      exemplar({
        id: 7,
        bezeichnung: 'Eigenbau',
        printing: '',
        druckID: '9',
        vorlaeufig: true,
      }),
    ]

    const { unmount } = render(
      <Sprachraum sprache="de">
        <Sammlungsliste
          exemplare={vorlaeufig}
          sicht={leereSicht}
          gewaehlt={new Set()}
          anhaken={vi.fn()}
          gehalten={undefined}
          waehlen={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(screen.getByText(marke)).toBeInTheDocument()
    unmount()

    render(
      <Sprachraum sprache="de">
        <Sammlungsliste
          exemplare={vorlaeufig}
          sicht={{ ...leereSicht, ansicht: 'galerie' }}
          gewaehlt={new Set()}
          anhaken={vi.fn()}
          gehalten={undefined}
          waehlen={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(screen.getByText(marke)).toBeInTheDocument()
  })
})
