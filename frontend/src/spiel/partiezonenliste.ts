// **Welche Zonen der Tisch zeigt** — PLAY-03.AK-01
// (#1711).
//
// # Was das Kriterium verlangt
//
// **„entsprechend aktivem Spielmodus"** — stünden alle sieben immer da, in
// der Zonenliste wie auf der Leinwand, wäre die Kommandozone an einem
// gewöhnlichen Zweiertisch ein leeres Feld je Sitz, das nie etwas trägt.
//
// # Was das nicht heißt
//
// **Nicht, dass es die Zone nicht gibt.** CR 408.1 kennt sie in jeder
// Partie; was hier entschieden wird, ist, ob ein **Feld dafür gezeichnet**
// wird. Das ist eine Frage der Oberfläche, und deshalb steht sie hier und
// nicht in der Domäne.
//
// **Und nicht, dass sie leer bleiben darf.** Ein Emblem entsteht in der
// Kommandozone (CR 114.2), auch im Sandkasten. Eine Zone auszublenden, in der
// etwas liegt, hieße eine Karte zu verstecken (§8) — deshalb die zweite
// Bedingung.
import type { Partie } from '../apipartie'

// Commanderkennung ist die Kennung der Variante — sie steht ebenso in
// `internal/domain/commander.go`.
//
// **Zwei Stellen, und das ist hier richtig so**: die Go-Seite entscheidet,
// was eine Partie **ist**; diese Zeile entscheidet, was das Fenster
// **zeichnet**. Ein Vertragstest hielte zwei Zeichenketten zusammen, die
// nichts miteinander zu tun haben — die Naht reicht die Kennung durch, und
// eine unbekannte führt zum Normalfall.
const Commanderkennung = 'commander'

// Grundzonen sind die sechs, die jede Partie hat — CR 400.1.
const Grundzonen = [
  'hand',
  'tisch',
  'stapel',
  'bibliothek',
  'friedhof',
  'exil',
] as const

// Kommandozone ist die siebte — CR 408.
const Kommandozone = 'kommando'

// zonenDerPartie gibt die Zonen, für die ein Feld gezeichnet wird.
//
// **Die Reihenfolge ist die des Blicks** und nicht die der Aufzählung in
// `domain`: von der Hand zum Tisch und dann in die Ablagen. Das Kommando
// steht zuletzt, weil es das seltenste ist.
export function zonenDerPartie(partie: Partie): string[] {
  const zonen: string[] = [...Grundzonen]

  const gebraucht =
    partie.regelwerk.kennung === Commanderkennung ||
    partie.objekte.some((eines) => eines.zone === Kommandozone)

  if (gebraucht) {
    zonen.push(Kommandozone)
  }

  return zonen
}
