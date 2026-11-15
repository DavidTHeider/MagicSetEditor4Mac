// **Das Suchfeld der Einstellungsseite** — Werkstück U11 von Phase 5.1
// (#1422),
// umgebaut für UX-04.AK-03
// (#1676).
//
// # Keine Trefferliste
//
// **Der Kasten wird an seinem Platz markiert**, und die Zahl steht im Kopf
// der Fläche — wie im Artboard. Eine Trefferliste zeichnete zu jedem Treffer
// eine **zweite Kopie** desselben Bedienelements, und wer in der Gruppe
// suchte, in der der Treffer liegt, sähe ihn zweimal.
//
// Wonach verglichen wird, steht in
// [einstellungstreffer.ts](./einstellungstreffer.ts) — zwei Verbraucher,
// eine Regel.
//
// # Die Frage kommt von außen
//
// **Das Feld steht in der Spalte, die Marken stehen in der Fläche**: zwei
// Orte, ein Zustand, und die Seite hält ihn.
import { useId } from 'react'

import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { useTexte } from '../sprache/kontext'

const feld =
  'w-full rounded-md border border-kante bg-flaeche px-3 py-2 ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Suchfeld({
  frage,
  setzeFrage,
  aufSprung,
}: {
  frage: string
  setzeFrage: (frage: string) => void

  // aufSprung läuft bei der Eingabetaste — sie führt auf das erste
  // getroffene Feld.
  //
  // **Und nicht jeder Anschlag tut das.** Wer bei jedem Buchstaben den Fokus
  // umsetzte, nähme ihn dem Suchfeld weg, und niemand könnte zu Ende tippen.
  aufSprung?: (() => void) | undefined
}) {
  const eigene = useTexte().seiten.einstellungen
  const feldID = useId()

  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={feldID} className="flex flex-col gap-1">
        <span className="text-sm font-medium">{eigene.suche}</span>
        <input
          id={feldID}
          type="search"
          value={frage}
          placeholder={eigene.suchePlatzhalter}
          className={feld}
          onChange={(ereignis) => {
            setzeFrage(ereignis.target.value)
          }}
          onKeyDown={(ereignis) => {
            if (ereignis.key === 'Enter') {
              ereignis.preventDefault()
              aufSprung?.()
            }
          }}
        />
      </label>

      {/* **Die Ebenen, erklärt, wo gesucht wird** — UX-04.AK-03
          (#1676).
          Am Feld steht, auf welcher Ebene ein Wert gilt; was die Ebenen
          unterscheidet, stand nirgends. **Außerhalb des `label`**: ein Knopf
          darin gehörte zur Beschriftung des Feldes. */}
      <Erklaerungsabsatz
        text={eigene.ebenenErklaerung}
        klasse="text-xs text-leise"
      />
    </div>
  )
}
