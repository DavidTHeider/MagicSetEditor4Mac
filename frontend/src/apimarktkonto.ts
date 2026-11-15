// **Das eigene Konto bei einem Marktplatz** — Werkstück T11 von Phase 5.6
// (#1476).
//
// **Ein Weg und nur einer**: gelesen wird, geschrieben nicht. Angebot
// anlegen, Warenkorb und Kauf kennt der Anbieter; gebaut ist davon nichts.
import type { Umschlag } from './umschlag'

// derMarktplatz ist der eine, den es heute gibt.
export const derMarktplatz = 'cardtrader'

// Eigenangebot ist ein Posten, den das eigene Konto anbietet.
export interface Eigenangebot {
  id: number
  blueprintID: number
  name: string
  menge: number

  // preisCent und waehrung gehören zusammen — beides oder keines.
  preisCent: number
  waehrung: string

  // eigennotiz ist der Faden zum eigenen Lager.
  eigennotiz: string
  beschreibung: string
}

// Marktbestellung ist ein Kauf, an dem das eigene Konto beteiligt ist.
export interface Marktbestellung {
  id: number
  code: string

  // zustand ist einer der acht des Anbieters — **unübersetzt**, damit ein
  // Vergleich ein Vergleich bleibt. Was er auf Deutsch heißt, entscheidet
  // die Anzeige.
  zustand: string

  ueberNullkonto: boolean
  rolle: string
  stueck: number
  bezahlt: string

  // bewegtBestand sagt, ob diese Bestellung nach der Regel des Anbieters ein
  // Stück aus dem eigenen Lager genommen hat — **eine Auskunft und keine
  // Buchung**.
  bewegtBestand: boolean
}

// Kontostand ist die Übersicht über das eigene Konto.
export interface Kontostand {
  name: string
  nutzerID: number
  angebote: Eigenangebot[]
  bestellungen: Marktbestellung[]
  befunde: string[]
}

// Kontostandantwort trägt die Übersicht samt ihrem Fehler.
export interface Kontostandantwort extends Umschlag {
  stand: Kontostand | null
}

// Marktkontoapi ist der eine Weg.
export interface Marktkontoapi {
  // Marktkonto gibt Konto, eigene Angebote und Bestellungen in einem Zug.
  //
  // **`seite` und `menge` sind die Blätterangaben des Anbieters**: null heißt
  // „nimm seine Vorgabe" und nicht „nimm alles".
  Marktkonto: (
    quelle: string,
    seite: number,
    menge: number,
  ) => Promise<Kontostandantwort>
}
