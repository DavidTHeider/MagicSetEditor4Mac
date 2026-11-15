// **Einen Feldwert schreiben.**
//
// Ein eigener Haken, weil `Karteneditor.tsx` mit ihm über 400 Zeilen ginge
// (§2) — und weil der Schnitt einer Frage folgt: dort steht, **wie** der
// Editor angeordnet ist, hier, **was** ein Schreibvorgang tut.
import type { Katalog } from '../sprache/kataloge'
import { api } from '../api'
import { auspacken } from '../umschlag'
import { meldungAus, verursachendesFeld } from '../meldung'
import { schreibenBeginnt, schreibenEndet } from '../speicherstand'
import type { Vorgang } from './useVorgang'

interface Schreibwege {
  setID: string
  karteID: string
  vorgang: Vorgang
  texte: Katalog

  melden: (befund: { feld: string; meldung: string } | undefined) => void
  geglueckt: () => void
}

// useFeldschreiben gibt die Handlung, die einen Feldwert schreibt.
//
// # Die Regeln aus §7
//
//   - **Regel 4**: ein Vorgang je Datensatz — der Schlüssel ist Karte **und**
//     Feld: zwei Felder derselben Karte dürfen gleichzeitig gehen.
//   - **Regel 3**: nachgeladen wird **nur** nach einem geglückten Schreiben.
//     Sonst überschriebe ein Lesefehler die Meldung, die sagt, was zu tun ist.
//
// **Der Fehler erscheint am verursachenden Feld** (§6): sagt die Bindung ein
// Feld, gilt das; sonst das eben bearbeitete.
//
// **`bekannt`** kommt vom Entwurf (U02b, #2179), und es ist Pflicht: jedes
// Kartenfeld im Editor hat einen, auch für einen Einschub aus der
// Symbolpalette. Bis zum Zweitblick über Phase 0.7.0-I war es freiwillig, und
// genau dieser Einschub schrieb ohne Bedingung über einen Konflikt hinweg.
export function useFeldschreiben({
  setID,
  karteID,
  vorgang,
  texte,
  melden,
  geglueckt,
}: Schreibwege): (
  feld: string,
  wert: string,
  bekannt: readonly string[],
) => Promise<Schreibergebnis> {
  return async (feld: string, wert: string, bekannt: readonly string[]) => {
    const ergebnis = await vorgang.starte(`${karteID}/${feld}`, async () => {
      // **Der Speicherstand hört mit** (UX-05.AK-03, #1677): er sagt über
      // der Karte, ob das Eingetippte im Bestand steht.
      schreibenBeginnt(setID)

      try {
        const antwort = await api().FeldSetzen(
          setID,
          karteID,
          feld,
          wert,
          bekannt,
        )
        const gespeichert = auspacken(antwort, () => antwort.gespeichert)

        schreibenEndet(setID, karteID, feld, true)

        return typeof gespeichert === 'string' ? { gespeichert } : true
      } catch (ursache: unknown) {
        schreibenEndet(setID, karteID, feld, false)

        melden({
          feld: verursachendesFeld(ursache, feld),
          meldung: meldungAus(ursache, texte),
        })

        return false
      }
    })

    if (ergebnis === undefined || ergebnis === false) {
      return false
    }

    geglueckt()

    // **Geglückt sagt der Entwurf weiter** (Gesamtreview über 0.9.0, Befund
    // 1), **und was gespeichert ist** (#2535): der Bestand speichert seit
    // #2396 den gerechneten Wert, und der Entwurf erkennt genau ihn als
    // seinen — einen fremden danach nicht.
    return ergebnis
  }
}

// Schreibergebnis sagt, ob das Schreiben glückte — und, wo bekannt, was im
// Feld gespeichert ist.
export type Schreibergebnis = boolean | { gespeichert: string }
