// **Eine Seite unter einer Schublade ist verdeckt** — UX-08
// (#1680).
//
// # Warum es das gibt
//
// **Weil die Kartenseite unter der Schublade stehen bleibt** — samt gewählter
// Karte, Filter und Formular, und das ist der Sinn der Schublade. Stehen
// bleiben aber auch ihre Horcher am **Fenster**: ⇟ blätterte die Karte
// darunter weiter, während jemand in der Konsole tippt, verbrauchte die Taste
// und zöge den Fokus aus der Schublade in ein Kartenfeld, das niemand sieht.
//
// **Ein Horcher am Fenster fragt deshalb, ob seine Seite verdeckt ist**, und
// hängt sich dann gar nicht erst ein. Wer einen neuen schreibt, stellt sich
// dieselbe Frage.
//
// # Was es nicht tut
//
// **Die Fläche selbst sperrt `inert`** in `Seiteninhalt.tsx` — Fokus, Klick
// und Vorleseprogramm. Ein Horcher am Fenster hört trotzdem jede Taste,
// denn die Ereignisse steigen aus der Schublade dorthin auf; `inert` hilft
// ihm nicht, und dieser Kontext ersetzt `inert` nicht.
import { createContext, use } from 'react'

export const Verdeckung = createContext(false)

// useVerdeckt sagt, ob die Seite gerade unter einer Schublade liegt.
export function useVerdeckt(): boolean {
  return use(Verdeckung)
}
