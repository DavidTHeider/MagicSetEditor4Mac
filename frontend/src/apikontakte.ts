// Die Fassade der **Kontakte aus einem Adressbuch** — ID-04
// (#1703).
import type { FehlerDTO } from './formen'
import type { Umschlag } from './umschlag'

export type Kontaktwegart = 'email' | 'telefon'

export interface KontaktwegDTO {
  art: Kontaktwegart
  wert: string
}

// KontaktDTO ist ein Kontakt — **das Bild reist als Base64**.
export interface KontaktDTO {
  id: number
  fassung: number
  name: string
  wege: KontaktwegDTO[]
  bild: string | null
  bildart: string
  ursprung: string
  uebernommen: string
}

export type Kontaktabgleich = 'neu' | 'gleich' | 'geaendert' | 'vermutlich'

export interface KartenvorschauDTO {
  kontakt: KontaktDTO
  abgleich: Kontaktabgleich
  vorhanden: KontaktDTO | null
}

export interface KontaktvorschauAntwort {
  karten: KartenvorschauDTO[]
  befunde: string[]
  abgebrochen: boolean
  fehler: FehlerDTO | null
}

export interface KontaktuebernahmeAntwort {
  angelegt: number
  ersetzt: number
  fehler: FehlerDTO | null
}

export interface KontakteAntwort {
  kontakte: KontaktDTO[]
  fehler: FehlerDTO | null
}

export interface Kontaktapi {
  // KontakteEinlesen öffnet den Dateidialog und zeigt die Karten — **ohne
  // zu schreiben**.
  KontakteEinlesen: () => Promise<KontaktvorschauAntwort>
  KontakteUebernehmen: (
    auswahl: KontaktDTO[],
  ) => Promise<KontaktuebernahmeAntwort>
  Kontakte: () => Promise<KontakteAntwort>
  KontaktLoeschen: (id: number, fassung: number) => Promise<Umschlag>
}
