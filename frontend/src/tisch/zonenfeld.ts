// **Wo eine Zone innerhalb eines Sitzes liegt** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Drei Eigenschaften, und alle drei sind nachgelesen
//
// **CR 400.2** (Fassung 2026-08-19) trennt öffentliche von verborgenen
// Zonen: Bibliothek und Hand sind verborgen, alles andere ist öffentlich.
// **CR 400.5** sagt, wo die Reihenfolge festliegt — Bibliothek, Friedhof und
// Stapel — und erlaubt für alles andere die freie Anordnung. **CR 401.2**
// nennt die Bibliothek einen einzelnen verdeckten Stapel, **CR 404.1** den
// Friedhof einen Ablagestapel.
//
// # Exil und Kommando gehören allen, liegen hier aber bei einem
//
// **CR 400.1 nennt sie geteilte Zonen** — sie gehören keinem Spieler. Auf
// einem Tisch legt trotzdem jeder seine exilierten Karten vor sich, und
// **CR 400.5 erlaubt das ausdrücklich**, solange klar bleibt, wer was
// kontrolliert. Hier steht der Beherrscher an jedem Objekt; die Zone bleibt
// eine, die Ablage ist nur der Ort.
import {
  Ablagebreite,
  Kartenabstand,
  Kartenbreite,
  Kartenhoehe,
  Kopfhoehe,
  Namenshoehe,
  Randbreite,
  Reihenhoehe,
  Sitzbreite,
} from './masse'

import type { Sitzflaeche } from './sitzflaeche'

// Zonengestalt sagt, wie die Objekte einer Zone zueinander liegen.
//
//   - `reihe` — nebeneinander, mit Umbruch: das Schlachtfeld.
//   - `faecher` — überlappend, jedes sichtbar: Hand und Stapel.
//   - `stapel` — übereinander, nur das oberste sichtbar: die vier Ablagen.
export type Zonengestalt = 'reihe' | 'faecher' | 'stapel'

// Zonenfeld ist der Platz einer Zone in Millimetern.
export interface Zonenfeld {
  zone: string
  sitz: string

  x: number
  y: number
  breite: number
  hoehe: number

  gestalt: Zonengestalt

  // verborgen sagt, ob nicht jeder hineinsehen darf — CR 400.2.
  verborgen: boolean

  // geordnet sagt, ob die Reihenfolge festliegt — CR 400.5.
  //
  // **Daran hängt, ob gezogen werden darf**: was geordnet ist, lässt sich
  // nicht umlegen, und ein Haufen in der Darstellung ist nie der Stapel der
  // Regeln.
  geordnet: boolean
}

// Die vier Ablagen, von der Hand nach außen.
//
// **Die Bibliothek zuerst**, weil von ihr gezogen wird; das Kommando zuletzt,
// weil es das seltenste ist. Das ist eine Entscheidung dieser Oberfläche und
// keine Regel.
const ablagen = [
  { zone: 'bibliothek', verborgen: true, geordnet: true },
  { zone: 'friedhof', verborgen: false, geordnet: true },
  { zone: 'exil', verborgen: false, geordnet: false },
  { zone: 'kommando', verborgen: false, geordnet: false },
]

// zonenfelder legt die Felder eines Sitzes.
//
// **`zonen` sagt, welche gezeichnet werden** — PLAY-03.AK-01
// (#1711).
// Ohne Angabe sind es alle; der Grund für eine Auswahl steht bei
// [zonenDerPartie], und er ist einer der Oberfläche und keiner der Regeln.
export function zonenfelder(
  flaeche: Sitzflaeche,
  zonen?: readonly string[],
): Zonenfeld[] {
  const x = flaeche.x + Randbreite
  const innen = Sitzbreite - 2 * Randbreite
  const handbreite = innen - Ablagebreite - Kartenabstand

  // **Was nicht gebraucht wird, bekommt kein Feld** — PLAY-03.AK-01. Die
  // übrigen rücken damit von selbst zusammen: ihre Stelle ist ihr Rang in
  // dieser Liste.
  const gezeigte =
    zonen === undefined
      ? ablagen
      : ablagen.filter((eine) => zonen.includes(eine.zone))

  // **Das Schlachtfeld liegt zur Mitte** — CR 403.1: die Permanenten eines
  // Spielers liegen vor ihm, und vor ihm ist zur Tischmitte hin.
  //
  // **Und jedes Feld steht unter seiner Namensleiste**, der Kopf des Sitzes
  // an seiner Außenkante (`masse.ts`, [Kopfhoehe] und [Namenshoehe]). Von
  // außen nach innen: Kopf, Leiste, Hand, Lücke, Leiste, Schlachtfeld.
  const zumStapel = flaeche.zurMitte === -1
  const tischY = zumStapel
    ? flaeche.y + Randbreite + Namenshoehe
    : flaeche.y +
      Randbreite +
      Kopfhoehe +
      2 * Namenshoehe +
      Kartenhoehe +
      Kartenabstand
  const handY = zumStapel
    ? tischY + Reihenhoehe + Kartenabstand + Namenshoehe
    : flaeche.y + Randbreite + Kopfhoehe + Namenshoehe

  return [
    {
      zone: 'tisch',
      sitz: flaeche.sitz,
      x,
      y: tischY,
      breite: innen,
      hoehe: Reihenhoehe,
      gestalt: 'reihe',
      verborgen: false,
      geordnet: false,
    },
    {
      zone: 'hand',
      sitz: flaeche.sitz,
      x,
      y: handY,
      breite: handbreite,
      hoehe: Kartenhoehe,
      gestalt: 'faecher',
      verborgen: true,
      geordnet: false,
    },
    ...gezeigte.map((eine, stelle): Zonenfeld => ({
      ...eine,
      sitz: flaeche.sitz,
      x:
        x +
        handbreite +
        Kartenabstand +
        stelle * (Kartenbreite + Kartenabstand),
      y: handY,
      breite: Kartenbreite,
      hoehe: Kartenhoehe,
      gestalt: 'stapel',
    })),
  ]
}

// zonenrahmen ist ein Feld samt seiner Namensleiste — das, was als Zone
// gezeichnet wird.
//
// **Das Feld selbst bleibt die Fläche der Karten**: `zonenlage.ts` und
// `schlachtfeld.ts` legen von seiner Oberkante aus, und so bleibt es. Die
// Leiste steht darüber, und wer die Zone zeichnet oder ihren Umfang
// braucht, nimmt den Rahmen.
export function zonenrahmen(feld: Zonenfeld): {
  x: number
  y: number
  breite: number
  hoehe: number
} {
  return {
    x: feld.x,
    y: feld.y - Namenshoehe,
    breite: feld.breite,
    hoehe: feld.hoehe + Namenshoehe,
  }
}
