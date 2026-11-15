// **Die Weite einer Schublade** — AP-12
// (#2369),
// Befund B-26.
//
// Die Schublade des Sets (`rahmen/Setschublade.tsx`) hält, ob sie weit
// aufgeht; die Seite darin bietet den Umschalter an, weil sie weiß, ob sie
// Platz braucht — heute nur der Symboleditor. Ohne Schublade gibt es keinen
// Wert, und die Seite zeigt keinen Umschalter.
import { createContext, useContext } from 'react'

export interface Schubladenweite {
  weit: boolean
  umschalten: () => void
}

export const SchubladenweiteKontext = createContext<
  Schubladenweite | undefined
>(undefined)

export function useSchubladenweite(): Schubladenweite | undefined {
  return useContext(SchubladenweiteKontext)
}
