// **Preise holen für die Zeilen eines Tauschs** — P10
// (#2378),
// Befund C-17.
//
// # Derselbe Weg wie in der Sammlung
//
// **`PreiseHolen` je Druck, nacheinander** — die Drossel steht im Klienten
// (§16, Regel 4), und ein zweiter Weg zu Scryfall hätte eine zweite. Gefragt
// wird jeder Druck einmal, auch wenn er in zwei Zeilen steht; ein eigener
// Druck ohne Scryfall-Kennung gar nicht.
//
// # Was ein Fehler und ein Abbruch nicht tun
//
// **Die Verhandlung bleibt**: geholt wird in den Preisbestand und nicht in
// den Vertrag, und kein Fehler fasst die Zeilen an. **Ein Fehler hält den
// Lauf nicht auf** — wer bei einem Druck keine Antwort bekommt, soll die
// übrigen trotzdem haben —, **ein Abbruch hält vor dem nächsten Druck**.
// **Geschrieben ist geschrieben** (§7, Regel 3): danach liest die Bilanz
// nach, auch nach einem Abbruch.
import { useEffect, useRef, useState } from 'react'

import { api } from '../fassade'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useLauf } from './useLauf'

import type { Handelsposition } from '../apihandelsleitung'

// Preislauf ist, wie weit das Holen ist.
export type Preislauf =
  | { art: 'laeuft'; fertig: number; gesamt: number }
  | { art: 'fertig'; zahl: number }
  | { art: 'teils'; zahl: number; gesamt: number; meldung: string }
  | { art: 'abgebrochen'; fertig: number; gesamt: number }

export function useHandelspreise(
  positionen: Handelsposition[],
  geholt: () => void,
) {
  const texte = useTexte()
  const lauf = useLauf()
  const abbruch = useRef(false)

  // **Gelesen und nicht geschlossen**: der Abbruch kommt während des Wartens
  // auf eine Antwort, und eine Typprüfung, die nur diese Funktion sieht, hielte
  // ihn für immer falsch.
  const abgebrochen = () => abbruch.current
  const [stand, setzeStand] = useState<Preislauf>()

  // **Ein Wechsel räumt** (§7, Regel 1): wer den Tausch verlässt, hält den
  // Lauf vor dem nächsten Druck an, und keine Antwort schreibt mehr hierher.
  useEffect(
    () => () => {
      abbruch.current = true
      lauf.naechster()
    },
    [lauf],
  )

  const drucke = [
    ...new Set(
      positionen
        .map((position) => position.printing)
        .filter((printing) => printing !== ''),
    ),
  ]
  const laeuft = stand?.art === 'laeuft'

  // **Ein Lauf zur Zeit** (§7, Regel 4) — der Knopf ruht, solange er läuft.
  const holen = async () => {
    const nummer = lauf.naechster()
    const gesamt = drucke.length
    abbruch.current = false
    setzeStand({ art: 'laeuft', fertig: 0, gesamt })

    let fertig = 0
    let fehlgeschlagen = 0
    let meldung = ''

    for (const printing of drucke) {
      if (abgebrochen()) {
        break
      }

      try {
        const antwort = await api().PreiseHolen(printing)
        auspacken(antwort, () => antwort.staende)
      } catch (ursache) {
        fehlgeschlagen += 1
        meldung = meldung === '' ? meldungAus(ursache, texte) : meldung
      }

      fertig += 1
      if (lauf.istAktuell(nummer)) {
        setzeStand({ art: 'laeuft', fertig, gesamt })
      }
    }

    if (!lauf.istAktuell(nummer)) {
      return
    }

    geholt()

    if (abgebrochen()) {
      setzeStand({ art: 'abgebrochen', fertig, gesamt })
    } else if (fehlgeschlagen > 0) {
      setzeStand({ art: 'teils', zahl: fehlgeschlagen, gesamt, meldung })
    } else {
      setzeStand({ art: 'fertig', zahl: fertig })
    }
  }

  return {
    stand,
    laeuft,
    kannHolen: drucke.length > 0,
    holen,
    abbrechen: () => {
      abbruch.current = true
    },
  }
}
