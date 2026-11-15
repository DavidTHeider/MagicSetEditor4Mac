// Aufmerksamkeit und Hintergrundaufgaben — Werkstück U11 von Phase 5.1
// (#1422).
// Die Gestalt kommt aus `internal/app/aufmerksamkeit.go` und `auftraege.go`.
import type { Umschlag } from './umschlag'
import type { ArbeitsobjektDTO } from './apiweiterarbeiten'
import type { Betriebsbandapi } from './apibetriebsband'
import type { Problemapi } from './apiprobleme'

export type Hinweisart = 'setVorbehalte' | 'deckFehlend' | 'paketBefund'

export interface HinweisDTO {
  art: Hinweisart
  objekt: ArbeitsobjektDTO
  anzahl: number

  // grund ist der deutsche Satz aus Go — der Rückfall, wenn der Katalog den
  // Schlüssel nicht kennt (§9).
  grund: string

  // ort ist der Bezeichner vor dem Satz — die Datei eines Skriptbefunds; er
  // wird nicht übersetzt. schluessel nennt den Satz im Katalog, werte füllen
  // seine Platzhalter `{0}`, `{1}` … — leer heißt: der Grund ist kein Satz,
  // sondern ein Name (#1963).
  ort: string
  schluessel: string
  werte: string[]
}

export interface AufmerksamkeitAntwort extends Umschlag {
  hinweise: HinweisDTO[]
  durchgesehen: { sets: number; decks: number; pakete: number }
  befunde: string[]
}

export type Auftragsstand = 'laeuft' | 'fertig' | 'gescheitert' | 'abgebrochen'

export interface AuftragDTO {
  id: string
  art: string
  was: string
  fortschritt: string
  stand: Auftragsstand
  grund: string
  begonnen: string
  sekunden: number

  // Der gezählte Stand eines Datei-Laufs (FMT-03) — `gesamt` 0 heißt „nicht
  // bekannt", `einheit` ist ein Schlüssel wie `karte`.
  erledigt: number
  gesamt: number
  einheit: string
  abbrechbar: boolean
}

export interface AuftraegeAntwort extends Umschlag {
  auftraege: AuftragDTO[]
  laufend: number
}

// **Das Betriebsband gehört dazu** (OPS02, #2348): seine Hinweise stehen in
// derselben Liste.
// **Und die Probleme der Sets** (AP-8, #2365) — dieselbe Übersicht.
export interface Aufmerksamkeitsapi extends Betriebsbandapi, Problemapi {
  Aufmerksamkeit: () => Promise<AufmerksamkeitAntwort>
  Auftraege: () => Promise<AuftraegeAntwort>
  AuftragAbbrechen: (id: string) => Promise<AuftraegeAntwort>
}
