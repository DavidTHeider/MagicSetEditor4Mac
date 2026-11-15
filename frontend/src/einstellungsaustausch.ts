// **Der Austausch der Einstellungen, wie das Fenster ihn sieht** — Werkstück
// O3 von Phase 5.2
// (#1427).
//
// **Eine eigene Datei**, weil `einstellungen.ts` daneben eine andere Frage
// beantwortet: dort steht, **was** eine Einstellung ist, hier, wie sie
// **reist**.
import type { Umschlag } from './umschlag'

// Einfuhrzeile ist eine Zeile des Unterschieds.
export interface Einfuhrzeile {
  ebene: string
  schluessel: string
  vorher: string
  nachher: string

  // herkunft sagt, wessen Entscheidung der **alte** Wert war — `profil`,
  // `nutzer` oder `vorgabe`.
  //
  // **Der alte und nicht der neue**: die Frage vor einer Übernahme ist
  // „überschreibe ich hier etwas, das jemand gewählt hat?"
  herkunft: string

  // wirkung sagt, wann der neue Wert gilt — `sofort`, `neustart` oder leer
  // für eine Einstellung des Menschen, die keinen Zeitpunkt nennt
  // (OPS-01.AK-02, #1718).
  wirkung: string
}

// Einfuhrbefund ist, was an einer Zeile nicht stimmt.
export interface Einfuhrbefund {
  ebene: string
  schluessel: string
  grund: string
}

// Einstellungssicht ist, was eine Datei ändern würde.
export interface Einstellungssicht {
  // inhalt ist der Text, über den entschieden wird — er geht bei der
  // Übernahme zurück.
  inhalt: string

  fassung: number

  // revision ist die Marke des Stands, gegen den verglichen wurde.
  revision: string

  zeilen: Einfuhrzeile[]
  befunde: Einfuhrbefund[]

  // unveraendert zählt die Zeilen, die schon so dastehen.
  unveraendert: number
}

// Einstellungsbericht ist, was eine Übernahme bewirkt hat.
export interface Einstellungsbericht {
  uebernommen: number
  revision: string
  zeilen: Einfuhrzeile[]
  befunde: Einfuhrbefund[]
}

// EinfuhrvorschauAntwort trägt die Sicht.
export interface EinfuhrvorschauAntwort extends Umschlag {
  sicht: Einstellungssicht
}

// EinfuhrAntwort trägt den Bericht.
export interface EinfuhrAntwort extends Umschlag {
  bericht: Einstellungsbericht
}
