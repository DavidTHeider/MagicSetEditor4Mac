// **Die Vorrichtung für die Bildfläche der Kartenvorschau.**
//
// Eigene Datei, seit #1112
// einen zweiten Lauf über denselben Aufbau gebracht hat und `bildflaeche.test.tsx`
// damit über 400 Zeilen ginge (§2).
//
// **Unter `src/test/`**: dort steht, womit Läufe anfangen, und die Abdeckung
// lässt es aus (`vite.config.ts`). Ein Aufbau, der selbst eine Schwelle tragen
// müsste, brächte Läufe hervor, die nur ihn prüfen.
import { render } from '@testing-library/react'
import { vi } from 'vitest'

import type { FeldbeschreibungDTO, KarteDTO, VorschauAntwort } from '../api'
import { Ablageraum } from '../ablage'
import { useAbgelegteDateien } from '../haken/useAbgelegteDateien'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from './api'
import { zuschnittantwort } from './zuschnitt'
import { Karteneditor } from '../seiten/Karteneditor'

export function vorschau(): VorschauAntwort {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild: '/vorschau/AAAA.png',
      breite: 100,
      hoehe: 140,
      dpi: 150,
      befunde: [],
      flaechen: [
        { feld: 'kunst', links: 0, oben: 0, breite: 100, hoehe: 80 },
        { feld: 'name', links: 0, oben: 90, breite: 100, hoehe: 20 },
      ],
      nennungen: [],
      vorgaben: [],
    },
  }
}

export function form(name: string, art: string): FeldbeschreibungDTO {
  return {
    name,
    aufschrift: name === 'kunst' ? 'Kunst' : 'Name',
    art,
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
  }
}

export const karte: KarteDTO = { id: '5', nummer: 0, werte: [] }

export function zeigeBildflaeche(teile = {}) {
  const BildSetzen = vi.fn(() => Promise.resolve(zuschnittantwort()))
  const BildAusDatei = vi.fn(() => Promise.resolve(zuschnittantwort()))
  const aufSet = vi.fn()

  belegeApi({
    Vorschau: () => Promise.resolve(vorschau()),
    Probevorschau: () => Promise.resolve(vorschau()),
    BildSetzen,
    BildAusDatei,
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      {/* **Der Ablageraum gehört dazu** (#891): der Horcher auf ein Ablegen
          steht seit dem 2026-08-25 eine Ebene über der Kartenseite — Wails
          meldet es **einmal fürs ganze Fenster**. Ohne ihn prüfte dieser Lauf
          eine Anmeldung, die es nicht mehr gibt. */}
      <Ablageraum>
        <Horcher aufSet={aufSet} />
        <Karteneditor
          setID="1"
          karte={karte}
          formen={[form('kunst', 'image'), form('name', 'text')]}
          geaendert={() => undefined}
        />
      </Ablageraum>
    </Sprachraum>,
  )

  return { BildSetzen, BildAusDatei, aufSet }
}

// Horcher meldet das Ablegen an — so wie `Anwendung` es tut.
export function Horcher({ aufSet }: { aufSet: (pfad: string) => void }) {
  useAbgelegteDateien(aufSet)

  return null
}

export interface MitLaufzeit {
  runtime?: {
    OnFileDrop?: (
      rueckruf: (x: number, y: number, pfade: string[]) => void,
      ueberZiel: boolean,
    ) => void
    OnFileDropOff?: () => void
  }
}
