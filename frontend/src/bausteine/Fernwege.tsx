// **Wie die Gegenüber angebunden sind** — NET-02.AK-03
// (#1706).
//
// # Warum auf Knopfdruck und nicht im Takt
//
// **Der Stand fragt tailcat nach seinen Gegenstellen.** Ein Takt, der das
// still täte, liefe auch, wenn niemand hinsieht — und der Kasten steht in
// den Einstellungen, nicht auf dem Tisch.
import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'

import type { Fernweg } from '../apipeerkontakt'

interface Eigenschaften {
  wege: Fernweg[]

  // angebunden zählt die offenen Ströme. **Mit leerer Wegliste heißt das:
  // der Weg ist unbestimmt**, nicht „niemand" (SEC-21, #2162).
  angebunden: number
  neuLesen: () => void
}

export function Fernwege({ wege, angebunden, neuLesen }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.technik

  const zeile = (weg: Fernweg): string => {
    switch (weg.art) {
      case 'direkt':
        return einsetzen(eigene.fernwege.direkt, { adresse: weg.adresse })
      case 'relais':
        return einsetzen(eigene.fernwege.relais, { adresse: weg.adresse })
      case 'aufbau':
        return eigene.fernwege.aufbau
    }
  }

  return (
    <div className="flex flex-col items-start gap-2">
      {wege.length === 0 ? (
        <p className="text-sm text-leise">
          {angebunden > 0
            ? einsetzen(eigene.fernwege.unbestimmt, {
                anzahl: String(angebunden),
              })
            : eigene.fernwege.keine}
        </p>
      ) : (
        <ul aria-label={eigene.fernwege.titel} className="flex flex-col gap-1">
          {wege.map((weg, stelle) => (
            <li key={stelle} className="font-mono text-sm break-all">
              {zeile(weg)}
            </li>
          ))}
        </ul>
      )}

      <Knopf gewicht="leise" aufKlick={neuLesen}>
        {texte.bausteine.zustaende.neuLaden}
      </Knopf>
    </div>
  )
}
