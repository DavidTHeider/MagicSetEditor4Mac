// **Die Antworten der Skriptkonsole** — Werkstück U13 von Phase 5.1 (#1424).
// Aus `antworten.ts` ausgelagert, weil die Datei an der Zeilengrenze stand (§2).
import type { Skriptstelle, Umschlag } from './api'

// KonsolenEintrag ist ein Eintrag einer Abbildung im Ergebnis der Konsole —
// mit seinen eigenen Einträgen, wenn er selbst eine ist (#1424).
export interface KonsolenEintrag {
  name: string
  quelltext: string
  art: string
  eintraege: KonsolenEintrag[]
}

export interface KonsolenAntwort extends Umschlag {
  // quelltext ist das Ergebnis als **Quelltext** — `"a"` und nicht `a`.
  quelltext: string

  // text ist dasselbe in gewöhnlicher Form.
  text: string

  // art ist die Art des Werts — `text`, `zahl`, `abbildung`.
  art: string

  // dauerMs ist, wie lange die Auswertung gedauert hat — Werkstück U13 von
  // Phase 5.1 (#1424).
  dauerMs: number

  // abgebrochen sagt, dass der Lauf beendet wurde, bevor ein Ergebnis da war.
  // **Kein Fehler**: es ist, was jemand wollte.
  abgebrochen: boolean

  // eintraege sind die Einträge einer Abbildung, rekursiv — leer für alles,
  // was keine ist. Das Fenster zeigt sie als Baum.
  eintraege: KonsolenEintrag[]

  // befunde sind die Vorbehalte beim Aufbau der Umgebung. Sie halten die
  // Auswertung **nicht** auf (§8).
  befunde: string[]

  // stelle ist der Ort eines Syntaxfehlers — **null**, wenn er keinen hat
  // (#1164).
  //
  // Eine Karte, die es nicht gibt, hat keine Zeile; dann wird nichts
  // angestrichen.
  stelle: Skriptstelle | null
}
