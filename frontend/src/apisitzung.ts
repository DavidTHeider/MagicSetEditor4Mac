// **Die lokale Sitzung** — ID-02
// (#1701).
//
// Wann gesperrt, entsperrt und abgemeldet werden darf, entscheidet Go
// (`service.Sitzung`); hier steht nur die Gestalt der Antworten.

import type { SchreibAntwort } from './api'
import type { Umschlag } from './umschlag'

export type Sitzungszustand = 'angemeldet' | 'gesperrt' | 'abgemeldet'

// SitzungsstandAntwort sagt, wie es um die Sitzung steht und was der
// Sperrbildschirm anbieten kann.
export interface SitzungsstandAntwort extends Umschlag {
  zustand: Sitzungszustand
  nutzerID: string

  // kennwort sagt, ob der Nutzer eines hat; sensor, ob er auf diesem Rechner
  // **lokal entsperren** kann — Touch ID, kein Passkey.
  kennwort: boolean
  sensor: boolean
}

// Sitzungsapi sind die Wege der Sitzung — **nur am Fenster**.
export interface Sitzungsapi {
  Sitzungsstand: () => Promise<SitzungsstandAntwort>
  Sperren: () => Promise<SchreibAntwort>
  Entsperren: (kennwort: string) => Promise<SchreibAntwort>
  EntsperrenMitSensor: () => Promise<SchreibAntwort>
  Abmelden: () => Promise<SchreibAntwort>

  // LokalAnmelden prüft das Kennwort und dann den zweiten Faktor des Ziels.
  LokalAnmelden: (
    nutzerID: string,
    kennwort: string,
    nachweis: string,
  ) => Promise<SchreibAntwort>
  KennwortWechseln: (alt: string, neu: string) => Promise<SchreibAntwort>

  // KennwortWiederherstellen nimmt einen Code vom Zettel und setzt nur das
  // Kennwort — keine Rolle, keine Freigabe, kein Gegenüber.
  KennwortWiederherstellen: (
    nutzerID: string,
    code: string,
    neu: string,
  ) => Promise<SchreibAntwort>
}
