// **Eine geschriebene Datei, auf Wunsch angesehen** — UX-13.AK-02/04
// (#1685).
//
// Steht unter der Erfolgsmeldung jeder Ausfuhr. **Gelesen wird erst auf
// Klick**: ein Druckbogen von zwanzig Megabyte wird nicht bei jedem Schreiben
// ein zweites Mal angefasst, und wer die Datei nicht sehen will, bekommt
// keinen Ruf, der scheitern könnte.
import { useState } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Dateiansicht } from './Dateiansicht'
import { Knopf } from './Knopf'

interface DateiverweisEigenschaften {
  pfad: string
}

export function Dateiverweis({ pfad }: DateiverweisEigenschaften) {
  const eigene = useTexte().bausteine.dateiansicht
  const [offen, setzeOffen] = useState(false)

  const name = pfad.slice(pfad.lastIndexOf('/') + 1)

  return (
    <div className="mt-2 flex flex-col gap-2">
      <div>
        <Knopf
          gewicht="leise"
          bezeichnung={
            offen ? eigene.zuklappen : einsetzen(eigene.ansehenFuer, { name })
          }
          aufKlick={() => {
            setzeOffen(!offen)
          }}
        >
          {offen ? eigene.zuklappen : eigene.ansehen}
        </Knopf>
      </div>

      {offen && <Dateiansicht pfad={pfad} />}
    </div>
  )
}
