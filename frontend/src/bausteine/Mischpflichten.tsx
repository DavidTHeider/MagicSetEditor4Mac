// **Wer seit seiner letzten Suche nicht gemischt hat** — Werkstück G1 von
// Phase 5.5
// (#1458).
//
// **Ein eigenes Bauteil**, weil `Partietisch.tsx` an der Zeilengrenze aus §2
// stand — dieselbe Bewegung wie bei `Legendenwahlen` und
// `Ausloeserreihung`: was der Tisch **zeigt**, steht daneben; was er
// **zusammenhält**, in ihm.
import { useTexte } from '../sprache/kontext'
import { sitzname, zonenwort } from '../seiten/partiewoerter'

import type { Partie } from '../apipartie'
import { Abschnittstitel } from './Abschnittstitel'

// Mischpflichten zeigt, wer seit seiner letzten Suche nicht gemischt hat.
//
// **Für jeden sichtbar, und das ist ihr Zweck.** Sie ist die öffentliche
// Hälfte eines Blicks; wer sie nur dem Betroffenen zeigte, verschwiege dem
// Gegenüber, was gerade geschehen ist.
//
// **Ohne offene Pflicht steht nichts da** — seit dem 2026-09-24. Ein Satz,
// dass keine aussteht, nennte keinen nächsten Schritt (§6), und das Artboard
// *Tisch* führt die Mischpflichten gar nicht. Wo eine entsteht, erscheint
// sie — dieselbe Regel wie bei der offenen Wahl und der Reihung darüber.
export function Mischpflichten({ partie }: { partie: Partie }) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  if (partie.mischpflichten.length === 0) {
    return undefined
  }

  return (
    <div className="flex flex-col gap-1">
      <Abschnittstitel className="text-base font-semibold">
        {eigene.mischpflichtTitel}
      </Abschnittstitel>

      <ul className="flex flex-col gap-1 text-sm">
        {partie.mischpflichten.map((pflicht) => (
          <li key={`${pflicht.sitz}:${pflicht.zone}`}>
            {sitzname(partie, pflicht.sitz)} · {zonenwort(eigene, pflicht.zone)}
          </li>
        ))}
      </ul>
    </div>
  )
}
