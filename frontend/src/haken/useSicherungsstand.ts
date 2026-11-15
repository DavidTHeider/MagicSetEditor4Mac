// Das Alter der letzten Sicherung und der freiwillige Plan — OPS01
// (#2347).
//
// # Die Regeln aus §7, die er hält
//
//  - **Ein Vorgang je Datensatz** (Regel 4): Einrichten und Beenden laufen
//    unter dem Vorgang des Bestands, denselben wie Sichern und Zurückspielen.
//  - **Eine Antwort gehört zu ihrem Stand** (Regel 2): gelesen wird über
//    `useLaden`, und der Nachlesestand ist die eigene Sicherung — ein neuer
//    Pfad, ein ersetzter Bestand.
//  - **Geschrieben ist geschrieben** (Regel 3): scheitert das Nachlesen nach
//    dem Einrichten, bleibt der Plan eingerichtet; der Fehler steht da.
//
// # Warum er nachfragt, solange der Plan fällig ist
//
// **Der Plan sichert im Prozess des Fensters**, und seine Sicherung ist kein
// fremdes Schreiben, das `useBestandsstand` bemerkte. Nach dem Einrichten ist
// er fällig und sichert gleich; gefragt wird deshalb im Takt der fremden
// Schreibvorgänge, bis er nicht mehr fällig ist oder einen Fehler meldet.
import { useCallback, useEffect, useState } from 'react'

import { api, auspacken } from '../api'
import type {
  Sicherungsstand,
  SicherungsstandAntwort,
} from '../antwortenbestand'
import { meldungAus, meldungAusAngabe } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { abstand } from './useBestandsstand'
import { useLaden } from './useLaden'
import type { Vorgang } from './useVorgang'

export interface Sicherungsstandlauf {
  stand: Sicherungsstand | undefined
  planfehler: string | undefined
  fehler: string | undefined
  einrichten: () => void
  beenden: () => void
  laeuft: boolean
}

export function useSicherungsstand(
  vorgang: Vorgang,
  nachlesen: unknown,
): Sicherungsstandlauf {
  const texte = useTexte()
  const [runde, setzeRunde] = useState(0)
  const [fehler, setzeFehler] = useState<string>()

  const geladen = useLaden(
    async () => {
      const antwort = await api().Sicherungsstand()

      return auspacken(antwort, () => antwort)
    },
    [],
    false,
    [nachlesen, runde],
  )

  const stand = geladen.daten?.stand
  const gemeldet = geladen.daten?.planfehler
  const planfehler = gemeldet
    ? meldungAusAngabe({ ...gemeldet, werte: gemeldet.werte ?? null }, texte)
    : undefined
  const wartet =
    stand !== undefined &&
    stand.plan.ordner !== '' &&
    stand.plan.faellig &&
    planfehler === undefined

  useEffect(() => {
    if (!wartet) {
      return undefined
    }

    const uhr = setTimeout(() => {
      setzeRunde((zahl) => zahl + 1)
    }, abstand)

    return () => {
      clearTimeout(uhr)
    }
  }, [wartet, geladen.daten])

  const handle = useCallback(
    (ruf: () => Promise<SicherungsstandAntwort>) => {
      void vorgang.starte('bestand', async () => {
        setzeFehler(undefined)

        try {
          const antwort = await ruf()
          auspacken(antwort, () => undefined)
          setzeRunde((zahl) => zahl + 1)
        } catch (ursache: unknown) {
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [texte, vorgang],
  )

  return {
    stand,
    planfehler,
    fehler: fehler ?? geladen.fehler,
    einrichten: useCallback(() => {
      handle(() => api().SicherungsplanEinrichten())
    }, [handle]),
    beenden: useCallback(() => {
      handle(() => api().SicherungsplanBeenden())
    }, [handle]),
    laeuft: vorgang.laeuft('bestand'),
  }
}
