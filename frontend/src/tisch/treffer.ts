// **Was an einem Punkt liegt** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Die umgekehrte Frage
//
// **Der Plan sagt, was wo liegt; hier wird gefragt, was an dieser Stelle
// ist.** Ohne diese Rechnung gäbe es kein Ziehen — und mit ihr in der
// Grafiknaht gäbe es keine Prüfung davon.
//
// **In Millimetern und nicht in Punkten.** Wer zieht, zeigt auf den
// Bildschirm; die Kamera rechnet das um, und ab hier gilt wieder das Maß des
// Tisches. Eine Trefferrechnung in Punkten hinge am Maßstab und wäre bei
// jedem Zoom eine andere.
import { umriss } from './zonenlage'
import { zonenrahmen } from './zonenfeld'

import type { Punkt } from './kamera'
import type { Tischplan } from './tischplan'

const trifft = (
  ort: Punkt,
  feld: { x: number; y: number; breite: number; hoehe: number },
) =>
  ort.x >= feld.x &&
  ort.x <= feld.x + feld.breite &&
  ort.y >= feld.y &&
  ort.y <= feld.y + feld.hoehe

// objektAn nennt die Karte unter dem Punkt — die **oberste**.
//
// **Bei einem Stapel überlappen die Karten**, und wer die unterste träfe,
// zöge etwas heraus, worauf er gar nicht gezeigt hat. Gesucht wird deshalb
// nach dem höchsten `hoehenrang`, nicht nach dem ersten Treffer.
export function objektAn(plan: Tischplan, ort: Punkt): string | undefined {
  let gefunden: { kennung: string; rang: number } | undefined

  for (const lage of plan.lagen) {
    // **Getroffen wird, was zu sehen ist** — bei einer getappten Karte die
    // quere Fläche und nicht die aufrechte Lage (`umriss`).
    if (!trifft(ort, umriss(lage))) continue
    if (gefunden && gefunden.rang >= lage.hoehenrang) continue

    gefunden = { kennung: lage.kennung, rang: lage.hoehenrang }
  }

  return gefunden?.kennung
}

// Zonenort ist eine Zone samt dem Sitz, dem sie gehört.
export interface Zonenort {
  zone: string
  sitz: string
}

// zoneAn nennt das Feld unter dem Punkt.
//
// **Das zuletzt gefundene gewinnt.** Die Felder eines Sitzes überschneiden
// sich nicht; der Stapel liegt zwischen den Sitzflächen und damit in keiner.
// Trotzdem wird bis zum Ende gesucht statt beim ersten Treffer abzubrechen —
// eine Überschneidung wäre ein Fehler im Plan, und beim ersten Treffer
// abzubrechen verbärge ihn.
export function zoneAn(plan: Tischplan, ort: Punkt): Zonenort | undefined {
  let gefunden: Zonenort | undefined

  // **Über den Rahmen und nicht über das Feld** — die Namensleiste gehört
  // zur Zone (`zonenrahmen`); ein Ablegen dort ginge sonst ins Leere.
  for (const feld of plan.felder)
    if (trifft(ort, zonenrahmen(feld)))
      gefunden = { zone: feld.zone, sitz: feld.sitz }

  return gefunden
}
