// **Die Zahl steht am Haufen** — PLAY-03.AK-04, grafische Hälfte
// (#1711).
//
// # Der Befund
//
// **Sprachlich sagt der Tisch seit dem 2026-09-21, was eine Zahl zählt** —
// „7 auf einem Haufen", „3 Marken +1/+1", „Spielstein". **Auf der Leinwand
// war die Haufenzahl weiterhin nirgends**: die Marken sitzen als Punkte an
// der Karte (`markenpunkte.ts`), und sieben Wälder übereinander sahen aus
// wie einer.
//
// Das Kriterium verlangt beides: „sprachlich **und grafisch**
// unterscheidbar".
//
// # Warum ein Schild und keine Punkte
//
// **Weil es etwas anderes zählt.** Marken gehören **zu** einer Karte und
// sitzen auf ihr; eine Haufenzahl zählt **Karten** und gehört an den Stapel,
// nicht auf die oberste. Wer beides als Punkte zeichnete, machte aus zwei
// Fragen eine — und genau das verbietet AK-04.
//
// # Warum hier und nicht im Zeichner
//
// **Dieselbe Trennung wie bei [zonenschilder]**: wo ein Schild sitzt und was
// darauf steht, lässt sich messen; `pixibeiwerk.ts` steht in `.covignore`,
// weil es **zeichnet**.
import { Schildabstand, Schildgroesse } from './zonenschilder'
import { umriss } from './zonenlage'

import type { Zonenschild } from './zonenschilder'
import type { Tischplan } from './tischplan'

// haufenschilder gibt je Haufen eine Zahl an seinem obersten Blatt.
//
// **Am obersten und nicht am untersten.** Ein Haufen liegt versetzt
// (`Haufenversatz`); die oberste Karte ist die, die man sieht, und an ihrer
// unteren rechten Ecke steht die Zahl — dort, wo keine Marke sitzt.
export function haufenschilder(plan: Tischplan): Zonenschild[] {
  const oberste = new Map<string, (typeof plan.lagen)[number]>()

  for (const lage of plan.lagen) {
    if (lage.haufen === '') {
      continue
    }

    const bisher = oberste.get(lage.haufen)

    if (bisher === undefined || lage.hoehenrang > bisher.hoehenrang) {
      oberste.set(lage.haufen, lage)
    }
  }

  return [...oberste].map(([haufen, lage]) => ({
    kennung: `haufen:${haufen}`,

    // **An der sichtbaren Ecke** — bei einem getappten Haufen liegt sie
    // anders als die der aufrechten Lage (`umriss`).
    x: umriss(lage).x + umriss(lage).breite - Schildabstand,
    y: umriss(lage).y + umriss(lage).hoehe - Schildabstand,
    text: String(lage.haufenzahl),
    groesse: Schildgroesse,

    // **Der Text hängt über seinem Punkt** — der Punkt liegt auf der
    // Unterkante der Karte; wer von dort nach unten schriebe, schriebe in
    // den Tisch.
    ankerY: 1,

    // **Und er endet an seinem Punkt**, statt dort anzufangen: der Punkt
    // liegt einen Schildabstand vor der rechten Kante, und bis zum
    // 2026-09-24 lief die Zahl von dort aus über die Karte hinaus.
    ankerX: 1,
    art: 'zahl' as const,
  }))
}
