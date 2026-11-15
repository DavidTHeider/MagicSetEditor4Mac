// **Wohin eine Karte außerhalb dieses Programms führt** — Werkstück T8 von
// Phase 5.6
// (#1473).
//
// **Zwei Wege und nicht einer**: die Liste entsteht, ohne dass etwas
// geschieht; das Öffnen verlässt dieses Programm — und geschieht nur auf
// einen Klick.
import type { Umschlag } from './umschlag'

// dieGenaueArt ist der Weg, der auf **diese** Karte führt.
export const dieGenaueArt = 'produkt'

// dieStartseite ist der Weg, der nur zum Anbieter führt.
export const dieStartseite = 'startseite'

// Marktlink ist ein Weg nach draußen.
export interface Marktlink {
  // anbieter ist `cardmarket`, `cardtrader`, `tcgplayer`, `cardkingdom`,
  // `gatherer` oder `mtggoldfish`.
  anbieter: string

  // art ist `produkt`, `suche` oder `startseite` — **die Beschriftung hängt
  // daran**: ein Suchlink verspricht nicht, was ein Produktlink hält.
  art: string

  // herkunft ist `kennung`, `name` oder `keine`.
  herkunft: string

  adresse: string
}

// Marktlinkantwort trägt die Wege samt ihrem Fehler.
export interface Marktlinkantwort extends Umschlag {
  links: Marktlink[]
}

// Schreibantwort ist, was ein Öffnen zurückgibt.
export interface Oeffnungsantwort extends Umschlag {
  geschrieben?: boolean
}

// Marktlinkapi sind die zwei Wege.
export interface Marktlinkapi {
  Marktlinks: (printing: string, name: string) => Promise<Marktlinkantwort>

  // MarktlinkOeffnen macht einen geprüften Weg im Standardbrowser auf.
  //
  // **Der Anbieter kommt mit**, denn ohne ihn wüsste die Prüfung nicht,
  // gegen welchen Wirt sie prüft.
  MarktlinkOeffnen: (
    anbieter: string,
    adresse: string,
  ) => Promise<Oeffnungsantwort>
}
