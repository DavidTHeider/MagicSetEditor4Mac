// **Das Tor der Sitzung** — ID-02
// (#1701).
//
// **Gesperrt steht der Sperrbildschirm statt der Anwendung da, abgemeldet die
// Anmeldung.** Wer entsperrt, bekommt die Anwendung frisch aufgebaut — mit dem
// Stand, den Go nach dem Entsperren wieder herausgibt.
import { useCallback, useState, type ReactNode } from 'react'

import { api } from '../fassade'
import { useLaden } from '../haken/useLaden'
import { Sitzungsraum } from '../haken/useSitzungsschritt'
import { auspacken } from '../umschlag'
import { LokaleAnmeldung } from './LokaleAnmeldung'
import { Sperrbildschirm } from './Sperrbildschirm'

interface Eigenschaften {
  children: ReactNode

  // geoeffnet lässt darüber neu lesen, was vom Bestand kommt — gesperrt
  // kamen für die Einstellungen eines Profils nur Vorgabewerte an.
  geoeffnet: () => void
}

export function Sitzungstor({ children, geoeffnet }: Eigenschaften) {
  const [runde, setzeRunde] = useState(0)
  const neuLesen = useCallback(() => {
    setzeRunde((bisher) => bisher + 1)
    geoeffnet()
  }, [geoeffnet])

  const stand = useLaden(async () => {
    const antwort = await api().Sitzungsstand()

    return auspacken(antwort, () => antwort)
  }, [runde])

  // **Bis die Antwort da ist, nichts** — die Anwendung spränge sonst einen
  // Augenblick auf und wieder zu. **Ein Fehler beim Fragen lässt sie stehen**:
  // die Sperre hält Go ohnehin, und ein Fenster ohne jede Antwort hülfe
  // niemandem.
  if (stand.daten === undefined && stand.fehler === undefined) {
    return null
  }

  if (stand.daten?.zustand === 'gesperrt') {
    return <Sperrbildschirm stand={stand.daten} geoeffnet={neuLesen} />
  }

  if (stand.daten?.zustand === 'abgemeldet') {
    return <LokaleAnmeldung angemeldet={neuLesen} />
  }

  return (
    <Sitzungsraum.Provider value={neuLesen}>{children}</Sitzungsraum.Provider>
  )
}
