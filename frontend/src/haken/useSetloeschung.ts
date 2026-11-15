// **Das Löschen eines Sets im Fenster** —
// #1218.
//
// # Warum ein eigener Haken
//
// Weil drei der vier Regeln aus §7 hier zusammenkommen und keine davon in die
// Komponente gehört:
//
//   - **Regel 4** — ein Vorgang je Set. Zwei Klicks auf „Endgültig löschen"
//     schickten zwei Aufrufe gegen dasselbe Set; der zweite fände es nicht
//     mehr und meldete einen Fehler für etwas, das geglückt ist.
//   - **Regel 3** — nach einem *gescheiterten* Löschen wird **nicht**
//     nachgeladen: ein Lesefehler überschriebe die Meldung, die sagt, was zu
//     tun ist. Nach einem geglückten schon, denn die Liste ist einen Eintrag
//     kürzer.
//   - **Regel 1** — wer das *offene* Set löscht, hat kein offenes Set mehr.
//     Was das heißt, weiß diese Ebene nicht; sie meldet nur, welches weg ist.
import { useCallback, useMemo, useState } from 'react'

import { api, auspacken } from '../api'
import { epocheBekannt, kennungGeloescht } from '../bestandsepoche'
import { meldungAus } from '../meldung'
import { type Katalog } from '../sprache/kataloge'
import { useVorgang } from './useVorgang'

export interface Setloeschung {
  loeschen: (setID: string, name: string) => Promise<void>
  laeuft: (setID: string) => boolean

  // befunde sind die Reste eines **geglückten** Löschens — je Set.
  befunde: (setID: string) => string[]

  fehler: (setID: string) => string | undefined
}

// useSetloeschung baut ihn über der Fassade.
//
// **`nachladen` und `geloescht` sind zwei Rückrufe und nicht einer.** Das
// Nachladen betrifft die **Liste**, das Melden das **offene Set** — zwei
// Ebenen, und die eine gehört nicht der anderen.
export function useSetloeschung(
  texte: Katalog,
  nachladen: () => void,
  geloescht: (setID: string) => void,
): Setloeschung {
  const vorgang = useVorgang()
  const [befunde, setzeBefunde] = useState<Record<string, string[]>>({})
  const [fehler, setzeFehler] = useState<Record<string, string>>({})

  const loeschen = useCallback(
    async (setID: string, name: string) => {
      await vorgang.starte(`loeschen:${setID}`, async () => {
        setzeFehler((bisher) => ohne(bisher, setID))
        const epoche = epocheBekannt()

        // **Eine Antwort aus dem vorigen Bestand ändert nichts** (§7, Regel 2;
        // Nachlauf über die Folgen vor 0.8.0, Befund 1): wurde inzwischen
        // zurückgespielt, gehört sie zu einem Bestand, den das Fenster nicht
        // mehr zeigt — und ihr Melden schlösse womöglich ein Set, das jetzt
        // offen ist. Für Erfolg wie Fehler.
        const veraltet = () => epocheBekannt() !== epoche

        try {
          const antwort = await api().SetLoeschen(setID, name)
          if (veraltet()) return

          const reste = auspacken(antwort, () => antwort.befunde)

          setzeBefunde((bisher) => ({ ...bisher, [setID]: reste }))
          kennungGeloescht(setID, epoche)

          // **Erst melden, dann nachladen.** Wer das offene Set löscht, soll
          // nicht erst eine Liste ohne es sehen und dann eine leere Kartenseite.
          geloescht(setID)
          nachladen()
        } catch (ursache: unknown) {
          if (veraltet()) return

          setzeFehler((bisher) => ({
            ...bisher,
            [setID]: meldungAus(ursache, texte),
          }))
        }
      })
    },
    [vorgang, texte, nachladen, geloescht],
  )

  return useMemo(
    () => ({
      loeschen,
      laeuft: (setID: string) => vorgang.laeuft(`loeschen:${setID}`),
      befunde: (setID: string) => befunde[setID] ?? [],
      fehler: (setID: string) => fehler[setID],
    }),
    [loeschen, vorgang, befunde, fehler],
  )
}

// ohne nimmt einen Schlüssel aus einer Abbildung.
//
// **Als eigene Funktion und nicht mit `delete`**: der Zustand einer Komponente
// wird ersetzt, nicht angefasst — sonst zeichnet React nicht neu.
function ohne(bisher: Record<string, string>, schluessel: string) {
  return Object.fromEntries(
    Object.entries(bisher).filter(([vorhanden]) => vorhanden !== schluessel),
  )
}
