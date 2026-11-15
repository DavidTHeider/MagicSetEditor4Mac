// **Rücknahme je Handgriff im Symboleditor** —
// #1340,
// Werkstück F von Phase 4.11.
//
// # Die Abnahme aus dem Issue
//
// > Ein Punkt wird verschoben, ⌘Z stellt **ihn** zurück und nicht die ganze
// > Zeichnung.
//
// Das ist die Zusage, und sie steht hier — nicht als „der Haken hält seinen
// Stapel", sondern als beobachtbares Verhalten an der Seite.
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

function zeige() {
  belegeApi({
    Grundform: vi.fn().mockResolvedValue(dreieck()),
    Setangaben: vi.fn().mockResolvedValue({
      felder: [],
      werte: [],
      kopf: leererSetkopf,
      fehler: null,
    }),
  })

  render(
    <Sprachraum sprache="de">
      <Symbol setID="7" />
    </Sprachraum>,
  )
}

// maesseStellen gibt der Zeichenfläche eine Größe.
//
// **In jsdom hat kein Element eine**; `getBoundingClientRect` gibt überall
// Nullen, und eine Umrechnung durch null gäbe `NaN`. Dieselbe Vorrichtung wie
// in `symbolziehen.test.tsx`.
function maesseStellen() {
  const nichts = { left: 0, top: 0, width: 0, height: 0 } as DOMRect

  vi.spyOn(Element.prototype, 'getBoundingClientRect').mockImplementation(
    function (this: Element) {
      return this.tagName.toLowerCase() === 'svg'
        ? ({ left: 0, top: 0, width: 320, height: 320 } as DOMRect)
        : nichts
    },
  )
}

// zeigerAuf schickt ein Zeigerereignis an ein Element.
function zeigerAuf(ziel: Element, art: string, x: number, y: number) {
  const ereignis = new MouseEvent(art, {
    bubbles: true,
    clientX: x,
    clientY: y,
  })

  Object.defineProperty(ereignis, 'pointerId', { value: 1 })
  ziel.dispatchEvent(ereignis)
}

// stelleVon liest die Lage eines Punktes aus seinem Kreis.
function stelleVon(nummer: number): [string, string] {
  const kreis = screen.getByRole('button', {
    name: `Punkt ${String(nummer)} von Dreieck`,
  })

  return [kreis.getAttribute('cx') ?? '', kreis.getAttribute('cy') ?? '']
}

