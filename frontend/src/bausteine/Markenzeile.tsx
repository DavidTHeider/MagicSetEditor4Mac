// **Marken an einem Eintrag jeder Liste** — UX-12.AK-01
// (#1684).
//
// Die Kacheln der Marken stehen am Eintrag; wer sie ändern will, klappt den
// Editor auf (`Objektmarken`). **Ein Knopf je Eintrag und kein Feld je
// Eintrag**: eine Liste aus zweihundert Sets mit zweihundert Eingaben wäre
// keine Liste mehr.
import { useState } from 'react'

import type { Markenart } from '../apiobjektmarken'
import { useSchreibsperre } from '../haken/schreibrecht'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Objektmarken } from './Objektmarken'

interface MarkenzeileEigenschaften {
  art: Markenart
  kennung: string

  // name nennt den Eintrag im Knopf — „Marken von Rot" und nicht achtmal
  // „Marken".
  name: string
  marken: readonly string[]
  nachladen: () => void

  // leise legt den Knopf hinter Zeiger und Fokus — UXR-24 (#1932): in einer
  // langen Liste stand „+ Marke" an jedem Eintrag und hob sie über R2. Er
  // erscheint, wenn die Zeile (`group/eintrag`, gesetzt von der Liste) unter
  // dem Zeiger steht oder den Fokus hat, und **bleibt im Baum** — für die
  // Tastatur und für VoiceOver. Offen ist er immer zu sehen.
  //
  // **Ausgeblendet nur, wo es Hover gibt** (N1 des Nachlaufs über 6.13,
  // #1906): auf einem Gerät ohne — Berührung, Stift — gäbe es keinen Weg,
  // ihn zu sehen, und ein Set ohne Marken zeigte keine Spur der Handlung.
  leise?: boolean
}

const knopfklassen =
  'rounded-full border border-kante px-2 py-0.5 text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-disabled:cursor-not-allowed aria-disabled:opacity-50'

// leiseklassen blenden den Knopf aus, bis die Zeile Zeiger oder Fokus hat —
// und nur, wo ein Zeiger Hover kennt.
const leiseklassen =
  '[@media(hover:hover)]:opacity-0 focus-visible:opacity-100 group-hover/eintrag:opacity-100 group-focus-within/eintrag:opacity-100'

export function Markenzeile({
  art,
  kennung,
  name,
  marken,
  nachladen,
  leise = false,
}: MarkenzeileEigenschaften) {
  const eigene = useTexte().bausteine.objektmarken
  const [offen, setzeOffen] = useState(false)

  // **Der Schalter öffnet nur, was schreibt** — G10 (#2276): ein Leser
  // sieht die Marken, bekommt aber kein Feld, das am Ende nicht geht.
  const schreibt = useSchreibsperre(true)

  return (
    <div className="flex flex-col gap-1 text-sm">
      <div className="flex flex-wrap items-center gap-1">
        {marken.length > 0 && (
          <ul className="flex flex-wrap gap-1">
            {marken.map((marke) => (
              <li
                key={marke}
                className="rounded-full border border-kante px-2 py-0.5 text-xs"
              >
                {marke}
              </li>
            ))}
          </ul>
        )}
        <button
          type="button"
          {...schreibt.knopf}
          title={schreibt.titel}
          aria-expanded={offen}
          aria-label={
            offen
              ? eigene.zeileSchliessen
              : einsetzen(eigene.zeileBearbeiten, { name })
          }
          onClick={() => {
            if (schreibt.gesperrt) return

            setzeOffen(!offen)
          }}
          className={[knopfklassen, leise && !offen ? leiseklassen : ''].join(
            ' ',
          )}
        >
          {offen ? eigene.zeileSchliessen : eigene.setzen}
          {schreibt.satz}
        </button>
      </div>
      {offen && (
        <Objektmarken
          art={art}
          kennung={kennung}
          marken={marken}
          geaendert={nachladen}
        />
      )}
    </div>
  )
}
