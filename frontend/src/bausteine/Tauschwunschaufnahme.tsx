// **Was einem Deck fehlt, steht als Wunsch im Handel** — UX-03.AK-01
// (#1675).
//
// **Ein Knopf und nicht gleich**: die Positionen gehören dem Vorschlag, den
// ein Mensch schickt; wer eintragen lässt, sieht vorher, was hineinkommt.
import { Knopf } from './Knopf'
import { useAnschluss } from '../anschlussraum'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Stueck } from '../anschluss'

export function Tauschwunschaufnahme({
  eintragen,
}: {
  eintragen: (stuecke: readonly Stueck[]) => void
}) {
  const eigene = useTexte().seiten.handeln
  const anschluss = useAnschluss()

  const wunsch =
    anschluss.offen?.art === 'tauschwunsch' ? anschluss.offen : undefined

  if (wunsch === undefined) {
    return null
  }

  const liste = wunsch.stuecke
    .map((stueck) => `${stueck.name} × ${String(stueck.menge)}`)
    .join(', ')

  return (
    <div
      role="group"
      aria-label={eigene.wunschTitel}
      className="flex flex-col gap-2 rounded-sm border border-kante p-3"
    >
      <p className="text-sm">
        {einsetzen(eigene.wunschWartet, { name: wunsch.name, liste })}
      </p>
      <div>
        <Knopf
          gewicht="haupt"
          aufKlick={() => {
            eintragen(wunsch.stuecke)

            // **Verbraucht ist verbraucht** (§7, Regel 3).
            anschluss.verbrauche()
          }}
        >
          {eigene.wunschEintragen}
        </Knopf>
      </div>
    </div>
  )
}
