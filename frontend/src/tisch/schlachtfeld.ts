// **Das Schlachtfeld in drei Gruppen** — PLAY-03.AK-02
// (#1711).
//
// # Was das Kriterium verlangt
//
// *„Länder, Kreaturen und sonstige bleibende Karten können auf dem
// Schlachtfeld visuell gruppiert werden; Hexereien und Spontanzauber erhalten
// keine falsche dauerhafte Schlachtfeld-Unterzone."*
//
// **Beides zusammen ist der Punkt.** Eine Gruppierung, die für jede Art eine
// eigene Fläche aufmacht, behauptet Zonen, die das Regelwerk nicht kennt —
// CR 403.1 nennt **eine** Zone, und wie darin gelegt wird, ist frei
// (CR 400.5). Hier entsteht deshalb **keine zweite Zone**: das Feld bleibt
// eines, und nur die **Reihenfolge und der Platz** darin folgen den Gruppen.
//
// # Zwei Reihen, drei Gruppen
//
// Ein Schlachtfeld ist zwei Kartenhöhen hoch (`Reihenhoehe` in `masse.ts`).
// Daraus folgt der Schnitt:
//
//   - **Zur Mitte die Kreaturen** — sie greifen an, und wer angreift, steht
//     vorn.
//   - **Zur eigenen Kante die Länder**, und daneben, hinter einer Lücke, das
//     übrige Bleibende.
//
// **Welche Reihe zur Mitte zeigt, sagt der Sitz** (`Sitzflaeche.zurMitte`) —
// dieselbe Angabe, aus der `zonenfelder` schon die Lage von Hand und
// Schlachtfeld ableitet.
//
// # Warum hier gerechnet und nicht in `zonenlage.ts`
//
// **Weil `zonenlage.ts` nichts über Karten weiß.** Es rechnet über `Gelegtes`
// — Kennung, getappt, verdeckt — und ist damit ohne eine Partie prüfbar. Was
// eine Kreatur ist, steht in den Kartentypen; diese Datei ist die Naht
// zwischen beidem.

import { Kartenabstand, Kartenbreite, Kartenhoehe } from './masse'
import { gruppenordnung, permanentengruppe } from './permanentengruppe'
import { zonenlagen } from './zonenlage'

import type { Gelegtes, Zonenlagen } from './zonenlage'
import type { Permanentengruppe } from './permanentengruppe'
import type { Spielobjekt } from '../apipartie'
import type { Zonenfeld } from './zonenfeld'

// Gruppenluecke trennt Länder und sonstiges Bleibendes in der hinteren Reihe.
//
// **Drei Kartenabstände.** Ein einzelner sähe aus wie der Abstand zwischen
// zwei Karten derselben Gruppe — dann wäre die Gruppierung keine. Breiter
// nähme der Platz, den die Karten brauchen.
export const Gruppenluecke = 3 * Kartenabstand

// Bespieltes ist ein gelegtes Objekt, das seine Karte mitbringt.
export interface Bespieltes extends Gelegtes {
  objekt: Spielobjekt
}

// schlachtfeldlagen legt die Permanenten eines Sitzes in ihre Gruppen.
//
// **`zurMitte` ist -1, wenn die Mitte oben liegt**, und +1, wenn sie unten
// liegt — dieselbe Angabe wie an der Sitzfläche.
export function schlachtfeldlagen<T extends Bespieltes>(
  feld: Zonenfeld,
  gelegte: T[],
  zurMitte: -1 | 1,
): Zonenlagen<T> {
  // **Ein vollständiger Verbund und keine Map.** Der Schlüsseltyp zählt die
  // drei Gruppen auf; damit gibt es weder ein `?.` noch ein `?? []`, das kein
  // Aufruf betritt — und `permanentengruppe` kann nichts anderes liefern.
  const gruppen: Record<Permanentengruppe, T[]> = {
    land: [],
    sonstige: [],
    kreatur: [],
  }

  for (const eines of gelegte) {
    gruppen[permanentengruppe(eines.objekt)].push(eines)
  }

  const felder = gruppenfelder(feld, zurMitte, {
    land: gruppen.land.length,
    sonstige: gruppen.sonstige.length,
  })

  const lagen: Zonenlagen<T>['lagen'] = []

  for (const gruppe of gruppenordnung) {
    const gerechnet = zonenlagen(felder[gruppe], gruppen[gruppe])

    // **Der Höhenrang zählt über die Gruppen hinweg weiter.** Jede Gruppe
    // rechnet bei null an; ohne Versatz trügen drei Lagen denselben Rang, und
    // der Zeichner entschiede nach Zufall, was oben liegt.
    for (const lage of gerechnet.lagen) {
      lagen.push({ ...lage, hoehenrang: lagen.length })
    }
  }

  return { lagen, zahl: gelegte.length }
}

// gruppenfelder schneidet das Schlachtfeld in die drei Plätze.
function gruppenfelder(
  feld: Zonenfeld,
  zurMitte: -1 | 1,
  zahlen: { land: number; sonstige: number },
): Record<Permanentengruppe, Zonenfeld> {
  const reihe = { hoehe: Kartenhoehe }

  // **Die Reihe zur Mitte ist die obere, wenn die Mitte oben liegt.**
  const naheY = zurMitte === -1 ? feld.y : feld.y + Kartenhoehe + Kartenabstand
  const ferneY = zurMitte === -1 ? feld.y + Kartenhoehe + Kartenabstand : feld.y

  const [landbreite, sonstigenbreite] = hintereBreiten(feld.breite, zahlen)

  return {
    kreatur: { ...feld, ...reihe, y: naheY },
    land: { ...feld, ...reihe, y: ferneY, breite: landbreite },
    sonstige: {
      ...feld,
      ...reihe,
      y: ferneY,
      x: feld.x + feld.breite - sonstigenbreite,
      breite: sonstigenbreite,
    },
  }
}

// hintereBreiten teilt die hintere Reihe zwischen Ländern und Sonstigem.
//
// **Nach Anzahl und nicht halbe-halbe.** Wer zwölf Länder und ein Artefakt
// hat, bekommt die Reihe fast ganz für die Länder; eine feste Hälfte zwänge
// sie in eine Überlappung, während daneben Platz stünde.
//
// **Ist eine Gruppe leer, gehört die Reihe der anderen** — und die Lücke
// entfällt mit ihr. Ein Platzhalter für nichts ist verlorener Raum.
function hintereBreiten(
  breite: number,
  zahlen: { land: number; sonstige: number },
): [number, number] {
  if (zahlen.sonstige === 0) {
    return [breite, 0]
  }

  if (zahlen.land === 0) {
    return [0, breite]
  }

  const frei = breite - Gruppenluecke
  const anteil = zahlen.land / (zahlen.land + zahlen.sonstige)

  // **Jede Gruppe behält mindestens eine Kartenbreite.** Sonst bekäme eine
  // einzelne Karte neben zwanzig anderen ein Feld, das schmaler ist als sie
  // selbst — und `reihenlagen` setzte sie über den Rand.
  const land = Math.min(
    Math.max(frei * anteil, Kartenbreite),
    frei - Kartenbreite,
  )

  return [land, frei - land]
}
