// **Ein Doppelklick im Finder** —
// #891.
//
// # Was Go tut und was hier geschieht
//
// macOS übergibt die Datei als Apple-Event `odoc`; Wails reicht sie an
// `mac.Options.OnFileOpen`, und Go liest das Set in den Bestand. Was danach
// **zu sehen** ist, entscheidet das Fenster — dieselbe Aufteilung wie bei der
// Menüleiste (`fenster.ts`): Go meldet, die Oberfläche handelt (§2).
//
// # Warum der Satz hier entsteht und nicht in Go
//
// Go kennt den **Schlüssel** eines Fehlers, nicht die Sprache der Sitzung
// (§9). Ein deutscher Satz aus Go wäre in einer englischen Sitzung deutsch —
// und genau das war #133.
import { useCallback, useMemo, useState } from 'react'

import type { Fehlerangabe } from '../fenster'
import { api, auspacken } from '../api'
import type { Lauf } from './useLauf'
import {
  epocheGesehen,
  epocheVeraltet,
  kennungIstGeloescht,
} from '../bestandsepoche'
import { meldungAus, meldungAusAngabe } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge, type Sprache } from '../sprache/kataloge'

export interface Dateioeffnung {
  dateiGeoeffnet: (setID: string, epoche?: number) => void

  // dateiGezogen öffnet ein Setpaket, das jemand ins Fenster gezogen hat
  // (#891).
  //
  // **Der Weg ist ein anderer als beim Doppelklick**: dort liest Go die Datei
  // und meldet die Kennung, hier ruft das Fenster selbst. Der Grund ist die
  // Antwort — ein Ablegen ist eine Geste **im** Fenster, und der Fehler
  // gehört dorthin, wo sie stattfand.
  dateiGezogen: (pfad: string) => void
  dateiFehler: (pfad: string, angabe: Fehlerangabe) => void

  // meldung ist der Satz für das Band — undefined, wenn nichts schiefging.
  meldung: string | undefined
}

