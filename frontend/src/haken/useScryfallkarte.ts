// **Eine Karte aus einem Scryfall-Namen anlegen**
// (#1032).
//
// Ein eigener Haken, weil `Karten.tsx` mit ihm über 400 Zeilen ginge (§2) —
// und weil der Schnitt einer Frage folgt: dort steht, **wie** die Lasche
// angeordnet ist, hier, **was** dieser Vorgang tut.
import type { Katalog } from '../sprache/kataloge'
import { api } from '../api'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import type { Vorgang } from './useVorgang'
import { useKartenwunsch } from './useKartenwunsch'

interface Anlagewege {
  setID: string
  vorgang: Vorgang
  texte: Katalog

  // nochAktuell sagt, ob die Antwort noch zu dem Stand gehört, aus dem sie
  // stammt — §7, Regel 2.
  nochAktuell: (beim: string) => boolean

  aufgeschlagen: (karteID: string) => void
  gemeldet: (meldung: string | undefined) => void
}

// useScryfallkarte gibt die Handlung, die eine Karte aus einem Namen anlegt.
//
// # Die vier Regeln aus §7
//
//   - **Regel 4**: ein Vorgang je Datensatz — der Schlüssel ist
//     `scryfallkarte`, und der Knopf hängt daran.
//   - **Regel 2**: die Antwort gehört zu dem Set, aus dem sie stammt.
//   - **Regel 3**: geschrieben ist geschrieben — die angelegte Karte wird
//     aufgeschlagen, **bevor** die Meldung kommt.
//   - **Regel 1** liegt beim Aufrufer: ein Setwechsel räumt den Kontext.
//
// # Warum die Karte auch im Fehlerfall aufgeschlagen wird
//
// Weil sie **da** ist (#155).
// Scheitert das Füllen nach dem Anlegen, bleibt die Karte im Bestand; wer sie
// nicht zeigt, lässt sie dort zurück, ohne dass jemand davon weiß.
export function useScryfallkarte({
  setID,
  vorgang,
  texte,
  nochAktuell,
  aufgeschlagen,
  gemeldet,
}: Anlagewege): (name: string, art: string) => Promise<void> {
  return async (name: string, art: string) => {
    await vorgang.starte('scryfallkarte', async () => {
      const beim = setID

      // **Das Bildfeld heißt `image`**, so wie es jede M15-Vorlage führt. Ohne
      // Bildart wird es nicht benutzt.
      const antwort = await api().ScryfallKarteAnlegen(
        setID,
        name,
        'image',
        art,
      )

      if (!nochAktuell(beim)) {
        return
      }

      if (antwort.karteID !== '') {
        aufgeschlagen(antwort.karteID)
      }

      try {
        auspacken(antwort, () => antwort.karteID)
        gemeldet(undefined)
      } catch (ursache: unknown) {
        gemeldet(meldungAus(ursache, texte))
      }
    })
  }
}

// useKarteAnlegen gibt die Handlung, die eine **leere** Karte anhängt.
//
// **Sie steht neben der darüber, weil es dieselbe Frage ist**: wie kommt eine
// Karte in dieses Set? Zwei Antworten, dieselben vier Regeln — und eine Datei,
// in der man beide nebeneinander liest.
export function useKarteAnlegen({
  setID,
  vorgang,
  texte,
  nochAktuell,
  aufgeschlagen,
  gemeldet,
}: Anlagewege): () => Promise<void> {
  return async () => {
    await vorgang.starte('anlegen', async () => {
      const beim = setID

      try {
        const antwort = await api().KarteAnlegen(setID)
        const neue = auspacken(antwort, () => antwort.karteID)

        // **Regel 2**: die Antwort gehört zu dem Set, aus dem sie stammt.
        if (!nochAktuell(beim)) {
          return
        }

        gemeldet(undefined)
        aufgeschlagen(neue)
      } catch (ursache: unknown) {
        if (!nochAktuell(beim)) {
          return
        }

        gemeldet(meldungAus(ursache, texte))
      }
    })
  }
}

// Kartenanlage sind die beiden Wege, auf denen eine Karte entsteht.
export interface Kartenanlage {
  anlegen: ReturnType<typeof useKarteAnlegen>
  ausScryfall: ReturnType<typeof useScryfallkarte>
}

// useKartenanlage gibt beide Wege aus **einem** Verbund.
//
// # Warum sie zusammen stehen
//
// Sie bekommen dasselbe: dieselbe Setkennung, denselben Vorgang, denselben
// Abschluss — die neue Karte wird aufgeschlagen, die Liste lädt nach, der
// Bestand meldet sich. Zweimal derselbe Verbund nebeneinander stand in
// `Karten.tsx` und war die Stelle, an der die Datei über die Zeilengrenze aus
// §2 ging (#1339,
// beim Anschluss des Symbolfeldes).
//
// **Zwei Haken bleiben es trotzdem.** Wer nur einen braucht — und
// `Kartenwerkzeuge` braucht nur einen —, nimmt ihn einzeln.
export function useKartenanlage(wege: Anlagewege): Kartenanlage {
  const anlegen = useKarteAnlegen(wege)

  // **Und derselbe Weg für „Karte erstellen" aus der Palette** (UX-02.AK-02).
  useKartenwunsch(wege.setID, anlegen)

  return {
    anlegen,
    ausScryfall: useScryfallkarte(wege),
  }
}
