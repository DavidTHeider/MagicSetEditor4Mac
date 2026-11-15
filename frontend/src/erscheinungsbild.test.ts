// Erscheinungsbild und Schriftgröße am Wurzelelement —
// #245.
//
// **Beides steht nicht in React**, weil beides für alles gilt — auch für den
// Rollbalken, den die Anwendung gar nicht zeichnet. Geprüft wird deshalb am
// Element und nicht an einer Komponente.
import { beforeEach, describe, expect, it } from 'vitest'

import {
  anzeigeschriften,
  groessen,
  istStufe,
  setzeAnzeigeschrift,
  setzeErscheinungsbild,
  setzeSchriftgroesse,
} from './erscheinungsbild'

let wurzel: HTMLElement

beforeEach(() => {
  wurzel = document.createElement('html')
})

describe('setzeErscheinungsbild', () => {
  it.each(['hell', 'dunkel'])('schreibt %s ans Element', (wahl) => {
    setzeErscheinungsbild(wurzel, wahl)

    expect(wurzel.getAttribute('data-erscheinungsbild')).toBe(wahl)
  })

  // **`system` entfernt das Attribut, statt es zu setzen.** Der Selektor in
  // `stil.css` fragt `html:not([data-erscheinungsbild])`; ein Attribut mit
  // irgendeinem Wert schlösse die Systemeinstellung aus — und `system` wäre
  // dann der einzige Wert, der nicht wirkt.
  it('nimmt das Attribut für die Systemeinstellung weg', () => {
    setzeErscheinungsbild(wurzel, 'dunkel')
    setzeErscheinungsbild(wurzel, 'system')

    expect(wurzel.hasAttribute('data-erscheinungsbild')).toBe(false)
  })

  // Dasselbe für einen Wert, den es nicht gibt: er kann nur aus einem Bestand
  // kommen, den jemand von Hand geändert hat — `domain.PruefeEinstellung`
  // lässt ihn nicht durch.
  it('nimmt das Attribut auch für einen unbekannten Wert weg', () => {
    setzeErscheinungsbild(wurzel, 'dunkel')
    setzeErscheinungsbild(wurzel, 'grau')

    expect(wurzel.hasAttribute('data-erscheinungsbild')).toBe(false)
  })
})

describe('setzeSchriftgroesse', () => {
  it.each(Object.keys(groessen))('setzt die Stufe %s', (stufe) => {
    setzeSchriftgroesse(wurzel, stufe)

    expect(wurzel.style.getPropertyValue('--schriftgroesse')).toBe(
      groessen[stufe as keyof typeof groessen],
    )
  })

  // **Eine unbekannte Stufe fällt auf `normal` zurück und nicht auf nichts.**
  // Eine Oberfläche, die daraufhin gar keine Schriftgröße setzte, erbte die des
  // Browsers und sähe zufällig richtig aus — bis jemand seine Vorgabe ändert.
  it('fällt bei einer unbekannten Stufe auf normal zurück', () => {
    setzeSchriftgroesse(wurzel, 'riesig')

    expect(wurzel.style.getPropertyValue('--schriftgroesse')).toBe(
      groessen.normal,
    )
  })

  // **Die Schritte sind 12,5 %** — groß genug, dass man sie sieht, klein
  // genug, dass keine Zeile springt. Der Test hält die Abstände fest, nicht
  // die Zahlen: wer eine Stufe verschiebt, verschiebt sie gleichmäßig.
  it('hält gleichmäßige Schritte', () => {
    const zahlen = Object.values(groessen).map((wert) => parseFloat(wert))

    for (let i = 1; i < zahlen.length; i++) {
      expect((zahlen[i] ?? 0) - (zahlen[i - 1] ?? 0)).toBeCloseTo(12.5)
    }
  })
})

describe('istStufe', () => {
  it('erkennt eine Stufe', () => {
    expect(istStufe('gross')).toBe(true)
  })

  it('weist ab, was keine ist', () => {
    expect(istStufe('riesig')).toBe(false)
  })
})

// Die Anzeigeschrift — Werkstück J zu
// #1325.
//
// **Sie steht neben der Schriftgröße und nicht in einer eigenen Datei**: es
// ist dieselbe Frage — was gilt für jede Zeile, ohne dass eine Komponente
// davon weiß.
describe('setzeAnzeigeschrift', () => {
  it.each(Object.keys(anzeigeschriften))('setzt die Schrift %s', (wahl) => {
    setzeAnzeigeschrift(wurzel, wahl)

    expect(wurzel.style.getPropertyValue('--font-anzeige')).toBe(
      anzeigeschriften[wahl as keyof typeof anzeigeschriften],
    )
  })

  // **Jede Kette endet bei `serif`.** Die vier Schriften liegen auf jedem Mac
  // — aber Windows und Linux werden übersetzt (§12), und dort entscheidet der
  // Rückfall, ob eine Überschrift eine Serifenschrift ist oder was der Browser
  // gerade für richtig hält.
  it.each(Object.entries(anzeigeschriften))(
    'endet bei %s auf einer allgemeinen Familie',
    (_wahl, kette) => {
      expect(kette.endsWith('serif')).toBe(true)
    },
  )

  // **Eine unbekannte Wahl fällt auf Pergament zurück** — aus demselben Grund
  // wie bei der Schriftgröße: sie kann nur aus einem von Hand geänderten
  // Bestand kommen, und gar keine Schrift zu setzen sähe zufällig richtig aus.
  it('fällt bei einer unbekannten Wahl auf Pergament zurück', () => {
    setzeAnzeigeschrift(wurzel, 'comic')

    expect(wurzel.style.getPropertyValue('--font-anzeige')).toBe(
      anzeigeschriften.pergament,
    )
  })
})
