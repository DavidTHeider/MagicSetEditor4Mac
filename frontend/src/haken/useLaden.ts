// Der Wächter jedes Ladeeffekts (AGENTS.md §7).
//
// Er setzt Regel 2 um, ohne dass jede Seite denselben Zähler von Hand führt:
// laden, und die Antwort nur eintragen, wenn sie noch zum aktuellen Stand
// gehört.
import { useCallback, useEffect, useRef, useState } from 'react'

import { Bindungsfehler } from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useLauf } from './useLauf'

export interface Ladezustand<T> {
  daten: T | undefined
  laedt: boolean
  fehler: string | undefined

  // nachladen ist Bestenfalls-Arbeit (Regel 3): es ist da, weil eine geglückte
  // Änderung den Stand verschiebt — nicht, damit jemand bei einem Fehler
  // wiederholt.
  nachladen: () => void
}

interface Stand<T> {
  schluessel: string

  // frage ist der Schlüssel **ohne** die Runde — wonach zuletzt gefragt wurde.
  //
  // **Sie unterscheidet den Wechsel vom Nachladen.** Ein Nachladen erhöht die
  // Runde und damit den Schlüssel; wer daran das Räumen festmachte, räumte
  // auch bei jeder Aktualisierung desselben Bezugsobjekts.
  frage: string
  daten?: T
  fehler?: string
}

