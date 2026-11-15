// **Marken an jedem Fachobjekt** — UX-12
// (#1684),
// Vermerk 0135. Was eine Marke ist und was ein Filter heißt, entscheidet Go
// (`service.Objektmarken`); hier steht die Gestalt, in der es ankommt.

import type { Umschlag } from './umschlag'

// Markenart ist die Art eines Objekts, das eine Marke trägt.
export type Markenart =
  | 'karte'
  | 'set'
  | 'exemplar'
  | 'deck'
  | 'paket'
  | 'turnier'
  | 'partie'
  | 'handel'
  | 'symbol'

// Verknuepfung sagt, wie mehrere Marken eines Filters zusammengehen.
export type Verknuepfung = 'und' | 'oder'

export interface MarkenpostenDTO {
  marke: string
  anzahl: number
}

export interface MarkenJeObjektAntwort extends Umschlag {
  marken: Record<string, string[]>
}

export interface MarkenfilterAntwort extends Umschlag {
  kennungen: string[]
}

export interface MarkenkatalogAntwort extends Umschlag {
  marken: MarkenpostenDTO[]
}

export interface MarkenzahlAntwort extends Umschlag {
  anzahl: number
}

export interface Objektmarkenapi {
  ObjektmarkeSetzen: (
    art: Markenart,
    kennung: string,
    marke: string,
  ) => Promise<Umschlag>
  ObjektmarkeEntfernen: (
    art: Markenart,
    kennung: string,
    marke: string,
  ) => Promise<Umschlag>

  // MarkeVergeben hängt eine Marke an mehrere Objekte — ganz oder gar nicht.
  MarkeVergeben: (
    art: Markenart,
    kennungen: string[],
    marke: string,
  ) => Promise<Umschlag>
  MarkenJeObjekt: (art: Markenart) => Promise<MarkenJeObjektAntwort>

  // MarkenFiltern nennt die Objekte, die durch den Filter gehen — **Go**
  // entscheidet, was „und" und „oder" heißt, nicht dieses Fenster.
  MarkenFiltern: (
    art: Markenart,
    marken: string[],
    verknuepfung: Verknuepfung,
  ) => Promise<MarkenfilterAntwort>
  Markenkatalog: () => Promise<MarkenkatalogAntwort>
  MarkeUmbenennen: (
    alt: string,
    neu: string,
    zusammenfuehren: boolean,
  ) => Promise<MarkenzahlAntwort>
  MarkeUeberallEntfernen: (marke: string) => Promise<MarkenzahlAntwort>
}
