// **Die Symbolbibliothek** — DOM-02
// (#1697).
//
// Symbole eines Profils, die sich in jedes Set einsetzen lassen. Was ein
// verwendetes Symbol ist und wann es nicht gelöscht wird, entscheidet Go
// (`service.Symbolbibliothek`); hier steht die Gestalt, in der es ankommt.

import type { SymbolAntwort, SymbolteilDTO } from './antworten'
import type { Umschlag } from './umschlag'

export interface BibliothekssymbolDTO {
  id: string
  name: string

  // abdruck ist die Prüfsumme der Zeichnung — zwei gleiche sind dasselbe
  // Symbol.
  abdruck: string

  // angelegt und geaendert stehen nach RFC 3339, leer wo unbekannt.
  angelegt: string
  geaendert: string

  // vorschau ist ein PNG als Data-URI — leer, wo es sich nicht packen ließ.
  vorschau: string

  // verwendet nennt die Sets, die dieses Symbol tragen.
  verwendet: string[]

  // vorige sagt, dass sich eine vorige Fassung zurückholen lässt (AK-02).
  vorige: boolean
}

export interface SymbolbibliothekAntwort extends Umschlag {
  symbole: BibliothekssymbolDTO[]
}

export interface BibliothekssymbolAntwort extends Umschlag {
  id: string
  name: string
  pfad: string

  // abgebrochen sagt, dass niemand eine Datei gewählt hat — kein Fehler.
  abgebrochen: boolean
}

export interface SymbolfolgenAntwort extends Umschlag {
  sets: string[]
}

export interface SymbolEinsatzAntwort extends Umschlag {
  datei: string
}

// SymbolaenderungAntwort ist eine neue Fassung samt den Sets, die die vorige
// tragen — sie behalten ihre Kopie, bis jemand sie nachzieht.
export interface SymbolaenderungAntwort extends Umschlag {
  id: string
  name: string
  veraltet: string[]
}

export interface FassungsbildDTO {
  // groesse ist die Kantenlänge in Punkten.
  groesse: number
  vorschau: string
}

export interface SymbolfassungDTO {
  name: string
  bilder: FassungsbildDTO[]
}

export interface SymbolfassungenAntwort extends Umschlag {
  fassungen: SymbolfassungDTO[]
}

export interface Symbolbibliotheksapi {
  // Symbolbibliothek gibt die Symbole, deren Name `suche` enthält.
  Symbolbibliothek: (suche: string) => Promise<SymbolbibliothekAntwort>

  // BibliothekssymbolAufnehmen legt das Symbol eines Sets in die Bibliothek.
  BibliothekssymbolAufnehmen: (
    setID: string,
    name: string,
  ) => Promise<BibliothekssymbolAntwort>

  BibliothekssymbolKopieren: (
    id: string,
    name: string,
  ) => Promise<BibliothekssymbolAntwort>

  BibliothekssymbolUmbenennen: (
    id: string,
    name: string,
  ) => Promise<BibliothekssymbolAntwort>

  // BibliothekssymbolLoeschen nennt die Sets, die das Symbol tragen. **Ohne
  // `bestaetigt` geht ein verwendetes nicht** — die Absage nennt dieselben.
  BibliothekssymbolLoeschen: (
    id: string,
    bestaetigt: boolean,
  ) => Promise<SymbolfolgenAntwort>

  // BibliothekssymbolEinsetzen legt eine Kopie neben das Set und macht sie
  // zu dessen Symbol.
  BibliothekssymbolEinsetzen: (
    id: string,
    setID: string,
  ) => Promise<SymbolEinsatzAntwort>

  // Ein- und Ausgeben fangen mit einem Dateidialog an.
  BibliothekssymbolEinlesen: () => Promise<BibliothekssymbolAntwort>
  BibliothekssymbolAusgeben: (
    id: string,
    name: string,
  ) => Promise<BibliothekssymbolAntwort>

  // **Die Bearbeitung** — DOM-02.AK-02. Die Teile haben dieselbe Gestalt wie
  // im Symboleditor.
  BibliothekssymbolTeile: (id: string) => Promise<SymbolAntwort>
  BibliothekssymbolSichern: (
    id: string,
    teile: SymbolteilDTO[],
  ) => Promise<SymbolaenderungAntwort>
  BibliothekssymbolZuruecknehmen: (
    id: string,
  ) => Promise<SymbolaenderungAntwort>
  BibliothekssymbolNachziehen: (id: string) => Promise<SymbolfolgenAntwort>

  // BibliothekssymbolFassungen zeichnet ein Symbol in den Fassungen eines
  // Sets — in drei Größen.
  BibliothekssymbolFassungen: (
    id: string,
    setID: string,
  ) => Promise<SymbolfassungenAntwort>
}
