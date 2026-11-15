// Das Sichern des ganzen Bestands — hinter *Ablage → Bestand sichern …*
// (#799).
//
// # Warum ein eigener Haken
//
// `useSetSpeichern` daneben sichert **ein Set** in eine `.mse-set`-Datei; hier
// geht es um die Datenbank selbst. Zwei Gegenstände, zwei Vorgänge — und die
// Meldungen unterscheiden sich: dort steht ein Set im Bestand und kommt heraus,
// hier wird der Bestand weggetragen.
//
// # Die Regeln aus §7, die er hält
//
//  1. **Ein Vorgang je Datensatz** (Regel 4): der Datensatz ist der Bestand,
//     und es gibt genau einen. Ein zweiter Aufruf, während der erste schreibt,
//     ginge gegen eine Datei, die halb dasteht.
//  2. **Geschrieben ist geschrieben** (Regel 3): nach einem Fehler bleibt die
//     Meldung stehen.
//
// **Regel 1 und 2 greifen hier nicht**: es gibt kein Bezugsobjekt, das
// wechseln könnte. Der Bestand ist der Bestand.
import { useCallback, useState } from 'react'

import { api, auspacken } from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import type { Vorgang } from './useVorgang'

export interface Sicherungslauf {
  // sichern schreibt die Sicherung — mit Kennwort, wenn eines kommt; aus einem
  // verschlüsselten Bestand ohne Kennwort nur mit `klartext` (OPS-06, #1723).
  sichern: (kennwort: string, klartext: boolean) => void

  // ausMenue ist *Ablage → Bestand sichern …* — **ohne Kennwort**. Ein
  // verschlüsselter Bestand antwortet dann im Band, wo es eines gibt.
  ausMenue: () => void
  laeuft: boolean

  // pfad ist der Ort der geglückten Sicherung — leer nach einem Abbruch.
  pfad: string | undefined
  fehler: string | undefined

  raeumen: () => void
}

// **Der Vorgang kommt von außen**, seit es einen zweiten Weg auf denselben
// Bestand gibt (#940).
//
// Regel 4 aus AGENTS.md §7 gilt je **Datensatz**, und der Datensatz ist hier
// der ganze Bestand. Zwei Haken mit je eigenem `useVorgang` sperrten einander
// nicht — und ein Sichern, das während eines Zurückspielens läuft, liest eine
// Datei, die gerade ausgetauscht wird.
export function useBestandSichern(vorgang: Vorgang): Sicherungslauf {
  const texte = useTexte()

  const [pfad, setzePfad] = useState<string>()
  const [fehler, setzeFehler] = useState<string>()

  const raeumen = useCallback(() => {
    setzePfad(undefined)
    setzeFehler(undefined)
  }, [])

  const sichern = useCallback(
    (kennwort: string, klartext: boolean) => {
      void vorgang.starte('bestand', async () => {
        setzePfad(undefined)
        setzeFehler(undefined)

        try {
          const antwort = await api().BestandSichern(kennwort, klartext)

          // **Über `auspacken` und nicht über `antwort.fehler` unmittelbar.**
          // `meldungAus` erwartet einen `Bindungsfehler` oder einen `Error`; ein
          // rohes Fehlerobjekt fällt bei ihm auf `String(…)` durch und ergibt
          // „[object Object]". Der erste Anlauf tat genau das, und der Test hat
          // es gefangen.
          auspacken(antwort, () => undefined)

          // **Ein Abbruch hinterlässt nichts.** Wails gibt dafür einen leeren
          // Pfad ohne Fehler; eine Erfolgsmeldung über eine Sicherung, die es
          // nicht gibt, wäre schlimmer als keine Meldung.
          if (antwort.pfad !== '') {
            setzePfad(antwort.pfad)
          }
        } catch (ursache: unknown) {
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [texte, vorgang],
  )

  return {
    sichern,
    ausMenue: useCallback(() => {
      sichern('', false)
    }, [sichern]),
    laeuft: vorgang.laeuft('bestand'),
    pfad,
    fehler,
    raeumen,
  }
}
