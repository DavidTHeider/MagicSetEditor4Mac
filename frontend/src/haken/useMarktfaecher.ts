// **Die zwei Fächer des Markttischs** — P12
// (#2380).
//
// **Links die eigene Sammlung samt Wunschliste**, aus dem Bestand; **rechts
// das freigegebene Fach des Gegenübers**, über die Leitung (P11) — **gefragt
// beim Aufschlagen** (Befund C-15, H3 = 0): wer den Tisch mit Bert öffnet,
// will sehen, was Bert zeigt. Ein zweites Fragen geht über „Neu fragen".
//
// **Ein Gegenüber ist der Schlüssel** (§7, Regel 1 und 2): wechselt es, gilt
// die Antwort des vorigen nicht.
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden, type Ladezustand } from './useLaden'
import {
  ausExemplar,
  ausSichtkarte,
  type Fachkarte,
} from '../bausteine/marktlegen'

export function useMarktfaecher(gegenueber: string) {
  const meine = useLaden(async () => {
    const antwort = await api().Sammlung()

    return auspacken(antwort, () => antwort.exemplare).map(ausExemplar)
  }, [])

  const sicht = useLaden(async () => {
    const antwort = await api().HandelSichtAnfragen(gegenueber)

    return auspacken(antwort, () => antwort.zeilen)
  }, [gegenueber])

  // **Was drüben gesucht wird, liegt nicht in seinem Fach** — eine
  // Wunschzeile ist kein Angebot (P11). **Die ganze Sicht geht trotzdem in den
  // Abgleich** (#2494): was zusammenpasst, rechnet er über Angebot und
  // Wünsche, und gefragt wird einmal.
  const seine: Ladezustand<Fachkarte[]> = {
    ...sicht,
    daten: sicht.daten
      ?.filter((zeile) => zeile.wunsch !== true)
      .map(ausSichtkarte),
  }

  const alle: Fachkarte[] = [...(meine.daten ?? []), ...(seine.daten ?? [])]

  return { meine, seine, alle, geholt: { zeilen: sicht.daten } }
}
