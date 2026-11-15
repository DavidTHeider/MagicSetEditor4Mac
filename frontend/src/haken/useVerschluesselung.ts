// **Die Verschlüsselung des Bestands** — Werkstück 14 von Phase 4.12
// (#1351),
// nach Vermerk 0072.
//
// # Die Regeln aus §7, die er hält
//
//  1. **Ein Vorgang je Datensatz** (Regel 4): der Datensatz ist der Bestand,
//     und es gibt genau einen. Der Vorgang kommt deshalb von **außen** —
//     dieselbe Begründung wie bei `useBestandSichern`: Sichern,
//     Zurückspielen und Umschalten greifen alle drei auf dieselbe Datei, und
//     zwei Haken mit je eigenem `useVorgang` sperrten einander nicht.
//  2. **Geschrieben ist geschrieben** (Regel 3): die Antwort trägt den neuen
//     Stand selbst — nachgeladen wird nicht. Ein Nachladen nach einem
//     geglückten Umschalten liefe gegen eine Verbindung, die gerade neu
//     aufgebaut wurde.
//
// **Regel 1 und 2 greifen nicht**: es gibt kein Bezugsobjekt, das wechseln
// könnte. Der Bestand ist der Bestand.
import { useCallback, useEffect, useRef, useState } from 'react'

import { api, auspacken } from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import type { Vorgang } from './useVorgang'

export interface Verschluesselungslauf {
  // verschluesselt ist der Stand — `undefined`, solange er nicht feststeht.
  verschluesselt: boolean | undefined

  umschalten: (an: boolean) => void
  laeuft: boolean
  fehler: string | undefined
}

export function useVerschluesselung(vorgang: Vorgang): Verschluesselungslauf {
  const texte = useTexte()

  const [verschluesselt, setzeStand] = useState<boolean>()
  const [fehler, setzeFehler] = useState<string>()

  // **Einmal beim Aufbau gelesen.** Der Stand steht in einer Datei und ändert
  // sich nur durch diesen Haken selbst — ein Takt darauf wäre eine Frage ohne
  // Anlass.
  //
  // **Der Wächter ist eine Referenz und keine lokale Variable.** Eine lokale
  // sähe TypeScript beim Abbau nicht verändert — es verengt sie im Abschluss
  // auf `true`, und `@typescript-eslint/no-unnecessary-condition` weist die
  // Abfrage dann als überflüssig ab. Sie ist es nicht: die Antwort kommt nach
  // dem Abbau, und ohne die Abfrage trüge sie in einen Baum ein, den es nicht
  // mehr gibt.
  const gilt = useRef(true)

  useEffect(() => {
    gilt.current = true

    void (async () => {
      try {
        const antwort = await api().Verschluesselungsstand()
        const stand = auspacken(antwort, () => antwort.verschluesselt)

        if (gilt.current) {
          setzeStand(stand)
        }
      } catch (ursache: unknown) {
        if (gilt.current) {
          setzeFehler(meldungAus(ursache, texte))
        }
      }
    })()

    return () => {
      gilt.current = false
    }
  }, [texte])

  const umschalten = useCallback(
    (an: boolean) => {
      void vorgang.starte('bestand', async () => {
        setzeFehler(undefined)

        try {
          const antwort = await api().VerschluesselungUmschalten(an)

          // **Der Stand kommt aus der Antwort** (§7, Regel 3). Die Bindung
          // liest ihn nach dem Tausch von der Platte; ihn hier auf `an` zu
          // setzen hieße, das Ergebnis zu behaupten statt es zu lesen.
          setzeStand(auspacken(antwort, () => antwort.verschluesselt))
        } catch (ursache: unknown) {
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [texte, vorgang],
  )

  return {
    verschluesselt,
    umschalten,
    laeuft: vorgang.laeuft('bestand'),
    fehler,
  }
}
