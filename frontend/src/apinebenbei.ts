// **Was neben der Partie herläuft** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Was hier steht
//
// **Reden und Zeigen**, die **Einwilligung** zur späteren Offenlegung und die
// **Diagnose der Leitung**. Dreierlei, und dasselbe ist ihnen gemeinsam: es
// gehört zum Tisch, ohne die Partie zu sein.
//
// **Und `apipartie.ts` stand mit alldem bei 407 Zeilen** (§2). Der Schnitt
// folgt trotzdem der Frage: eine Partie ist ihr Zustand und ihr Strom; was
// hier steht, ist flüchtig, nachträglich oder technisch.

import type { FehlerDTO } from './formen'
import type { Umschlag } from './umschlag'

// Nebenbei ist Chat oder Präsenz, wie das Fenster sie gemeldet bekommt.
//
// **Flach und nicht mit offenem Inhalt**: `art` sagt, welche Felder etwas
// bedeuten. Der erste Anlauf reichte eine offene Abbildung durch — damit stand
// die Bedeutung jedes Feldes in keinem Typ, und der **Sprecher** kam aus dem
// falschen: aus dem Umschlag statt aus der Nutzlast.
export interface Nebenbei {
  // art ist `chat` oder `praesenz`.
  art: string

  // von ist der Sitz, der geredet oder gezeigt hat — **nicht** der, der die
  // Nachricht geschickt hat.
  von: string

  // text gilt bei `chat`.
  text: string

  // zeigerart, objekt, x und y gelten bei `praesenz`. **x und y sind
  // Weltkoordinaten in Millimetern**, keine Bildpunkte: zwei Fenster sind
  // verschieden groß.
  zeigerart: string
  objekt: string
  x: number
  y: number
}

// NebenbeiAntwort sagt, ob ein Satz oder ein Zeiger hinausging.
//
// **Und sonst nichts.** Was aus einem Satz wird, ist, dass ihn jemand liest;
// eine Antwort mit Inhalt behauptete mehr, als diese Naht weiß.
export type NebenbeiAntwort = Umschlag

// Offenlegung ist die Einwilligung **eines** Sitzes, dass seine privaten
// Ergänzungen später auch andere sehen dürfen.
//
// **Später und nicht jetzt**: am laufenden Tisch wirkt sie nicht — wer sein
// Blatt für die Nachlese freigibt, hat nicht zugesagt, es **während** des
// Spiels zu zeigen.
export interface Offenlegung {
  // von ist der Sitz, dessen Ergänzungen es sind.
  von: string

  // an sind die Sitze, die mitlesen dürfen — `*` heißt: jeder.
  an: string[]
}

// PartieoffenlegungAntwort trägt alle Einwilligungen einer Partie.
export interface PartieoffenlegungAntwort extends Umschlag {
  offenlegungen: Offenlegung[]

  // nachladefehler heißt: **der Schritt gilt**, nur die Liste danach war
  // nicht zu lesen (§7, Regel 3; #1970). `offenlegungen` ist dann leer und
  // sagt nichts.
  nachladefehler?: FehlerDTO
}

// Leitungszeile ist eine Leitung dieses Tisches.
//
// **Ohne jede Zahl über Karten** (G5,
// #1462):
// weder der Abdruck des Decks noch der des Stroms steht darin. Beide sind
// Hashes über Karten, und gegen einen Hash lässt sich raten.
export interface Leitungszeile {
  // sitz ist, wer am anderen Ende sitzt — als Gast leer, denn dort ist es
  // der Wirt und kein Sitz.
  sitz: string

  // weg ist der Kanal: die Route.
  weg: string

  // gegenueber ist der Fingerabdruck.
  gegenueber: string

  seitSekunden: number

  wartend: number
  verworfen: number
}

// PartiediagnoseAntwort sagt, wie es um die Leitungen steht.
export interface PartiediagnoseAntwort extends Umschlag {
  laeuft: boolean

  // rolle ist `wirt` oder `gast` — leer, wenn nichts läuft.
  rolle: string

  partie: string
  pausiert: boolean

  // stand ist die Folge, bei der dieses Programm im Strom steht.
  stand: number

  // standBekannt trennt „bei null" von „weiß ich nicht".
  standBekannt: boolean

  rueckkehrMillis: number
  rueckkehrBekannt: boolean

  leitungen: Leitungszeile[]
}
