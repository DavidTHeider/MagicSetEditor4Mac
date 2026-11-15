// **Die Zugriffstoken eines Profils** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13.
//
// **Eine eigene Datei**, weil `apinutzer.ts` mit ihnen über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage — derselbe wie auf der Go-Seite: dort
// steht, **wer** hier arbeitet, hier, **womit sich ein Automat ausweist**.
import type { Umschlag } from './umschlag'

// Tokenapi sind die vier Bindungen dazu.
//
// **Verwaltet wird an jeder Fläche, geprüft nur am REST-Zugang**: er ist
// die einzige, die auf einem Port lauscht; das Fenster weiß, wer
// davorsitzt.
export interface Tokenapi {
  Token: (profilID: string) => Promise<TokenAntwort>

  // TokenAnlegen legt eines an — **der Klartext steht nur in dieser
  // Antwort**.
  TokenAnlegen: (angaben: TokenangabenDTO) => Promise<EinTokenAntwort>

  // TokenAendern setzt Bezeichnung, Erlaubnis und Ablauf; das Geheimnis
  // bleibt dasselbe.
  TokenAendern: (angaben: TokenangabenDTO) => Promise<EinTokenAntwort>

  // TokenLoeschen zieht eines zurück.
  TokenLoeschen: (tokenID: string) => Promise<Umschlag>

  // **Rotation, Sammelwiderruf und Ablaufwahlen** — Werkstück O2 von Phase
  // 5.2 (#1426).
  // Die Rotation gibt den neuen Klartext **genau einmal**, wie das Anlegen.
  TokenRotieren: (tokenID: string) => Promise<EinTokenAntwort>
  TokenAlleWiderrufen: (profilID: string) => Promise<TokenwiderrufAntwort>
  Tokenablaufwahlen: () => Promise<AblaufwahlenAntwort>
}

// AblaufwahlDTO ist ein benannter Vorschlag für den Ablauf — gerechnet in
// Go, damit alle vier Flächen dieselben Zeitpunkte nennen.
export interface AblaufwahlDTO {
  name: string
  ablauf: string
}

export interface AblaufwahlenAntwort extends Umschlag {
  wahlen: AblaufwahlDTO[]
}

// TokenwiderrufAntwort sagt, wie viele Token ein Sammelwiderruf gezogen hat.
export interface TokenwiderrufAntwort extends Umschlag {
  gezogen: number
}

// TokenDTO ist ein Zugriffstoken, wie das Fenster es sieht — **ohne sein
// Geheimnis**.
export interface TokenDTO {
  id: string
  profilID: string

  bezeichnung: string

  // principal sagt, für wen das Token arbeitet — Werkstück O2 (#1426).
  principal: string

  // scope ist `lesen` oder `schreiben`; `schreiben` schließt Lesen ein.
  scope: string

  // ablauf ist ein Zeitpunkt in RFC 3339 — leer heißt: läuft nicht ab.
  ablauf: string
  angelegt: string

  // abgelaufen sagt, ob der Ablauf **jetzt** überschritten ist.
  //
  // **Ausgerechnet und nicht hier**: sonst rechnete jede der vier Flächen
  // dieselbe Zeitspanne aus, und die Ergebnisse liefen an der Zeitzone
  // auseinander.
  abgelaufen: boolean

  // letzteNutzung ist RFC 3339 — leer heißt: noch nie angenommen.
  letzteNutzung: string

  // vorgaengerBis ist die Frist, bis zu der das vorige Geheimnis nach einer
  // Rotation noch gilt — leer heißt: es gibt keines.
  vorgaengerBis: string

  // partie und sitz nennen die Bindung an einen Sitz (#1819) — leer heißt:
  // ungebunden.
  partie: string
  sitz: string

  fassung: string
}

// TokenangabenDTO ist, was das Fenster an einem Token setzen darf.
export interface TokenangabenDTO {
  // id ist beim Anlegen leer und beim Ändern die des Tokens.
  id: string

  // profilID ist beim Anlegen nötig — **ein Token wechselt sein Profil nicht**.
  profilID: string

  bezeichnung: string
  principal: string
  scope: string
  ablauf: string

  // partie und sitz binden ein neues Token an einen Sitz (#1819); das
  // Fenster legt keine gebundenen an und lässt beide leer.
  partie: string
  sitz: string

  fassung: string
}

// TokenAntwort trägt die Token eines Profils.
export interface TokenAntwort extends Umschlag {
  token: TokenDTO[]
}

// EinTokenAntwort trägt ein einzelnes Token.
export interface EinTokenAntwort extends Umschlag {
  token: TokenDTO | null

  // klartext ist **nur beim Anlegen** gefüllt.
  klartext: string
}
