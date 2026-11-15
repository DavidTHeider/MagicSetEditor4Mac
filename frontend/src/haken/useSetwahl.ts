// **Wie ein Set das offene wird** —
// #1111.
//
// # Warum das ein eigener Haken ist
//
// Es gibt **vier** Wege zu demselben Ziel: eine Ablage, ein Doppelklick im
// Finder, ein Klick in der Setliste und ein frisch angelegtes Set. Alle vier
// tun dasselbe — Set setzen, Karte räumen, zur Kartenseite —, und alle vier
// müssen denselben **Laufstand** erhöhen.
//
// Solange das in `Anwendung.tsx` verteilt stand, tat es nur einer: der
// Laufstand lag in `useDateioeffnung` und stieg beim Ablegen. Wer zog und
// währenddessen ein Set aus der Liste wählte, bekam am Ende das gezogene.
//
// **§7 Regel 2 spricht vom Bezugsobjekt**, nicht von einem Weg dorthin. Hier
// laufen sie zusammen.
import { useCallback, useMemo } from 'react'

import type { Lauf } from './useLauf'
import { api } from '../api'
import { useLauf } from './useLauf'
import { wegschreiben } from '../feldwegschreiben'

export interface Setwahl {
  // waehleSet schlägt ein Set auf und vermerkt es als benutzt — mit
  // `karteID` gleich bei dieser Karte (AP-8, Teil 2b: die erste Karte mit
  // Befund). **Die Karte reist mit der Wahl** und nicht hinterher: ein Effekt
  // danach verlor gegen die erste Karte der Liste, die die Kartenseite beim
  // Aufbau schon gemeldet hatte (gesehen am Fenster, `ap8befunde09`).
  waehleSet: (setID: string, karteID?: string) => void

  // aufschlagen schlägt ein Set auf, **ohne** es zu vermerken — der Weg des
  // Doppelklicks und der Ablage: Go hat es schon gelesen.
  aufschlagen: (setID: string) => void

  // lauf ist der Stand, den jeder dieser Wege erhöht.
  lauf: Lauf
}

// useSetwahl bündelt die vier Wege zu einem offenen Set.
export function useSetwahl(
  setzeSetID: (setID: string | undefined) => void,
  setzeKarteID: (karteID: string | undefined) => void,
  zurKartenseite: () => void,
  eigenerstandErhoehen: () => void,
): Setwahl {
  const lauf = useLauf()

  // **Der gemeinsame Kern.** Ein Wechsel des Bezugsobjekts räumt den offenen
  // Kontext (§7, Regel 1): wer ein Set wählt, landet bei dessen Karten, und
  // die aufgeschlagene Karte des vorigen gehört geräumt.
  const wechsle = useCallback(
    (gewaehlt: string, karteID?: string) => {
      // **Erst das offene Feld verlassen** (#1290, Nachlauf über Phase 4.3,
      // Befund 1). Dieser Kern trägt *Zuletzt verwendetes Set*, das Öffnen
      // einer Datei und eine abgelegte Datei — alle drei bauen die Karte
      // unmittelbar ab, und ein programmgesteuerter Abbau erzeugt **kein**
      // `blur`. Der zuletzt getippte Wert wäre damit weg.
      wegschreiben()

      lauf.naechster()
      setzeSetID(gewaehlt)
      setzeKarteID(karteID)
      zurKartenseite()
    },
    [lauf, setzeSetID, setzeKarteID, zurKartenseite],
  )

  const aufschlagen = useCallback(
    (gewaehlt: string) => {
      wechsle(gewaehlt)

      // **Die Setliste hat einen Eintrag mehr.** Ohne diesen Zähler zeigte sie
      // den Stand von vor dem Doppelklick.
      eigenerstandErhoehen()
    },
    [wechsle, eigenerstandErhoehen],
  )

  // **Und es wird als benutzt vermerkt** (#891): daraus entsteht die Liste der
  // jüngsten Sets, im Menü wie auf der Startseite. Ein Fehler dabei bleibt
  // ohne Folgen — die Reihenfolge einer Liste ist keine Arbeit, die verloren
  // ginge, und eine Meldung darüber verdeckte die Seite, die gerade aufgeht.
  const waehleSet = useCallback(
    (gewaehlt: string, karteID?: string) => {
      wechsle(gewaehlt, karteID)

      // **In zwei Zeilen und nicht in einer Kette.** Der Wächter über
      // `docs/SCHREIBWEGE.md` sucht den Aufruf der Fassade **in einer Zeile**;
      // eine Kette, die der Formatierer umbricht, wäre für ihn unlesbar — und
      // eine Bindung, die er nicht sieht, fehlt in der Liste, ohne dass jemand
      // es merkt.
      const gemerkt = api().SetBenutzt(gewaehlt)

      void gemerkt.catch(() => undefined)
    },
    [wechsle],
  )

  return useMemo(
    () => ({ waehleSet, aufschlagen, lauf }),
    [waehleSet, aufschlagen, lauf],
  )
}
