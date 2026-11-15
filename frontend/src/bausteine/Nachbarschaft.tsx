// **Wer in der Nähe ist** — der Abschnitt drumherum. Werkstück N2 von
// Phase 5.4
// (#1455).
//
// **Ein eigener Abschnitt und kein Teil des Horchpostens**, weil er eine
// andere Richtung hat: dort geht es darum, **gefunden zu werden**, hier
// darum, **zu finden**. Das eine geht ohne das andere.
import { useId } from 'react'

import { Nachbarn } from './Nachbarn'
import { useTexte } from '../sprache/kontext'

interface Nachbarschaftseigenschaften {
  // aufGeaendert meldet eine Aufnahme — die Liste der Gegenüber daneben
  // lädt dann nach (§7, Regel 3).
  aufGeaendert?: (() => void) | undefined
}

export function Nachbarschaft({
  aufGeaendert,
}: Nachbarschaftseigenschaften = {}) {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const ueberschrift = useId()

  return (
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-3">
      <h3 id={ueberschrift} className="text-lg font-semibold">
        {eigene.nachbarnTitel}
      </h3>

      <Nachbarn aufGeaendert={aufGeaendert} />
    </section>
  )
}
