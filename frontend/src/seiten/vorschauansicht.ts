// **Die Ansicht der Vorschau** — Maßstab und Überlagerungen, UXR-06
// (#1914).
//
// # Warum sie über dem Editor liegt
//
// Der Karteneditor wurde **nach jedem Schreiben** neu gebaut: sein Schlüssel
// trug den Verlaufsstand (`Kartenarbeitsflaeche.tsx`). Eine Ansicht, die
// darin lebte, fiel auf „Anpassen" und „keine Linie" zurück — wer auf 200 %
// prüfte und ein Feld verließ, fing von vorn an. Seit U01 (#2119) baut ihn
// nur noch ein Kartenwechsel neu; die Ansicht bleibt darüber, damit auch der
// sie nicht zurücksetzt. Dieselbe Klasse wie die Feldwahl in
// #1075,
// und dieselbe Antwort: die Wahl liegt in `Karten`, gereicht als das Paar,
// das `useState` gibt.
//
// **Wahlfrei**, wie dort: wer den Editor allein fährt, braucht keine Seite
// darüber; dann hält er sie selbst.
import { useState, type Dispatch, type SetStateAction } from 'react'

import type { FeldflaecheDTO } from '../formen'

import { keineOverlays, type Overlaywahl } from './Vorschauoverlays'
import type { Zoomwahl } from './zoom'

export interface Vorschauansicht {
  zoom: Zoomwahl
  overlays: Overlaywahl

  // flaechenJeStil sind die Flächen, nach denen das Formular seine Felder
  // ordnet (#674) — **je eigenem Stil der Karte**, `''` für den des Sets.
  // **Sie liegen hier, damit sie den Kartenwechsel überstehen** (#2301): der
  // Editor wird je Karte neu gebaut.
  //
  // **Die zuerst gesehene Lage eines Feldes bleibt** (#2309): beim Blättern
  // ändert sich die Leseordnung im selben Stil in 601 von 1.033 Schritten des
  // Kanons, und ein Formular, das sich nach 130 ms umsortiert, ist keines,
  // in dem man arbeitet. Was eine Karte neu bringt, kommt dazu
  // (`formularruhe.ts`).
  flaechenJeStil?: Readonly<Record<string, readonly FeldflaecheDTO[]>>

  // ordnungJeStil ist die Reihenfolge der Felder je Stil, **wie sie einmal
  // stand** (Zweitblick über die Folgen vor 0.8.0, Befund 9): wächst die
  // Menge der Flächen, rechnet die Leseordnung ihre Zeilen neu, und ein hohes
  // neues Feld fasste zwei bekannte in eine Zeile und vertauschte sie. Neues
  // wird hier nur eingereiht.
  ordnungJeStil?: Readonly<Record<string, readonly string[]>>

  // gezeichnetJeStil sagt je geschriebenem Stil, mit welchem Paket zuletzt
  // gezeichnet wurde (#2314). **Gelernt und geordnet wird unter dem Paket**:
  // `''` meint nach einem fremden Setstilwechsel ein anderes.
  gezeichnetJeStil?: Readonly<Record<string, string>>
}

// grundansicht ist die Ansicht, mit der eine Seite anfängt: die Karte passt
// in ihre Spalte, und nichts liegt darüber.
export const grundansicht: Vorschauansicht = {
  zoom: 'anpassen',
  overlays: keineOverlays,
}

// Ansichtswahl ist die Ansicht samt ihrem Setzer — die Gestalt von
// `useState<Vorschauansicht>()`.
export type Ansichtswahl = [
  Vorschauansicht,
  Dispatch<SetStateAction<Vorschauansicht>>,
]

// useVorschauansicht nimmt die Ansicht von oben — oder hält eine eigene, wenn
// niemand eine reicht. **Beide Zustände entstehen**, damit die Reihenfolge
// der Haken nicht davon abhängt, ob eine gereicht wird.
export function useVorschauansicht(
  von: Ansichtswahl | undefined,
): Ansichtswahl {
  const eigene = useState(grundansicht)

  return von ?? eigene
}
