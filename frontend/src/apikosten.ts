// **Das Kostenjournal externer Aufträge** — Werkstück O12 von Phase 5.2
// (#1439).
// Die Gestalt kommt aus `internal/app/kostenjournal.go`.
import type { Umschlag } from './umschlag'

// BetragDTO ist Geld in der **kleinsten Einheit** seiner Währung.
//
// Als Zahl und Währung und nicht als Text: „12,50 €" wäre eine Formatierung,
// und wer sie zurückrechnet, rundet.
export interface BetragDTO {
  wert: number
  waehrung: string
}

export interface HaushaltsstandAntwort extends Umschlag {
  konto: string

  // zeitraum ist das Fenster, über das gezählt wurde — die Antwort sagt es
  // selbst, denn wer nichts nennt, bekommt den Tag.
  zeitraum: string

  grenze: BetragDTO
  reserviert: BetragDTO
  verbraucht: BetragDTO

  // unklar zählt wie verbraucht, bis ein Mensch es klärt.
  unklar: BetragDTO

  verfuegbar: BetragDTO
}

export interface KostenpostenDTO {
  kennung: string
  konto: string

  // aufgabe ist eine Marke — kein Prompt.
  aufgabe: string

  hoechstbetrag: BetragDTO
  betrag: BetragDTO

  // stand ist eine Marke — der Satz dazu steht im Sprachkatalog.
  stand: string
  grund: string

  begonnen: string
  beendet: string
}

export interface KostenpostenAntwort extends Umschlag {
  posten: KostenpostenDTO[]
}

// NachweisAntwort ist die erteilte Freigabe eines Menschen.
//
// **Sie gilt einem Auftrag und einer Frist** — wer zwei Aufträge schickt,
// braucht zwei ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)).
export interface NachweisAntwort extends Umschlag {
  kennung: string
  laeuftab: string
}

export interface Kostenapi {
  Kostenstand: (
    konto: string,
    zeitraum: string,
  ) => Promise<HaushaltsstandAntwort>
  Kostenposten: (
    konto: string,
    zeitraum: string,
  ) => Promise<KostenpostenAntwort>
  KostenbudgetSetzen: (
    konto: string,
    zeitraum: string,
    wert: number,
    waehrung: string,
  ) => Promise<Umschlag>

  // **Die Freigabe eines bezahlten KI-Auftrags** (#1539). Sie steht hier und
  // nicht bei den KI-Bindungen, weil sie von Geld handelt: der Höchstbetrag
  // ist derselbe, den das Kostenjournal bucht.
  //
  // **Und es gibt sie nur am Fenster** — `Nachweise.Ausstellen` weist jede
  // andere Fläche ab, denn ein Modell ist kein Mensch.
  KIFreigabeErteilen: (
    konto: string,
    betrag: number,
    waehrung: string,
  ) => Promise<NachweisAntwort>
}
