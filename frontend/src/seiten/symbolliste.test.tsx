// **Die Teilliste des Symboleditors und das Ablegen** —
// #890,
// Werkstück 3 von Phase X.
//
// # Was hier geprüft wird
//
// **Hier steht die Teilliste und das Ablegen** — die Zeichenfläche und die
// Tastatur stehen in `symbol.test.tsx`. Der Schnitt folgt der Frage: dort
// *wie gezeichnet wird*, hier *was mit dem Gezeichneten geschieht*.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Symbol } from './Symbol'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.symbol

// dreieck ist die Antwort, die die Go-Seite auf einen Grundformknopf gibt.
//
// **Drei Punkte, und die Zahlen sind die echten** — `service.Grundform` gibt
// sie für ein Dreieck über der Mitte (0,5|0,5) mit halber Größe 0,4.
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

// bei liest eine Koordinate eines Punktes als Zahl.
function bei(name: string, achse: 'cx' | 'cy') {
  return Number(screen.getByRole('button', { name }).getAttribute(achse))
}

// ohneSymbol ist die Antwort eines Sets, das noch keines hat.
function ohneSymbol() {
  return { felder: [], werte: [], fehler: null }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Symbol setID="7" />
    </Sprachraum>,
  )
}

describe('Die Teilliste', () => {
  it('legt die Zeichnung ins Set und sagt, wie sie heißt', async () => {
    const SymbolSetzen = vi.fn().mockResolvedValue({
      teile: [],
      name: 'symbol1.mse-symbol',
      fehler: null,
    })

    belegeApi({
      Grundform: vi.fn().mockResolvedValue(dreieck()),
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
      SymbolSetzen,
      SetwertSetzen: vi.fn().mockResolvedValue({ fehler: null }),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))
    await screen.findByRole('button', { name: 'Punkt 1 von Dreieck' })

    await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

    await waitFor(() => {
      expect(SymbolSetzen).toHaveBeenCalled()
    })

    expect(await screen.findByText(/symbol1\.mse-symbol/)).toBeInTheDocument()
  })

  // **Zwei Teile, und nur eines bewegt sich.**
  //
  // Wer alle Teile durch dieselbe Verschiebung schickte, verschöbe die ganze
  // Zeichnung — und es sähe aus, als bewegte sich nichts.
  it('lässt das andere Teil in Ruhe', async () => {
    const Grundform = vi
      .fn()
      .mockResolvedValueOnce(dreieck())
      .mockResolvedValueOnce({
        teile: dreieck().teile.map((teil) => ({ ...teil, name: 'Rechteck' })),
        name: '',
        fehler: null,
      })

    belegeApi({
      Grundform,
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))
    await screen.findByRole('button', { name: 'Punkt 1 von Dreieck' })

    await userEvent.click(screen.getByRole('button', { name: texte.rechteck }))
    await screen.findByRole('button', { name: 'Punkt 1 von Rechteck' })

    screen.getByRole('button', { name: 'Punkt 1 von Rechteck' }).focus()
    await userEvent.keyboard('{Shift>}{ArrowRight}{/Shift}')

    await waitFor(() => {
      expect(bei('Punkt 1 von Rechteck', 'cx')).toBeCloseTo(192, 6)
    })

    // **Das Dreieck steht, wo es stand** — 0,5 × 320.
    expect(bei('Punkt 1 von Dreieck', 'cx')).toBe(160)
  })

  // **Ein Klick in der Teilliste wählt das Teil.**
  //
  // Die Auswahl entscheidet, welches Teil hervorgehoben ist — und
  // `aria-pressed` sagt es einem Vorleseprogramm. Ohne diesen Lauf wäre die
  // Liste eine Aufzählung, auf die niemand zeigen kann.
  //
  // **Bis zum 2026-08-27 stand hier `aria-current`**, und das war richtig,
  // solange es eine Einzelwahl war: „current" ist genau **eines** aus einer
  // Menge. Mit der Mehrfachwahl (#1168) nennte ein Vorleser sonst drei Zeilen
  // zugleich „aktuell".
  it('wählt ein Teil über die Liste', async () => {
    const Grundform = vi
      .fn()
      .mockResolvedValueOnce(dreieck())
      .mockResolvedValueOnce({
        teile: dreieck().teile.map((teil) => ({ ...teil, name: 'Rechteck' })),
        name: '',
        fehler: null,
      })

    belegeApi({
      Grundform,
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))
    await screen.findByRole('button', { name: 'Punkt 1 von Dreieck' })

    await userEvent.click(screen.getByRole('button', { name: texte.rechteck }))
    await screen.findByRole('button', { name: 'Punkt 1 von Rechteck' })

    const dreieckszeile = screen.getByRole('button', {
      name: /Dreieck — 3/,
    })

    // Das zweite Teil ist gewählt, weil es zuletzt kam.
    expect(dreieckszeile).toHaveAttribute('aria-pressed', 'false')

    await userEvent.click(dreieckszeile)

    expect(screen.getByRole('button', { name: /Dreieck — 3/ })).toHaveAttribute(
      'aria-pressed',
      'true',
    )
  })

  // **Ein Teil lässt sich löschen, und die Auswahl rutscht mit.**
  //
  // Ohne das Nachrücken zeigte die Liste auf ein Teil, das es nicht mehr gibt —
  // und die Zeichenfläche hätte nichts hervorzuheben.
  it('löscht ein Teil', async () => {
    belegeApi({
      Grundform: vi.fn().mockResolvedValue(dreieck()),
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))
    await screen.findByRole('button', { name: 'Punkt 1 von Dreieck' })

    await userEvent.click(
      screen.getByRole('button', { name: 'Dreieck löschen' }),
    )

    expect(await screen.findByText(texte.keineTeile)).toBeInTheDocument()
  })

  // **Und die Auswahl rutscht auf das vorige Teil.**
  //
  // Ohne das Nachrücken zeigte sie auf eine Stelle, die es nicht mehr gibt —
  // und die Zeichenfläche hätte nichts hervorzuheben.
  it('rückt die Auswahl nach, wenn ein späteres Teil geht', async () => {
    const Grundform = vi
      .fn()
      .mockResolvedValueOnce(dreieck())
      .mockResolvedValueOnce({
        teile: dreieck().teile.map((teil) => ({ ...teil, name: 'Rechteck' })),
        name: '',
        fehler: null,
      })

    belegeApi({
      Grundform,
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))
    await screen.findByRole('button', { name: 'Punkt 1 von Dreieck' })

    await userEvent.click(screen.getByRole('button', { name: texte.rechteck }))
    await screen.findByRole('button', { name: 'Punkt 1 von Rechteck' })

    // Das zweite ist gewählt; es wird gelöscht.
    //
    // **Der Knopf des Rechtecks, nicht irgendeiner** — die Zusage ist, dass
    // die Auswahl auf das **vorige** rutscht, und die prüft nur, wer das
    // spätere löscht. **Seit dem 2026-08-27 nennt jede Zeile ihr Teil**
    // (#1168); vorher hießen alle drei Knöpfe jeder Zeile gleich, und der Test
    // musste über den Index zeigen.
    await userEvent.click(
      screen.getByRole('button', { name: 'Rechteck löschen' }),
    )

    // **Das Dreieck steht wieder da, und es ist gewählt.**
    expect(
      await screen.findByRole('button', { name: /Dreieck — 3/ }),
    ).toHaveAttribute('aria-pressed', 'true')
  })

  // **Ein Fehler beim Anlegen steht am Bildschirm und nicht im Log.**
  it('meldet einen Fehler der Grundform', async () => {
    belegeApi({
      Grundform: vi
        .fn()
        .mockRejectedValue(new Error('Der Bestand ist gesperrt.')),
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Bestand ist gesperrt.',
    )
  })

  // **Und die abgelegte Zeichnung wird zum Setsymbol.**
  //
  // Die Datei allein tut nichts: gezeichnet wird, was in `set.symbol` steht.
  // Der erste Anlauf legte nur ab — die Zeichnung lag neben dem Set, und auf
  // der Karte blieb MSEs Standardquadrat. Angestrichen vom Zweitblick über
  // Phase X.
  it('macht die Zeichnung zum Setsymbol', async () => {
    const SetwertSetzen = vi.fn().mockResolvedValue({ fehler: null })

    belegeApi({
      Grundform: vi.fn().mockResolvedValue(dreieck()),
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
      SymbolSetzen: vi.fn().mockResolvedValue({
        teile: [],
        name: 'symbol1.mse-symbol',
        fehler: null,
      }),
      SetwertSetzen,
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))
    await screen.findByRole('button', { name: 'Punkt 1 von Dreieck' })

    await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

    await waitFor(() => {
      expect(SetwertSetzen).toHaveBeenCalledWith(
        '7',
        'symbol',
        'symbol1.mse-symbol',
        null,
      )
    })
  })

  // **Ein gescheitertes Ablegen setzt das Feld gar nicht erst.**
  //
  // §7, Regel 3: ein Verweis auf eine Datei, die nicht liegt, wäre schlimmer
  // als kein Verweis — die Karte zeigte dann nichts und niemand wüsste, warum.
  it('setzt das Feld nicht, wenn das Ablegen scheitert', async () => {
    const SetwertSetzen = vi.fn()

    belegeApi({
      Grundform: vi.fn().mockResolvedValue(dreieck()),
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
      SymbolSetzen: vi
        .fn()
        .mockRejectedValue(new Error('Die Platte ist voll.')),
      SetwertSetzen,
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))
    await screen.findByRole('button', { name: 'Punkt 1 von Dreieck' })

    await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

    await screen.findByRole('alert')

    expect(SetwertSetzen).not.toHaveBeenCalled()
  })

  // **Und ein Fehler beim Ablegen ebenso.**
  //
  // **Der Name bleibt dabei leer** — eine Zeichnung, die nicht abgelegt wurde,
  // darf nicht dastehen, als wäre sie es (§7, Regel 3).
  it('meldet einen Fehler beim Ablegen', async () => {
    belegeApi({
      Grundform: vi.fn().mockResolvedValue(dreieck()),
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
      SymbolSetzen: vi
        .fn()
        .mockRejectedValue(new Error('Die Platte ist voll.')),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))
    await screen.findByRole('button', { name: 'Punkt 1 von Dreieck' })

    await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Die Platte ist voll.',
    )

    expect(screen.queryByText(/mse-symbol/)).not.toBeInTheDocument()
  })
})
