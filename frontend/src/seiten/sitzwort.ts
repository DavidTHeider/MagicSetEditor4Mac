// **Wie ein Sitz heißt** — Werkstück G8 von Phase 5.5
// (#1465).
//
// **Eine eigene Datei**, weil `Partietisch.tsx` mit ihr über die Zeilengrenze
// aus §2 ginge — und weil der Schnitt der Frage folgt: dort steht, wie der
// Tisch **aussieht**, hier, **wie ein Platz heißt**.
import { useEigeneProfile } from '../haken/useEigeneProfile'

import type { Sitz } from '../apipartie'

// sitzwort gibt den Namen eines Sitzes — mit seinem Profil, wenn es eines hat.
//
// # Warum das Profil dabeisteht
//
// **Weil zwei Sitze denselben Namen tragen dürfen** (G8,
// #1465).
// Wer gegen sich selbst spielt, tippt zweimal seinen eigenen — und dann sagt
// die Liste zweimal dasselbe. Das Profil daneben trennt sie.
//
// **Und wenn es das Profil nicht mehr gibt**, steht nur der Name da; was
// fehlt, sagt der Befund darüber und nicht diese Zeile.
export function sitzwort(sitz: Sitz, namen: Map<string, string>): string {
  const profil = namen.get(sitz.profilID)

  return profil === undefined ? sitz.name : `${sitz.name} (${profil})`
}

// profilnamenAus gibt die Namen der Profile dieses Nutzers, nach Kennung.
//
// **Eine leere Abbildung, solange nichts geladen ist** — dann steht in der
// Sichtwahl nur der Name, und das ist der Stand von vorher.
function profilnamenAus(
  stand: ReturnType<typeof useEigeneProfile>,
): Map<string, string> {
  return new Map(stand.profile.map((profil) => [profil.id, profil.name]))
}

// useProfilnamen gibt die Namen der eigenen Profile, nach Kennung.
//
// **Und holt sie neu, sobald ein Befund dasteht** (G8, #1465): einer von
// ihnen sagt, dass eine Bindung ins Leere zeigt — dann darf die Sichtwahl den
// Namen nicht mehr führen. Warum der Anlass die Befunde sind und nicht die
// Fassung, steht bei [useEigeneProfile].
export function useProfilnamen(
  befunde: readonly string[],
): Map<string, string> {
  return profilnamenAus(useEigeneProfile(befunde.join('\u0000')))
}
