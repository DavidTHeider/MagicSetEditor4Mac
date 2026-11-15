// **Das eigene Fensterbild** — UX-09.AK-01
// (#1681).
//
// **Dieselben drei Wege wie beim Tisch**, an den Bindungen des Fensters —
// was der Haken tut und warum, steht in `useEigenesBild.ts`.
import { api } from '../api'
import { useEigenesBild, type Eigenbildstand } from './useEigenesBild'

import type { Bildwege } from './useEigenesBild'

const fensterwege: Bildwege = {
  lesen: (wem) => api().Fensterbild(wem),
  waehlen: (wem) => api().FensterbildWaehlen(wem),
  entfernen: (wem) => api().FensterbildEntfernen(wem),
}

// **Der Mensch kommt vom Aufrufer** — der `Gestaltungsraum` kennt ihn.
export function useFensterbild(nutzerID: string): Eigenbildstand {
  return useEigenesBild(fensterwege, nutzerID)
}
