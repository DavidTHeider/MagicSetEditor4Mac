// **Was ein Batch und eine Gruppe zurückgeben** — Werkstück A13.1b von Phase
// 5.3 (#1452).
//
// # Eine eigene Datei
//
// **Weil `antwortenmarkt.ts` mit ihnen über die Zeilengrenze aus §2 ging** —
// und weil der Schnitt der Frage folgt: dort steht, was **eine** Karte
// zurückbekommt, hier, was **mehrere** zurückbekommen.
import type { FehlerDTO } from './formen'
import type { KIFeldwertDTO } from './antwortenmarkt'
import type { Umschlag } from './umschlag'

// KIBatchlage ist, was aus **einem** Ziel eines Batches geworden ist
// (A13.1b, #1452).
export interface KIBatchlage {
  zielkarte: number

  // kennung ist die des Auftrags — leer, wo keiner entstand.
  kennung: string

  modell: string
  felder: KIFeldwertDTO[]

  // befunde sind die Beanstandungen des Schemas — sie stehen **neben** den
  // Feldern und nicht statt ihrer.
  befunde: string[]

  ablauf: string

  // fehler ist der Grund **an diesem Ziel** — null, wo es geklappt hat.
  fehler: FehlerDTO | null
}

// KIBatchAntwort trägt die Lagen aller Ziele.
export interface KIBatchAntwort extends Umschlag {
  kennung: string
  anbieter: string
  ziele: KIBatchlage[]
}

// KIGruppenwahl ist ein Vorschlag mit den Feldern, die gelten sollen
// (A13.1b, #1452).
export interface KIGruppenwahl {
  kennung: string
  felder: string[]
}

// KIGruppenkarte nennt eine Karte und ihre geschriebenen Felder.
export interface KIGruppenkarte {
  karteId: number
  felder: string[]
}

// KIGruppenantwort sagt, was aus der Gruppe wurde.
//
// **Ganz oder gar nicht**: ist `fehler` gesetzt, wurde **keine** Karte
// geschrieben und kein Vorschlag verbraucht — ein zweiter Versuch ist
// möglich.
export interface KIGruppenantwort extends Umschlag {
  karten: number
  felder: number
  namen: KIGruppenkarte[]
}
