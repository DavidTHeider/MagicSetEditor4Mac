// **Die Vorrichtung für jeden Lauf über den Tischplan** — Werkstück G3 von
// Phase 5.5 (#1460).
//
// **Eine eigene Datei**, seit zwei Prüfdateien sie brauchen und
// `tischplan.test.ts` über die 400 Zeilen aus §2 lief. Der Schnitt folgt der
// Frage: hier steht, **woraus ein Plan entsteht** — dort und in
// `planschlachtfeld.test.ts`, **was von ihm verlangt wird**.
//
// **Und sie liegt unter `test/` und nicht bei den Tischdateien.** Dort steht
// jede Vorrichtung dieses Fensters, und `vite.config.ts` nimmt `src/test/**`
// von der Abdeckung aus — eine Wache, die nur ein kaputter Lauf betritt, wäre
// sonst eine unbetretene Zeile, die niemand betreten *soll*.
import { einObjekt, einSitz } from '../seiten/partieprobe'

import type { Spielobjekt } from '../apipartie'
import type { Tischlage, Tischplan } from '../tisch/tischplan'
import type { Sitzflaeche } from '../tisch/sitzflaeche'

// zweiSitze ist der gewöhnliche Tisch: ich unten, du oben.
export const zweiSitze = {
  sitze: [
    einSitz({ kennung: 'ich', name: 'Ich', deckabdruck: 'a', karten: 60 }),
    einSitz({ kennung: 'du', name: 'Du', deckabdruck: 'b', karten: 60 }),
  ],
  fuer: 'ich',
}

// meines gibt ein Objekt, das dem rechnenden Sitz gehört — `einObjekt` setzt
// dafür `sitz1`, und die Sitze hier heißen anders.
export const meines = (teile: Partial<Spielobjekt>) =>
  einObjekt({ eigner: 'ich', beherrscher: 'ich', ...teile })

// sitzflaeche gibt die Fläche eines Sitzes — oder scheitert laut.
export function sitzflaeche(plan: Tischplan, sitz: string): Sitzflaeche {
  const eine = plan.sitze.find((jede) => jede.sitz === sitz)
  if (!eine) {
    // keinfehlertext: eine Vorrichtung für Läufe, nie im Fenster
    throw new Error(`keine Fläche für ${sitz}`)
  }

  return eine
}

// lage gibt die Lage eines Objekts — oder scheitert laut.
//
// **Laut und nicht still**: ein `undefined` liefe in die Zusicherung und
// meldete dort etwas anderes, als fehlt.
export function lage(plan: Tischplan, kennung: string): Tischlage {
  const eine = plan.lagen.find((jede) => jede.kennung === kennung)
  if (!eine) {
    // keinfehlertext: eine Vorrichtung für Läufe, nie im Fenster
    throw new Error(`keine Lage für ${kennung}`)
  }

  return eine
}

// mitTypen gibt eine Permanente mit gesetzten Kartentypen.
//
// **Alle Felder, und zwar hier**: `Karteneigenschaften` hat zwölf, und jede
// von Hand gebaute Kopie wich beim nächsten ab (§3).
export function mitTypen(kennung: string, typen: string[]): Spielobjekt {
  return meines({
    kennung,
    zone: 'tisch',
    eigenschaften: {
      name: kennung,
      farben: [],
      manawert: 0,
      supertypen: [],
      typen,
      untertypen: [],
      staerke: { steht: false, stern: false, zahl: 0 },
      widerstandskraft: { steht: false, stern: false, zahl: 0 },
      loyalitaet: { steht: false, stern: false, zahl: 0 },
      verteidigung: { steht: false, stern: false, zahl: 0 },
      unbekannteTypen: [],
    },
  })
}
