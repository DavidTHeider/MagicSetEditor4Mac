// **Mit wem man im Netz spielen kann** — P9c
// (#2377),
// Artboard `Netzpartie09`.
//
// **Bestätigt und mit dem Recht am Tisch** — dieselbe Auswahl, die
// `Kontaktwahl` für den Beitritt trifft und `service.Peersitzungen.Oeffne` vor
// dem Wählen prüft. Ein Fehler beim Laden heißt hier: niemand — die kurzen
// Wege fehlen dann, der lange in der Schublade bleibt und sagt den Fehler.
import { api, auspacken } from '../api'
import { useLaden } from './useLaden'

// Partiekontakt ist ein Gegenüber, wie ein Mensch es liest.
export interface Partiekontakt {
  kennung: string
  name: string
}

const derBestaetigteStand = 'bestaetigt'
const dasRechtAmTisch = 'partie'

export function usePartiekontakte(): Partiekontakt[] {
  const liste = useLaden(async () => {
    const antwort = await api().Peerkontakte()

    return auspacken(antwort, () => antwort.kontakte)
      .filter(
        (kontakt) =>
          kontakt.stand === derBestaetigteStand &&
          kontakt.rechte.includes(dasRechtAmTisch),
      )
      .map((kontakt) => ({
        kennung: kontakt.kennung,
        name: kontakt.eigenerName || kontakt.selbstgewaehlterName,
      }))
  }, [])

  return liste.daten ?? []
}
