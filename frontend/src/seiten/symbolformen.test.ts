// **Die Zeichenhilfen des Symboleditors** —
// #890.
//
// **Sie rechnen keine Formen.** Das tut `service.Grundform` auf der Go-Seite;
// hier steht, wie aus fertigen Punkten ein SVG-Pfad wird — und was passiert,
// wenn keine da sind.
import { describe, expect, it } from 'vitest'

import type { SymbolpunktDTO, SymbolteilDTO } from '../antworten'
import { alsPfad, verschoben } from './symbolformen'

function punkt(x: number, y: number, kurve = false): SymbolpunktDTO {
  return {
    x,
    y,
    griffDavorX: kurve ? -0.1 : 0,
    griffDavorY: 0,
    griffDanachX: kurve ? 0.1 : 0,
    griffDanachY: 0,
    kurveDanach: kurve,
    sperre: 'free',
  }
}

function teil(punkte: SymbolpunktDTO[]): SymbolteilDTO {
  return {
    art: 'Form',
    name: 'Probe',
    kombination: 'overlap',
    punkte,
    spiegelart: '',
    kopien: 0,
    mitteX: 0,
    mitteY: 0,
    griffX: 0,
    griffY: 0,
    teile: [],
  }
}

describe('Der Pfad einer Form', () => {
  // **Ein Teil ohne Punkte gibt keinen Pfad.**
  //
  // Er kommt vor: eine Symmetrie hat keine eigenen Punkte, sondern Kinder.
  // Ein `M NaN NaN` daraus wäre ein SVG, das der Browser stillschweigend
  // verwirft — und niemand wüsste, warum nichts dasteht.
  it('ist leer, wenn es keine Punkte gibt', () => {
    expect(alsPfad(teil([]), 320)).toBe('')
  })

  it('zieht Geraden und schließt den Zug', () => {
    const pfad = alsPfad(teil([punkt(0, 0), punkt(1, 0), punkt(1, 1)]), 100)

    expect(pfad).toBe('M 0.00 0.00 L 100.00 0.00 L 100.00 100.00 L 0.00 0.00 Z')
  })

  // **Eine Kurve nimmt beide Griffe, und sie sind relativ.**
  //
  // Wer sie absolut läse, bekäme eine Kurve, die am Ursprung klebt.
  it('zieht eine Kurve über beide Griffe', () => {
    // **Der zweite Punkt trägt seinen Griff davor**, obwohl er selbst keine
    // Kurve *danach* setzt: eine Strecke hat zwei Enden, und der Griff des
    // Endpunkts gehört zu ihr. Der erste Anlauf ließ ihn weg und erwartete
    // ihn trotzdem.
    const zweiter = { ...punkt(1, 0), griffDavorX: -0.1 }

    const pfad = alsPfad(teil([punkt(0, 0, true), zweiter]), 100)

    // Der Griff nach dem ersten Punkt liegt bei 0 + 0,1; der vor dem zweiten
    // bei 1 − 0,1.
    expect(pfad).toContain('C 10.00 0.00 90.00 0.00 100.00 0.00')
  })
})

describe('Ein verschobener Punkt', () => {
  it('wandert um den angegebenen Betrag', () => {
    const bewegt = verschoben(teil([punkt(0.5, 0.5)]), 0, 0.1, -0.2)

    expect(bewegt.punkte[0]?.x).toBeCloseTo(0.6, 6)
    expect(bewegt.punkte[0]?.y).toBeCloseTo(0.3, 6)
  })

  // **Und nicht über den Rand.**
  //
  // Eine `.mse-symbol` rechnet in `[0…1]`; ein Punkt außerhalb wäre eine Form,
  // die aus ihrem eigenen Rahmen fällt.
  it('bleibt im Einheitsquadrat', () => {
    const oben = verschoben(teil([punkt(0.5, 0.05)]), 0, 0, -0.5)
    expect(oben.punkte[0]?.y).toBe(0)

    const rechts = verschoben(teil([punkt(0.95, 0.5)]), 0, 0.5, 0)
    expect(rechts.punkte[0]?.x).toBe(1)
  })

  // **Die anderen Punkte bleiben, wo sie sind.**
  it('lässt die anderen Punkte stehen', () => {
    const bewegt = verschoben(teil([punkt(0, 0), punkt(1, 1)]), 1, 0.1, 0)

    expect(bewegt.punkte[0]?.x).toBe(0)
    expect(bewegt.punkte[1]?.x).toBeCloseTo(1, 6)
  })
})
