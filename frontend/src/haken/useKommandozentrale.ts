// **Was die Kommandozentrale tut** — Werkstück U12 von Phase 5.1 (#1423);
// wie sie aussieht, steht in `rahmen/Kommandozentrale.tsx` (§2).
import { useCallback, useMemo, useState } from 'react'
import { api } from '../api'
import { Bindungsfehler } from '../bindungsfehler'
import {
  argumenteAus,
  bereinigt,
  parameterAus,
  type Argumentbefund,
  type Parameter,
  type Werte,
} from '../befehle'
import { merke, type Ausgang } from '../befehlsverlauf'
import type { BefehlDTO, BefehlsergebnisAntwort } from '../apibefehle'
import type { FehlerDTO } from '../formen'

export interface Befehlsergebnis {
  ausgang: Ausgang
  text: string
  dauerMs: number
  fehler: FehlerDTO | null
}

export interface Kommandozentrale {
  parameter: Parameter[]
  werte: Werte
  setzeWert: (name: string, wert: string | boolean) => void
  befunde: Argumentbefund[]
  vorschau: string
  bestaetigt: boolean
  setzeBestaetigt: (wert: boolean) => void
  bereit: boolean
  laeuft: boolean
  ergebnis: Befehlsergebnis | undefined
  ausfuehren: () => void
  abbrechen: () => void
  uebernehmen: (argumente: Record<string, unknown>) => void
}

// ausgangVon liest den Ausgang aus der Antwort — ein Abbruch ist keiner der
// beiden anderen (§7).
function ausgangVon(antwort: BefehlsergebnisAntwort): Ausgang {
  if (antwort.fehler === null || antwort.fehler === undefined) {
    return 'gelungen'
  }

  return antwort.fehler.code === 'abgebrochen' ? 'abgebrochen' : 'gescheitert'
}

export function useKommandozentrale(befehl: BefehlDTO): Kommandozentrale {
  const parameter = useMemo(() => parameterAus(befehl.schema), [befehl.schema])
  const [werte, setzeWerte] = useState<Werte>({})
  const [bestaetigt, setzeBestaetigt] = useState(false)
  const [laeuft, setzeLaeuft] = useState(false)
  const [ergebnis, setzeErgebnis] = useState<Befehlsergebnis | undefined>(
    undefined,
  )

  const { argumente, befunde } = useMemo(
    () => argumenteAus(parameter, werte),
    [parameter, werte],
  )
  const vorschau = useMemo(
    () => JSON.stringify(argumente, null, 2),
    [argumente],
  )

  const setzeWert = useCallback((name: string, wert: string | boolean) => {
    setzeWerte((bisher) => ({ ...bisher, [name]: wert }))
  }, [])

  // uebernehmen füllt die Felder aus einem Verlaufseintrag — Zahlen und
  // Objekte als Text, Wahrheiten als Haken.
  const uebernehmen = useCallback((vorige: Record<string, unknown>) => {
    const neu: Werte = {}
    for (const [name, wert] of Object.entries(vorige)) {
      if (typeof wert === 'boolean') {
        neu[name] = wert
      } else if (typeof wert === 'string') {
        neu[name] = wert
      } else {
        neu[name] = JSON.stringify(wert)
      }
    }
    setzeWerte(neu)
    setzeErgebnis(undefined)
  }, [])

  const bereit = befunde.length === 0 && (!befehl.gefaehrlich || bestaetigt)

  // **Ein Lauf zur Zeit** (§7, Regel 4): der Knopf ist angehalten, und Go
  // weist einen zweiten ohnehin ab.
  const ausfuehren = useCallback(() => {
    if (!bereit || laeuft) {
      return
    }

    setzeLaeuft(true)
    setzeErgebnis(undefined)

    const gestartet = new Date().toISOString()

    const lauf = api().BefehlAusfuehren(
      befehl.name,
      JSON.stringify(argumente),
      bestaetigt,
    )

    void lauf
      .then((antwort) => {
        const ausgang = ausgangVon(antwort)
        setzeErgebnis({
          ausgang,
          text: antwort.text,
          dauerMs: antwort.dauerMs,
          fehler: antwort.fehler ?? null,
        })
        merke({
          name: befehl.name,
          argumente: bereinigt(argumente),
          ausgang,
          dauerMs: antwort.dauerMs,
          zeitpunkt: gestartet,
        })
      })
      // **Eine abgewiesene Zusage ist ein Ausgang wie jeder andere** —
      // Befund 8 des Zweitblicks über Phase 5.1 (2026-09-10). Ohne `catch`
      // ginge bei einer Ablehnung der Naht nur `laeuft` zurück, und sonst
      // geschähe **nichts** — kein Ergebnis, keine Meldung, kein
      // Verlaufseintrag, dazu eine unbehandelte Zusage im Protokoll.
      //
      // **Die Meldung kommt roh durch**, wie die aus Go daneben: diese Fläche
      // zeigt `fehler.meldung` unübersetzt, und ein zweiter Weg für denselben
      // Kasten wären zwei Wahrheiten über dieselbe Sache.
      .catch((ursache: unknown) => {
        setzeErgebnis({
          ausgang: 'gescheitert',
          text: '',
          dauerMs: Date.now() - Date.parse(gestartet),
          fehler: {
            code:
              ursache instanceof Bindungsfehler ? ursache.code : 'unbekannt',
            schluessel:
              ursache instanceof Bindungsfehler
                ? ursache.schluessel
                : 'unbekannt',
            meldung:
              ursache instanceof Error ? ursache.message : String(ursache),
          },
        })
        merke({
          name: befehl.name,
          argumente: bereinigt(argumente),
          ausgang: 'gescheitert',
          dauerMs: Date.now() - Date.parse(gestartet),
          zeitpunkt: gestartet,
        })
      })
      .finally(() => {
        setzeLaeuft(false)
      })
  }, [bereit, laeuft, befehl.name, argumente, bestaetigt])

  // abbrechen ist Bestenfalls-Arbeit: die Antwort des laufenden Rufs sagt,
  // ob es geklappt hat (§7, Regel 3).
  const abbrechen = useCallback(() => {
    const abbruch = api().BefehlAbbrechen()
    void abbruch.catch(() => undefined)
  }, [])

  return {
    parameter,
    werte,
    setzeWert,
    befunde,
    vorschau,
    bestaetigt,
    setzeBestaetigt,
    bereit,
    laeuft,
    ergebnis,
    ausfuehren,
    abbrechen,
    uebernehmen,
  }
}
