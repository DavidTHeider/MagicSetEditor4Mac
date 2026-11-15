// **Der Fokus kehrt zurück — wenn es den Ort noch gibt** — Werkstück U2 von
// Phase 5.1 (#1413).
import { fireEvent } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { fokusZurueck, merkeRueckkehr, relativZu } from './fokus'

describe('fokusZurueck', () => {
  it('setzt den Fokus auf ein Element im Baum', () => {
    const knopf = document.createElement('button')
    document.body.append(knopf)

    expect(fokusZurueck(knopf)).toBe(true)
    expect(knopf).toHaveFocus()

    knopf.remove()
  })

  it('lässt den Fokus, wo er ist, wenn das Element den Baum verlassen hat', () => {
    const knopf = document.createElement('button')

    expect(fokusZurueck(knopf)).toBe(false)
    expect(document.activeElement).toBe(document.body)
  })

  // **`null` und ein Element ohne `focus()`** — `document.activeElement` ist
  // beides, je nach Lage: `null` in einem Dokument ohne Körper, ein
  // SVG-Knoten, wenn jemand in eine Grafik geklickt hat.
  it('tut nichts bei null und bei einem Element, das keinen Fokus nimmt', () => {
    expect(fokusZurueck(null)).toBe(false)

    const grafik = document.createElementNS('http://www.w3.org/2000/svg', 'svg')
    document.body.append(grafik)

    expect(fokusZurueck(grafik)).toBe(false)

    grafik.remove()
  })
})

// **Wohin der Fokus beim Schließen zurück soll** — UX-08.AK-03 (#1680):
// WebKit gibt einem geklickten Knopf keinen Fokus.
describe('merkeRueckkehr', () => {
  it('nimmt nach einer Taste das Element, das den Fokus hat', () => {
    const knopf = document.createElement('button')
    const feld = document.createElement('input')
    document.body.append(knopf, feld)

    fireEvent.pointerDown(knopf)
    feld.focus()
    fireEvent.keyDown(feld, { key: 'Tab' })

    expect(merkeRueckkehr()).toBe(feld)

    knopf.remove()
    feld.remove()
  })

  // **Und geht der Fokus danach woanders hin**, gilt der Druck nicht mehr —
  // ein Knopf, der beim Klick selbst den Fokus nimmt, wie die der Schubladen.
  it('vergisst den Druck, wenn der Fokus danach woanders hingeht', () => {
    const knopf = document.createElement('button')
    const anderer = document.createElement('button')
    document.body.append(knopf, anderer)

    fireEvent.pointerDown(knopf)
    anderer.focus()

    expect(merkeRueckkehr()).toBe(anderer)

    knopf.remove()
    anderer.remove()
  })

  // **Geht er in den gedrückten Knopf**, bleibt der Druck — so klickt jeder
  // Browser außer WebKit.
  it('behält den Druck, wenn der Fokus in den gedrückten Knopf geht', () => {
    const knopf = document.createElement('button')
    document.body.append(knopf)

    fireEvent.pointerDown(knopf)
    knopf.focus()

    expect(merkeRueckkehr()).toBe(knopf)

    knopf.remove()
  })

  // **Und geht er in einen Behälter um den Knopf**, bleibt der Druck
  // ebenso: WebKit gibt den Fokus beim Klick dem nächsten fokussierbaren
  // Vorfahren — in der Anwendung dem Inhaltsbereich mit `tabindex="-1"`.
  it('behält den Druck, wenn der Fokus in einen Behälter um den Knopf geht', () => {
    const behaelter = document.createElement('div')
    behaelter.tabIndex = -1
    const knopf = document.createElement('button')
    behaelter.append(knopf)
    document.body.append(behaelter)

    fireEvent.pointerDown(knopf)
    behaelter.focus()

    expect(merkeRueckkehr()).toBe(knopf)

    behaelter.remove()
  })

  // **Der Fall aus WebKit**: der Klick lässt den Fokus, wo er war.
  it('nimmt den gedrückten Knopf, auch wenn der Fokus woanders steht', () => {
    const feld = document.createElement('input')
    const knopf = document.createElement('button')
    document.body.append(feld, knopf)

    feld.focus()
    fireEvent.pointerDown(knopf)

    expect(merkeRueckkehr()).toBe(knopf)

    feld.remove()
    knopf.remove()
  })

  it('nimmt ohne Fokus den zuletzt gedrückten Knopf — auch über sein Inneres', () => {
    const knopf = document.createElement('button')
    const aufschrift = document.createElement('span')
    knopf.append(aufschrift)
    document.body.append(knopf)

    fireEvent.pointerDown(aufschrift)

    expect(merkeRueckkehr()).toBe(knopf)

    // **Und nicht mehr, sobald es ihn nicht mehr gibt**: ein Fokus auf einem
    // losgelösten Knoten wäre einer im Nirgendwo.
    knopf.remove()
    expect(merkeRueckkehr()).toBe(document.body)
  })

  it('vergisst den Knopf, wenn der Druck nicht auf einem Element landet', () => {
    const knopf = document.createElement('button')
    document.body.append(knopf)
    fireEvent.pointerDown(knopf)

    fireEvent.pointerDown(document)

    expect(merkeRueckkehr()).toBe(document.body)
    knopf.remove()
  })
})

