// **Tisch oder Liste** — P12
// (#2380):
// der Markttisch ist die Vorgabe, **der bisherige Assistent bleibt als Liste
// erreichbar** — für alle, die sich an ihn gewöhnt haben (Risiko des Issues:
// disruptiv für Bestandsnutzende).
//
// **Gemerkt je Rechner und nicht im Bestand**: eine Ansicht ist eine
// Vorliebe dessen, der davor sitzt, kein Teil eines Handels. Ohne Speicher
// gilt die Vorgabe.
import { useState } from 'react'

export type Handelsansicht = 'tisch' | 'liste'

const schluessel = 'mse4mac.handel.ansicht'

function gemerkt(): Handelsansicht {
  try {
    return localStorage.getItem(schluessel) === 'liste' ? 'liste' : 'tisch'
  } catch {
    return 'tisch'
  }
}

export function useHandelsansicht() {
  const [ansicht, setzeAnsicht] = useState<Handelsansicht>(gemerkt)

  const waehle = (neu: Handelsansicht) => {
    setzeAnsicht(neu)
    try {
      localStorage.setItem(schluessel, neu)
    } catch {
      // **Ohne Speicher gilt sie bis zum nächsten Öffnen** — mehr nicht.
    }
  }

  return { ansicht, waehle }
}
