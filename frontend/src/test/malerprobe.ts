// **Ein Maler, der nichts malt** — die gemeinsame Vorrichtung der Tischläufe.
//
// # Warum es sie gibt
//
// **Weil dieselbe Vorrichtung fünfmal dastand** — in `maler.test.ts`, in
// `useTisch.test.ts`, in `tischzielen.test.ts` und in zwei Läufen über die
// Tischfläche. Sie sind beim ersten neuen Feld auseinandergelaufen: `abzug`
// kam mit PLAY-05.AK-03 dazu, und **fünf** Dateien wurden rot. Genau davor
// warnt §10 — „eine Vorrichtung, die zweimal dastünde, liefe beim ersten
// neuen Feld auseinander".
//
// # Was sie nicht ist
//
// **Kein Ersatz für einen echten Maler.** Sie sagt zu jedem Ruf „ja" und
// zeichnet nichts; was PixiJS daraus macht, sieht nur, wer vor dem Fenster
// sitzt (§3, Sichtprüfung).
import { vi } from 'vitest'

import type { Tischmaler, Zeichenwunsch } from '../tisch/maler'
import type { Rechteck } from '../tisch/tischausschnitt'
import type { Tischfarben } from '../tisch/tischfarben'
import type { Mock } from 'vitest'

// Malerprobe ist ein Maler samt der Griffe, die ein Lauf braucht.
//
// **Die Spione stehen als solche da.** Ein `Tischmaler` allein genügte nicht:
// die Läufe greifen an `zeichne.mock`, und ein schlichter Funktionstyp trägt
// das nicht.
export interface Malerprobe extends Tischmaler {
  zeichne: Mock<(wunsch: Zeichenwunsch) => void>
  beiVerlust: Mock<(melde: () => void) => void>
  faerbe: Mock<(farben: Tischfarben) => void>
  raeume: Mock<() => void>
  abzug: Mock<(ausschnitt: Rechteck, punkte: number) => Promise<string>>
  passeAn: Mock<() => { breite: number; hoehe: number }>

  // verlieren ruft jeden angemeldeten Horcher auf den Kontextverlust.
  verlieren: () => void

  // behaelterWird ändert die Größe des Behälters. **Die Fläche folgt erst
  // beim Anpassen** — wie PixiJS, das die Leinwand im nächsten Bild anpasst
  // und nicht in dem Augenblick, in dem der Behälter sich ändert.
  behaelterWird: (neu: { breite: number; hoehe: number }) => void
}

// einMaler gibt einen Maler, der nichts tut — mit den Maßen, die der Aufrufer
// nennt.
export function einMaler(
  masse: { breite: number; hoehe: number } = { breite: 0, hoehe: 0 },
): Malerprobe {
  const verlust: (() => void)[] = []
  let flaeche = masse
  let behaelter = masse

  return {
    zeichne: vi.fn(),
    masse: () => flaeche,
    passeAn: vi.fn(() => {
      flaeche = behaelter

      return flaeche
    }),
    beiVerlust: vi.fn((melde: () => void) => {
      verlust.push(melde)
    }),
    faerbe: vi.fn(),
    raeume: vi.fn(),

    // **Ein fester Data-URL und kein echtes Bild**: was daraus entsteht,
    // entscheidet die Grafikkarte, und die gibt es hier nicht.
    abzug: vi.fn(() => Promise.resolve('data:image/png;base64,AAA')),

    verlieren: () => {
      for (const melde of verlust) melde()
    },
    behaelterWird: (neu) => {
      behaelter = neu
    },
  }
}
