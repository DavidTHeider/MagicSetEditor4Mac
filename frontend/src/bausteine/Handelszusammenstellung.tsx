// **Was hin und her geht, als Tisch oder als Liste** — P12
// (#2380).
//
// **Bearbeitbar steht der Markttisch da**, und eine Wahl darüber führt zur
// bisherigen Liste; **nach der Einigung** gibt es nichts mehr zu ziehen, und
// die Zeilen stehen nur da.
import { Handelspositionen } from './Handelspositionen'
import { Knopf } from './Knopf'
import { Markttisch } from './Markttisch'
import { useHandelsansicht } from '../haken/useHandelsansicht'
import { useTexte } from '../sprache/kontext'

import type { Handelsposition } from '../apihandelsleitung'

interface Eigenschaften {
  ich: string
  kennung: string
  name: string
  positionen: Handelsposition[]
  setzePositionen: (positionen: Handelsposition[]) => void
  bearbeitbar: boolean
}

export function Handelszusammenstellung({
  ich,
  kennung,
  name,
  positionen,
  setzePositionen,
  bearbeitbar,
}: Eigenschaften) {
  const tisch = useTexte().seiten.handeln.markttisch
  const { ansicht, waehle } = useHandelsansicht()

  const liste = (
    <Handelspositionen
      ich={ich}
      gegenueber={kennung}
      positionen={positionen}
      setzePositionen={setzePositionen}
      bearbeitbar={bearbeitbar}
    />
  )

  if (!bearbeitbar) {
    return liste
  }

  return (
    <div className="flex flex-col gap-2">
      <div role="group" aria-label={tisch.ansicht} className="flex gap-1">
        <Knopf
          gewicht="leise"
          gedrueckt={ansicht === 'tisch'}
          aufKlick={() => {
            waehle('tisch')
          }}
        >
          {tisch.alsTisch}
        </Knopf>
        <Knopf
          gewicht="leise"
          gedrueckt={ansicht === 'liste'}
          aufKlick={() => {
            waehle('liste')
          }}
        >
          {tisch.alsListe}
        </Knopf>
      </div>

      {ansicht === 'tisch' ? (
        <Markttisch
          ich={ich}
          gegenueber={kennung}
          name={name}
          positionen={positionen}
          setzePositionen={setzePositionen}
        />
      ) : (
        liste
      )}
    </div>
  )
}
