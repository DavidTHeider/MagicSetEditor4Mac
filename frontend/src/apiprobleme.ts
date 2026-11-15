// **Die Probleme jedes Sets** — AP-8, Teil 2
// (#2365),
// über den Vertrag aus PERF02 (#2352): gelesen wird ohne zu prüfen, geprüft
// auf Anstoß. Eine Zahl gilt nur, wenn `sicher` sie trägt.
import type { Umschlag } from './umschlag'

export type Problemurteil =
  'ungeprueft' | 'aktuell' | 'veraltet' | 'unvollstaendig'

export interface SetproblemstandDTO {
  setID: string
  name: string
  urteil: Problemurteil

  // zahl sind die Karten mit Befund — **nur bei `sicher` eine, die gilt**.
  zahl: number
  sicher: boolean
  geprueft: number
  gesamt: number

  // beendet ist der Zeitpunkt des Stands (RFC 3339), alter seine Zeit in
  // Sekunden — leer und null, solange nie geprüft wurde.
  beendet: string
  alter: number
  karten: { karteID: string; befunde: string[] }[]
}

export interface SetproblemeAntwort extends Umschlag {
  staende: SetproblemstandDTO[]
}

export interface Problemapi {
  Setprobleme: () => Promise<SetproblemeAntwort>
  SetproblemePruefen: () => Promise<SetproblemeAntwort>
}
