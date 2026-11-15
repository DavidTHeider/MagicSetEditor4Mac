// **Rechtsklick und ⇧F10 an einer Paketzeile** — AP-5
// (#2362),
// Befund B-01: dieselben Handlungen wie im Menü der Zeile, an der Zeile
// selbst. Eine andere als die gewählte Zeile wird zuerst gewählt, und das
// Menü gehört dann ihr. Löschen und Zurück öffnen den **Plan** — er bleibt
// die Rückfrage (DOM-01, #1696).
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { PaketDTO } from '../api'
import { Schreibrechtraum } from '../haken/schreibrecht'
import type { Markenfilterstand } from '../haken/useMarkenfilter'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { Pakettabelle } from './Pakettabelle'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.pakete

function paket(name: string): PaketDTO {
  return {
    art: 'mse-style',
    name,
    anzeigename: name,
    version: '',
    pfad: '',
    herkunft: '',
    pruefsumme: '',
    braucht: [],
    gruppe: '',
    rang: 1000000,
    artrang: 30,
    herkunftsart: 'mitgeliefert' as const,
    hauptdatei: 'style',
    geaendert: false,
    gebrauchtVon: [],
    benutzt: 0,
  }
}

const erstes = paket('magic-m15')
const zweites = paket('magic-m15-planeswalker')

const marken: Markenfilterstand = {
  filter: { marken: [], verknuepfung: 'und' },
  setzeFilter: vi.fn(),
  je: {},
  getroffen: undefined,
  fehler: undefined,
  nachladen: vi.fn(),
  runde: 0,
}

function Tabelle({
  gewaehlt,
  waehlen,
  planen,
  nurLesend = false,
}: {
  gewaehlt: string | undefined
  waehlen: (paket: PaketDTO) => void
  planen: (vorgang: 'entfernen' | 'zurueck', paket: PaketDTO) => void
  nurLesend?: boolean
}) {
  return (
    <Sprachraum sprache="de">
      <Schreibrechtraum nurLesend={nurLesend}>
        <Pakettabelle
          zeilen={[
            { schluessel: 'g', aufschrift: 'magic', stufe: 0 },
            {
              schluessel: 'a',
              aufschrift: erstes.name,
              stufe: 0,
              paket: erstes,
            },
            {
              schluessel: 'b',
              aufschrift: zweites.name,
              stufe: 0,
              paket: zweites,
            },
          ]}
          eingeklappt={new Set()}
          suchend={false}
          umschalten={vi.fn()}
          vorbehalte={new Map()}
          gewaehlt={gewaehlt}
          waehlen={waehlen}
          planen={planen}
          laeuft={false}
          marken={marken}
        />
      </Schreibrechtraum>
    </Sprachraum>
  )
}

const menueVon = (name: string) =>
  screen.getByRole('menu', { name: einsetzen(eigene.handlungenAn, { name }) })

describe('der Rechtsklick an einer Paketzeile', () => {
  it('öffnet an der gewählten Zeile ihr Menü, und Löschen öffnet den Plan', async () => {
    const planen = vi.fn()
    render(
      <Tabelle
        gewaehlt={`mse-style/${erstes.name}`}
        waehlen={vi.fn()}
        planen={planen}
      />,
    )

    const ereignis = fireEvent.contextMenu(
      screen.getByRole('row', { name: erstes.name }),
      { clientX: 10, clientY: 10 },
    )

    expect(ereignis).toBe(false)
    expect(menueVon(erstes.name)).toBeInTheDocument()
    await userEvent.click(
      screen.getByRole('menuitem', { name: eigene.plan.loeschen }),
    )
    expect(planen).toHaveBeenCalledWith('entfernen', erstes)
  })

  it('wählt eine andere Zeile zuerst, und das Menü gehört dann ihr', async () => {
    const waehlen = vi.fn()
    const planen = vi.fn()
    const { rerender } = render(
      <Tabelle
        gewaehlt={`mse-style/${erstes.name}`}
        waehlen={waehlen}
        planen={planen}
      />,
    )

    fireEvent.contextMenu(screen.getByRole('row', { name: zweites.name }), {
      clientX: 10,
      clientY: 30,
    })
    expect(waehlen).toHaveBeenLastCalledWith(zweites)

    rerender(
      <Tabelle
        gewaehlt={`mse-style/${zweites.name}`}
        waehlen={waehlen}
        planen={planen}
      />,
    )
    expect(menueVon(zweites.name)).toBeInTheDocument()
    await userEvent.click(
      screen.getByRole('menuitem', { name: eigene.plan.zurueck }),
    )
    expect(planen).toHaveBeenCalledWith('zurueck', zweites)
  })

  it('öffnet mit ⇧F10 an der Zeile mit dem Fokus', () => {
    const waehlen = vi.fn()
    render(<Tabelle gewaehlt={undefined} waehlen={waehlen} planen={vi.fn()} />)

    fireEvent.keyDown(screen.getByRole('row', { name: erstes.name }), {
      key: 'F10',
      shiftKey: true,
    })

    expect(waehlen).toHaveBeenLastCalledWith(erstes)
  })

  it('zeigt einem Leser das Schreibende gesperrt, mit Grund', () => {
    render(
      <Tabelle
        gewaehlt={`mse-style/${erstes.name}`}
        waehlen={vi.fn()}
        planen={vi.fn()}
        nurLesend
      />,
    )

    fireEvent.contextMenu(screen.getByRole('row', { name: erstes.name }), {
      clientX: 10,
      clientY: 10,
    })

    const marke = screen.getByRole('menuitem', {
      name: kataloge.de.bausteine.objektmarken.ueberschrift,
    })
    expect(marke).toHaveAttribute('aria-disabled', 'true')
    expect(marke).toHaveAccessibleDescription()
  })

  it('zeigt an einer Gruppenzeile nicht das Menü des gewählten Pakets', () => {
    render(
      <Tabelle
        gewaehlt={`mse-style/${erstes.name}`}
        waehlen={vi.fn()}
        planen={vi.fn()}
      />,
    )

    const gruppe = screen.getByRole('row', { name: 'Gruppe magic' })
    expect(fireEvent.contextMenu(gruppe)).toBe(true)
    fireEvent.keyDown(gruppe, { key: 'F10', shiftKey: true })

    expect(screen.queryByRole('menu')).toBeNull()
  })
})

