// **Die digitale Deckleihe an der Naht** — Werkstück G6 von Phase 5.5
// (#1463).
//
// **Eine eigene Datei und nicht in `apipartie.ts`.** Eine Leihe ist keine
// Partie: sie entsteht davor, überlebt sie und gilt auch, wenn keine läuft.
// Der Schnitt folgt der Frage (§2) — und `apipartie.ts` stand mit ihr bei 450
// Zeilen.

import type { FehlerDTO } from './formen'
import type { Umschlag } from './umschlag'

// Leihe ist eine digitale Deckleihe, wie das Fenster sie sieht.
//
// **Eine Nutzungsfreigabe und keine Eigentumsübertragung** (G6,
// #1463):
// wer verleiht, behält sein Deck; wer leiht, darf damit **antreten**. Der
// physische Bestand ändert sich um kein Exemplar.
export interface Leihe {
  kennung: string

  // verleiher und entleiher sind **Peerkennungen** — kein lokales Profil
  // geht über die Grenze.
  verleiher: string
  entleiher: string

  // abdruck sagt, dass zwei Seiten denselben Abzug meinen — **ohne ihn zu
  // zeigen**.
  abdruck: string

  // bindungsart ist `partie` oder `turnier`, woran die Kennung — leer heißt:
  // die nächste dieser Art.
  bindungsart: string
  woran: string

  // rechte sind `spielen` und `einreichen`.
  rechte: string[]

  ausgestellt: string
  bis: string

  // modus sagt, was mit einer laufenden Partie geschieht: `zuende` oder
  // `abbruch`.
  modus: string

  // verliehen sagt, ob **dieses** Profil sie ausgestellt hat.
  verliehen: boolean

  // verbraucht ist die Partie, an der sie eingelöst wurde — leer heißt: noch
  // nicht.
  verbraucht: string

  widerrufen: boolean

  // karten ist die Zahl der Karten im Abzug.
  karten: number
}

// Leihwunsch ist, was jemand verleihen will.
export interface Leihwunsch {
  entleiher: string
  deckID: string
  bindungsart: string
  woran: string
  rechte: string[]

  // stunden ist die Dauer ab jetzt.
  stunden: number

  modus: string
}

// DeckleiheAntwort trägt die Leihen dieses Profils.
export interface DeckleiheAntwort extends Umschlag {
  leihen: Leihe[]

  // pfad ist der abgelegte Umschlag — nur beim Anbieten.
  pfad: string

  // nachladefehler heißt: **der Schritt ist geschrieben**, nur die Liste
  // danach war nicht zu lesen (§7, Regel 3; #1960). `leihen` ist dann leer
  // und sagt nichts — die Seite zeigt den Grund an der Liste.
  nachladefehler?: FehlerDTO
}

// Leihziel ist der Ort, den jemand für den Umschlag einer Leihe gewählt hat
// (UXR-F15, #2030).
//
// **Der Dialog schreibt nichts**: erst `DeckleiheAnbieten` legt den Umschlag
// dort ab, wo das Fenster den Pfad bis dahin zeigt.
export interface Leihziel extends Umschlag {
  pfad: string

  // abgebrochen heißt: niemand hat einen Ort gewählt. **Kein Fehler** — das
  // Fenster lässt den vorigen stehen.
  abgebrochen: boolean
}