// **`aufschlagen` bekommt die Kennung des Sets, das schon im Bestand steht.**
//
// Go hat es eingelesen, bevor es meldet; hier ist nur noch zu entscheiden, was
// das Fenster zeigt.
// **Der Laufstand kommt von außen** —
// #1111.
//
// Er lag bis zum 2026-08-25 in diesem Haken und stieg deshalb **nur** beim
// Ablegen. Ein Set aufzuschlagen gibt es aber an vier Stellen: Ablegen,
// Doppelklick im Finder, die Setliste und ein neu angelegtes Set. Wer zog und
// währenddessen eines aus der Liste wählte, bekam am Ende das gezogene.
//
// §7 Regel 2 spricht vom **Bezugsobjekt**, nicht von einem Weg dorthin —
// also gehört der Stand dorthin, wo alle Wege zusammenlaufen.
export function useDateioeffnung(
  sprache: Sprache,
  aufschlagen: (setID: string) => void,
  lauf: Lauf,
): Dateioeffnung {
  // **Eine Meldung je Pfad, nicht eine für alle** —
  // #1110.
  //
  // Mehrere Dateien kommen auf zwei Wegen zusammen an: der Finder ruft
  // `OnFileOpen` je Datei, und ein Ablegen reicht jedes Setpaket einzeln
  // durch. Mit **einer** Meldung räumte der Erfolg der einen den Fehler der
  // anderen — und wer drei Dateien zog, sah den Fehler nicht, den er suchte.
  //
  // **Ein Erfolg räumt nur seinen eigenen Pfad.** Was übrig bleibt, sind die
  // Dateien, die gescheitert sind und es seither nicht wieder versucht haben.
  const [meldungen, setzeMeldungen] = useState<Map<string, string>>(
    () => new Map(),
  )

  const merke = useCallback((pfad: string, satz: string | undefined) => {
    setzeMeldungen((bisher) => {
      if (satz === undefined && !bisher.has(pfad)) {
        return bisher
      }

      const naechste = new Map(bisher)

      if (satz === undefined) {
        naechste.delete(pfad)
      } else {
        naechste.set(pfad, satz)
      }

      return naechste
    })
  }, [])

  // **Regel 2 aus §7 für das Ablegen** — angestrichen vom Zweitblick über
  // Phase Y als Befund 4.
  //
  // Jedes Ablegeereignis stößt einen eigenen `SetAusDatei` an, und zwei
  // Ereignisse kommen einzeln aus Wails. Wer erst ein großes und gleich darauf
  // ein kleines Set ablegt, bekommt die Antworten in der Reihenfolge, in der
  // **Go** sie fertig hat: das kleine schlägt auf, danach überschreibt das
  // große es wieder — und ein alter Fehler stünde über einem Set, das längst
  // offen ist.
  //
  // **Hier steht bewusst kein `useVorgang`.** Ein zweites Ablegen soll nicht
  // *abgewiesen* werden; es ist eine neue Absicht, und die gilt. Was nicht
  // gelten darf, ist die Antwort der vorigen.

  const dateiGeoeffnet = useCallback(
    (setID: string, epoche?: number) => {
      // **Eine Kennung aus einem älteren Bestand wird verworfen** (J05,
      // #2269, zu #2292): Wails stellt das Ereignis in eine eigene Schlange,
      // und die Antwort eines Neuanfangs kann es überholen. Eine höhere
      // Epoche merkt sich das Fenster dabei — sie sagt, dass Go weiter ist.
      if (epocheVeraltet(epoche)) {
        return
      }

      epocheGesehen(epoche)

      // **Ein Set, das inzwischen gelöscht ist, schlägt nicht auf** (#2306):
      // der Finder las es vor dem Löschen ein, und seine Meldung kam danach.
      if (kennungIstGeloescht(setID)) {
        return
      }

      // **Ohne Pfad räumt hier nichts** — [#1110]. Go meldet beim Doppelklick
      // die **Kennung**, nicht den Weg; welche Meldung zu dieser Datei gehört,
      // ist von hier aus nicht zu sagen. Ein pauschales Räumen nähme dem
      // Nachbarn seine, und genau das war der Befund.
      aufschlagen(setID)
    },
    [aufschlagen],
  )

  const dateiFehler = useCallback(
    (pfad: string, angabe: Fehlerangabe) => {
      // **Der Katalog unmittelbar über die Sprache** — dieser Haken steht
      // unter `App` und über den Seiten; `useTexte` läse einen Raum, den er
      // nicht garantiert hat.
      const texte = kataloge[sprache]

      merke(
        pfad,
        einsetzen(texte.allgemein.dateiNichtGeoeffnet, {
          pfad,
          grund: meldungAusAngabe(angabe, texte),
        }),
      )
    },
    [sprache, merke],
  )

  const dateiGezogen = useCallback(
    (pfad: string) => {
      const meiner = lauf.naechster()

      void (async () => {
        try {
          const antwort = await api().SetAusDatei(pfad)
          const kennung = auspacken(antwort, () => antwort.setID)

          // **Eine überholte Antwort schlägt nichts mehr auf.** Das Set liegt
          // im Bestand — geschrieben ist geschrieben (Regel 3) —, es gewinnt
          // nur die Anzeige nicht mehr.
          if (!lauf.istAktuell(meiner)) {
            return
          }

          // **Ein Abbruch gibt es hier nicht** — die Datei ist genannt. Eine
          // leere Kennung hieße, dass nichts angelegt wurde, und das wäre ein
          // Fehler, den `auspacken` schon geworfen hätte.
          merke(pfad, undefined)
          aufschlagen(kennung)
        } catch (ursache: unknown) {
          // **Ein Fehler wird immer vermerkt, auch überholt** —
          // #1110.
          //
          // Hier stand eine zweite Abweisung: ein überholter Fehler kam nicht
          // mehr an, weil „Set nicht lesbar" über einem gerade aufgegangenen
          // Set wie ein Fehler des aktuellen aussah.
          //
          // **Das stimmt nicht mehr, seit die Meldung ihren Pfad nennt** —
          // und es kostete zu viel: wer zwei Sets zusammen zieht, erzeugt
          // damit zwei Läufe, und der erste war immer überholt. Eine Datei,
          // die nicht aufging, verschwand **wortlos**.
          //
          // Der Laufstand entscheidet weiterhin, **was aufschlägt**. Ob etwas
          // schiefging, entscheidet er nicht.
          const texte = kataloge[sprache]

          merke(
            pfad,
            einsetzen(texte.allgemein.dateiNichtGeoeffnet, {
              pfad,
              grund: meldungAus(ursache, texte),
            }),
          )
        }
      })()
    },
    [aufschlagen, lauf, merke, sprache],
  )

  // **Alle offenen Meldungen in einem Satz.** Wer drei Dateien zieht und bei
  // zweien scheitert, sieht beide — und nicht die zweite anstelle der ersten.
  const meldung = useMemo(
    () =>
      meldungen.size === 0 ? undefined : [...meldungen.values()].join(' '),
    [meldungen],
  )

  return { dateiGeoeffnet, dateiGezogen, dateiFehler, meldung }
}
