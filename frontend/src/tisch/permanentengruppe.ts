// **Was auf dem Schlachtfeld zusammengehört** — PLAY-03.AK-02
// (#1711).
//
// # Drei Gruppen, und keine vierte
//
// **Länder, Kreaturen, sonstige bleibende Karten.** So steht es im
// Abnahmekriterium, und so sieht ein Tisch aus: die Länder hinten, weil man
// sie tappt und selten ansieht; die Kreaturen vorn, weil sie angreifen; alles
// andere dazwischen.
//
// # Und ausdrücklich keine Unterzone für Hexereien und Spontanzauber
//
// **Eine Permanente ist eine Karte oder ein Token auf dem Schlachtfeld** —
// CR 110.1 (Fassung 2026-08-19), dieselbe Stelle, auf die sich `zonenlage.ts`
// beruft. Die fünfzehn Kartentypen zählt CR 205.2a auf
// (`internal/domain/kartentyp.go`); welche davon bleibend sind, entscheidet
// hier **nicht** eine Nummer, sondern die Frage, was gruppiert werden soll:
// Länder, Kreaturen, alles übrige.
//
// **Das Kriterium sagt es wörtlich**: Hexereien und Spontanzauber *„erhalten
// keine falsche dauerhafte Schlachtfeld-Unterzone."* Sie bekommen hier
// deshalb **keine eigene Gruppe** — sie fallen zu `sonstige`, und das ist
// kein Widerspruch: wer durch einen fremden Zustand doch auf dem
// Schlachtfeld landet, wird **gezeigt** statt verschluckt (§8), nur eben ohne
// einen Platz, den er behielte.
//
// **Keine erfundene Nummer für die Aufzählung der bleibenden Typen.** Ein
// erster Entwurf nannte hier CR 110.4a und CR 608.2m; beide standen ungeprüft
// da, weil das Regelwerk auf diesem Rechner nicht liegt. §11 ist dazu
// eindeutig — eine erfundene Nummer ist schlimmer als keine, weil sie wie ein
// Beleg aussieht. Wer die Aufzählung braucht, schlägt sie nach und trägt sie
// mit ihrer Fassung ein.

import type { Spielobjekt } from '../apipartie'

// Permanentengruppe ist eine der drei Reihen des Schlachtfelds.
export type Permanentengruppe = 'land' | 'kreatur' | 'sonstige'

// gruppenordnung ist die Reihenfolge, in der die Gruppen liegen.
//
// **Von hinten nach vorn**, also von der eigenen Tischkante zur Mitte: erst
// die Länder, dann das Übrige, dann die Kreaturen. CR 403.1 sagt nur, dass
// die Permanenten eines Spielers **vor ihm** liegen; welche zuerst, ist eine
// Entscheidung dieser Oberfläche und keine Regel.
export const gruppenordnung: readonly Permanentengruppe[] = [
  'land',
  'sonstige',
  'kreatur',
]

// permanentengruppe sagt, in welche Reihe ein Objekt gehört.
//
// **Aus den Kartentypen und nicht aus dem Namen.**
//
// # Ein verdecktes Permanent braucht hier keinen eigenen Zweig
//
// **Weil der Kern es schon beantwortet.** Ein verdecktes Permanent ist eine
// 2/2-Kreatur ohne Namen und ohne Typen — CR 708.2a —, und
// `domain.VerdeckteEigenschaften` gibt genau das: `Typen: [creature]`. Was
// hier ankommt, ist also bereits eine Kreatur; ein `if verdeckt` daneben wäre
// ein zweiter Ort für dieselbe Regel.
//
// # Und ohne Eigenschaften zählt es zu den sonstigen
//
// **`eigenschaften` ist `null`, wo niemand hineinsehen darf** (M2, #1604) —
// auf dem Schlachtfeld kommt das nicht vor, denn es ist eine **öffentliche**
// Zone (CR 400.2). Der Rückfall steht trotzdem da, weil der Typ ihn verlangt:
// wer ohne Typen dasteht, ist keine Kreatur und kein Land, und `sonstige`
// heißt genau das.
export function permanentengruppe(objekt: Spielobjekt): Permanentengruppe {
  const typen = objekt.eigenschaften?.typen ?? []

  if (typen.includes('creature')) {
    return 'kreatur'
  }

  return typen.includes('land') ? 'land' : 'sonstige'
}
