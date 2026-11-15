// **Ein Gegenüber von Hand eintragen** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum es diesen Weg gibt
//
// **Weil es die Begegnung noch nicht gibt.** Der Transport kommt mit N2 und
// N3; bis dahin ist der Austausch der Kennung von Hand der einzige Weg — und
// er ist kein Notbehelf, sondern genau das, wofür der Zweck `paarung` da ist.
//
// **Was hier entsteht, darf nichts.** Ein eingetragenes Gegenüber ist
// unbestätigt; erst der Vergleich des Fingerabdrucks und danach eine Freigabe
// je Zweck erlauben etwas.
import { useId, useState } from 'react'

import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { kennungseingabe } from './kennungseingabe'

interface Kontaktaufnahmeeigenschaften {
  laeuft: boolean
  aufnehmen: (kennung: string, name: string) => void
}

export function Kontaktaufnahme({
  laeuft,
  aufnehmen,
}: Kontaktaufnahmeeigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const kennungsfeld = useId()
  const namensfeld = useId()

  const [kennung, setzeKennung] = useState('')
  const [name, setzeName] = useState('')

  return (
    <div className="flex flex-col gap-2 rounded-sm border border-kante p-3">
      <Erklaerungsabsatz
        text={eigene.kontakteAufnehmenErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      <label className="flex flex-col gap-1 text-sm" htmlFor={kennungsfeld}>
        {eigene.kontakteAufnehmenKennung}
        <input
          {...kennungseingabe}
          id={kennungsfeld}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 font-mono text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          value={kennung}
          onChange={(ereignis) => {
            setzeKennung(ereignis.target.value)
          }}
        />
      </label>

      <label className="flex flex-col gap-1 text-sm" htmlFor={namensfeld}>
        {eigene.kontakteAufnehmenName}
        <input
          id={namensfeld}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          value={name}
          onChange={(ereignis) => {
            setzeName(ereignis.target.value)
          }}
        />
      </label>

      <Knopf
        gewicht="haupt"
        angehalten={laeuft || kennung === ''}
        aufKlick={() => {
          aufnehmen(kennung, name)
          setzeKennung('')
          setzeName('')
        }}
      >
        {eigene.kontakteAufnehmen}
      </Knopf>
    </div>
  )
}
