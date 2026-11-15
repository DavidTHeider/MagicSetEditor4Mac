// Rückgängig und Wiederherstellen — ⌘Z und ⇧⌘Z (#52).
//
// # Die vier Regeln aus AGENTS.md §7
//
//  1. **Ein Wechsel des Bezugsobjekts räumt den offenen Kontext.** Wer das Set
//     wechselt, verliert Stand und Meldung des vorigen.
//  2. **Eine Antwort gehört zu dem Stand, aus dem sie stammt.** `useLauf` zählt
//     die Wechsel mit; eine Antwort aus einem vorigen Set wird verworfen.
//  3. **Geschrieben ist geschrieben.** Der Stand steigt nur nach einer
//     geglückten Bewegung — nach einem Fehler bleibt er stehen, und die Meldung
//     ist das, was zu sehen ist.
//  4. **Ein Vorgang je Datensatz.** `useVorgang` hält den zweiten Aufruf an,
//     solange der erste läuft; sonst wanderte der Verlauf zwei Schritte, wo
//     jemand einen gemeint hat.
import { useCallback, useState } from 'react'

import { api, auspacken } from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useLauf } from './useLauf'
import { useVorgang } from './useVorgang'

export interface Verlaufssteuerung {
  // stand steigt nach jeder geglückten Bewegung.
  //
  // **Als Zahl und nicht als Rückruf**, weil die Kartenseite ihn in ihre
  // Ladeabhängigkeiten legt: was sich geändert hat, steht im Bestand, und
  // gelesen wird es von dort. Ein Rückruf verlangte, dass der Rahmen die
  // Kartenseite kennt.
  stand: number

  fehler: string | undefined

  // **Mit `gezielt` genau diesen Schritt** — Werkstück U6 von Phase 5.1
  // (#1417): die Verlaufsliste nennt den Vorgang; ⌘Z lässt ihn leer und
  // trifft den jüngsten.
  zuruecknehmen: (gezielt?: string) => void
  wiederherstellen: (gezielt?: string) => void
}

export function useVerlauf(setID: string | undefined): Verlaufssteuerung {
  const lauf = useLauf()
  const vorgang = useVorgang()
  const texte = useTexte()

  const [stand, setzeStand] = useState(0)
  const [fehler, setzeFehler] = useState<string>()
  const [vorlage, setzeVorlage] = useState(setID)

  // Regel 1: der Setwechsel räumt mit. **Beim Zeichnen und nicht im Effekt** —
  // ein Effekt zeigte für einen Durchgang die Meldung des vorigen Sets über dem
  // neuen.
  if (vorlage !== setID) {
    setzeVorlage(setID)
    setzeStand(0)
    setzeFehler(undefined)

    // **Der Wechsel zählt den Lauf weiter** — sonst greift Regel 2 hier nicht.
    //
    // `useLaden` bekommt das geschenkt: dort löst der Wechsel einen neuen
    // Ladelauf aus, und der nimmt seine Marke. Hier löst er gar nichts aus, und
    // ohne diese Zeile trüge sich die Antwort einer laufenden Rücknahme in das
    // **neue** Set ein — mit einem Stand, der eine Liste nachlädt, die mit der
    // Antwort nichts zu tun hat, und einer Meldung über ein Set, das niemand
    // mehr offen hat. Der Test hat es gefunden; die erste Fassung hatte die
    // Zeile nicht und war grün, weil kein Fall sie verlangte.
    lauf.naechster()
  }

  // nurAktuell führt eine Eintragung aus, wenn sie noch zum heutigen Stand
  // gehört — Regel 2, an **einer** Stelle für Erfolg und Fehler.
  //
  // **Synchron und ausgelagert**, aus zwei Gründen. Erstens stand die Prüfung
  // zweimal da, einmal im `try` und einmal im `catch`; die Meldung ist die
  // unauffälligere Hälfte, und zwei Fassungen derselben Regel laufen
  // auseinander. Zweitens misst die Abdeckung einen `if` ohne `else` im Rumpf
  // einer `async`-Funktion falsch — sie kam auf einen **negativen** Zähler für
  // den nicht genommenen Zweig. Hier zählt sie richtig, und die Regel ist
  // dieselbe.
  const nurAktuell = useCallback(
    (meiner: number, eintragen: () => void) => {
      if (lauf.istAktuell(meiner)) {
        eintragen()
      }
    },
    [lauf],
  )

  const bewege = useCallback(
    (richtung: 'zurueck' | 'vor', gezielt: string) => {
      if (setID === undefined) {
        return
      }

      // Regel 4: **je Set**, nicht je Fenster. Zwei Sets nebeneinander gibt es
      // heute nicht; der Schlüssel sagt trotzdem, worauf sich die Sperre
      // bezieht, statt es dem Zufall zu überlassen.
      void vorgang.starte(setID, async () => {
        // **Die Laufmarke wird hier genommen, nicht davor.**
        //
        // Davor zählte auch ein Aufruf mit, den Regel 4 gar nicht ausführt —
        // und entwertete damit die Antwort des laufenden: der zweite ⌘Z-Druck
        // machte aus der ersten Rücknahme eine, die niemand einträgt. Der Test
        // „lässt keinen zweiten Aufruf zu" hat es gefunden.
        const meiner = lauf.naechster()

        try {
          const antwort =
            richtung === 'zurueck'
              ? await api().Zuruecknehmen(setID, gezielt)
              : await api().Wiederherstellen(setID, gezielt)

          const bewegung = auspacken(antwort, () => antwort)

          nurAktuell(meiner, () => {
            setzeFehler(undefined)

            // **Nur eine wirkliche Bewegung verschiebt den Stand.** ⌘Z am
            // Anfang des Verlaufs ist eine gewöhnliche Lage; wer dabei nachlüde,
            // ließe die ganze Liste neu holen, ohne dass sich etwas geändert
            // hat.
            if (bewegung.bewegt) {
              setzeStand((zahl) => zahl + 1)
            }
          })
        } catch (ursache: unknown) {
          nurAktuell(meiner, () => {
            setzeFehler(meldungAus(ursache, texte))
          })
        }
      })
    },
    [setID, lauf, vorgang, nurAktuell, texte],
  )

  const zuruecknehmen = useCallback(
    (gezielt = '') => {
      bewege('zurueck', gezielt)
    },
    [bewege],
  )

  const wiederherstellen = useCallback(
    (gezielt = '') => {
      bewege('vor', gezielt)
    },
    [bewege],
  )

  return { stand, fehler, zuruecknehmen, wiederherstellen }
}
