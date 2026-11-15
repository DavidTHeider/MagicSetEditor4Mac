// **Ein Einstellungsstand für das ganze Fenster** — Befund G14 des
// Zweitblicks über Phase 6.3
// (#1665).
//
// # Warum ein Raum
//
// **Weil jeder Haken, der `useEinstellungen()` rief, einen eigenen Stand
// hielt.** Die Feldtiefe lud bei jedem Kartenwechsel neu und zeigte bis dahin
// die Vorgabe; was sie schrieb, erfuhr die Einstellungsseite nicht, denn die
// las aus dem Stand der Anwendung. Hier steht der eine Stand, den `App`
// ohnehin hält, und `useEinstellungen` gibt ihn jedem darunter.
//
// **Ohne Raum** — in Läufen, die eine Seite allein aufbauen — hält ein Haken
// weiter seinen eigenen, wie bisher.
import { createContext, useContext, type ReactNode } from 'react'

import type { Einstellungsstand } from './useEinstellungen'

const Kontext = createContext<Einstellungsstand | undefined>(undefined)

export function Einstellungsraum({
  stand,
  children,
}: {
  // stand fehlt in Läufen, die eine Fläche ohne geteilten Stand aufbauen —
  // dann ist es, als gäbe es keinen Raum.
  stand: Einstellungsstand | undefined
  children: ReactNode
}) {
  return <Kontext.Provider value={stand}>{children}</Kontext.Provider>
}

// useEinstellungsraum gibt den geteilten Stand — ohne Raum `undefined`.
export function useEinstellungsraum(): Einstellungsstand | undefined {
  return useContext(Kontext)
}
