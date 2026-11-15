// **Wie dieses Programm ins Netz geht — an der Naht** — Werkstück N5 von
// Phase 5.4
// (#1571).
//
// **Rein lesend.** Geändert wird über `EinstellungSetzen` wie jede andere
// Einstellung; ein zweiter Schreibweg wäre eine zweite Fassung derselben
// Frage.
import type { Umschlag } from './umschlag'

// NetzfeldDTO ist ein wirksamer Wert samt seiner Herkunft.
export interface NetzfeldDTO {
  wert: string

  // herkunft ist „einstellung", „umgebung" oder „ohne".
  herkunft: string
}

// NetzstandAntwort sagt, wie dieses Programm **jetzt** nach außen geht.
export interface NetzstandAntwort extends Umschlag {
  proxyHTTP: NetzfeldDTO
  proxyHTTPS: NetzfeldDTO
  proxyAus: NetzfeldDTO

  // Die drei Fristen in **Sekunden**.
  fristVerbindung: number
  fristKopf: number
  fristStillstand: number

  proxyBenutzt: boolean
}

// Netzapi ist der eine Weg.
export interface Netzwegapi {
  Netzstand: () => Promise<NetzstandAntwort>
}
