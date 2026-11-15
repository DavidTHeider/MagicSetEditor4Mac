// **Die zwei Dienste an der Naht** — Werkstück N6 von Phase 5.4
// (#1572).
//
// **Drei Wege je Dienst und kein Schalter**: Anfangen und Aufhören sind zwei
// Entscheidungen, der Stand ist eine Frage. Ein Wahrheitswert nähme die
// Adresse auch beim Ausschalten entgegen.
import type { Umschlag } from './umschlag'

// DienstZustand ist, was eine Fläche über einen Posten erfährt.
export interface DienstZustand {
  art: string
  laeuft: boolean

  // adresse ist die **wirkliche** Adresse samt Schema — leer, solange nichts
  // läuft.
  adresse: string

  // tls sagt, ob der Lauscher verschlüsselt ist. Jenseits der Schleife immer.
  tls: boolean

  // marke ist der Zugang, den dieser Posten verlangt — eine je Öffnung.
  marke: string

  // namen und adressen sind, was ein Gegenüber eintippen kann.
  namen: string[]
  adressen: string[]

  fingerabdruck: string

  // zeugnisAussteller und zeugnisBis nennt nur ein eigenes Zeugnis (NET-03,
  // #1707) — leer beim Zeugnis auf der Peeridentität. zeugnisBis in RFC 3339.
  zeugnisAussteller: string
  zeugnisBis: string

  // lauscher sind alle Adressen, auf denen der Posten lauscht — die erste
  // ist `adresse`; fehlende die gewählten, auf denen keiner entstand, mit
  // dem Grund (G08, #2272).
  lauscher: string[]
  fehlende: Lauschbefund[]
}

// Lauschbefund ist eine gewählte Adresse ohne Lauscher.
export interface Lauschbefund {
  adresse: string
  grund: string
}

// DienstAntwort ist der Ausgang eines Dienstwegs.
export interface DienstAntwort extends Umschlag {
  stand: DienstZustand
}

// Diensteapi sind die sechs Wege.
export interface Diensteapi {
  RESTStarten: (adresse: string, port: number) => Promise<DienstAntwort>
  RESTBeenden: () => Promise<DienstAntwort>
  RESTStand: () => Promise<DienstAntwort>

  MCPStarten: (adresse: string, port: number) => Promise<DienstAntwort>
  MCPBeenden: () => Promise<DienstAntwort>
  MCPStand: () => Promise<DienstAntwort>
}
