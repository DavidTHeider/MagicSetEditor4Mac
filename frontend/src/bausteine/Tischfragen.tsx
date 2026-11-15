// **Was das Spiel anhält, steht über dem Tisch** — der Beginn (#1839), die
// offene Legendenwahl (M12, #1614) und die Reihung wartender Fähigkeiten
// (M9, #1611).
//
// # Warum an einer Stelle
//
// **Weil sie dieselbe Frage beantworten**: worauf wartet der Tisch gerade?
// Wer eine davon unter den Befehlen suchen müsste, hielte den Tisch für
// hängengeblieben — CR 704.5j lässt die Partie bis zur Wahl stehen, CR 603.3
// gibt niemandem Priorität, solange die Reihung offen ist, und vor dem Ende
// des Beginns kommt kein erster Zug (CR 103.8).
//
// **Bis #1839 standen die zwei älteren einzeln in `Partietisch`**, und die
// Seite stand damit an der Zeilengrenze aus §2.
import { Ausloeserreihung } from './Ausloeserreihung'
import { Eroeffnung } from './Eroeffnung'
import { Legendenwahlen } from './Legendenwahlen'
import { partievorgang } from '../haken/partiewege'

import type { Partie } from '../apipartie'
import type { Partielauf } from '../haken/usePartie'

export function Tischfragen({
  lauf,
  partie,
}: {
  lauf: Partielauf
  partie: Partie
}) {
  const angehalten = lauf.laeuft(partievorgang(partie.id))

  // **Nach dem Ende fragt der Tisch nichts mehr** — P4 (#2372): endete eine
  // Partie mitten im Beginn, stand die Starthand-Frage weiter da. Das
  // Ergebnis steht oben (`Partieausgang`).
  if (partie.stand === 'beendet') {
    return undefined
  }

  return (
    <>
      <Eroeffnung
        partie={partie}
        erlaubt={lauf.erlaubt}
        angehalten={angehalten}
        befehlen={lauf.befehlen}
      />

      <Legendenwahlen
        partie={partie}
        angehalten={angehalten}
        befehlen={lauf.befehlen}
      />

      <Ausloeserreihung
        partie={partie}
        angehalten={angehalten}
        befehlen={lauf.befehlen}
      />
    </>
  )
}
