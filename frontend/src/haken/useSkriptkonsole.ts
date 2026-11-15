// **Was die Skriptkonsole tut** — Werkstück U13 von Phase 5.1 (#1424): ein
// Lauf je Karte, Abbruch, Dauer, und jeder Lauf geht in den Verlauf. Wie sie
// aussieht, steht in `seiten/Skriptkonsole.tsx` (§2).
import { useCallback, useState } from 'react'

import type { KonsolenEintrag } from '../antworten'
import type { Skriptstelle } from '../api'
import { api } from '../fassade'
import { meldungAus } from '../meldung'
import { merke } from '../skriptverlauf'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useVorgang } from './useVorgang'

// Konsolenergebnis ist, was nach einem gelungenen Lauf dasteht.
export interface Konsolenergebnis {
  quelltext: string
  art: string
  befunde: string[]
  dauerMs: number
  eintraege: KonsolenEintrag[]
}

export interface Skriptkonsolenstand {
  ausdruck: string
  setzeAusdruck: (ausdruck: string) => void
  ergebnis: Konsolenergebnis | undefined
  fehler: string | undefined
  stelle: Skriptstelle | null
  // abgebrochen sagt, dass der letzte Lauf endete, bevor ein Ergebnis da war.
  abgebrochen: boolean
  laeuft: boolean
  auswerten: () => Promise<void>
  abbrechen: () => void
}

export function useSkriptkonsole(karteID: string): Skriptkonsolenstand {
  const texte = useTexte()

  const [ausdruck, setzeAusdruck] = useState('')
  const [ergebnis, setzeErgebnis] = useState<Konsolenergebnis>()
  const [fehler, setzeFehler] = useState<string>()
  const [stelle, setzeStelle] = useState<Skriptstelle | null>(null)
  const [abgebrochen, setzeAbgebrochen] = useState(false)

  // **Ein Vorgang je Karte** (§7, Regel 4): zwei gleichzeitige Auswertungen
  // gäben zwei Antworten, und die zweite überschriebe die erste ohne Anlass.
  // Go weist einen zweiten Lauf an derselben Karte ohnehin mit `konflikt` ab;
  // hier kommt er gar nicht erst los.
  //
  // **Und eine verspätete Antwort erreicht keinen neueren Stand**: die Seite
  // ist je Karte geschlüsselt (`Seiteninhalt`), und an einer Karte läuft nur
  // einer — der Abbruch beendet ihn, bevor der nächste beginnt.
  const vorgang = useVorgang()

  const auswerten = useCallback(async () => {
    const gestartet = new Date().toISOString()

    const stand = await vorgang.starte(karteID, async () => {
      // **Die Antwort steht außerhalb des Versuchs**, damit Stelle und Dauer
      // auch im Fehlerfall zu lesen sind: `auspacken` wirft, und der Wurf
      // trägt nur die Meldung (#1164).
      let gemeldet: Skriptstelle | null = null
      let dauer = 0

      try {
        const antwort = await api().Skript(karteID, ausdruck)

        gemeldet = antwort.stelle
        dauer = antwort.dauerMs

        return auspacken(antwort, () => antwort)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
        setzeStelle(gemeldet)
        setzeAbgebrochen(false)
        merke({
          ausdruck,
          ausgang: 'gescheitert',
          dauerMs: dauer,
          zeitpunkt: gestartet,
        })

        return undefined
      }
    })

    if (stand === undefined) {
      return
    }

    setzeFehler(undefined)
    setzeStelle(null)

    if (stand.abgebrochen) {
      setzeErgebnis(undefined)
      setzeAbgebrochen(true)
      merke({
        ausdruck,
        ausgang: 'abgebrochen',
        dauerMs: stand.dauerMs,
        zeitpunkt: gestartet,
      })

      return
    }

    setzeAbgebrochen(false)
    setzeErgebnis({
      quelltext: stand.quelltext,
      art: stand.art,
      befunde: stand.befunde,
      dauerMs: stand.dauerMs,
      eintraege: stand.eintraege,
    })
    merke({
      ausdruck,
      ausgang: 'gelungen',
      dauerMs: stand.dauerMs,
      zeitpunkt: gestartet,
    })
  }, [ausdruck, karteID, texte, vorgang])

  // **Abbrechen ist Bestenfalls-Arbeit** (§7, Regel 3): die Antwort des
  // laufenden Rufs sagt, ob er abgebrochen wurde; ein Fehler beim Abbrechen
  // selbst wird geschluckt, denn der Lauf endet so oder so.
  const abbrechen = useCallback(() => {
    const abbruch = api().SkriptAbbrechen(karteID)
    void abbruch.catch(() => undefined)
  }, [karteID])

  return {
    ausdruck,
    setzeAusdruck,
    ergebnis,
    fehler,
    stelle,
    abgebrochen,
    laeuft: vorgang.laeuft(karteID),
    auswerten,
    abbrechen,
  }
}
