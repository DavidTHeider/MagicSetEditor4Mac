// **Liest wie die Paketseite** (U14, #2159): die Abhängigkeiten einmal, und
// reicht sie an Liste oder Abschnitt — für Läufe, die nur einen davon
// zeichnen.
import { useAbhaengigkeiten } from '../haken/useAbhaengigkeiten'
import { Abhaengigkeiten } from '../seiten/Abhaengigkeiten'
import { Paketliste } from '../seiten/Paketliste'
import { usePakete } from '../haken/usePakete'

export function PaketlisteMitBefunden({ runde = 0 }: { runde?: number }) {
  const { daten } = useAbhaengigkeiten(runde)
  const pakete = usePakete(runde)

  return <Paketliste pakete={pakete} befunde={daten} />
}

export function AbhaengigkeitenGelesen({ runde = 0 }: { runde?: number }) {
  const { daten, fehler } = useAbhaengigkeiten(runde)

  return <Abhaengigkeiten daten={daten} fehler={fehler} />
}