export function useLaden<T>(
  holen: () => Promise<T>,
  abhaengig: readonly unknown[],

  // behalten lässt den vorigen Stand stehen, während ein **neues**
  // Bezugsobjekt lädt.
  //
  // **Ohne es wird geräumt** — §7 Regel 1: ein Wechsel des Bezugsobjekts räumt
  // den offenen Kontext. Bis zum 2026-08-22 behielt der Haken die alten Daten
  // immer, und wer das Set wechselte, sah die Karten des vorigen unter dem
  // Namen des neuen (Zweitblick über Phase L).
  //
  // **Die Vorschau will das Gegenteil**, und deshalb ist es eine Wahl: ihre
  // Abhängigkeiten ändern sich mit jedem Tastendruck, und ein Bild, das dabei
  // verschwindet und wiederkommt, ist schlechter als ein einen Augenblick
  // altes.
  behalten = false,

  // nachlesen ist ein **Nachlesestand** (#2069): ändert er sich, wird
  // nachgeladen wie mit `nachladen` — derselbe Bezug, ein neuer Stand, und
  // was dasteht, bleibt stehen. **Ein Zähler gehört hierher und nicht in die
  // Abhängigkeiten**: dort wäre er eine neue Frage und räumte. Bis zum
  // 2026-10-05 standen der Stand fremder Schreibvorgänge, eigene Runden und
  // der Verlaufsstand dort, und die Paketliste ersetzte ihre Tabelle bei
  // jedem fremden Schreiben für einen Augenblick durch „Wird geladen" — die
  // Seite sprang um 37 Punkte.
  nachlesen: unknown = 0,
): Ladezustand<T> {
  const lauf = useLauf()
  const texte = useTexte()
  const [runde, setzeRunde] = useState(0)
  const [stand, setzeStand] = useState<Stand<T>>({ schluessel: '', frage: '' })

  // nachgeholt merkt, für welche **Frage** schon einmal nachgeschlagen wurde.
  //
  // **An den Abhängigkeiten und nicht am Schlüssel.** Der Nachschlag erhöht
  // die Runde und damit den Schlüssel; wer daran merkte, merkte sich beim
  // zweiten Anlauf nichts mehr und drehte.
  const nachgeholt = useRef<string>('')

  // Die Frage ist, **wonach** gerade gefragt wird — ohne die Runde.
  const frage = JSON.stringify(abhaengig)

  // Der Schlüssel sagt, **welchen** Stand die Anzeige gerade meint — mit
  // Runde und Nachlesestand, die beide nachladen, ohne zu räumen.
  const schluessel = `${frage}#${String(runde)}#${JSON.stringify(nachlesen)}`

  const nachladen = useCallback(() => {
    setzeRunde((zahl) => zahl + 1)
  }, [])

  useEffect(() => {
    const meiner = lauf.naechster()

    holen()
      .then((antwort) => {
        // **Der Wächter.** Ohne ihn trägt die Antwort eines vorigen Sets sich
        // in den neuen Stand ein — und dort steht dann die Kartenliste des
        // vorigen, ohne Fehlermeldung und ohne erkennbaren Anlass.
        if (!lauf.istAktuell(meiner)) {
          return
        }

        setzeStand({ schluessel, frage, daten: antwort })
      })
      .catch((ursache: unknown) => {
        if (!lauf.istAktuell(meiner)) {
          return
        }

        // **Ein Abbruch ist kein Fehler, sondern die Absicht** (§7).
        //
        // Die Vorschau bricht den vorigen Lauf ab, sobald ein neuer beginnt.
        // Über der gezeichneten Karte stand dabei „Es ist ein unerwarteter
        // Fehler aufgetreten" — die Meldung des Laufs, der genau so
        // abgebrochen wurde, wie es vorgesehen ist
        // (#288).
        //
        // **Das Vorige bleibt stehen.** Wer den Stand leerte, zeigte statt der
        // Karte eine Ladeanzeige, die nie endet: den abgebrochenen Lauf holt
        // niemand nach.
        if (
          ursache instanceof Bindungsfehler &&
          ursache.code === 'abgebrochen'
        ) {
          // **Ohne Vorheriges wird nachgeholt — genau einmal.**
          //
          // Zwei Läufe gehen gleichzeitig hinaus — beim Kartenwechsel, oder
          // weil React einen Effekt im Entwicklungsmodus zweimal ausführt.
          // `App.Vorschau` bricht den **vorigen** ab; welcher das ist,
          // entscheidet die Reihenfolge, in der Go sie annimmt, und die muss
          // nicht die sein, in der das Fenster sie losgeschickt hat.
          //
          // Trifft es den Lauf, auf den hier gewartet wird, und gibt es kein
          // voriges Bild, bliebe **nichts** stehen: keine Ladeanzeige, keine
          // Meldung, kein Bild — für immer
          // (#566).
          //
          // **Einmal und nicht bis es klappt.** Ein Nachschlag ohne Schranke
          // dreht, sobald etwas jeden Lauf abbricht; einer genügt, denn der
          // Nachschlag ist allein unterwegs und kann das Rennen nicht
          // verlieren.
          if (stand.daten === undefined && nachgeholt.current !== frage) {
            nachgeholt.current = frage

            setzeRunde((zahl) => zahl + 1)
          }

          setzeStand((bisher) => ({ ...bisher, schluessel, frage }))

          return
        }

        // **Bestätigte Daten bleiben beim Nachlesen stehen** — U07 (#2123):
        // derselbe Bezug ist bekannt, nur der Weg dorthin ging schief. Bis
        // dahin warf ein Fehler die Daten weg, und das Arbeitsobjekt
        // verschwand hinter einer Meldung. **Ein neuer Bezug bekommt keine
        // fremden** — außer die Stelle will sie ausdrücklich (`behalten`).
        setzeStand((bisher) => {
          const naechster: Stand<T> = {
            schluessel,
            frage,
            fehler: meldungAus(ursache, texte),
          }

          if (
            (bisher.frage === frage || behalten) &&
            bisher.daten !== undefined
          ) {
            naechster.daten = bisher.daten
          }

          return naechster
        })
      })
    // `holen` ändert sich bei jedem Zeichnen und gehört deshalb nicht dazu; die
    // Abhängigkeiten stecken im Schlüssel.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [schluessel])

  // **`laedt` wird abgeleitet, nicht gesetzt.**
  //
  // Der erste Anlauf rief `setzeLaedt(true)` im Effektrumpf — eine
  // Zustandsänderung als Folge einer Zustandsänderung, und React rät zu Recht
  // davon ab. Abgeleitet ist es zugleich richtiger: „lädt" heißt genau, dass
  // der eingetragene Stand nicht der gewünschte ist.
  const laedt = stand.schluessel !== schluessel

  return {
    // **Beim Wechsel geräumt, beim Nachladen behalten.**
    //
    // Der Unterschied ist die **Frage**, nicht der Schlüssel: ein Nachladen
    // erhöht die Runde, fragt aber nach demselben Bezugsobjekt. Ein Stand, der
    // zu einem **anderen** gehört, ist kein Zwischenstand, sondern eine
    // falsche Auskunft.
    daten: stand.frage !== frage && !behalten ? undefined : stand.daten,
    laedt,
    fehler: laedt ? undefined : stand.fehler,
    nachladen,
  }
}
