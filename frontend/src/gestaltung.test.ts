// **Farbwelt, Fensterbild und eigene Merkmale am Wurzelelement** — UX-09
// (#1681).
import { afterEach, describe, expect, it } from 'vitest'

import {
  eigeneMerkmale,
  schleierAnteil,
  setzeEigeneMerkmale,
  setzeFarbwelt,
  setzeHintergrund,
} from './gestaltung'

const wurzel = document.documentElement

afterEach(() => {
  wurzel.removeAttribute('data-farbwelt')
  wurzel.removeAttribute('data-hintergrund')
  wurzel.removeAttribute('style')
})

describe('setzeFarbwelt', () => {
  it('setzt eine Farbwelt und nimmt sie für Pergament wieder weg', () => {
    setzeFarbwelt(wurzel, 'schiefer', false)
    expect(wurzel).toHaveAttribute('data-farbwelt', 'schiefer')

    setzeFarbwelt(wurzel, 'pergament', false)
    expect(wurzel).not.toHaveAttribute('data-farbwelt')
  })

  it('setzt keine unbekannte Farbwelt', () => {
    setzeFarbwelt(wurzel, 'grau', false)
    expect(wurzel).not.toHaveAttribute('data-farbwelt')
  })

  // **Hoher Kontrast sticht die Farbwelt.**
  it('setzt bei hohem Kontrast keine Farbwelt', () => {
    setzeFarbwelt(wurzel, 'moos', true)
    expect(wurzel).not.toHaveAttribute('data-farbwelt')
  })
})

describe('setzeHintergrund', () => {
  it('legt das Fensterbild mit seiner Skalierung unter den Grund', () => {
    setzeHintergrund(wurzel, 'fuellen', '/fensterbild/7.png?stand=2', false)

    expect(wurzel).toHaveAttribute('data-hintergrund', 'fuellen')
    expect(wurzel.style.getPropertyValue('--mse-fensterbild')).toBe(
      'url("/fensterbild/7.png?stand=2")',
    )
  })

  // **Keine fremde Adresse** (AK-04): was nicht der eigene Bildweg ist, wird
  // kein `url()`.
  it('nimmt keine fremde Adresse an', () => {
    for (const fremd of [
      'https://example.org/a.png',
      '/fensterbild/7.png") , url("https://example.org/a.png',
      '/tischbild/7.png',
      '',
    ]) {
      setzeHintergrund(wurzel, 'fuellen', fremd, false)
      expect(wurzel).not.toHaveAttribute('data-hintergrund')
      expect(wurzel.style.getPropertyValue('--mse-fensterbild')).toBe('')
    }
  })

  it('nimmt das Bild bei »aus«, bei unbekannter Stellung und bei hohem Kontrast weg', () => {
    for (const [modus, hoch] of [
      ['aus', false],
      ['riesig', false],
      ['kacheln', true],
    ] as const) {
      setzeHintergrund(wurzel, 'einpassen', '/fensterbild/7.png', false)
      setzeHintergrund(wurzel, modus, '/fensterbild/7.png', hoch)
      expect(wurzel).not.toHaveAttribute('data-hintergrund')
    }
  })
})

describe('eigeneMerkmale', () => {
  it('liest bekannte Merkmale mit Hexfarben, eine je Zeile', () => {
    expect(eigeneMerkmale('--mse-grund: #ABC;\n\n --mse-text:#112233')).toEqual(
      {
        angaben: [
          ['--mse-grund', '#abc'],
          ['--mse-text', '#112233'],
        ],
      },
    )
  })

  // **Dieselbe Form wie in Go** — eine Zeile, die sie verfehlt, setzt
  // nichts, auch die übrigen nicht.
  it('weist alles andere ab und nennt die Zeile', () => {
    for (const [falsch, zeile] of [
      ['--mse-grund: url(x);', 1],
      ['--mse-tisch-grund: #111111;', 1],
      ['color: #111111;', 1],
      ['--mse-grund: #111111;\n--mse-grund: #222222;', 2],
      ['--mse-grund: #111111;\n\n--mse-grund: #111; } body { x', 3],
    ] as const) {
      expect(eigeneMerkmale(falsch), falsch).toEqual({ zeile })
    }
  })
})

