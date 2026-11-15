// **Das eigene Tischbild im Fenster** — Werkstück G3 von Phase 5.5
// (#1460).
//
// **Die drei Wege des Tisches an den gemeinsamen Haken** — was er tut und
// warum, steht in `useEigenesBild.ts`.
import { api } from '../api'
import { useEigenesBild, type Eigenbildstand } from './useEigenesBild'

import type { Bildwege } from './useEigenesBild'

// tischwege stehen außerhalb des Hakens, damit sie bei jedem Zeichnen
// dieselben sind — sonst lüde `useLaden` bei jedem Zeichnen neu.
const tischwege: Bildwege = {
  lesen: (wem) => api().Tischbild(wem),
  waehlen: (wem) => api().TischbildWaehlen(wem),
  entfernen: (wem) => api().TischbildEntfernen(wem),
}

export type Tischbildstand = Eigenbildstand

export function useTischbild(): Tischbildstand {
  return useEigenesBild(tischwege)
}
