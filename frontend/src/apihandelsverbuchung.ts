// **Die Verbuchung eines Handels** — Werkstück T6 von Phase 5.6
// (#1471).
//
// **Drei Rufe und nicht einer**: Verbuchen bewegt den Bestand, Abgleichen
// meldet der Gegenseite, dass hier fertig ist, Bestätigen nimmt ihre Stimme
// entgegen. Ein Ruf, der alles täte, müsste die zweite Unterschrift erfinden.
import type { Umschlag } from './umschlag'

// derOffeneAbgleich ist der Zustand, in dem die Gegenseite noch fehlt.
export const derOffeneAbgleich = 'abgleichOffen'

// derBeidseitigeAbschluss ist der Zustand, in dem beide unterschrieben haben.
export const derBeidseitigeAbschluss = 'beidseitigVerbucht'

// Handelsbeleg ist das Papier zu einem Abschluss.
export interface Handelsbeleg {
  vorgang: string

  // abdruck ist der Stand, der abgeschlossen wurde.
  abdruck: string

  unterschrift: string

  // beidseitig sagt, ob auch die Gegenseite unterschrieben hat — und es
  // sagt es erst, seit es stimmt (#1650).
  beidseitig: boolean

  // gegenzeichner ist die Peerkennung dessen, dessen geprüfter Beleg hier
  // liegt — leer, solange keiner da ist.
  gegenzeichner: string

  // gegengezeichnet ist der Augenblick, in dem er ankam — leer, solange
  // keiner da ist.
  gegengezeichnet: string

  gebucht: string

  // befund ist der jüngste abgewiesene Gegenbeleg — null, solange keiner
  // abgewiesen wurde (#2500).
  befund: GegenbelegbefundDTO | null
}

// GegenbelegbefundDTO sagt, dass ein Gegenbeleg nicht passte.
export interface GegenbelegbefundDTO {
  von: string
  abdruck: string

  // mangel ist eines aus `domain.DieGegenbelegmaengel` — ein unbekanntes
  // wird roh genannt (§16).
  mangel: string

  // grund ist der Satz der Prüfung, deutsch — für Protokoll und Skript.
  grund: string

  empfangen: string
}

// Journalzeile ist eine Bewegung im eigenen Bestand.
export interface Journalzeile {
  id: number
  art: string
  anlass: string
  vorgang: string
  bezug: number
  grund: string
  printing: string

  // bezeichnung ist, wie die Karte heißt — beim Lesen nachgeschlagen und
  // **nicht** im Journal gespeichert: ein eingefrorener Name wäre falsch,
  // sobald jemand den Druck berichtigt.
  bezeichnung: string

  menge: number
  zustand: string
  sprache: string
  finish: string
  exemplarId: number
  lagerort: string
  notiz: string
  gebucht: string
}

// Verbuchungsantwort trägt den Stand samt seinem Beleg.
export interface Verbuchungsantwort extends Umschlag {
  vorgang: string
  zustand: string
  beleg: Handelsbeleg
  hatBeleg: boolean
}

// Journalantwort trägt die Bewegungen samt ihrem Fehler.
export interface Journalantwort extends Umschlag {
  vorgang: string
  zeilen: Journalzeile[]
}

// Freigabeantwort ist die erteilte Zustimmung eines Menschen (T10, #1475).
export interface Freigabeantwort extends Umschlag {
  // kennung ist, was der nächste Schritt vorzeigt — leer bei einem Fehler.
  kennung: string

  // laeuftab sagt, bis wann sie gilt.
  laeuftab: string
}

// derVerbuchungsschritt und derAbschlussschritt sind die zwei Schritte, für
// die es eine Freigabe gibt.
//
// **Das Abgleichen ist keiner**: es bindet nichts.
export const derVerbuchungsschritt = 'verbuchen'
export const derAbschlussschritt = 'bestaetigen'

// Handelsverbuchungapi sind die sechs Wege der Verbuchung.
export interface Handelsverbuchungapi {
  // HandelFreigabeErteilen stellt die Zustimmung für **einen** Schritt aus —
  // gebunden an diesen Vorgang in seinem **jetzigen** Stand (T10, #1475).
  //
  // **Sie gilt fünf Minuten und einmal.** Ändert sich der Vorgang dazwischen,
  // gilt sie nicht mehr: niemand hat den neuen Stand gesehen.
  // **`wer` sagt, wessen Unterschrift sie setzen darf** — leer für
  // `verbuchen` und `abgleichen`, die keine Stimme setzen (Befund 5 des
  // Zweitblicks über Phase 5.6).
  HandelFreigabeErteilen: (
    vorgang: string,
    schritt: string,
    wer: string,
  ) => Promise<Freigabeantwort>

  Handelsverbuchung: (
    vorgang: string,
    freigabe: string,
  ) => Promise<Verbuchungsantwort>
  Handelsabgleich: (vorgang: string) => Promise<Verbuchungsantwort>
  // Handelsbestaetigung setzt die Stimme von `wer` — in dieser Reihenfolge,
  // wie die Bindung in Go (`vorgang, wer, nachweis`). Hier standen die Namen
  // `freigabe, wer` vertauscht, gerufen wurde richtig (#2026).
  Handelsbestaetigung: (
    vorgang: string,
    wer: string,
    freigabe: string,
  ) => Promise<Verbuchungsantwort>
  Handelsjournal: (vorgang: string) => Promise<Journalantwort>
  // Handelsbeleg liest Stand und Beleg, **ohne** einen Schritt zu gehen
  // (Hüllenprobe `nat04handel`, #1672).
  Handelsbeleg: (vorgang: string) => Promise<Verbuchungsantwort>
}
