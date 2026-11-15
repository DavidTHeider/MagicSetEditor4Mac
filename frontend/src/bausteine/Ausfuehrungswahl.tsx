// **Die Wahl der physischen Ausführung** — Werkstück T2 von Phase 5.6
// (#1467,
// [0119](../../../docs/entscheidungen/0119-eine-ausfuehrung-statt-eines-wahrheitswerts.md)).
//
// # Warum ein eigener Baustein
//
// **Weil es zwei Stellen gibt, die sie brauchen** — das Formular und der
// Inspector. Bis T2 stand dort je ein Häkchen, zweimal derselbe Code; mit drei
// Ausführungen wären es zweimal dieselbe Aufzählung, und die läuft
// auseinander (§16, derselbe Grund wie auf der Go-Seite).
//
// # Warum kein Häkchen mehr
//
// **Weil ein Häkchen zwei Zustände kennt und es drei gibt.** Was `etched` ist,
// ließe sich mit zwei Kästchen nur so ausdrücken, dass beide angekreuzt etwas
// bedeuten — und das rät niemand.
import { useId } from 'react'

import type { Finish } from '../api'
import { useTexte } from '../sprache/kontext'

interface AusfuehrungswahlEigenschaften {
  wert: Finish
  setzeWert: (wert: Finish) => void
}

export function Ausfuehrungswahl({
  wert,
  setzeWert,
}: AusfuehrungswahlEigenschaften) {
  const eigene = useTexte().seiten.sammlung
  const kennung = useId()

  return (
    <label className="flex flex-col gap-1">
      <span className="text-sm font-medium" id={kennung}>
        {eigene.ausfuehrung}
      </span>
      <select
        aria-labelledby={kennung}
        // **Der Nullwert wird hier beantwortet** — leer heißt `nonfoil`
        // (0119, Satz 1). Ohne diese Zeile zeigte das Feld den ersten
        // Eintrag, und der Entwurf sagte etwas anderes als das Feld.
        value={wert === ('' as Finish) ? 'nonfoil' : wert}
        onChange={(ereignis) => {
          setzeWert(ereignis.target.value as Finish)
        }}
        className={wahlfeld}
      >
        <option value="nonfoil">{eigene.nonfoil}</option>
        <option value="foil">{eigene.foil}</option>
        <option value="etched">{eigene.etched}</option>
      </select>
    </label>
  )
}

const wahlfeld =
  'rounded-sm border border-kante bg-transparent px-2 py-1 text-sm ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'
