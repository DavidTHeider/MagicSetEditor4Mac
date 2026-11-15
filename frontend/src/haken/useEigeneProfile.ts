// **Die Profile des Menschen, der arbeitet** — Werkstück G8 von Phase 5.5
// (#1465).
//
// # Warum ein eigener Haken neben `useNutzer`
//
// **Weil jener sechs Wege geht.** `useNutzer` holt Nutzer, gezeigtes Profil,
// Anbieter, Konten, Bilder und Abzeichen — er trägt die ganze Nutzerseite.
// Wer hier nur zwei Namen braucht, zöge damit fünf Rufe mit, die niemand
// ansieht; und jeder Lauf über den Startassistenten müsste sie belegen.
//
// **Zwei Rufe und sonst nichts**: wer arbeitet, und welche Profile er hat.
//
// # Und warum die des Arbeitenden und nicht die des gezeigten Profils
//
// **Weil ein Sitz einem Profil desselben Nutzers gehört.** Wer in ein fremdes
// Profil wechselt, bleibt, wer er ist
// (#1400)
// — und darf einen Sitz nicht an ein Profil binden, das ihm nicht gehört.
import { api, auspacken } from '../api'
import { useLaden } from './useLaden'

import type { ProfilDTO } from '../apinutzer'

export interface Profilstand {
  // profile sind die des arbeitenden Menschen — leer, solange nichts geladen
  // ist.
  profile: ProfilDTO[]

  fehler: string | undefined
}

// anlass ist der Grund, die Liste **noch einmal** zu holen — leer heißt: nie.
//
// # Wozu einer nötig ist
//
// **Weil ein gelöschtes Profil sonst weiter in der Sichtwahl steht.** Der
// Befund kommt bei jedem Lesen frisch aus dem Bestand („dieses Profil gibt es
// nicht mehr"), die Namen daneben aus dieser Liste — und die stand seit dem
// Öffnen des Tisches. Gesehen am 2026-09-17 an der laufenden Anwendung: der
// Hinweis sagte, das Profil sei weg, und einen Zentimeter darüber stand sein
// Name. Zwei Sätze, die einander widersprechen, sind schlimmer als einer
// weniger (§10).
//
// **Und nicht die Fassung der Partie**: die ändert sich mit jedem Zug, und
// zwei Rufe an den Bestand je gelegter Karte wären Arbeit für nichts. Gefragt
// wird, wenn etwas sagt, dass eine Bindung ins Leere zeigt.
export function useEigeneProfile(anlass = ''): Profilstand {
  const stand = useLaden(async () => {
    const [nutzer, arbeitet] = await Promise.all([
      api().Nutzer(),
      api().WerArbeitet(),
    ])

    const wer = auspacken(arbeitet, () => arbeitet.nutzerID)

    return (
      auspacken(nutzer, () => nutzer.nutzer).find((einer) => einer.id === wer)
        ?.profile ?? []
    )
  }, [anlass])

  return { profile: stand.daten ?? [], fehler: stand.fehler }
}