describe('Die Rücknahme im Symboleditor', () => {
  it('steht still, solange nichts geschehen ist', async () => {
    zeige()

    const knopf = await screen.findByRole('button', {
      name: texte.zuruecknehmen,
    })

    expect(knopf).toBeDisabled()
  })

  it('nimmt genau den letzten Handgriff zurück', async () => {
    zeige()

    const dreieckKnopf = await screen.findByRole('button', {
      name: texte.dreieck,
    })

    await userEvent.click(dreieckKnopf)

    const erster = await screen.findByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    const vorher = stelleVon(1)

    erster.focus()
    await userEvent.keyboard('{ArrowRight}')

    expect(stelleVon(1)).not.toEqual(vorher)

    await userEvent.click(
      screen.getByRole('button', { name: texte.zuruecknehmen }),
    )

    // **Der Punkt steht wieder, wo er war — und das Dreieck steht noch da.**
    // Genau das ist der Unterschied zwischen einer Rücknahme je Handgriff und
    // einer, die die ganze Zeichnung zurücknimmt.
    expect(stelleVon(1)).toEqual(vorher)
    expect(
      screen.getByRole('button', { name: 'Punkt 2 von Dreieck' }),
    ).toBeInTheDocument()
  })

  it('legt für jeden Tastendruck einen eigenen Schritt an', async () => {
    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.dreieck }),
    )

    const erster = await screen.findByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    const vorher = stelleVon(1)

    erster.focus()
    await userEvent.keyboard('{ArrowRight}')

    const nachEinem = stelleVon(1)

    await userEvent.keyboard('{ArrowRight}')

    await userEvent.click(
      screen.getByRole('button', { name: texte.zuruecknehmen }),
    )

    // **Ein Schritt, nicht beide** — `point_editor.cpp` Zeilen 355–380 legt je
    // Tastendruck eine neue Handlung an, und `ControlPointMoveAction` kennt
    // kein `merge`.
    expect(stelleVon(1)).toEqual(nachEinem)
    expect(stelleVon(1)).not.toEqual(vorher)
  })

  it('stellt wieder her, was zurückgenommen wurde', async () => {
    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.dreieck }),
    )

    const erster = await screen.findByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    erster.focus()
    await userEvent.keyboard('{ArrowRight}')

    const gezogen = stelleVon(1)

    await userEvent.click(
      screen.getByRole('button', { name: texte.zuruecknehmen }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.wiederherstellen }),
    )

    expect(stelleVon(1)).toEqual(gezogen)
  })

  // **Ein Zug ist ein Schritt — und zwei Züge sind zwei.**
  //
  // # Warum dieser Fall eigens dasteht
  //
  // Die vielen Bewegungen **eines** Zuges müssen zusammenfallen, sonst nähme
  // ⌘Z einen Punkt um ein Pixel zurück. Was zwei Züge **trennt**, ist das
  // Aufsetzen — MSEs `resetActions()` (`point_editor.cpp` Zeile 358).
  //
  // **Die Gegenprobe dazu blieb erst grün**: kein Lauf enthielt zwei Züge,
  // und das Entfernen der Zählung fiel deshalb niemandem auf (§10). Dieser
  // Fall ist die Antwort darauf.
  it('trennt zwei Züge, hält aber jeden zusammen', async () => {
    maesseStellen()
    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.dreieck }),
    )

    const erster = await screen.findByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    Object.assign(erster, {
      setPointerCapture: vi.fn(),
      releasePointerCapture: vi.fn(),
    })

    const anfang = stelleVon(1)

    // **Der erste Zug — zwei Bewegungen, ein Schritt.**
    zeigerAuf(erster, 'pointerdown', 160, 32)
    zeigerAuf(erster, 'pointermove', 200, 64)
    zeigerAuf(erster, 'pointermove', 288, 96)
    await waitFor(() => {
      expect(Number(stelleVon(1)[0])).toBeCloseTo(288, 6)
    })
    zeigerAuf(erster, 'pointerup', 288, 96)

    const nachErstem = stelleVon(1)

    // **Der zweite Zug an demselben Punkt.**
    zeigerAuf(erster, 'pointerdown', 288, 96)
    zeigerAuf(erster, 'pointermove', 64, 224)
    await waitFor(() => {
      expect(Number(stelleVon(1)[0])).toBeCloseTo(64, 6)
    })
    zeigerAuf(erster, 'pointerup', 64, 224)

    await userEvent.click(
      screen.getByRole('button', { name: texte.zuruecknehmen }),
    )

    // **Zurück auf den Stand nach dem ersten Zug** — nicht auf den Anfang und
    // nicht um eine Zeigerbewegung.
    expect(stelleVon(1)).toEqual(nachErstem)
    expect(stelleVon(1)).not.toEqual(anfang)

    await userEvent.click(
      screen.getByRole('button', { name: texte.zuruecknehmen }),
    )

    expect(stelleVon(1)).toEqual(anfang)
  })

  // **Das Anlegen einer Grundform ist auch ein Handgriff.**
  it('nimmt eine hinzugefügte Grundform zurück', async () => {
    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.dreieck }),
    )

    expect(
      await screen.findByRole('button', { name: 'Punkt 1 von Dreieck' }),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.zuruecknehmen }),
    )

    expect(
      screen.queryByRole('button', { name: 'Punkt 1 von Dreieck' }),
    ).not.toBeInTheDocument()
  })
})
