// **Was ein erzeugtes Bild mitbringt** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Eine eigene Datei
//
// **Weil `antwortenmarkt.ts` mit ihr über die Zeilengrenze aus §2 ginge.**
// Der Schnitt folgt der Frage: dort steht, was **Märkte und Anbieter** sagen,
// hier, was aus einem **Bildauftrag** zurückkommt.
import type { Umschlag } from './umschlag'

// KIBildangabeDTO ist eine Angabe über die Entstehung eines Bildes.
//
// **Als Paar und nicht als Verbund mit festen Feldern**: was ein Anbieter
// nennt, ist bei jedem ein anderes. Was **nicht** dasteht, wurde nicht
// genannt — und wird nicht erfunden.
export interface KIBildangabeDTO {
  name: string
  wert: string
}

// KIBildvorschlagAntwort ist ein wartender Bildvorschlag.
//
// **Er ändert die Karte nicht.** Erst `KIBildEinsetzen` schreibt.
export interface KIBildvorschlagAntwort extends Umschlag {
  // kennung ist die gewürfelte Kennung im Vorrat — leer, wenn nichts kam.
  kennung: string

  anbieter: string
  modell: string

  // umschrift ist der Prompt, wie der Dienst ihn umgeschrieben hat.
  //
  // **Sie ist der Unterschied zwischen Wunsch und Auftrag**: gezeichnet
  // wurde, was hier steht, und nicht, was jemand getippt hat.
  umschrift: string

  breite: number
  hoehe: number
  bytes: number

  // vorschau ist das Bild als Daten-URI.
  vorschau: string

  // ablauf sagt, bis wann der Vorschlag gilt (RFC 3339).
  ablauf: string

  angaben: KIBildangabeDTO[]
}

// KIBildherkunftAntwort sind die Angaben zu einem übernommenen Bild.
export interface KIBildherkunftAntwort extends Umschlag {
  angaben: KIBildangabeDTO[]
}
