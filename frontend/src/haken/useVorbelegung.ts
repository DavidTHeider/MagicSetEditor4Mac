// **Woher die Vorbelegung der Sitze kommt** — P6
// (#2374):
// das gezeigte Profil (der Name am ersten Sitz) und die zuletzt geöffneten
// Decks. Den Namen am zweiten Sitz bringt die Partieliste der Seite mit — sie
// steht schon da, ein zweites Lesen wäre doppelte Arbeit. Gerechnet wird in
// `spiel/vorbelegung.ts`.
//
// **Ein Lesefehler kostet nur die Vorbelegung** — das Formular bleibt, wie es
// ohne sie war, und niemandem wird ein Deck genommen.
import { api, auspacken } from '../api'
import { useLaden } from './useLaden'

export interface Vorbelegungsdaten {
  profilname: string
  zuletzt: string[]
}

const ohne: Vorbelegungsdaten = { profilname: '', zuletzt: [] }

// useVorbelegung gibt die Quellen — `undefined`, solange sie laden.
export function useVorbelegung(): Vorbelegungsdaten | undefined {
  const stand = useLaden(async () => {
    const [profil, weiter] = await Promise.all([
      api().GezeigtesProfil(),
      api().Weiterarbeiten(),
    ])

    return {
      profilname: auspacken(profil, () => profil.profil?.name ?? ''),
      zuletzt: auspacken(weiter, () => weiter.geoeffnet)
        .filter((schritt) => schritt.objekt.art === 'deck')
        .map((schritt) => schritt.objekt.kennung),
    }
  }, [])

  return stand.fehler === undefined ? stand.daten : ohne
}
