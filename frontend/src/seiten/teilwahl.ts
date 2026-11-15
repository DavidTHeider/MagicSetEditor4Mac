// **Die Mehrfachauswahl in der Teilliste des Symboleditors** — Werkstück 5 von
// Phase 3.7
// (#1168,
// Befund Z5-02).
//
// **Eine eigene Datei und eine reine Funktion**, weil sie eine Regel ist und
// keine Darstellung (§2): welche Teile nach einem Klick gewählt sind, hängt
// von der Taste ab, nicht vom Aussehen der Liste.

// Wahl ist die Menge der gewählten Stellen — **nie leer**.
//
// **Der Typ sagt es, nicht ein `?? 0` beim Lesen.** Die erste Fassung gab ein
// gewöhnliches Array zurück, und jeder Leser brauchte einen Rückfall für den
// leeren Fall — einen Zweig, den kein Lauf betreten kann und den die Abdeckung
// deshalb anstrich (§3).
export type Wahl = readonly [number, ...number[]]

// gewaehltNach gibt die Wahl nach einem Klick auf eine Stelle.
//
// Portiert aus `SymbolPartsSelection::select` (`src/gui/symbol/selection.cpp`,
// Zeilen 25–56, Fassung `f9d9356` vom 2021-07-06). Von MSEs drei Weisen
// braucht es hier zwei: `SELECT_OVERRIDE` für den gewöhnlichen Klick und
// `SELECT_TOGGLE` für den mit gedrückter Befehlstaste. Die dritte,
// `SELECT_IF_OUTSIDE`, gehört zum Ziehen auf der Zeichenfläche — dort wird
// eine bestehende Mehrfachwahl **nicht** zerstört, wenn man eines ihrer Teile
// anfasst.
//
// **Dieselbe Liste zurück, wenn sich nichts ändert.** MSE gibt dafür `false`
// und zeichnet nicht neu; hier hält React still, solange die Referenz dieselbe
// ist.
export function gewaehltNach(
  bisher: Wahl,
  stelle: number,
  hinzu: boolean,
): Wahl {
  if (!hinzu) {
    if (bisher.length === 1 && bisher[0] === stelle) {
      return bisher
    }

    return [stelle]
  }

  if (!bisher.includes(stelle)) {
    return [...bisher, stelle].sort((a, b) => a - b) as unknown as Wahl
  }

  // **Die letzte Wahl bleibt stehen.** Das ist eine Abweichung von MSE, wo die
  // Wahl leer sein darf: dort steht die Zeichenfläche im Vordergrund und
  // arbeitet ohne Wahl weiter. Hier hängen die Griffe daran — ein Klick, der
  // die Bedienung wegnimmt, ist keine.
  if (bisher.length === 1) {
    return bisher
  }

  return bisher.filter((wo) => wo !== stelle) as unknown as Wahl
}

// nachEinemWegfall rückt eine Wahl nach, wenn ein Teil verschwunden ist.
//
// **Ohne das zeigte die Wahl auf ein anderes Teil als vorher** — die Stellen
// hinter dem gelöschten rutschen alle um eins vor. §7, Regel 1 im Kleinen: ein
// Wechsel des Bezugsobjekts räumt, was sich darauf bezog.
export function nachEinemWegfall(bisher: Wahl, gefallen: number): Wahl {
  const geblieben = bisher
    .filter((wo) => wo !== gefallen)
    .map((wo) => (wo > gefallen ? wo - 1 : wo))

  const [erste, ...weitere] = geblieben

  // **Bleibt nichts übrig, steht die erste Stelle** — dieselbe Zusage wie in
  // [gewaehltNach]: ohne Wahl gäbe es keine Griffe, mit denen man weiterarbeitet.
  return erste === undefined ? [0] : [erste, ...weitere]
}
