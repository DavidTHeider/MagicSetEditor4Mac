// **Die Vorrichtungen der Symbolläufe** —
// #890.
//
// # Warum sie hier stehen und nicht je Datei
//
// Die Griffläufe sind auf zwei Dateien verteilt — Tastatur und Zeiger —, weil
// eine Datei über 400 Zeilen gegen §2 liefe. Zwei Kopien derselben
// Vorrichtung liefen mit der Zeit auseinander, und dann prüften zwei Läufe
// verschiedene Kreise, ohne dass es jemandem auffiele.
//
// **Unter `src/test/`, weil das kein Produktivcode ist** — die Abdeckung
// schließt das Verzeichnis aus (`vite.config.ts`).
import { vi } from 'vitest'

import type { SymbolpunktDTO, SymbolteilDTO } from '../antworten'

// punkt ist ein Kontrollpunkt ohne Griffe.
export function punkt(x: number, y: number): SymbolpunktDTO {
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

// **Die Maße werden gestellt.** In jsdom hat kein Element eine Größe;
// `getBoundingClientRect` gibt überall Nullen, und eine Umrechnung durch null
// gäbe `NaN`. Der Lauf stellt deshalb ein 320 × 320 großes SVG an der linken
// oberen Ecke — genau die Maße, die die Zeichenfläche im Browser hat.
export function maesseStellen(): void {
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

// mitKurve ist die Antwort der Go-Seite auf den Kreisknopf, gekürzt auf zwei
// Punkte — der erste trägt beide Griffe und die Sperre `size`.
export function mitKurve(): {
  teile: SymbolteilDTO[]
  name: string
  fehler: null
} {
  return {
    teile: [
      {
        art: 'Form',
        name: 'Kreis',
        kombination: 'overlap',
        punkte: [
          {
            ...punkt(0.5, 0.1),
            kurveDanach: true,
            griffDanachX: 0.2,
            griffDanachY: 0,
            griffDavorX: -0.2,
            griffDavorY: 0,
            sperre: 'size',
          },
          {
            ...punkt(0.9, 0.5),
            kurveDanach: true,
            griffDavorX: 0,
            griffDavorY: -0.2,
          },
        ],
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

// mitSymmetrie ist eine Zeichnung, deren einziges Teil eine Symmetrie ist
// (#1092).
export function mitSymmetrie(): {
  teile: SymbolteilDTO[]
  name: string
  fehler: null
} {
  return {
    teile: [
      {
        art: 'Symmetrie',
        name: 'Quadrat',
        kombination: '',
        punkte: [],
        spiegelart: 'rotation',
        kopien: 2,
        mitteX: 0.5,
        mitteY: 0.5,
        griffX: 0.25,
        griffY: 0,
        teile: [
          {
            art: 'Form',
            name: 'Quadrat',
            kombination: 'overlap',
            punkte: [punkt(0.6, 0.4), punkt(0.8, 0.4), punkt(0.8, 0.6)],
            spiegelart: '',
            kopien: 0,
            mitteX: 0,
            mitteY: 0,
            griffX: 0,
            griffY: 0,
            teile: [],
          },
        ],
      },
    ],
    name: '',
    fehler: null,
  }
}

// zweiKopien ist die Antwort von `Symmetriekopien` zu [mitSymmetrie].
//
// **Die zweite steht gegenüber der Mitte** — eine Drehung um 180°.
export function zweiKopien(): {
  kopien: { punkte: SymbolpunktDTO[] }[][]
  fehler: null
} {
  return {
    kopien: [
      [{ punkte: [punkt(0.6, 0.4), punkt(0.8, 0.4), punkt(0.8, 0.6)] }],
      [{ punkte: [punkt(0.4, 0.6), punkt(0.2, 0.6), punkt(0.2, 0.4)] }],
    ],
    fehler: null,
  }
}