describe('relativZu', () => {
  it('rechnet die Stelle gegen den Rand des Kastens', () => {
    const kasten = document.createElement('div')
    kasten.getBoundingClientRect = () => ({ left: 100, top: 40 }) as DOMRect

    expect(relativZu(kasten, 130, 55)).toEqual({ left: 30, top: 15 })
  })

  it('lässt die Stelle ohne Kasten, wie sie ist', () => {
    expect(relativZu(null, 130, 55)).toEqual({ left: 130, top: 55 })
  })
})

// **Nach einem nativen Dialog kehrt der Fokus zum Auslöser zurück** —
// UX-08.AK-03 (#1680).
//
// Ein Dateidialog gehört dem Betriebssystem: solange er offen ist, hat das
// Fenster den Fokus abgegeben, und wenn er schließt, bekommt es ihn zurück.
// **Stand er vorher auf `body`** — ein Klick fokussiert in WebKit nicht —,
// wüsste niemand mehr, wo er war.
describe('der Fokus nach einem nativen Dialog', () => {
  // **Jeder Lauf räumt, auch ein roter** — sonst hielte ein Behälter von
  // vorhin den Fokus, und der nächste Lauf prüfte dessen Rest.
  afterEach(() => {
    document.body.replaceChildren()
  })

  it('kehrt zum gedrückten Knopf zurück, wenn das Fenster ihn wiederbekommt', () => {
    const knopf = document.createElement('button')
    document.body.append(knopf)

    fireEvent.pointerDown(knopf)
    expect(document.activeElement).toBe(document.body)

    fireEvent.blur(window)
    fireEvent.focus(window)

    expect(knopf).toHaveFocus()
  })

  // **Auch wenn WebKit den Fokus dem Behälter gab** — ein Klick fokussiert
  // dort den nächsten fokussierbaren Vorfahren, in der Anwendung den
  // Inhaltsbereich. Nach dem Dialog steht er also **nicht** auf `body`.
  it('kehrt auch aus dem Behälter um den Knopf zurück', () => {
    const behaelter = document.createElement('div')
    behaelter.tabIndex = -1
    const knopf = document.createElement('button')
    behaelter.append(knopf)
    document.body.append(behaelter)

    fireEvent.pointerDown(knopf)
    behaelter.focus()

    fireEvent.focus(window)

    expect(knopf).toHaveFocus()
  })

  // **Nicht, wenn der Fokus schon woanders steht** — ein Feld, in das jemand
  // danach geklickt hat, behält ihn.
  it('nimmt einem Feld den Fokus nicht', () => {
    const knopf = document.createElement('button')
    const feld = document.createElement('input')
    document.body.append(knopf, feld)

    // **Dass jemand zuhört** (§3): steht der Fokus auf `body`, kehrt er zum
    // Knopf zurück — sonst bewiese das Nein unten nichts.
    fireEvent.pointerDown(knopf)
    fireEvent.focus(window)
    expect(knopf).toHaveFocus()

    fireEvent.pointerDown(knopf)
    feld.focus()

    fireEvent.focus(window)

    expect(feld).toHaveFocus()
  })

  // **Und nicht nach einer Taste** — wer danach mit der Tastatur weiterging,
  // meint nicht mehr den Knopf von vorhin.
  it('vergisst den Knopf nach einer Taste', () => {
    const knopf = document.createElement('button')
    document.body.append(knopf)

    // **Dass jemand zuhört** (§3) — derselbe Weg ohne Taste bringt ihn hin.
    fireEvent.pointerDown(knopf)
    fireEvent.focus(window)
    expect(knopf).toHaveFocus()
    knopf.blur()

    fireEvent.pointerDown(knopf)
    fireEvent.keyDown(document.body, { key: 'Tab' })

    fireEvent.focus(window)

    expect(document.activeElement).toBe(document.body)
  })
})
