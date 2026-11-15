// **Das offene Set, einmal gelesen** — U13
// (#2158).
//
// # Der Befund
//
// Die Brotkrumen über dem Kopf (`useBezugsobjekt`) und der Arbeitsplatz
// darunter (`Setarbeitsplatz`) fragten **je für sich** nach allen Sets, um
// das offene zu finden — zwei Rufe für dieselbe Frage, und mit zwei
// verschiedenen Ständen: die Brotkrumen lasen nach ⌘Z nicht nach, der
// Arbeitsplatz räumte bei jedem Schreiben seinen Kopf, bis die Antwort kam.
//
// # Der Schnitt
//
// **Das Fenster liest einmal** ([useOffenesSet]) — mit dem vollen Stand
// (Verlauf, fremde und eigene Schreiber) — und reicht das Ergebnis über
// diesen Raum hinunter. **Kein Zwischenspeicher über das Fenster hinaus**:
// ein anderes Set oder ein anderes Profil ist ein anderer Ruf, und ein
// Wechsel räumt (`useLaden` ohne `behalten`, §7 Regel 1).
import { createContext, useCallback, useContext, type ReactNode } from 'react'

import { api, auspacken } from '../api'
import type { SetDTO } from '../api'
import type { Bezugsobjekt } from '../brotkrumen'
import { useFenstertitel } from './useFenstertitel'
import { useLaden } from './useLaden'

const Kontext = createContext<SetDTO | undefined>(undefined)

// useOffenesSet liest das offene Set — **einmal** für das ganze Fenster.
//
// **Ein Fehler gibt keines**: die Brotkrumen sind eine Orientierungshilfe,
// und der Arbeitsplatz nennt dann „ohne Namen"; was schiefging, sagt die
// Fläche darunter (§8).
export function useOffenesSet(
  setID: string | undefined,
  stand: number,
): SetDTO | undefined {
  const { daten } = useLaden<SetDTO | undefined>(
    useCallback(async () => {
      if (setID === undefined) {
        return undefined
      }

      const antwort = await api().Sets()

      return auspacken(antwort, () => antwort.sets).find(
        (eines) => eines.id === setID,
      )
    }, [setID]),
    [setID],
    false,
    stand,
  )

  return daten
}

// useFensterset gibt dem Fenster das offene Set und die Station, die der Weg
// über dem Kopf dafür zeigt.
export function useFensterset(
  setID: string | undefined,
  stand: number,
): { offenesSet: SetDTO | undefined; objekt: Bezugsobjekt | undefined } {
  const offenesSet = useOffenesSet(setID, stand)

  return {
    offenesSet,
    objekt:
      offenesSet === undefined
        ? undefined
        : { name: offenesSet.name, ziel: 'karten' },
  }
}

// OffenesSet reicht das gelesene Set an die Flächen darunter.
export function OffenesSet({
  wert,
  children,
}: {
  wert: SetDTO | undefined
  children: ReactNode
}) {
  // **Der Titel folgt dem offenen Set** (AP-7, #2364) — hier, wo es
  // hinuntergereicht wird.
  useFenstertitel(wert?.name)

  return <Kontext value={wert}>{children}</Kontext>
}

// useSetImFenster gibt das Set, das im Fenster offen ist — oder keines.
//
// **Für Vorbelegungen** (AP-6, #2363): wer ein Deck anlegt oder eine Karte
// sucht, arbeitet meist an dem Set, das gerade offen ist.
export function useSetImFenster(): SetDTO | undefined {
  return useContext(Kontext)
}

// useGelesenesSet gibt das offene Set — **nur**, wenn es das genannte ist.
// Ein Arbeitsplatz für ein anderes Set bekommt keines, statt eines fremden.
export function useGelesenesSet(setID: string | undefined): SetDTO | undefined {
  const offen = useContext(Kontext)

  return offen !== undefined && offen.id === setID ? offen : undefined
}
