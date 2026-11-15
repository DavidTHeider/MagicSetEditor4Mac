// **Die Sprache dem Fenster und Go melden** — eine Frage, eine Datei.
//
// **Herausgelöst aus `Anwendung.tsx`**, weil sie über die Zeilengrenze wuchs
// (§2). Der Schnitt folgt der Frage: dort steht, **was** im Fenster dasteht,
// hier, **in welcher Sprache** — für den Vorleser und für die native
// Menüleiste.
import { useEffect, useState } from 'react'

import { bcp47, kataloge, type Sprache } from '../sprache/kataloge'
import { meldungAus } from '../meldung'
import { spracheMelden } from '../sprache/menue'
import { useLauf } from './useLauf'

// useSprachmeldung meldet die Sprache und gibt zurück, was dabei schieflief.
//
// **`undefined` heißt: noch nicht.** Der Haken steht in `App` über beiden
// Startwegen und muss dort vor jeder frühen Rückkehr gerufen werden — auch,
// solange die Sprache noch nicht feststeht.
export function useSprachmeldung(
  sprache: Sprache | undefined,

  // eigene ist die Kennung einer eigenen Übersetzung — sie steht dann im
  // `lang`-Attribut, die Menüleiste spricht ihre Basis (LOC-02, #1688).
  eigene?: string,
): string | undefined {
  const lauf = useLauf()
  const [meldefehler, setzeMeldefehler] = useState<string>()

  // **Auch das `lang`-Attribut des Dokuments.**
  //
  // `index.html` steht fest auf `de`. Für einen Vorleser ist das die
  // maßgebliche Seitensprache — eine englische Oberfläche, die sich als deutsch
  // ausgibt, wird mit deutscher Aussprache vorgelesen. Gefunden hat es der
  // Zweitblick über Phase 5; gesehen hat es kein Test, weil keiner danach sah.
  //
  // **Mit der Schrift, wo sie die Glyphen entscheidet** (LOC-01, #1687):
  // `zh-Hans` statt `zh`, sonst darf WebKit für ein Han-Zeichen die
  // traditionelle Form nehmen.
  useEffect(() => {
    if (sprache !== undefined) {
      document.documentElement.lang = eigene ?? bcp47[sprache]
    }
  }, [sprache, eigene])

  // **Und die native Menüleiste.**
  //
  // Ohne das stünde das Fenster in der gemerkten Sprache und die Leiste in
  // der, mit der Go gestartet ist — eine Anwendung in zwei Sprachen, und zwar
  // genau für die Leute, die einmal umgestellt haben.
  //
  // `useLauf` hält §7, Regel 2 ein: wer zweimal schnell umstellt, bekommt die
  // Antwort auf die **letzte** Wahl eingetragen, nicht die zuletzt
  // eintreffende.
  useEffect(() => {
    if (sprache === undefined) {
      return
    }

    const meiner = lauf.naechster()

    spracheMelden(sprache).then(
      () => {
        if (lauf.istAktuell(meiner)) {
          setzeMeldefehler(undefined)
        }
      },
      (ursache: unknown) => {
        if (lauf.istAktuell(meiner)) {
          // **Der Katalog wird hier unmittelbar geholt, nicht über `useTexte`.**
          // `App` steht **über** dem `Sprachraum` — sie stellt ihn her. Ein
          // Haken daraus läse einen Katalog, den es an dieser Stelle noch nicht
          // gibt.
          setzeMeldefehler(meldungAus(ursache, kataloge[sprache]))
        }
      },
    )
  }, [sprache, lauf])

  return meldefehler
}
