// **Was die Kartenseite lädt** —
// #891,
// Werkstück 5 von Phase Y.
//
// **Ein eigener Haken**, weil `Karten.tsx` mit ihm über 400 Zeilen ginge (§2) —
// und weil der Schnitt einer Frage folgt: dort steht, **wie** die Seite
// angeordnet ist, hier, **was** sie holt.
import { useCallback } from 'react'

import {
  api,
  auspacken,
  type FeldbeschreibungDTO,
  type KarteDTO,
  type KartenzeileDTO,
} from '../api'
import { useLaden, type Ladezustand } from './useLaden'

interface Kartenfrage {
  setID: string

  // verlaufsstand steigt nach jeder geglückten Rücknahme (#52) — Liste und
  // aufgeschlagene Karte hängen beide daran.
  verlaufsstand: number

  // gewaehlt ist die aufgeschlagene Karte; `undefined` heißt „keine".
  gewaehlt: string | undefined
}

export interface Kartendaten {
  karten: Ladezustand<KartenzeileDTO[]>

  // liste ist `karten.daten ?? []` — **einmal ausgepackt, nicht dreimal**.
  //
  // Drei Stellen brauchen sie, und drei `?? []` wären drei Zweige für dieselbe
  // Frage; zwei davon betritt kein Test, weil geladen ist, sobald eine Karte
  // aufgeschlagen ist.
  liste: KartenzeileDTO[]

  formen: Ladezustand<FeldbeschreibungDTO[]>
  geoeffnete: Ladezustand<KarteDTO | undefined>
}

export function useKartendaten({
  setID,
  verlaufsstand,
  gewaehlt,
}: Kartenfrage): Kartendaten {
  // **Die Liste holt keine Feldwerte** (#891).
  //
  // `Black.mse-set` hat 3.788 Karten; sie mit jedem Wert zu laden waren
  // **4.761.182 Bytes JSON** je Aufruf und 35,0 ms allein auf der Go-Seite —
  // danach über die Wails-Naht und durch den JSON-Zerteiler des Fensters.
  // Über die Liste sind es **245.641 Bytes** und 10,0 ms: 19,4-mal kleiner,
  // 3,5-mal schneller (gemessen am 2026-08-25, `BenchmarkKartenliste`).
  //
  // Was die **eine** aufgeschlagene Karte trägt, holt `holenKarte` darunter.
  const holenKarten = useCallback(async () => {
    const antwort = await api().Kartenzeilen(setID)

    return auspacken(antwort, () => antwort.zeilen)
  }, [setID])

  // **Ohne gewählte Karte wird nicht gefragt.** Ein Aufruf mit leerer Kennung
  // wäre eine Frage nach einer Karte, die niemand gemeint hat.
  const holenKarte = useCallback(async () => {
    if (gewaehlt === undefined) {
      return undefined
    }

    const antwort = await api().Karte(gewaehlt)

    return auspacken(antwort, () => antwort.karte)
  }, [gewaehlt])

  const holenFormen = useCallback(async () => {
    const antwort = await api().Kartenfelder(setID)

    return auspacken(antwort, () => antwort.felder)
  }, [setID])

  const karten = useLaden<KartenzeileDTO[]>(
    holenKarten,
    [setID],
    false,
    verlaufsstand,
  )

  // **Die Formen der Felder gehören zum Set**, nicht zur Karte
  // (#560):
  // hier geladen, damit der Editor sie beim Kartenwechsel schon hat.
  //
  // **Und am Stand wie die Werte** — Befund 5 des Nachlaufs zum
  // Gesamtreview über 0.6.0 (#1660): tauschte ein anderer Prozess den Stil,
  // standen neue Werte neben alten Formen. **Behalten**, damit der Editor
  // beim Nachladen nicht abbaut; ein Setwechsel räumt ohnehin, denn die
  // Seite trägt das Set als `key`.
  const formen = useLaden<FeldbeschreibungDTO[]>(
    holenFormen,
    [setID, verlaufsstand],
    true,
  )

  // **Die aufgeschlagene Karte holt ihre Feldwerte selbst** (#891).
  //
  // Bis zum 2026-08-25 kamen sie mit der Liste — für **jede** Karte. Bei 3.788
  // waren das 4,7 MB je Aufruf, damit eine davon bearbeitet werden konnte.
  //
  // **Der Verlaufsstand gehört in den Schlüssel**: eine Rücknahme ändert die
  // Werte, und ohne ihn stünde die Karte von vorhin da.
  const geoeffnete = useLaden<KarteDTO | undefined>(
    holenKarte,
    [gewaehlt],
    false,
    verlaufsstand,
  )

  return { karten, liste: karten.daten ?? [], formen, geoeffnete }
}
