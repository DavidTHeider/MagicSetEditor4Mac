// **Der portable Umschlag an der Naht** — Werkstück N4 von Phase 5.4
// (#1457).
//
// # Warum drei Wege und nicht zwei
//
// **Weil Ansehen und Übernehmen getrennt sind.** Ein Ansehen, das schon
// schriebe, machte die Freigabe zu einer Frage nach etwas, das schon geschehen
// ist — und #1457 verlangt sie: „Import zeigt Objekte/Abhängigkeiten/Konflikte
// und benötigt fachliche Übernahmefreigabe."
import type { AnlegeAntwort, OeffnenAntwort } from './antworten'
import type { PaketAntwort } from './pakete'
import type { DeckAntwort } from './apidecks'
import type { Sichtvorschauantwort } from './apisichtfreigabe'
import type { Umschlag } from './umschlag'

// UmschlagobjektDTO ist ein Stück Inhalt eines Umschlags.
export interface UmschlagobjektDTO {
  // art ist „set", „karte", „vorlage", „deck" oder „sammlungssicht".
  art: string

  // bezeichnung ist, was ein Mensch liest — **kein Nachweis**.
  bezeichnung: string

  groesse: number

  // braucht nennt, was dieses Objekt zum Vollständigsein braucht.
  braucht: string[]
}

// UmschlagAntwort ist, was in einem Umschlag steht.
export interface UmschlagAntwort extends Umschlag {
  objekte: UmschlagobjektDTO[]

  // urheber ist die Peerkennung dessen, der unterschrieben hat.
  urheber: string

  // fingerabdruck ist die Form, die zwei Menschen vergleichen — und der
  // einzige Nachweis, der zählt.
  fingerabdruck: string

  // pfad ist, wo der angesehene Umschlag liegt — leer beim Schnüren.
  //
  // **Er kommt zurück, weil das Übernehmen ihn braucht**: gewählt hat ihn
  // der native Dialog, und das Fenster kennt ihn sonst nicht.
  pfad: string

  // abgebrochen heißt: der Dialog wurde weggeklickt — **kein Fehler**.
  abgebrochen: boolean
}

// Umschlagapi sind die drei Wege.
export interface Umschlagapi {
  // UmschlagSchnueren fragt selbst nach dem Ort — wie `SetSpeichernUnter`.
  //
  // **Der Pfad kommt nicht von hier.** Wo etwas hingelegt wird, entscheidet
  // ein Mensch im nativen Dialog; ein hier zusammengesetzter Pfad wäre eine
  // zweite Fassung derselben Frage.
  // **Zwei Angaben seit T7** (#1472): `art` ist `set`, `karte`, `vorlage`,
  // `deck` oder `sammlungssicht` — leer heißt `set` —, und was `kennung`
  // bedeutet, entscheidet die Art.
  UmschlagSchnueren: (art: string, kennung: string) => Promise<UmschlagAntwort>

  // UmschlagWaehlen fragt nach einem Umschlag und sagt, was drin steht —
  // **ohne** etwas in den Bestand zu schreiben.
  UmschlagWaehlen: () => Promise<UmschlagAntwort>

  UmschlagUebernehmen: (pfad: string) => Promise<OeffnenAntwort>

  // UmschlagDeck nimmt ein Deck aus einem Umschlag an — als **neues** Deck.
  UmschlagDeck: () => Promise<DeckAntwort>

  // UmschlagSicht liest eine empfangene Sammlungssicht — **ohne** sie in die
  // eigene Sammlung zu schreiben.
  //
  // **Ein Empfang ist keine Buchung.** Sie sagt, was ein Gegenüber hat; in
  // den eigenen Bestand geht sie nie (T7, #1472).
  UmschlagSicht: () => Promise<Sichtvorschauantwort>

  // UmschlagKarte nimmt eine Karte aus einem Umschlag in ein **gewähltes**
  // Set — als **neue** Karte
  // (#1639).
  //
  // **Die Kennung aus dem Umschlag ist hier keine**: sie stammt aus einem
  // fremden Bestand, und wer sie übernähme, schriebe über eine Karte, von
  // der der Absender nichts weiß.
  UmschlagKarte: (setID: string) => Promise<AnlegeAntwort>

  // UmschlagVorlage nimmt eine Vorlage als **gewähltes** Paket an — der Name
  // trägt seine Endung, etwa `meinspiel.mse-game` (#1639).
  //
  // **Ein vorhandenes Paket ist ein Konflikt und kein Überschreiben.**
  UmschlagVorlage: (name: string) => Promise<PaketAntwort>
}
