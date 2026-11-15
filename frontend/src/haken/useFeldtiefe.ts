// **Wie viel vom Kartenformular dasteht** — UX-05.AK-01
// (#1677).
//
// # Warum ein Haken und keine Eigenschaft
//
// **Weil die Tiefe eine Einstellung ist und kein Zustand der Seite.** Das
// Kriterium sagt „ein je Nutzer oder Profil änderbarer Standardmodus" — wer
// sie in einem `useState` hielte, fände beim nächsten Aufschlagen wieder
// `normal` vor.
//
// # Und warum sie trotzdem hier steht und nicht im Bestand allein
//
// **Weil ein Wechsel sofort gilt.** Der Schreibweg geht über
// `useEinstellungen`; was er zurückgibt, ist die Vorgabe des nächsten Mals.
// Bis die Antwort da ist, steht die gewählte Tiefe — sonst spränge das
// Formular beim Umschalten kurz zurück (§7, Regel 3: geschrieben ist
// geschrieben).
import { useState } from 'react'

import { feldtiefen, type Feldtiefe } from '../seiten/feldtiefe'
import { useEinstellungen } from './useEinstellungen'

// Feldtiefeschluessel ist der Name der Einstellung — derselbe wie in
// `internal/domain/einstellung.go`.
export const Feldtiefeschluessel = 'feldtiefe'

export interface Feldtiefewahl {
  tiefe: Feldtiefe
  aufTiefe: (tiefe: Feldtiefe) => void
}

export function useFeldtiefe(): Feldtiefewahl {
  const einstellungen = useEinstellungen()
  const [gewaehlt, setzeGewaehlt] = useState<Feldtiefe>()

  return {
    tiefe: gewaehlt ?? gelesen(einstellungen.wert(Feldtiefeschluessel)),

    aufTiefe: (tiefe) => {
      setzeGewaehlt(tiefe)
      einstellungen.setze(Feldtiefeschluessel, tiefe)
    },
  }
}

// gelesen macht aus dem Wert der Einstellung eine Tiefe.
//
// **Ein fremder Wert ist `normal`** (§8): ein von Hand geänderter Bestand
// kann dort alles stehen haben, und ein leeres Formular wäre schlechter als
// das flachste.
function gelesen(wert: string): Feldtiefe {
  return (feldtiefen as readonly string[]).includes(wert)
    ? (wert as Feldtiefe)
    : 'normal'
}
