// **Was da sein soll, bevor die Karte kommt** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Was vorgeladen wird
//
// **Der Rücken und das eigene Deck.** Das Issue nennt sechs Dinge — „Eigene
// Decks, Rückseite, Commander, bekannte Tokens, Hintergrund und Counters" —,
// und zwei davon gibt es heute:
//
//   - **Der Rücken** liegt auf jeder verdeckten Karte und ist für alle
//     dieselbe Textur. Ein Tisch mit vierzig verdeckten Karten holt **ein**
//     Bild.
//   - **Das eigene Deck** steht als [Partie.vorrat] in der Sicht — und zwar
//     nur in der **eigenen**. Welche Karten im eigenen Deck liegen, hat der
//     Eigner selbst aufgeschrieben; geheim ist die **Reihenfolge** der
//     Bibliothek, und die steht nirgends.
//
// **Die anderen vier haben heute keinen Gegenstand**: Commander und Tokens
// entstehen mit G5 (#1462),
// Zähler und Hintergrund mit G3 und G4. Sie gehen **denselben** Weg, sobald
// es sie gibt: der Vorrat kommt aus der Partiesicht, und was dort steht,
// wird geholt.
//
// # In welcher Güte
//
// **`tisch-niedrig`** — die Stufe, die es genau dafür gibt: „der Tisch,
// solange noch geladen wird". Sechzig Karten in voller Tischgüte vorzuladen
// hieße, beim Öffnen sechzig Rasterläufe anzustoßen, von denen die meisten
// niemand sieht. In der niedrigen steht sofort etwas da, und `Kartenbild`
// zeigt sie, bis die volle ankommt.
import { useEffect } from 'react'

import type { Partie } from '../apipartie'
import type { Texturlauf } from './useSpieltexturen'
import type { Texturwunsch } from '../apitextur'

// Vorladeguete ist die Stufe, in der vorgeladen wird.
export const Vorladeguete = 'tisch-niedrig'

export function useVorladen(texturen: Texturlauf, partie: Partie): void {
  const { vorladen } = texturen

  // **Aus der Liste eine Zeichenkette**, damit der Effekt nicht bei jedem
  // Zeichnen läuft: `partie.vorrat` ist nach jeder Antwort ein neues Feld
  // mit demselben Inhalt.
  const vorrat = partie.vorrat.join(',')
  const spiel = partie.spiel

  useEffect(() => {
    const wuensche: Texturwunsch[] = [
      { art: 'ruecken', spiel, guete: Vorladeguete },
    ]

    for (const karteID of vorrat.split(',')) {
      if (karteID !== '') {
        wuensche.push({
          art: 'karte',
          karteID,
          seite: 'vorne',
          guete: Vorladeguete,
        })
      }
    }

    vorladen(wuensche)
  }, [vorladen, spiel, vorrat])
}
