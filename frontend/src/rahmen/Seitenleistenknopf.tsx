// **Der Knopf am Rand der Seitenleiste** — Werkstück U14 von Phase 5.1
// (#1490): er sagt, was er tut und wie die Leiste steht (`aria-expanded`),
// und er trägt den Fokus, wenn die Leiste unter ihm wegklappt.
import type { RefObject } from 'react'

import { useTexte } from '../sprache/kontext'

const spalte =
  'relative z-20 flex shrink-0 flex-col items-center border-r border-kante bg-flaeche px-1 py-2'

interface KnopfEigenschaften {
  offen: boolean
  umschalten: () => void
  griff: RefObject<HTMLButtonElement | null>
}

export function Seitenleistenknopf({
  offen,
  umschalten,
  griff,
}: KnopfEigenschaften) {
  const texte = useTexte().allgemein.navigation
  const beschriftung = offen
    ? texte.seitenleisteAusblenden
    : texte.seitenleisteEinblenden

  // **Der Knopf steht außerhalb der Leiste**, damit er bleibt, wenn sie geht
  // (#1490). Eingeklappt ist der Streifen alles, was von ihr übrig ist — die
  // Flächen bleiben über ⌘1–⌘8, ⌘?, ⌘, und ⌘K erreichbar. **Die Spalte gehört
  // deshalb zum Knopf** und nicht zum Rahmen (UX-14.AK-02, #1686).
  return (
    <div className={spalte}>
      <button
        ref={griff}
        type="button"
        aria-label={beschriftung}
        title={`${beschriftung} (${texte.seitenleisteKuerzel})`}
        aria-expanded={offen}
        aria-controls="seitenleiste"
        onClick={umschalten}
        className="flex h-8 w-8 items-center justify-center rounded-md text-leise hover:bg-betont hover:text-text focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      >
        {/* Das Zeichen zeigt die Richtung; der Name steht im `aria-label`. */}
        <span aria-hidden="true" className="text-base leading-none">
          {offen ? '⟨' : '⟩'}
        </span>
      </button>
    </div>
  )
}
