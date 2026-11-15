// **Den Fingerabdruck von Hand vergleichen** — aus `Peerkontaktzeile.tsx`
// herausgelöst, als NET-01
// (#1705)
// ihn an zweiter Stelle brauchte: hinter dem Vergleichscode als Angabe für
// Fachleute, und vorn, wo es keinen Code gibt.
//
// **Ein Feld und kein Häkchen**: wer bestätigt, tippt ab, was er gesehen
// hat, und stimmt es nicht, sagt es die Anwendung.
import { useId, useState } from 'react'

import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'
import { kennungseingabe } from './kennungseingabe'

interface Abdruckvergleichseigenschaften {
  kennung: string
  laeuft: boolean
  bestaetigen: (kennung: string, verglichen: string) => void
}

export function Abdruckvergleich({
  kennung,
  laeuft,
  bestaetigen,
}: Abdruckvergleichseigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const feld = useId()

  const [verglichen, setzeVerglichen] = useState('')

  return (
    <div className="flex flex-wrap items-end gap-2">
      <label className="flex flex-col gap-1 text-sm" htmlFor={feld}>
        {eigene.kontakteVergleichen}
        <input
          {...kennungseingabe}
          id={feld}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 font-mono text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          value={verglichen}
          onChange={(ereignis) => {
            setzeVerglichen(ereignis.target.value)
          }}
        />
      </label>

      <Knopf
        gewicht="haupt"
        angehalten={laeuft || verglichen === ''}
        aufKlick={() => {
          bestaetigen(kennung, verglichen)
        }}
      >
        {eigene.kontakteBestaetigen}
      </Knopf>
    </div>
  )
}
