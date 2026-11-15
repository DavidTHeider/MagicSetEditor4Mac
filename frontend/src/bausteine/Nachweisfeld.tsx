// **Das Feld für den zweiten Faktor** — Werkstück 7 von Phase 4.12
// (#1351).
//
// # Warum es erst erscheint, wenn es verlangt wird
//
// **Weil die meisten keinen zweiten Faktor haben** — und ein Feld, das
// dauernd danebensteht, ist eine Frage an jemanden, der sie nicht beantworten
// muss (§6). Der Aufrufer versucht es ohne, und erst wenn die Anwendung einen
// verlangt, klappt er dieses Feld auf.
//
// # Warum es beide Sorten Code nimmt
//
// **Weil der Nutzer den Unterschied nicht kennen soll.** Sechs Ziffern aus der
// App oder zwölf Zeichen vom Zettel — beides ist „mein zweiter Faktor", und
// der Dienst unterscheidet sie selbst (`service.Zweiterfaktor`).
import { useId, useState } from 'react'

import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'

interface Eigenschaften {
  angehalten?: boolean

  // abgeschickt bekommt den Code, den jemand eingegeben hat.
  abgeschickt: (nachweis: string) => void
}

export function Nachweisfeld({ angehalten, abgeschickt }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.bausteine.nachweis
  const feldID = useId()

  const [wert, setzeWert] = useState('')

  return (
    <span className="flex items-center gap-2">
      <label htmlFor={feldID} className="text-sm">
        {eigene.beschriftung}
      </label>

      <input
        id={feldID}
        type="text"
        value={wert}
        disabled={angehalten}
        // **Kein Ausfüllen durch den Browser**: ein Einmalkennwort ist
        // einmalig, und ein gemerktes wäre bereits verbraucht.
        autoComplete="one-time-code"
        inputMode="text"
        className="w-40 rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
        onChange={(ereignis) => {
          setzeWert(ereignis.target.value)
        }}
        onKeyDown={(ereignis) => {
          // **Die Eingabetaste schickt ab** — wer sechs Ziffern tippt, will
          // nicht danach noch zur Maus greifen (§6, Tastaturbedienbarkeit).
          if (ereignis.key === 'Enter' && wert.trim() !== '') {
            abgeschickt(wert.trim())
          }
        }}
      />

      <Knopf
        gewicht="leise"
        angehalten={angehalten === true || wert.trim() === ''}
        aufKlick={() => {
          abgeschickt(wert.trim())
        }}
      >
        {eigene.abschicken}
      </Knopf>
    </span>
  )
}
