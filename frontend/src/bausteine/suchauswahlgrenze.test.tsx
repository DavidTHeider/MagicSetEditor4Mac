// **Eine Auswahl zeichnet nicht jede Zeile** — UXR-F1
// (#2007).
//
// Ein Set trägt tausende Karten, und jede Zeile der Liste ist ein Knoten. Die
// Auswahl zeichnet höchstens `zeilengrenze` davon und sagt, dass es mehr gibt
// — mit einem Satz aus dem Katalog, sichtbar unter der Liste und angesagt in
// der Live-Region. **Gesucht wird in allen**, nicht nur in den gezeichneten:
// wer eingrenzt, findet auch, was hinter der Grenze stand.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Probe, ansage, feld } from '../test/suchauswahlprobe'
import { zeilengrenze } from './suchwahl'

import type { Suchwahl } from './suchwahl'

const worte = kataloge.de.bausteine.suchauswahl

// zeilen baut so viele Zeilen, wie verlangt sind: „Karte 001", „Karte 002" …
function zeilen(anzahl: number): Suchwahl[] {
  return Array.from({ length: anzahl }, (_, stelle) => {
    const nummer = String(stelle + 1).padStart(3, '0')

    return { wert: `k${nummer}`, name: `Karte ${nummer}` }
  })
}

const begrenzt = (gezeigt: number, anzahl: number) =>
  worte.begrenzt
    .replace('{gezeigt}', String(gezeigt))
    .replace('{anzahl}', String(anzahl))

const gezeichnet = () => screen.queryAllByRole('option')

describe('Suchauswahl — die Zeilengrenze', () => {
  it('liegt bei zweihundert', () => {
    expect(zeilengrenze).toBe(200)
  })

  it('zeichnet höchstens die Grenze und sagt, dass es mehr gibt', async () => {
    render(<Probe eintraege={zeilen(250)} anfang="" />)

    await userEvent.click(feld())

    expect(gezeichnet()).toHaveLength(zeilengrenze)
    expect(gezeichnet()[0]).toHaveTextContent('Karte 001')
    expect(gezeichnet()[199]).toHaveTextContent('Karte 200')

    // **Sichtbar unter der Liste und angesagt** — zweimal im Baum; in der Liste
    // selbst steht er nicht: ein Satz ist keine Zeile, die man wählen könnte.
    expect(screen.getAllByText(begrenzt(200, 250))).toHaveLength(2)
    expect(ansage()).toHaveTextContent(begrenzt(200, 250))
    expect(screen.getByRole('listbox')).not.toHaveTextContent(
      worte.begrenzt.slice(0, 5),
    )
  })

  it('zeichnet bei genau der Grenze alles und schweigt', async () => {
    render(<Probe eintraege={zeilen(200)} anfang="" />)

    await userEvent.click(feld())

    expect(gezeichnet()).toHaveLength(200)
    expect(screen.queryByText(begrenzt(200, 200))).toBeNull()
    expect(ansage()).toHaveTextContent(worte.treffer.replace('{anzahl}', '200'))
  })

  it('findet auch, was hinter der Grenze steht', async () => {
    const aendern = vi.fn()
    render(<Probe eintraege={zeilen(250)} anfang="" aendern={aendern} />)

    await userEvent.click(feld())
    await userEvent.keyboard('karte 250{Enter}')

    expect(aendern).toHaveBeenCalledExactlyOnceWith('k250')
    expect(feld()).toHaveValue('Karte 250')
  })

  it('schweigt, sobald die Suche unter die Grenze fällt', async () => {
    render(<Probe eintraege={zeilen(250)} anfang="" />)

    await userEvent.click(feld())
    await userEvent.keyboard('karte 24')

    // 240 bis 249: zehn Treffer.
    expect(gezeichnet()).toHaveLength(10)
    expect(screen.queryByText(/Suche eingrenzen/)).toBeNull()
    expect(ansage()).toHaveTextContent(worte.treffer.replace('{anzahl}', '10'))
  })

  it('zählt die Treffer der Suche, nicht die Zeilen insgesamt', async () => {
    render(<Probe eintraege={zeilen(450)} anfang="" />)

    await userEvent.click(feld())
    await userEvent.keyboard('karte 3')

    // 300 bis 399: hundert Treffer, alle gezeichnet.
    expect(gezeichnet()).toHaveLength(100)

    await userEvent.clear(feld())
    await userEvent.keyboard('karte')

    expect(gezeichnet()).toHaveLength(zeilengrenze)
    expect(ansage()).toHaveTextContent(begrenzt(200, 450))
  })

  it('zeigt einen Wert hinter der Grenze im Feld, hebt aber nichts hervor', async () => {
    render(<Probe eintraege={zeilen(250)} anfang="k250" />)

    expect(feld()).toHaveValue('Karte 250')

    await userEvent.click(feld())

    expect(gezeichnet()).toHaveLength(zeilengrenze)
    expect(feld()).not.toHaveAttribute('aria-activedescendant')
  })

  it('wandert mit den Pfeilen nur durch die gezeichneten Zeilen', async () => {
    render(<Probe eintraege={zeilen(250)} anfang="k001" />)
    await userEvent.tab()
    await userEvent.keyboard('{ArrowDown}')

    await userEvent.keyboard('{ArrowUp}')

    expect(feld()).toHaveAttribute(
      'aria-activedescendant',
      gezeichnet()[199]?.id,
    )
  })
})
