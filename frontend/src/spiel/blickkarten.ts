// **Was ein Blick gezeigt hat** —
// #1836.
//
// # Warum eine eigene Datei
//
// **Weil die private Ergänzung eines Ereignisses unbekannt getypt ankommt**
// (`Record<string, unknown>`): sie ist, was der Kern für genau einen
// Empfänger dazulegt, und die Naht verspricht keine Gestalt. Was hier steht,
// liest sie so weit, wie sie trägt — und was nicht trägt, fällt heraus, statt
// die Liste umzuwerfen (§8, fremde Daten sind der Normalfall).
//
// # Wie der Kern sie legt
//
// `Sandkasten.blicken` (`internal/service/sandkastenzonen.go`): unter
// `karten` je Karte `objekt`, `karte`, `name` und `ordnung` — die Eins liegt
// oben (CR 401.1 hält die Bibliothek geordnet).

// Gesehene ist eine Karte aus einem Blick.
export interface Gesehene {
  objekt: string
  name: string
  ordnung: number
}

// gesehene liest die Karten eines Blicks aus der privaten Ergänzung —
// der Reihe nach, von oben.
//
// **Der Name fällt auf die Kartenkennung zurück und die auf die
// Objektkennung**: ein leerer Eintrag sähe aus wie ein Fehler der Seite,
// und eine Kennung ist besser als nichts (§8).
export function gesehene(privat: Record<string, unknown>): Gesehene[] {
  const karten = privat.karten

  if (!Array.isArray(karten)) {
    return []
  }

  return karten
    .filter(
      (eine): eine is Record<string, unknown> =>
        typeof eine === 'object' && eine !== null,
    )
    .map((eine) => ({
      objekt: text(eine.objekt),
      name: text(eine.name) || text(eine.karte) || text(eine.objekt),
      ordnung: typeof eine.ordnung === 'number' ? eine.ordnung : 0,
    }))
    .sort((eine, andere) => eine.ordnung - andere.ordnung)
}

// text gibt einen Wert als Zeichenkette — leer, wo er keine ist.
function text(wert: unknown): string {
  return typeof wert === 'string' ? wert : ''
}
