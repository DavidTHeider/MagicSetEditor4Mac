// **Die Griffe am Zeiger** —
// #890.
//
// **Eine eigene Datei**, weil `symbolgriffe.test.tsx` über die Zeilengrenze
// wuchs, als der Nachlauf des Zweitblicks eine zweite Bewegung verlangte
// (§2). Der Schnitt folgt der Frage: dort **die Tastatur**, hier **der
// Zeiger**.
import { act, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { leererSetkopf } from '../test/setkopfprobe'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { maesseStellen, mitKurve } from '../test/symbolproben'
import { Symbol } from './Symbol'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.symbol

describe('Die Griffe am Zeiger', () => {
  it('folgen dem Zeiger', async () => {
    maesseStellen()

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
      name: 'Griff nach Punkt 1',
    })

    Object.assign(griff, {
      setPointerCapture: vi.fn(),
      releasePointerCapture: vi.fn(),
    })

    const zeiger = (art: string, x: number, y: number) => {
      const ereignis = new MouseEvent(art, {
        bubbles: true,
        clientX: x,
        clientY: y,
      })
      Object.defineProperty(ereignis, 'pointerId', { value: 1 })
      griff.dispatchEvent(ereignis)
    }

    zeiger('pointerdown', 224, 32)
    zeiger('pointermove', 256, 32)

    // **Geprüft wird die Stelle, nicht der bloße Aufruf.**
    //
    // `toHaveBeenCalled()` stand hier zuerst und hätte jede Zahl
    // durchgelassen — auch die falsche, die der Nachlauf des Zweitblicks
    // gefunden hat (#890).
    await waitFor(() => {
      expect(GriffZiehen).toHaveBeenCalledWith(expect.anything(), {
        teil: 0,
        punkt: 0,
        davor: false,
        x: 0.8,
        y: 0.1,
      })
    })

    // **Und die zweite Bewegung landet auch dort, wo der Zeiger steht.**
    // Eine Differenz gegen den gemerkten Anfang gäbe hier 0,975.
    zeiger('pointermove', 288, 32)

    await waitFor(() => {
      expect(GriffZiehen).toHaveBeenLastCalledWith(expect.anything(), {
        teil: 0,
        punkt: 0,
        davor: false,
        x: 0.9,
        y: 0.1,
      })
    })

    zeiger('pointerup', 288, 32)
  })

  // **Und der vordere Griff ebenso.**
  //
  // Die Seitenwahl steht in `zeigerBewegt` und ist eine andere Zeile als die in
  // `griffTaste`; der Tastaturlauf daneben deckt sie nicht.
  it('folgen dem Zeiger auch davor', async () => {
    maesseStellen()

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

    Object.assign(griff, { setPointerCapture: vi.fn() })

    const zeiger = (art: string, x: number, y: number) => {
      const ereignis = new MouseEvent(art, {
        bubbles: true,
        clientX: x,
        clientY: y,
      })

      Object.defineProperty(ereignis, 'pointerId', { value: 1 })
      griff.dispatchEvent(ereignis)
    }

    zeiger('pointerdown', 96, 32)
    zeiger('pointermove', 64, 32)

    await waitFor(() => {
      expect(GriffZiehen).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({ davor: true }),
      )
    })

    zeiger('pointerup', 64, 32)
  })

  // **Eine Taste ohne Richtung tut nichts.**
  //
  // Am Punkt gibt es diesen Zweig auch; am Griff hat er seinen eigenen
  // Behandler, und ein Zweig, den kein Aufruf betritt, ist eine Behauptung
  // (§3).
})

