// **Ein Rechtsklick auf eine andere Karte** — AP-5
// (#2362),
// Befund B-01: er wählt sie, und das Menü geht erst mit **ihren** Handlungen
// auf. Kommt stattdessen eine dritte Karte, verfällt er — sonst spränge das
// Menü später an einer Stelle auf, an der niemand mehr geklickt hat.
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KartenzeileDTO } from '../api'
import { useHandlungsliste } from '../haken/useHandlungsliste'
import { Objekthandlungsraum } from '../objekthandlungsraum'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Kartenliste } from './Kartenliste'

afterEach(() => {
  cleanup()
  raeumeApi()
})

function karte(id: string, nummer: number, name: string): KartenzeileDTO {
  return { id, nummer, name, felder: 1 }
}

const drei = [
  karte('1', 0, 'Erste'),
  karte('2', 1, 'Zweite'),
  karte('3', 2, 'Dritte'),
]

// Handlungen meldet, was `Kartenhandlungen` für die offene Karte meldet.
function Handlungen({ kennung }: { kennung: string }) {
  useHandlungsliste({
    titel: 'Karte',
    art: 'karte',
    kennung,
    eintraege: [
      { schluessel: 'a', name: `Verdoppeln ${kennung}`, aufWahl: vi.fn() },
    ],
  })

  return null
}

function Liste({
  gewaehlt,
  waehlen,
}: {
  gewaehlt: string
  waehlen: (karteID: string) => void
}) {
  return (
    <Sprachraum sprache="de">
      <Objekthandlungsraum>
        <Handlungen kennung={gewaehlt} />
        <Kartenliste
          spalten={[]}
          marken={{}}
          setID="7"
          stand={0}
          karten={drei}
          gewaehlt={gewaehlt}
          waehlen={waehlen}
          angehakt={[]}
          anhaken={vi.fn()}
          mitKaestchen={false}
        />
      </Objekthandlungsraum>
    </Sprachraum>
  )
}

async function gemeldet() {
  await act(async () => {
    await Promise.resolve()
  })
}

describe('der Rechtsklick auf eine andere Karte', () => {
  it('öffnet ihr Menü, sobald sie gewählt ist', async () => {
    belegeApi({
      Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
      ObjekthandlungenMelden: () => Promise.resolve({ fehler: null }),
    })
    const waehlen = vi.fn()
    const { rerender } = render(<Liste gewaehlt="1" waehlen={waehlen} />)
    await gemeldet()

    fireEvent.contextMenu(screen.getByRole('row', { name: 'Zweite' }), {
      clientX: 5,
      clientY: 5,
    })
    expect(waehlen).toHaveBeenLastCalledWith('2')
    expect(screen.queryByRole('menu')).toBeNull()

    rerender(<Liste gewaehlt="2" waehlen={waehlen} />)
    await gemeldet()

    expect(
      screen.getByRole('menuitem', { name: 'Verdoppeln 2' }),
    ).toBeInTheDocument()
  })

  it('verfällt, wenn stattdessen eine dritte gewählt wird', async () => {
    belegeApi({
      Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
      ObjekthandlungenMelden: () => Promise.resolve({ fehler: null }),
    })
    const waehlen = vi.fn()
    const { rerender } = render(<Liste gewaehlt="1" waehlen={waehlen} />)
    await gemeldet()
    fireEvent.contextMenu(screen.getByRole('row', { name: 'Zweite' }), {
      clientX: 5,
      clientY: 5,
    })

    rerender(<Liste gewaehlt="3" waehlen={waehlen} />)
    await gemeldet()
    rerender(<Liste gewaehlt="2" waehlen={waehlen} />)
    await gemeldet()

    expect(screen.queryByRole('menu')).toBeNull()
  })

  it('öffnet mit ⇧F10 auf einer anderen Zeile ebenso deren Menü', async () => {
    belegeApi({
      Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
      ObjekthandlungenMelden: () => Promise.resolve({ fehler: null }),
    })
    const waehlen = vi.fn()
    const { rerender } = render(<Liste gewaehlt="1" waehlen={waehlen} />)
    await gemeldet()

    fireEvent.keyDown(screen.getByRole('row', { name: 'Dritte' }), {
      key: 'F10',
      shiftKey: true,
    })
    expect(waehlen).toHaveBeenLastCalledWith('3')

    rerender(<Liste gewaehlt="3" waehlen={waehlen} />)
    await gemeldet()

    expect(
      screen.getByRole('menuitem', { name: 'Verdoppeln 3' }),
    ).toBeInTheDocument()
  })
})
