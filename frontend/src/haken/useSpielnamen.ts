// **Welche Spiele hier installiert sind** — als Vorschläge für ein Feld, das
// ein Spiel nennt: UX-06.AK-01
// (#1678).
//
// # Vorschläge und keine Auswahl
//
// **Ein Format oder ein Turnier darf ein Spiel nennen, das hier nicht
// installiert ist** — ein eingelesenes Turnier kommt von einem anderen
// Rechner. Eine Auswahl aus den installierten schlösse das aus, und genau
// das verbietet das Kriterium: *„fachlich offene Werte behalten Freitext mit
// Vorschlägen"*.
//
// **Ein Fehler beim Laden nimmt nur die Vorschläge weg**, nicht das Feld:
// wer ein Spiel kennt, tippt es.
import { api, auspacken } from '../api'
import { useLaden } from './useLaden'

import type { Vorschlag } from '../bausteine/Vorschlagsfeld'

export function useSpielnamen(): readonly Vorschlag[] {
  const pakete = useLaden(holePakete, [])

  return (pakete.daten ?? [])
    .filter((paket) => paket.art === 'mse-game')
    .map((paket) => ({ wert: paket.name }))
}

async function holePakete() {
  const antwort = await api().Pakete()

  return auspacken(antwort, () => antwort.pakete)
}
