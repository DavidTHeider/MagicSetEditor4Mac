// **Die Griffe einer Kurve** —
// #890.
//
// **Eine eigene Datei**, weil `symbolziehen.test.tsx` über die Zeilengrenze
// wuchs (§2). Der Schnitt folgt der Frage: dort **der Punkt am Zeiger**, hier
// **die Griffe**, die ihn zu einer Kurve machen.
import { render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { leererSetkopf } from '../test/setkopfprobe'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { mitKurve } from '../test/symbolproben'
import { Symbol } from './Symbol'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.symbol

describe('Die Griffe eines Kurvenpunktes', () => {
  it('stehen da und lassen sich mit der Tastatur bewegen', async () => {
    const GriffZiehen = vi.fn().mockResolvedValue({
      ...mitKurve(),
      teile: mitKurve().teile.map((teil) => ({
        ...teil,
        punkte: teil.punkte.map((p, wo) =>
          wo === 0 ? { ...p, griffDanachX: 0.3 } : p,
        ),
      })),
    })

    belegeApi({
      Grundform: vi.fn().mockResolvedValue(mitKurve()),
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [],
        kopf: leererSetkopf,
        fehler: null,
      }),
      GriffZiehen,
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="7" />
      </Sprachraum>,
    )

    ;(await screen.findByRole('button', { name: texte.kreis })).click()

    const griff = await screen.findByRole('button', {
      name: 'Griff nach Punkt 1',
    })

    griff.focus()

    const taste = new KeyboardEvent('keydown', {
      key: 'ArrowRight',
      bubbles: true,
    })

    griff.dispatchEvent(taste)

    await waitFor(() => {
      // **Die Spitze steht bei 0,7**: der Punkt bei 0,5 plus sein Griff von
      // 0,2. Ein Schritt nach rechts bringt sie auf 0,71 — die Umrechnung von
      // Schritt auf Stelle geschieht dort, wo der Punkt vorliegt (#890).
      expect(GriffZiehen).toHaveBeenCalledWith(expect.anything(), {
        teil: 0,
        punkt: 0,
        davor: false,
        x: 0.71,
        y: 0.1,
      })
    })
  })

  // **Ein Griff lässt sich auch ziehen.**
  it('lassen eine fremde Taste liegen', async () => {
    const GriffZiehen = vi.fn()

    belegeApi({
      Grundform: vi.fn().mockResolvedValue(mitKurve()),
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [],
        kopf: leererSetkopf,
        fehler: null,
      }),
      GriffZiehen,
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="7" />
      </Sprachraum>,
    )

    ;(await screen.findByRole('button', { name: texte.kreis })).click()

    const griff = await screen.findByRole('button', {
      name: 'Griff nach Punkt 1',
    })

    griff.dispatchEvent(
      new KeyboardEvent('keydown', { key: 'a', bubbles: true }),
    )

    expect(GriffZiehen).not.toHaveBeenCalled()
  })

  // **Ein Fehler beim Ziehen steht am Bildschirm.**
  it('melden einen Fehler', async () => {
    belegeApi({
      Grundform: vi.fn().mockResolvedValue(mitKurve()),
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [],
        kopf: leererSetkopf,
        fehler: null,
      }),
      GriffZiehen: vi.fn().mockRejectedValue(new Error('Das ging nicht.')),
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="7" />
      </Sprachraum>,
    )

    ;(await screen.findByRole('button', { name: texte.kreis })).click()

    const griff = await screen.findByRole('button', {
      name: 'Griff nach Punkt 1',
    })

    griff.dispatchEvent(
      new KeyboardEvent('keydown', { key: 'ArrowRight', bubbles: true }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Das ging nicht.',
    )
  })

  // **Auch der Griff davor lässt sich bewegen, und mit Umschalt weiter.**
  //
  // Zwei Zweige in einem Lauf: die Seitenwahl und die Schrittweite. Beide
  // gehören zusammen — wer den einen Griff bedienen kann, erwartet den anderen
  // und dieselben Tasten.
  it('lassen sich auch davor und mit Umschalt bewegen', async () => {
    const GriffZiehen = vi.fn().mockResolvedValue(mitKurve())

    belegeApi({
      Grundform: vi.fn().mockResolvedValue(mitKurve()),
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [],
        kopf: leererSetkopf,
        fehler: null,
      }),
      GriffZiehen,
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="7" />
      </Sprachraum>,
    )

    ;(await screen.findByRole('button', { name: texte.kreis })).click()

    const griff = await screen.findByRole('button', {
      name: 'Griff vor Punkt 1',
    })

    griff.dispatchEvent(
      new KeyboardEvent('keydown', {
        key: 'ArrowDown',
        shiftKey: true,
        bubbles: true,
      }),
    )

    await waitFor(() => {
      expect(GriffZiehen).toHaveBeenCalledWith(expect.anything(), {
        teil: 0,
        punkt: 0,
        davor: true,

        // **Die vordere Spitze steht bei 0,3** — der Punkt bei 0,5 plus sein
        // Griff von −0,2. Ein großer Schritt nach unten bringt sie auf 0,2.
        x: 0.3,
        y: 0.2,
      })
    })
  })
})
