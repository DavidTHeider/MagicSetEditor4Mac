// **Punkte mit dem Zeiger ziehen** —
// #890.
//
// # Warum es einen eigenen Lauf braucht
//
// Die Tastatur kam zuerst und ist der Entwurf (§6). **Aber die Maus ist
// dieselbe Zusage von der anderen Seite**: die Punkte trugen `cursor-pointer`
// und ließen sich nicht ziehen. Angestrichen vom Zweitblick über Phase X.
//
// **Zeigerereignisse und keine Mausereignisse**: `onPointerDown` bedient Maus,
// Trackpad und Stift mit einem Weg — und `setPointerCapture` hält den Punkt
// auch dann, wenn der Zeiger den Kreis verlässt. Ohne das bliebe er am Rand
// kleben.
import { render, screen, waitFor } from '@testing-library/react'
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

// **Die Maße werden gestellt.** In jsdom hat kein Element eine Größe;
// `getBoundingClientRect` gibt überall Nullen, und eine Umrechnung durch null
// gäbe `NaN`. Der Lauf stellt deshalb ein 320 × 320 großes SVG an der linken
// oberen Ecke — genau die Maße, die die Zeichenfläche im Browser hat.
function maesseStellen() {
  // **Ein eigener Nullkasten statt der echten Methode.** Ein Aufruf der
  // ursprünglichen `getBoundingClientRect` gäbe in jsdom ohnehin überall
  // Nullen; sie festzuhalten, um sie später zu rufen, wäre eine Bindung an
  // `this`, die der Prüfer zu Recht anstreicht.
  const nichts = { left: 0, top: 0, width: 0, height: 0 } as DOMRect

  vi.spyOn(Element.prototype, 'getBoundingClientRect').mockImplementation(
    function (this: Element) {
      return this.tagName.toLowerCase() === 'svg'
        ? ({ left: 0, top: 0, width: 320, height: 320 } as DOMRect)
        : nichts
    },
  )
}

