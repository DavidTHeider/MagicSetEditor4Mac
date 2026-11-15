// **Die Gestalt der Kartenliste holen** — Werkstück 3 von Phase 3.7
// (#1168).
//
// # Warum ein eigener Ladeweg neben `useKartendaten`
//
// **Weil er eine andere Frage stellt.** Jener holt die Karten; dieser holt,
// **wie sie zu zeigen sind** — und das hängt am Spiel, nicht am Set. Beides in
// einer Antwort zu bündeln hieße, die Kartenliste jedes Mal neu zu laden, wenn
// jemand eine Spalte ein- oder ausblendet.
import { useMemo } from 'react'

import { api, type ListenspalteDTO } from '../api'
import type { Listenzusatz } from '../seiten/Kartenliste'
import type { Listengestalt } from '../seiten/listengestalt'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'
import type { Vorgang } from './useVorgang'

interface Gestaltlage {
  setID: string

  // stand steigt bei jeder Änderung am Bestand — die **Werte** der Spalten
  // hängen daran: wer ein Feld ändert, ändert eine Zelle.
  stand: number

  vorgang: Vorgang
  aufFehler: (meldung: string | undefined) => void
}

export function useListengestalt({
  setID,
  stand,
  vorgang,
  aufFehler,
}: Gestaltlage): Listengestalt {
  const geladen = useLaden(async () => {
    const antwort = await api().Kartenlistenaufbau(setID)

    return auspacken(antwort, () => antwort)
  }, [setID, stand])

  const aufbau = geladen.daten

  // **Die Zusätze werden hier gerechnet und nicht in der Liste**: sie hängen
  // an der Antwort, und die Liste zeichnet zwanzigmal je Rollschritt.
  const zusaetze = useMemo<Listenzusatz[]>(() => {
    if (aufbau === undefined) {
      return []
    }

    return aufbau.spalten
      .filter((spalte) => spalte.sichtbar && spalte.feld !== nameSpalte)
      .map((spalte) => ({
        feld: spalte.feld,
        aufschrift: spalte.aufschrift,
        ausrichtung: spalte.ausrichtung,
        werte: Object.fromEntries(
          aufbau.zeilen.map((zeile) => [
            zeile.karteID,
            zeile.werte[spalte.feld] ?? '',
          ]),
        ),
      }))
  }, [aufbau])

  // **Die Marken je Karte** (#1208) — aus derselben Antwort, in der die Zeilen
  // stehen. Ein zweiter Ruf hätte einen zweiten Stand.
  const marken = useMemo<Record<string, readonly string[]>>(
    () =>
      Object.fromEntries(
        (aufbau?.zeilen ?? []).map((zeile) => [zeile.karteID, zeile.marken]),
      ),
    [aufbau],
  )

  // **Die Reihenfolge ist die der Zeilen** — der Dienst hat sie sortiert.
  const ordnung = useMemo<string[]>(
    () => (aufbau?.zeilen ?? []).map((zeile) => zeile.karteID),
    [aufbau],
  )

  return {
    spalten: aufbau?.spalten ?? leereSpalten,
    zusaetze,
    ordnung,
    marken,
    sortierfeld: aufbau?.sortierfeld ?? '',
    aufsteigend: aufbau?.aufsteigend ?? true,
    fehler: geladen.fehler,
    ohneVorlage: aufbau?.ohneVorlage ?? false,
    vorgang,
    geaendert: () => {
      geladen.nachladen()
      aufFehler(undefined)
    },
  }
}

// nameSpalte ist das Feld, das die Liste **schon** zeigt.
//
// **Es steht nicht ein zweites Mal darunter.** Der Name ist die Zeile; eine
// Spalte »Name: Feuerball« unter »Feuerball« wäre dieselbe Auskunft zweimal.
const nameSpalte = 'name'

// leereSpalten ist eine geteilte leere Liste — **eine** und nicht je Aufruf
// eine neue, damit ein Vergleich nach Gleichheit nicht bei jedem Zeichnen
// scheitert.
const leereSpalten: readonly ListenspalteDTO[] = []