// **Spalten nach Inhalt** — AP-12
// (#2369),
// Befund B-28: mit festen Prozentspalten lief die Fassung im schmalen Zentrum
// in den Chip daneben („2025-06-18eingebunden“). Jetzt nehmen Art, Fassung,
// Stand und Menü, was ihr Inhalt braucht, und der Name den Rest. Ob es am
// Fenster wirklich nicht mehr überlappt, misst `ap12feinschliff09`.
describe('die Spalten der Pakettabelle', () => {
  it('folgen dem Inhalt: Fassung und Stand ungebrochen und nur so breit wie nötig', () => {
    render(<Tabelle gewaehlt={undefined} waehlen={vi.fn()} planen={vi.fn()} />)

    const tabelle = screen.getByRole('treegrid')
    expect(tabelle).not.toHaveClass('table-fixed')
    expect(tabelle.querySelector('colgroup')).toBeNull()

    // **Nur der Kopf** — die Gruppenzeilen im Rumpf sind auch `th`.
    const kopf = [...tabelle.querySelectorAll('thead th')]
    expect(kopf).toHaveLength(5)
    for (const spalte of kopf.slice(1, 4)) {
      expect(spalte).toHaveClass('w-px', 'whitespace-nowrap')
    }
    // **Das Menü hat seinen Platz, bevor es erscheint** — es steht nur an
    // der gewählten Zeile; ohne feste Breite schob seine Spalte beim Wählen
    // alle anderen, und am Fenster landete der Klick daneben (`dom01`).
    expect(kopf[4]).toHaveClass('w-12')
    expect(kopf[4]).not.toHaveClass('w-px')
    // **Der Name nimmt den Rest** (`w-full`) — ohne das verteilte WebKit
    // den Platz nach eigenem Maß, und am Fenster brach „probe-spielskript“
    // in drei Zeilen, während die Art Luft hatte.
    expect(kopf[0]).toHaveClass('w-full')
    expect(kopf[0]).not.toHaveClass('w-px')
  })
})

// **Was nur an der gewählten Zeile steht, schiebt nichts** — das Menü liegt
// absolut in seiner Zelle. Am Fenster wuchs sie beim Fokus um zwei Punkte,
// die Namen brachen um, die Zeilen wurden höher, und der Klick landete
// zwischen zwei Zeilen (`dom01`).
describe('das Menü der gewählten Zeile', () => {
  it('liegt in seiner Zelle, ohne ihre Breite zu bestimmen', async () => {
    render(
      <Tabelle
        gewaehlt="mse-style/magic-m15"
        waehlen={vi.fn()}
        planen={vi.fn()}
      />,
    )

    const knopf = await screen.findByRole('button', {
      name: einsetzen(eigene.handlungenAn, { name: erstes.name }),
    })
    const huelle = knopf.closest('td > div')
    expect(huelle).toHaveClass('absolute')
    expect(huelle?.parentElement).toHaveClass('relative')
    // **Und die Zelle behält ihre 3rem** — ohne Inhalt im Fluss schrumpfte
    // sie neben dem Namen (`w-full`) auf ihren Rand, und der Knopf lag über
    // dem Chip daneben.
    expect(
      huelle?.parentElement?.querySelector('[aria-hidden="true"]'),
    ).toHaveClass('w-12')
  })
})

// **Die Art unter dem Namen, wenn die Tabelle schmal ist** — bei 1.024 × 700
// brachen sonst alle Namen mitten im Wort: die Tabelle maß 464 Punkte, und
// „mse-export-template“ nahm davon 170. Unter 42rem Breite der Tabelle
// verschwindet ihre Spalte, und sie steht klein unter dem Namen.
describe('die schmale Pakettabelle', () => {
  it('führt die Art je Zeile zweimal: als Spalte ab 42rem, sonst unter dem Namen', () => {
    render(<Tabelle gewaehlt={undefined} waehlen={vi.fn()} planen={vi.fn()} />)

    const tabelle = screen.getByRole('treegrid')
    expect(tabelle.parentElement).toHaveClass('@container')
    const kopf = [...tabelle.querySelectorAll('thead th')]
    expect(kopf[1]).toHaveClass('hidden', '@2xl:table-cell')

    const zeile = screen.getByRole('row', { name: erstes.name })
    const [name, art] = [...zeile.querySelectorAll('td')]
    expect(art).toHaveClass('hidden', '@2xl:table-cell')
    expect(name?.querySelector('.\\@2xl\\:hidden')).toHaveTextContent(
      erstes.art,
    )
  })
})
