// **Was die Kartenliste zeigt, bevor Marke und Schnellfilter eingreifen** —
// herausgelöst, weil `Kartenspalte.tsx` mit dem zweiten Filter über 400
// Zeilen ginge (§2).
//
// Zwei Auswahlen liegen unter allen anderen Filtern, und beide kommen von
// außen: die Karten eines Balkens der Statistik (UXR-F6, #2013) und die
// Karten mit Befund aus dem Überblick (AP-8, Teil 2b,
// #2365).
// Gelten beide, gilt ihre Schnittmenge.
import { useEffect, useMemo, useState } from 'react'

import { useAnschluss, useAufgenommen } from '../anschlussraum'
import { useStatistikfilter } from '../statistikfilter'

// Befundauswahl ist die Auswahl, die der Überblick mitgab.
export interface Befundauswahl {
  // name ist der Satz, der sie nennt — „2 Karten mit Befund“.
  name: string
  karteIDs: ReadonlySet<string>
}

// useKartenvorauswahl gibt beide Auswahlen und die Karten, die nach ihnen
// übrig sind.
export function useKartenvorauswahl<K extends { id: string }>(
  setID: string,
  karten: readonly K[],
  gewaehlt: string | undefined,
  waehlen: (karteID: string) => void,
) {
  // **Als Menge und nicht als Liste**: bei 3.761 Karten und 1.500 Karten im
  // Balken wären es sonst über fünf Millionen Vergleiche je Zeichnen.
  const statistik = useStatistikfilter()
  const imBalken = useMemo(
    () => statistik.wahl && new Set(statistik.wahl.kartenIDs),
    [statistik.wahl],
  )

  const anschluss = useAnschluss()
  const mitgereist =
    anschluss.offen?.art === 'kartenauswahl' ? anschluss.offen : undefined
  const [befund, setzeBefund] = useState<Befundauswahl>()

  // **Nur für das Set, dem sie gehört** (§7, Regel 2): wer inzwischen ein
  // anderes geöffnet hat, sähe sonst Kennungen, die es dort nicht gibt — und
  // eine leere Liste unter einem Satz, der Befunde verspricht.
  useAufgenommen(
    mitgereist?.setID === setID ? mitgereist : undefined,
    ({ name, karteIDs }) => {
      setzeBefund({ name, karteIDs: new Set(karteIDs) })
    },
  )

  // **Verbraucht, sobald entschieden ist** — aufgenommen oder verworfen.
  // Sonst nähme die Spalte, die nach einem Setwechsel und der Rückkehr neu
  // entsteht, die Auswahl ein zweites Mal auf: ein Filter, den niemand mehr
  // angeordnet hat. Dieselbe Lücke wie beim Kartenwunsch (Zweitblick über
  // Phase 6.3, S7).
  //
  // **Und aufgeschlagen wird die erste Karte mit Befund**, wenn die offene
  // keinen trägt — unter dem Satz „1 Karte mit Befund“ stand sonst die
  // vorige Karte mit „Keine Probleme an dieser Karte“ (gesehen am Fenster,
  // `ap8befunde09`). **Im Effekt und nicht beim Zeichnen**: die Wahl ist
  // der Zustand der Kartenlasche, nicht dieser Spalte.
  useEffect(() => {
    if (mitgereist === undefined) {
      return
    }

    anschluss.verbrauche()
    // **Mit `undefined` als möglichem Element**, damit „nichts gewählt“
    // dieselbe Frage ist wie „eine Karte ohne Befund gewählt“.
    const mitBefund = new Set<string | undefined>(mitgereist.karteIDs)
    const erste = karten.find((karte) => mitBefund.has(karte.id))
    if (
      mitgereist.setID === setID &&
      erste !== undefined &&
      !mitBefund.has(gewaehlt)
    ) {
      waehlen(erste.id)
    }
  }, [mitgereist, anschluss, setID, karten, gewaehlt, waehlen])

  const vorab = karten.filter(
    (karte) =>
      (imBalken === undefined || imBalken.has(karte.id)) &&
      (befund === undefined || befund.karteIDs.has(karte.id)),
  )

  return {
    statistik,
    befund,
    loeseBefund: () => {
      setzeBefund(undefined)
    },
    karten: vorab,
  }
}
