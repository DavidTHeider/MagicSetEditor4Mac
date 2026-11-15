// **Was der Editor gerade zeigt** — die geladene Vorschau, die Probe beim
// Tippen und die daraus gerechneten Vorgabewerte.
//
// # Warum ein eigener Haken
//
// Es ist **eine** Frage: welches Bild steht jetzt da? Sie wird aus zwei
// Läufen beantwortet, die einander ablösen, und aus einer Regel darüber,
// welcher gilt. Im Editor stand das zwischen dem Formular und den Feldern —
// und ließ ihn über die Zeilengrenze aus §2 wachsen.
import { useCallback } from 'react'

import type { VorschauDTO } from '../api'
import { api, auspacken } from '../api'
import { Katalogfehler } from '../meldung'
import { useLaden } from './useLaden'

// Bildstand ist ein gezeigtes Bild samt der Karte, zu der es gehört.
export interface Bildstand {
  karteID: string
  bild: VorschauDTO
}

export interface Kartenvorschau {
  // gezeigt ist das Bild, das gerade dasteht — undefined, solange keines da
  // war.
  gezeigt: VorschauDTO | undefined

  // abgeloest ist das **letzte** Bild, solange die neue Antwort lädt — meist
  // das der vorigen Karte, nur zum Ansehen (AP-4, #2361): sonst stand bis zur Antwort gar keines
  // da, und die Spalte fiel zusammen. **Daraus gilt nichts**: Felder,
  // Befunde und Vorgaben kommen nur aus `gezeigt` (§7, Regel 1).
  abgeloest: VorschauDTO | undefined

  // vorgaben sind die gerechneten Vorgabewerte je Feld.
  vorgaben: Map<string, string>

  // nachladen holt die geschriebene Vorschau neu.
  nachladen: () => void

  laedt: boolean
  fehler: string | undefined
}

// useKartenvorschau lädt die Vorschau einer Karte und die Probe beim Tippen.
export function useKartenvorschau(
  karteID: string,
  getippt: { feld: string; wert: string } | undefined,

  // vergroesserung ist der Faktor, mit dem der Go-Teil rastert — die Zoomwahl
  // der Vorschau (U5, #1416). 1 ist die Vorlage in ihren Punkten.
  vergroesserung = 1,

  // stand ist der Verlaufsstand: eine Änderung anderswo liest nach, ohne die
  // Karte zu räumen (U01, #2119).
  stand = 0,

  // voriges ist das zuletzt gezeigte Bild — **über dem Schlüssel des
  // Editors gehalten** (`Kartenarbeitsflaeche`), denn der Editor beginnt mit
  // jeder Karte frisch (§7, Regel 1).
  voriges?: Bildstand,
): Kartenvorschau {
  const holen = useCallback(async () => {
    const antwort = await api().Vorschau(karteID, vergroesserung)

    return auspacken(antwort, () => {
      if (!antwort.vorschau) {
        throw new Katalogfehler('vorschauLeer')
      }

      return antwort.vorschau
    })
  }, [karteID, vergroesserung])

  // **Der Maßstab liest nach und fragt nicht neu** (U12, #2126): als Teil
  // der Frage räumte er, und beim Zoomen verschwand die Karte, bis neu
  // gerastert war. Eine andere Karte ist eine neue Frage und räumt.
  const vorschau = useLaden<VorschauDTO>(holen, [karteID], false, {
    vergroesserung,
    stand,
  })

  const holenProbe = useCallback(async () => {
    if (getippt === undefined) {
      return undefined
    }

    const antwort = await api().Probevorschau(
      karteID,
      getippt.feld,
      getippt.wert,
      vergroesserung,
    )

    return auspacken(antwort, () => antwort.vorschau ?? undefined)
  }, [karteID, getippt, vergroesserung])

  // **Kein Warten, sondern Abbrechen** — die Sonderregel dieses Projekts (§7).
  //
  // Jeder Tastendruck stößt einen Lauf an; der vorige wird im Backend über
  // seinen `context` abgebrochen (`App.Probevorschau`), und `useLaden` verwirft
  // eine überholte Antwort. Ein Zeitgeber davor wäre eine zweite Antwort auf
  // dieselbe Frage — und im Test ein Wackelkandidat.
  //
  // **Und das Getippte liest nach** (U12): als Teil der Frage räumte es, und
  // zwischen zwei Tastendrücken sprang das Bild auf den gespeicherten Stand
  // zurück — ein Stand, den gerade niemand meint.
  const probe = useLaden<VorschauDTO | undefined>(
    holenProbe,
    [karteID],
    false,
    { getippt, vergroesserung },
  )

  // **Die Probe schlägt die geladene Vorschau**, solange sie ein Bild hat.
  //
  // **Und sonst bleibt das letzte stehen.** Zwischen zwei Tastendrücken hat die
  // Probe noch keine Antwort; wer dann auf sie umschaltete, ließe die Karte bei
  // jedem Buchstaben verschwinden und wiederkommen.
  const gezeigt = probe.daten ?? vorschau.daten

  // **Das letzte Bild**, solange die neue Antwort lädt — und nur dann.
  //
  // **Auch wenn es dieselbe Karte ist**: auf dem schnellen Rückweg A → B → A
  // gehört das letzte Bild A selbst, und eine Bedingung „nur das einer
  // anderen Karte“ ließ die Vorschau leer (Zweitblick über Phase 6 von 0.9.0,
  // Befund 7). Es steht gedämpft, `inert` und als „Wird geladen …“ da.
  const abgeloest =
    gezeigt === undefined && vorschau.laedt && voriges !== undefined
      ? voriges.bild
      : undefined

  // **Die gerechneten Vorgabewerte kommen mit der Vorschau** (#563): ein
  // `default:` ist ein Skript über `card`, und dieselbe Vorlage gibt für zwei
  // Karten zwei Werte — die Auskunft gehört zur **Karte**, nicht zum Set.
  //
  // **Aus `gezeigt` und nicht aus `vorschau.daten`**: während des Tippens gilt
  // die Probevorschau, und sie rechnet dieselben Werte mit dem eben getippten
  // Text. Wer die geladene nähme, zeigte die Herleitung von vorhin.
  const vorgaben = new Map(
    (gezeigt?.vorgaben ?? []).map((vorgabe) => [vorgabe.feld, vorgabe.wert]),
  )

  return {
    gezeigt,
    abgeloest,
    vorgaben,
    nachladen: vorschau.nachladen,
    laedt: vorschau.laedt,
    // **Auch der Fehler der Probevorschau** — E01 (#2238). Scheitert sie,
    // behält `useLaden` das Bild des vorigen Entwurfs; ohne Meldung sah es
    // aus wie das zum neuen Text (am nativen Fenster gefunden, `e01vorschau`).
    fehler: probe.fehler ?? vorschau.fehler,
  }
}
