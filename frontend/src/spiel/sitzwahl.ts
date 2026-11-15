// **Was in der Deckwahl eines Sitzes steht** — ein eigenes Deck, eine
// angenommene Leihe (P1b, #2370) oder eine Einladung (P9c,
// #2377).
//
// **Ein Wert und ein Vorsatz** je Art, damit keiner einer Deckkennung gleichen
// kann; und **genau eines** gilt — wer einlädt, legt kein Deck an den Sitz,
// denn das bringt der Gast mit (Vermerk 0199).
import type { Sitzwunsch } from '../seiten/Sitzliste'

const leihvorsatz = 'leihe:'
const einladevorsatz = 'einladen:'

export const leihwert = (kennung: string) => `${leihvorsatz}${kennung}`
export const einladewert = (kennung: string) => `${einladevorsatz}${kennung}`

// deckwert ist, was die Deckwahl dieses Sitzes zeigt.
export function deckwert(wunsch: Sitzwunsch): string {
  if (wunsch.eingeladen !== undefined && wunsch.eingeladen !== '') {
    return einladewert(wunsch.eingeladen)
  }

  return wunsch.leiheID === '' ? wunsch.deckID : leihwert(wunsch.leiheID)
}

// gewaehlterSitz ist der Sitz nach einer Wahl — **mit einer Einladung auch
// mit dem Namen des Gastes**, denn den liest der Wirt danach am Tisch.
export function gewaehlterSitz(
  wunsch: Sitzwunsch,
  wert: string,
  kontakte: readonly { kennung: string; name: string }[],
): Sitzwunsch {
  if (wert.startsWith(einladevorsatz)) {
    const kennung = wert.slice(einladevorsatz.length)
    const kontakt = kontakte.find((einer) => einer.kennung === kennung)

    return {
      ...wunsch,
      name: kontakt?.name ?? wunsch.name,
      deckID: '',
      leiheID: '',
      eingeladen: kennung,
    }
  }

  if (wert.startsWith(leihvorsatz)) {
    return {
      ...wunsch,
      deckID: '',
      leiheID: wert.slice(leihvorsatz.length),
      eingeladen: '',
    }
  }

  return { ...wunsch, deckID: wert, leiheID: '', eingeladen: '' }
}
