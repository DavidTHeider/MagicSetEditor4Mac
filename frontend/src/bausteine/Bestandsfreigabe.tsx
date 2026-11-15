// **Der eigene Bestand geht nur mit Kreuz hinaus** — Werkstück A13.4 von
// Phase 5.3
// (#1452).
//
// # Warum ein eigenes Bauteil
//
// **Weil drei Stellen dieselbe Entscheidung stellen** — die Recherche, die
// Ähnlichkeit und die Kontextfreigabe daneben. Drei Fassungen desselben
// Satzes liefen beim ersten Sonderfall auseinander, und jemand läse an der
// einen Stelle etwas anderes über dieselbe Sache.
//
// # Und warum es ein Kreuz ist und keine Voreinstellung
//
// **Weil das Issue es verlangt**: „Referenzkarten/Decks/private Setnotizen
// **nur ausdrücklich ausgewählt**". Ein Schalter, der aufgeräumt vorbelegt
// wäre, ist keine Auswahl — er ist eine Vorgabe, an die sich niemand
// erinnert.
import { useId } from 'react'

import { derEigeneBestand } from '../seiten/recherchefeld'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

interface FreigabeEigenschaften {
  // freigegeben ist die Auswahl — leer heißt: nichts.
  freigegeben: string[]

  waehlen: (neu: string[]) => void
}

export function Bestandsfreigabe({
  freigegeben,
  waehlen,
}: FreigabeEigenschaften) {
  const eigene = useTexte().seiten.recherche

  // **Eine Kennung je Einbau** (#1553): zwei Kästen auf derselben Seite
  // trügen sonst dieselbe `id`, und die Beschriftung zeigte auf den falschen.
  const feldID = useId()

  const an = freigegeben.includes(derEigeneBestand)

  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={feldID} className="flex items-center gap-2 text-sm">
        <input
          id={feldID}
          type="checkbox"
          className="focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          checked={an}
          onChange={(ereignis) => {
            waehlen(
              ereignis.target.checked
                ? [...freigegeben, derEigeneBestand]
                : freigegeben.filter((einer) => einer !== derEigeneBestand),
            )
          }}
        />
        {eigene.freigabeEigeneKarten}
      </label>

      <Erklaerungsabsatz
        text={eigene.freigabeBestandErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />
    </div>
  )
}