// **Eine überholte Antwort trägt den Griff nicht zurück** (§7, Regel 2).
//
// # Warum das hier eine eigene Prüfung braucht
//
// Ein Zug schickt eine Anfrage je Bewegung. Sie gehen über dieselbe Naht und
// kommen **nicht zwingend in der Reihenfolge** zurück, in der sie hinausgingen
// — angestrichen vom Nachlauf des Zweitblicks über Phase X (#890).
//
// Der Lauf lässt die **erste** Antwort nach der zweiten eintreffen. Ohne den
// Zähler stünde danach der Stand von vor zwei Bewegungen da, während der
// Zeiger woanders steht.
describe('Zwei Antworten, die sich überholen', () => {
  function haltbar() {
    let loesen: (wert: unknown) => void = () => undefined
    let scheitern: (grund: unknown) => void = () => undefined

    const versprechen = new Promise((ja, nein) => {
      loesen = ja
      scheitern = nein
    })

    return { versprechen, loesen, scheitern }
  }

  function zug(griff: HTMLElement) {
    return (art: string, x: number, y: number) => {
      const ereignis = new MouseEvent(art, {
        bubbles: true,
        clientX: x,
        clientY: y,
      })

      Object.defineProperty(ereignis, 'pointerId', { value: 1 })
      griff.dispatchEvent(ereignis)
    }
  }

  async function gezogen(erste: Promise<unknown>, zweite: Promise<unknown>) {
    maesseStellen()

    const GriffZiehen = vi
      .fn()
      .mockReturnValueOnce(erste)
      .mockReturnValueOnce(zweite)

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

    Object.assign(griff, { setPointerCapture: vi.fn() })

    const zeiger = zug(griff)

    zeiger('pointerdown', 224, 32)
    zeiger('pointermove', 256, 32)
    zeiger('pointermove', 288, 32)

    await waitFor(() => {
      expect(GriffZiehen).toHaveBeenCalledTimes(2)
    })

    return griff
  }

  it('lässt die frühere liegen', async () => {
    const erste = haltbar()
    const zweite = haltbar()

    const griff = await gezogen(erste.versprechen, zweite.versprechen)

    // **Die zweite kommt zuerst an** und trägt sich ein.
    zweite.loesen({
      ...mitKurve(),
      teile: mitKurve().teile.map((teil) => ({
        ...teil,
        name: 'Zweite',
      })),
    })

    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: /^Zweite —/ }),
      ).toBeInTheDocument()
    })

    // **Und die erste danach.** Sie gehört zu einem Stand, den niemand mehr
    // offen hat.
    erste.loesen({
      ...mitKurve(),
      teile: mitKurve().teile.map((teil) => ({ ...teil, name: 'Erste' })),
    })

    // **Gewartet wird auf das Versprechen selbst, nicht auf eine Frist.**
    // Danach ist der Rückruf der Seite gelaufen — er hing vor diesem hier —,
    // und was jetzt dasteht, ist das Ergebnis und keine Momentaufnahme (§3).
    await act(async () => {
      await erste.versprechen
    })

    expect(
      screen.queryByRole('button', { name: /^Erste —/ }),
    ).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: /^Zweite —/ }),
    ).toBeInTheDocument()
    expect(griff).toBeDefined()
  })

  it('meldet auch ihren Fehler nicht mehr', async () => {
    const erste = haltbar()
    const zweite = haltbar()

    await gezogen(erste.versprechen, zweite.versprechen)

    // **Die zweite scheitert**, und das ist der heutige Stand — die Meldung
    // gehört auf die Seite. Sie ist zugleich der Beleg, dass der Rückruf
    // wirklich hängt: ohne ihn stünde hier nie ein `alert` (§3).
    zweite.scheitern(new Error('Der Griff ließ sich nicht ziehen.'))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Griff ließ sich nicht ziehen.',
    )

    // **Und die erste danach.** Ihre Meldung gehört zu einem Stand, den
    // niemand mehr offen hat; sie darf die heutige nicht überschreiben.
    erste.scheitern(new Error('Ein überholter Fehler.'))

    await act(async () => {
      await erste.versprechen.catch(() => undefined)
    })

    expect(screen.getByRole('alert')).toHaveTextContent(
      'Der Griff ließ sich nicht ziehen.',
    )
  })
})
