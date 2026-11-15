// **Der Zug und sein Ende** — Werkstücke G9 und M17 von Phase 5.5.
//
// **Eine eigene Datei**, weil `apipartie.ts` mit ihnen über die Zeilengrenze
// aus §2 ginge — und weil der Schnitt der Frage folgt: dort steht, **was** an
// einem Tisch liegt, hier, **wo die Partie im Zug steht** und **wie sie
// ausgegangen ist**.

// Zug ist, wo die Partie im Zug steht — G9,
// #1592.
//
// **Für jeden derselbe**: wer am Zug ist und wer Priorität hat, ist am Tisch
// öffentlich (CR 102.1, CR 117). **Null in `nummer` heißt: es geht nicht
// zugweise zu** — die Partien aus G1 bis G8 sind Sandkästen.
export interface Zug {
  nummer: number
  aktiv: string
  starter: string
  phase: string
  schritt: string

  // prioritaet ist, wer gerade handeln darf — leer heißt: niemand
  // (CR 502.4, CR 514.3).
  prioritaet: string

  // abwurf ist, wie viele Karten der aktive Spieler noch abwerfen muss
  // (CR 514.1).
  abwurf: number

  gepasst: string[]

  // angriffe und bloecke sind der Kampfstand — CR 508.1, CR 509.1
  // (#1623).
  //
  // **M6 hat den Kampf gebaut, und keine Fläche zeigte ihn.** Wer blocken
  // will, muss wissen, wer angreift.
  angriffe: Kampfpaar[]
  bloecke: Kampfpaar[]
}

// Kampfpaar ist eine Kreatur und das, wogegen sie steht.
//
// **Ein Verbund für beide Richtungen**: bei einem Angriff steht `ziel`
// (CR 508.1a), bei einem Block `angreifer` (CR 509.1a).
export interface Kampfpaar {
  kreatur: string
  ziel?: string
  angreifer?: string
}

// Ausgeschieden ist ein Sitz, der die Partie verlassen hat — CR 104.3
// (M17, #1618).
export interface Ausgeschieden {
  sitz: string

  // grund ist einer der vier aus CR 104.3 — `aufgabe`, `leben`, `bibliothek`
  // oder `gift`.
  grund: string
}

// Ausgang ist, wie die Partie ausgegangen ist — CR 104.
//
// **Für jeden derselbe**: ein Ergebnis ist öffentlich. **Leerer `sieger` und
// `unentschieden: false` heißen zusammen: sie läuft noch.**
export interface Ausgang {
  sieger: string
  unentschieden: boolean
  ausgeschieden: Ausgeschieden[]
}
