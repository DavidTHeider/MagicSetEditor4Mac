// **Eine Datei ansehen, ohne dass etwas geschieht** — UX-13
// (#1685).
//
// Was eine Datei zeigen darf, entscheidet Go (`domain.AnsichtsregelFuer`);
// hier steht nur die Gestalt, in der die Antwort ankommt.

import type { Umschlag } from './umschlag'

// DateieintragDTO ist ein Eintrag eines Pakets oder Archivs.
export interface DateieintragDTO {
  name: string
  groesse: number
  verzeichnis: boolean

  // unsicher ist ein Verweis oder ein Name, der aus dem Archiv hinausführte —
  // gezählt und nie betreten.
  unsicher: boolean
}

// DateiansichtDTO ist eine Datei, wie das Fenster sie zeigt.
export interface DateiansichtDTO extends Umschlag {
  name: string
  pfad: string
  art: 'paket' | 'bild' | 'vektor' | 'pdf' | 'text' | 'archiv' | 'unbekannt'
  vorschau: 'bild' | 'text' | 'liste' | 'keine'
  groesse: number
  geaendert: string
  verzeichnis: boolean

  // oeffnen sagt, ob die Datei dem Standardprogramm übergeben werden darf.
  oeffnen: boolean

  text: string
  gekuerzt: boolean

  // bild ist eine Datenadresse — erscheint über `<img>`, wo WebKit kein
  // Skript ausführt und keine fremde Adresse lädt.
  bild: string
  breite: number
  hoehe: number
  seiten: number

  eintraege: DateieintragDTO[]
  unsicher: number

  verschluesselt: boolean
  zuGross: boolean
  unlesbar: boolean
}

// Dateiansichtapi sind die drei Wege: ansehen, im Finder zeigen, öffnen.
//
// **Teilen fehlt, und mit Grund**: das Teilenblatt von macOS
// (`NSSharingServicePicker`) braucht einen Anker im nativen Fenster, und die
// Hülle reicht keinen durch (`docs/DATEIANSICHTEN.md`).
export interface Dateiansichtapi {
  DateiAnsehen: (pfad: string) => Promise<DateiansichtDTO>

  // DateiImFinderZeigen markiert die Datei in einem Finderfenster — für jede
  // Datei, denn es startet nichts.
  DateiImFinderZeigen: (pfad: string) => Promise<Umschlag>

  // DateiMitProgrammOeffnen übergibt sie dem Standardprogramm — nur, was
  // `oeffnen` erlaubt.
  DateiMitProgrammOeffnen: (pfad: string) => Promise<Umschlag>
}
