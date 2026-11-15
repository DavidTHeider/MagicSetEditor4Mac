// **Womit sich dieses Profil gegenüber anderen ausweist** — Werkstück N1 von
// Phase 5.4
// (#1454).
//
// **Eine eigene Datei**, weil sie eine eigene Frage beantwortet: nicht was
// dieses Programm **kann**, sondern wer es **ist**.
import type { Umschlag } from './umschlag'

// Peeridentitaetsantwort ist, was über die eigene Identität feststeht.
export interface Peeridentitaetsantwort extends Umschlag {
  // stand ist einer der vier: `ohne`, `bereit`, `verloren`, `fremd`.
  //
  // **Drei davon sind kein Fehler** — jeder verlangt eine andere Handlung,
  // und keiner darf still behoben werden.
  stand: string

  // kennung ist die öffentliche Peerkennung — leer heißt: es gibt keine.
  kennung: string

  // fingerabdruck ist die Form, die zwei Menschen vergleichen.
  //
  // **Er kommt fertig aus Go.** Eine Seite, die ihn selbst rechnete, könnte
  // ihn vergessen — und zwei Seiten, die ihn beide rechnen, laufen beim
  // ersten Sonderfall auseinander.
  fingerabdruck: string

  // seit ist der Zeitpunkt, an dem diese Kennung entstanden ist.
  seit: string

  // vorher ist die Kennung vor der letzten Ablösung — leer, wenn es keine
  // gab.
  vorher: string

  // vorherFingerabdruck ist ihr Abdruck — **das** hat ein Kontakt bestätigt.
  vorherFingerabdruck: string

  // befunde erklären, was geschehen ist — leer heißt: nichts Besonderes.
  befunde: string[]
}

// Peerapi sind die drei Wege zur eigenen Identität.
//
// **Drei und nicht einer**: fragen, anlegen, ablösen. Ein Ruf, der alles
// dreies je nach Lage täte, legte still eine Identität an — und genau das
// schließt N1 aus.
export interface Peerapi {
  Peeridentitaet: () => Promise<Peeridentitaetsantwort>
  PeeridentitaetSicherstellen: () => Promise<Peeridentitaetsantwort>
  PeeridentitaetRotieren: () => Promise<Peeridentitaetsantwort>
}