describe('setzeEigeneMerkmale', () => {
  const pergament: Record<string, string> = {
    '--mse-grund': '#f2e8d5',
    '--mse-flaeche': '#fdf9f0',
    '--mse-text': '#241c14',
    '--mse-leise': '#5c4a34',
    '--mse-kante': '#786248',
    '--mse-marke': '#6b4423',
    '--mse-betont': '#e8dcc2',
    '--mse-ring': '#6b4423',
    '--mse-marketext': '#fdf9f0',
    '--mse-warnung': '#8c2f16',
  }

  // hell ist die Lage der meisten Läufe: helles Bild, kein Fensterbild.
  const hell = { dunkel: false }

  // lies gibt, was am Wurzelelement gesetzt ist, sonst Pergament — jsdom
  // kennt die Stilblätter nicht.
  const lies = (merkmal: string) =>
    wurzel.style.getPropertyValue(merkmal) || (pergament[merkmal] ?? '')

  it('setzt lesbare eigene Farben', () => {
    const befunde = setzeEigeneMerkmale(
      wurzel,
      '--mse-marke: #4e2f16;',
      false,
      hell,
      lies,
    )

    expect(befunde).toEqual([])
    expect(wurzel.style.getPropertyValue('--mse-marke')).toBe('#4e2f16')
  })

  // **Die Sperre** (AK-03): unlesbar heißt nicht gesetzt — und der Befund
  // sagt, welches Paar es war.
  it('nimmt unlesbare eigene Farben zurück und nennt das Paar', () => {
    const befunde = setzeEigeneMerkmale(
      wurzel,
      '--mse-text: #efe5d0;\n--mse-marke: #4e2f16;',
      false,
      hell,
      lies,
    )

    expect(befunde).toContainEqual(
      expect.objectContaining({ vorn: '--mse-text', hinten: '--mse-grund' }),
    )
    expect(wurzel.style.getPropertyValue('--mse-text')).toBe('')
    expect(wurzel.style.getPropertyValue('--mse-marke')).toBe('')
  })

  it('setzt bei hohem Kontrast und bei ungültiger Form nichts — und räumt Früheres ab', () => {
    setzeEigeneMerkmale(wurzel, '--mse-marke: #4e2f16;', false, hell, lies)

    expect(
      setzeEigeneMerkmale(wurzel, '--mse-marke: #4e2f16;', true, hell, lies),
    ).toEqual([])
    expect(wurzel.style.getPropertyValue('--mse-marke')).toBe('')

    expect(
      setzeEigeneMerkmale(wurzel, 'color: red', false, hell, lies),
    ).toEqual([])
    expect(wurzel.style.getPropertyValue('--mse-marke')).toBe('')
  })

  // **Ohne mitgegebenes Lesen** misst sie am gerechneten Stil — unter jsdom
  // gibt es dort nur, was am Element selbst steht.
  it('misst ohne Vorgabe am gerechneten Stil des Wurzelelements', () => {
    expect(setzeEigeneMerkmale(wurzel, '--mse-grund: #ffffff;', false)).toEqual(
      [],
    )
    expect(wurzel.style.getPropertyValue('--mse-grund')).toBe('#ffffff')
  })

  // **Über eigenen Farben keine Faser** — angestrichen vom Zweitblick über
  // Phase 6.3 (S4): auf Pergament kostete sie rund 16 % des Verhältnisses,
  // und die Sperre misst reine Farben.
  it('nimmt über eigenen Farben die Faser weg und gibt sie danach zurück', () => {
    setzeEigeneMerkmale(wurzel, '--mse-marke: #4e2f16;', false, hell, lies)
    expect(wurzel.style.getPropertyValue('--mse-faser')).toBe('none')

    setzeEigeneMerkmale(wurzel, '', false, hell, lies)
    expect(wurzel.style.getPropertyValue('--mse-faser')).toBe('')
  })
})

// **Der Schleier kommt aus dem Stilblatt** (`--mse-schleier`) — und ohne
// lesbaren Wert gilt der ganze Grund, statt dass eine Zahl geraten wird.
describe('schleierAnteil', () => {
  it('liest den Anteil als Prozent, und ohne Wert gilt der ganze Grund', () => {
    expect(schleierAnteil(wurzel)).toBe(1)

    wurzel.style.setProperty('--mse-schleier', '90%')

    expect(schleierAnteil(wurzel)).toBeCloseTo(0.9, 5)
  })
})
