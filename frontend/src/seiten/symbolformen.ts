// **Die vier Grundformen, wie die Oberfläche sie anlegt** —
// #890.
//
// # Warum sie hier noch einmal stehen
//
// Sie stehen **nicht** noch einmal. Gerechnet wird in `internal/service`
// (`Grundform`, portiert aus `SymbolBasicShapeEditor::makeCenteredShape`);
// hier steht nur, **welche** vier es gibt und wie sie heißen — der Rest ist
// ein Aufruf.
//
// **Das ist die Trennung, die §2 verlangt**: eine Kreisformel im Frontend wäre
// eine zweite Antwort auf dieselbe Frage, und die beiden liefen auseinander.

import type { SymbolteilDTO } from '../antworten'

// Formart ist eine der vier Grundformen — dieselben Namen wie in `service`.
export type Formart = 'Kreis' | 'Rechteck' | 'Vieleck' | 'Stern'

// alsPfad macht aus den Punkten eines Teils einen SVG-Pfad.
//
// **Der Zug ist geschlossen** — die Strecke vom letzten zum ersten Punkt
// gehört dazu, so wie bei MSE (`SymbolShape::getPoint` läuft über den Rand,
// `src/data/symbol.hpp`, Zeilen 191–194, Fassung f9d9356 vom 2021-07-06).
//
// **Die Griffe sind relativ**, also werden sie addiert. Wer sie absolut läse,
// bekäme eine Kurve, die am Ursprung klebt.
export function alsPfad(teil: SymbolteilDTO, groesse: number): string {
  const punkte = teil.punkte

  // **Die Leerprüfung ist zugleich die Verengung.** Ein zweites
  // `if (erster === undefined)` daneben wäre ein Zweig, den kein Aufruf
  // betreten kann — und ein solcher ist eine Behauptung (§3). Der Prüfer
  // verlangt ihn nur, weil `punkte[0]` unter `noUncheckedIndexedAccess`
  // `undefined` sein darf; hier beantwortet **eine** Prüfung beide Fragen.
  const erster = punkte[0]
  if (erster === undefined) {
    return ''
  }

  const x = (wert: number) => (wert * groesse).toFixed(2)

  let pfad = `M ${x(erster.x)} ${x(erster.y)}`

  for (const [stelle, von] of punkte.entries()) {
    // **Der Nachfolger des letzten ist der erste** — der Zug ist geschlossen.
    // Das `??` ist deshalb kein Notnagel, sondern die Regel: es greift bei
    // **jedem** Aufruf genau einmal, an der Schlussstrecke.
    const nach = punkte[stelle + 1] ?? erster

    if (von.kurveDanach) {
      pfad +=
        ` C ${x(von.x + von.griffDanachX)} ${x(von.y + von.griffDanachY)}` +
        ` ${x(nach.x + nach.griffDavorX)} ${x(nach.y + nach.griffDavorY)}` +
        ` ${x(nach.x)} ${x(nach.y)}`
    } else {
      pfad += ` L ${x(nach.x)} ${x(nach.y)}`
    }
  }

  return `${pfad} Z`
}

// verschoben gibt einen Punkt, der um `dx`/`dy` gewandert ist.
//
// **Auf das Einheitsquadrat begrenzt.** Eine `.mse-symbol` rechnet in
// `[0…1]`; ein Punkt außerhalb wäre eine Form, die aus ihrem eigenen Rahmen
// fällt, und MSEs Editor lässt ihn nicht dorthin.
export function verschoben(
  teil: SymbolteilDTO,
  stelle: number,
  dx: number,
  dy: number,
): SymbolteilDTO {
  const punkte = teil.punkte.map((punkt, wo) =>
    wo === stelle
      ? { ...punkt, ...imQuadrat(punkt.x + dx, punkt.y + dy) }
      : punkt,
  )

  return { ...teil, punkte }
}

// gesetzt stellt einen Punkt **auf** eine Stelle.
//
// **Der Unterschied zu `verschoben` ist nicht die Rechnung, sondern wer
// führt.** Eine Verschiebung braucht einen Ausgangswert; wer ihn sich merkt,
// hat ihn ab der zweiten Bewegung falsch. Eine Stelle braucht keinen (#890).
export function gesetzt(
  teil: SymbolteilDTO,
  stelle: number,
  x: number,
  y: number,
): SymbolteilDTO {
  const punkte = teil.punkte.map((punkt, wo) =>
    wo === stelle ? { ...punkt, ...imQuadrat(x, y) } : punkt,
  )

  return { ...teil, punkte }
}

// imQuadrat klemmt eine Stelle auf das Einheitsquadrat.
function imQuadrat(x: number, y: number): { x: number; y: number } {
  return {
    x: Math.min(1, Math.max(0, x)),
    y: Math.min(1, Math.max(0, y)),
  }
}
