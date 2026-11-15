// **Die Recherche** — Werkstück A16 von Phase 5.3
// (#1515).
// Die Gestalt kommt aus `internal/app/recherche.go`.
import type { KorpusstandDTO } from './apiindex'
import type { Umschlag } from './umschlag'

// RecherchefilterDTO ist die geprüfte Einschränkung — **kein Suchtext**.
export interface RecherchefilterDTO {
  // sprache schränkt auf eine ein; leer heißt: jede.
  sprache: string

  // korpusstand verlangt einen bestimmten Stand; leer heißt: der aktive.
  korpusstand: string

  // freigegeben nennt die **privaten** Korpora, die ein Mensch ausdrücklich
  // ausgewählt hat — Werkstück A13.4 von Phase 5.3
  // (#1452).
  //
  // **Leer heißt: keine.** Der eigene Bestand bleibt dann außen, und die
  // Antwort sagt es unter `einschraenkungen`. Eine Freigabe ohne Wirkung
  // wird abgewiesen und nicht übergangen.
  freigegeben: string[]
}

export interface RecherchetrefferDTO {
  rang: number
  kennung: string
  quelle: string

  // kennungssystem, fachkennung und semantik sind die Identität nach A2 —
  // **kein Name als Identität**.
  kennungssystem: string
  fachkennung: string
  semantik: string

  // eigeneKarteID ist gesetzt, wo der Treffer eine Karte aus diesem Bestand
  // ist — dann führt der Weg dorthin und nicht nach außen.
  eigeneKarteID: number

  titel: string
  ausschnitt: string
  text: string

  sprache: string
  regelnummer: string
  urheber: string

  abschnitt: number
  von: number
  bis: number

  bezugsadresse: string

  // quellrevision, quelldatum und abgerufen sind die Herkunft — drei
  // Angaben, weil das Abrufdatum kein Gültigkeitsdatum ist.
  quellrevision: string
  quelldatum: string
  abgerufen: string

  // weitere nennt die gebündelten Abschnitte desselben Konzepts.
  weitere: string[]
}

export interface RechercheAntwort extends Umschlag {
  treffer: RecherchetrefferDTO[]

  // modus ist `hybrid`, `lexikalisch` oder `ohne-index`.
  modus: string

  // leere sagt, **warum** nichts kam — leer, wo etwas kam.
  leere: string
  grund: string

  modell: string
  modellrevision: string
  suchkonfiguration: string

  // indexfingerabdruck bindet die Liste, raumabdruck den Vektorraum —
  // zwei Fragen, zwei Abdrücke.
  indexfingerabdruck: string
  raumabdruck: string

  // ausgangsstand ist der Stand der Karte, zu der Ähnliches gesucht wurde —
  // die Antwort nennt ihn, weil ihn vorher niemand kennt. Wer die nächste
  // Seite holt, gibt ihn zurück.
  ausgangsstand: string

  korpusstaende: KorpusstandDTO[]

  // einschraenkungen nennt, was **nicht** getan wurde.
  einschraenkungen: string[]

  fortsetzung: string
}

export interface BelegDTO {
  kennung: string
  quellkennung: string
  quelle: string

  kennungssystem: string
  fachkennung: string
  semantik: string

  titel: string

  // ausschnitt ist **der Text, der wirklich mitgegeben wurde**.
  ausschnitt: string
  inhaltshash: string

  sprache: string
  regelnummer: string
  urheber: string

  bezugsadresse: string

  quellrevision: string
  quelldatum: string
  abgerufen: string

  veraltet: boolean
}

export interface KontextfreigabeDTO {
  zielanbieter: string
  zielmodell: string

  // ort sagt, **wo** gerechnet wird.
  ort: string

  // zieladresse sagt, wohin es ginge — leer heißt: die Vorgabe des
  // Anbieters. Hinter einem Anbieternamen kann ein Gateway stehen, und das
  // ist ein Datenexport an einen Dritten.
  zieladresse: string

  kategorien: string[]

  belege: number
  zeichen: number
  tokens: number

  // privat sagt, ob eigener Inhalt dabei ist.
  privat: boolean

  // satz ist die Zusammenfassung — an einer Stelle gebildet.
  satz: string
}

export interface KontextpaketAntwort extends Umschlag {
  paketID: string

  aufgabe: string
  zielrevision: string

  // zaehlweise nennt, wie Tokens geschätzt wurden.
  zaehlweise: string

  kontextgrenze: number
  ausgabereserve: number
  verbraucht: number

  kontextpolicy: string

  korpusstaende: KorpusstandDTO[]
  belege: BelegDTO[]

  freigabe: KontextfreigabeDTO

  laeuftab: string

  // ergebnis ist die Recherche, aus der die Belege stammen.
  ergebnis: RechercheAntwort
}

export interface Rechercheapi {
  Recherchieren: (
    fall: string,
    frage: string,
    grenze: number,
    filter: RecherchefilterDTO,
    fortsetzung: string,
  ) => Promise<RechercheAntwort>
  AehnlicheKarten: (
    karteID: number,
    stand: string,
    grenze: number,
    filter: RecherchefilterDTO,
    fortsetzung: string,
  ) => Promise<RechercheAntwort>
  KontextZusammenstellen: (
    aufgabe: string,
    zielrevision: string,
    fall: string,
    frage: string,
    grenze: number,
    anbieter: string,
    modell: string,
    fenster: number,
    reserve: number,
    hoechstens: number,

    // freigegeben nennt die **privaten** Korpora, die mitgesucht werden
    // dürfen — A13.4 (#1452). Leer heißt: keine.
    freigegeben: string[],
  ) => Promise<KontextpaketAntwort>
}