describe('Ein Punkt am Zeiger', () => {
  it('folgt ihm', async () => {
    maesseStellen()

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

    const knopf = await screen.findByRole('button', { name: texte.dreieck })
    knopf.click()

    const erster = await screen.findByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    // **`setPointerCapture` gibt es in jsdom nicht** — es ist Teil der
    // Zeiger-API, die dort nur halb nachgebaut ist. Der Lauf stellt sie, statt
    // den Aufruf im Quelltext zu vermeiden: er ist der Grund, warum das Ziehen
    // über den Rand hinaus funktioniert.
    Object.assign(erster, {
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
      erster.dispatchEvent(ereignis)
    }

    zeiger('pointerdown', 160, 32)
    zeiger('pointermove', 288, 96)

    // 288 / 320 = 0,9 — und zurück auf die Fläche: 288.
    await waitFor(() => {
      const gezogen = screen.getByRole('button', {
        name: 'Punkt 1 von Dreieck',
      })

      expect(Number(gezogen.getAttribute('cx'))).toBeCloseTo(288, 6)
      expect(Number(gezogen.getAttribute('cy'))).toBeCloseTo(96, 6)
    })

    zeiger('pointerup', 288, 96)

    // **Nach dem Loslassen bewegt sich nichts mehr.**
    zeiger('pointermove', 32, 32)

    expect(
      Number(
        screen
          .getByRole('button', { name: 'Punkt 1 von Dreieck' })
          .getAttribute('cx'),
      ),
    ).toBeCloseTo(288, 6)
  })

  // **Zwei Bewegungen hintereinander, und der Punkt steht am Zeiger.**
  //
  // # Der Befund
  //
  // Der erste Anlauf merkte sich den Punkt beim **Aufsetzen** und rechnete
  // jede Bewegung gegen diesen Anfangswert — addiert wurde die Differenz aber
  // auf den **schon verschobenen** Punkt.
  //
  // Von 0,5 auf 0,6 gibt `+0,1`; von 0,6 auf 0,7 gibt gegen den **alten** Wert
  // `+0,2`, addiert auf 0,6 also **0,8**. Der Punkt läuft dem Zeiger davon,
  // und mit jeder Bewegung weiter.
  //
  // **Der erste Lauf schickte genau eine Bewegung** und konnte es deshalb
  // nicht sehen — ein Prüfstück, das seinen Fall nicht enthält, prüft ihn
  // nicht. Angestrichen vom Nachlauf des Zweitblicks über Phase X.
  it('folgt ihm auch beim zweiten Mal', async () => {
    maesseStellen()

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

    ;(await screen.findByRole('button', { name: texte.dreieck })).click()

    const erster = await screen.findByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    Object.assign(erster, { setPointerCapture: vi.fn() })

    const zeiger = (art: string, x: number, y: number) => {
      const ereignis = new MouseEvent(art, {
        bubbles: true,
        clientX: x,
        clientY: y,
      })

      Object.defineProperty(ereignis, 'pointerId', { value: 1 })
      erster.dispatchEvent(ereignis)
    }

    zeiger('pointerdown', 160, 32)
    zeiger('pointermove', 192, 32)

    await waitFor(() => {
      expect(Number(erster.getAttribute('cx'))).toBeCloseTo(192, 6)
    })

    // **Die zweite Bewegung**, und der Punkt steht wieder genau am Zeiger.
    zeiger('pointermove', 224, 32)

    await waitFor(() => {
      expect(Number(erster.getAttribute('cx'))).toBeCloseTo(224, 6)
    })

    zeiger('pointerup', 224, 32)
  })

  // **Verschwindet das Teil unter dem Zeiger, geschieht nichts.**
  //
  // Der Fall ist erreichbar: „Teil löschen" ist ein eigener Knopf, und wer ihn
  // mit der Tastatur auslöst, während die Maus noch einen Punkt hält, hat
  // genau diesen Zustand. Ein `undefined` weiterzurechnen gäbe `NaN` in der
  // Zeichnung — und ein `NaN` fällt erst auf, wenn jemand die Datei öffnet.
  it('rechnet nicht weiter, wenn das Teil verschwindet', async () => {
    maesseStellen()

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

    ;(await screen.findByRole('button', { name: texte.dreieck })).click()

    const erster = await screen.findByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    Object.assign(erster, { setPointerCapture: vi.fn() })

    const zeiger = (art: string, x: number, y: number) => {
      const ereignis = new MouseEvent(art, {
        bubbles: true,
        clientX: x,
        clientY: y,
      })

      Object.defineProperty(ereignis, 'pointerId', { value: 1 })
      erster.dispatchEvent(ereignis)
    }

    zeiger('pointerdown', 160, 32)

    // **Das Teil geht, während der Zeiger noch hängt.**
    screen.getByRole('button', { name: 'Dreieck löschen' }).click()

    await screen.findByText(texte.keineTeile)

    // **Und die Bewegung tut nichts** — kein Absturz, keine Zeichnung.
    zeiger('pointermove', 288, 96)

    expect(screen.getByText(texte.keineTeile)).toBeInTheDocument()
  })
})

// **Die Griffe einer Kurve sind sichtbar und bedienbar** —
// #890.
//
// # Der Befund
//
// `service.GriffZiehen` und `service.SperreSetzen` standen gerechnet und
// geprüft da, **ohne Aufrufer** — und die Zeichenfläche zeichnete keinen
// einzigen Griff. Eine Rundung ließ sich nur über die Grundform herstellen und
// danach nie wieder ändern. Angestrichen vom Zweitblick über Phase X.
//
// **Nur am gewählten Teil**, so wie MSE es tut
// (`src/gui/symbol/point_editor.cpp`, Fassung f9d9356 vom 2021-07-06): alle
// Griffe aller Teile zugleich wären ein Netz, in dem niemand den einen findet.

// **Ein Zug rührt die anderen Teile nicht an.**
//
// # Warum das eine eigene Prüfung braucht
//
// `setzenAuf` läuft über **alle** Teile und ersetzt eines. Der Zweig, der ein
// fremdes Teil unverändert zurückgibt, wird von einem Blatt mit nur einer Form
// nie betreten — und ein Zweig, den kein Aufruf betritt, ist eine Behauptung
// (§3). Ein Symbol besteht regelmäßig aus mehreren Formen; wer die zweite
// zöge und dabei die erste verschöbe, sähe es erst im Set.
describe('Ein Zug an einem Teil von zweien', () => {
  it('lässt das andere stehen', async () => {
    maesseStellen()

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

    const knopf = await screen.findByRole('button', { name: texte.dreieck })

    knopf.click()
    await screen.findByRole('button', { name: 'Punkt 1 von Dreieck' })

    // **Zweimal dieselbe Grundform** — das gibt zwei Teile mit gleichen
    // Punkten, und ein Zug am zweiten muss den ersten in Ruhe lassen.
    knopf.click()

    await waitFor(() => {
      expect(
        screen.getAllByRole('button', { name: 'Punkt 1 von Dreieck' }),
      ).toHaveLength(2)
    })

    const punkte = screen.getAllByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    const [erstes, zweites] = punkte
    if (erstes === undefined || zweites === undefined) {
      throw new Error('es stehen nicht zwei Punkte da')
    }

    const vorher = erstes.getAttribute('cx')

    Object.assign(zweites, { setPointerCapture: vi.fn() })

    const zeiger = (art: string, x: number, y: number) => {
      const ereignis = new MouseEvent(art, {
        bubbles: true,
        clientX: x,
        clientY: y,
      })

      Object.defineProperty(ereignis, 'pointerId', { value: 1 })
      zweites.dispatchEvent(ereignis)
    }

    zeiger('pointerdown', 160, 32)
    zeiger('pointermove', 96, 32)

    await waitFor(() => {
      expect(Number(zweites.getAttribute('cx'))).toBeCloseTo(96, 6)
    })

    expect(erstes.getAttribute('cx')).toBe(vorher)

    zeiger('pointerup', 96, 32)
  })
})
