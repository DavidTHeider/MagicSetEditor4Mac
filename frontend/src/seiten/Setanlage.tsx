// **Ein Set anlegen — offen ohne Set, sonst in einer Schublade** — AP-10
// (#2367),
// Befund B-19, Entwurf `Wege09`: „In einer Schublade, ab dem ersten Set; ⌘N
// öffnet sie und setzt den Fokus in den Namen.“
//
// # Warum zwei Gestalten
//
// **Ohne Set ist das Anlegen der Einstieg** und steht offen neben der
// leeren Liste — es erklärt den ersten Schritt selbst. **Ab dem ersten Set**
// stand es bis AP-10 dauernd neben der Liste, drei Schritte samt Spiel- und
// Stilwahl, die man selten braucht; jetzt liegt es in einer Schublade, die
// „Neues Set“ und ⌘N öffnen. **Das Namensfeld bekommt den Fokus** — das tut
// `SetNeu` beim Anstoß, in beiden Gestalten (`menueablage.test.tsx` hat
// angestrichen, als ein erster Anlauf es hinter einem Aufklapp versteckte).
//
// # Was die Schublade behält
//
// **Einmal gezeichnet, bleibt sie im Baum** (`Activity`), wie die
// Paketschublade: eine halb getippte Eingabe übersteht das Schließen.
// **Nach dem Anlegen schließt sie** und das neue Set geht auf.
import { Activity, useState } from 'react'

import { Schublade } from '../bausteine/Schublade'
import { useTexte } from '../sprache/kontext'
import { SetNeu } from './SetNeu'

// Setanlagestand ist, was die Seite über die Schublade wissen muss — sie
// sperrt sich darunter.
export interface Setanlagestand {
  offen: boolean
  gezeichnet: boolean
  schliesse: () => void
}

// useSetanlage öffnet die Schublade auf einen Anstoß, **sobald es Sets
// gibt** — auch, wenn der Anstoß kam, während die Liste noch lud.
//
// **Auf den Anstoß, nicht auf seinen Stand**: ein Anstoß, der stehen bleibt,
// öffnet einmal — sonst ginge die Schublade nach dem Schließen gleich wieder
// auf, solange niemand ihn verbraucht hat.
export function useSetanlage(
  mitSets: boolean,
  anstoss: boolean,
): Setanlagestand {
  const [offen, setzeOffen] = useState(false)
  const [gezeichnet, setzeGezeichnet] = useState(false)
  const [vorher, setzeVorher] = useState(false)
  const [wartend, setzeWartend] = useState(false)

  if (anstoss !== vorher) {
    setzeVorher(anstoss)
    setzeWartend(anstoss)
  }

  if (wartend && mitSets) {
    setzeWartend(false)
    setzeOffen(true)
    setzeGezeichnet(true)
  }

  return {
    offen: offen && mitSets,
    gezeichnet,
    schliesse: () => {
      setzeOffen(false)
    },
  }
}

export function Setanlage({
  stand,
  gestalt,
  angelegt,
  anstoss,
  anstossVerbraucht,
  zuDenPaketen,
}: {
  stand: Setanlagestand
  // gestalt sagt, wo das Anlegen steht: offen ohne Set, in der Schublade mit
  // Sets, und gar nicht, solange die Liste lädt.
  gestalt: 'offen' | 'schublade' | 'keine'
  angelegt: (setID: string) => void
  anstoss: boolean
  anstossVerbraucht: () => void
  zuDenPaketen: () => void
}) {
  const texte = useTexte()

  if (gestalt === 'offen') {
    return (
      <SetNeu
        angelegt={angelegt}
        anstoss={anstoss}
        anstossVerbraucht={anstossVerbraucht}
        zuDenPaketen={zuDenPaketen}
      />
    )
  }

  if (gestalt === 'keine' || !stand.gezeichnet) {
    return null
  }

  return (
    <Activity mode={stand.offen ? 'visible' : 'hidden'}>
      <Schublade
        titel={texte.seiten.sets.neuTitel}
        schliessenBezeichnung={
          texte.allgemein.arbeitsplatz.schubladen.schliessen
        }
        schliessen={stand.schliesse}
      >
        <div className="px-4 pb-5 pt-3">
          <SetNeu
            angelegt={(setID) => {
              stand.schliesse()
              angelegt(setID)
            }}
            anstoss={anstoss}
            anstossVerbraucht={anstossVerbraucht}
            zuDenPaketen={zuDenPaketen}
            ueberschriftVerborgen
          />
        </div>
      </Schublade>
    </Activity>
  )
}
