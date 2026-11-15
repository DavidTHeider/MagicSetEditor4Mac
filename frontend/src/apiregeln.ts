// **Die Regelfreigabe an der Naht** — Werkstück G4 von Phase 5.5
// (#1461).
import type { Umschlag } from './umschlag'

// Matrixeintrag ist die Aussage über einen Regelbereich.
export interface Matrixeintrag {
  bereich: string

  // grad ist `unbekannt`, `manuell`, `begleitet` oder `automatisch`.
  grad: string

  // quelle ist leer, wo nichts zugesagt ist.
  //
  // **Und das ist Absicht**: eine Regelnummer neben einer Nichtzusage sähe
  // aus wie ein Beleg.
  quelle: string

  // grenzen nennt, was ausdrücklich **nicht** abgedeckt ist.
  grenzen: string[]
}

// RegelfreigabeDTO ist die Antwort auf „darf es so losgehen?".
export interface RegelfreigabeDTO {
  erlaubt: boolean

  regelwerk: string
  fassung: number

  // wirksam ist das Datum, ab dem diese Fassung gilt.
  wirksam: string

  // hoechste ist die strengste Stufe, die dieses Regelwerk trägt.
  //
  // **Sie steht auch bei einem Ja da** — wer „begleitet" bekommt, soll
  // sehen, dass „streng" nicht zur Wahl steht.
  hoechste: string

  matrix: Matrixeintrag[]
  luecken: string[]
  hindernisse: string[]

  // modus ist der Bedienmodus, über den gerechnet wurde — `anfaenger` oder
  // `frei` (PLAY-02.AK-01,
  // #1710).
  //
  // **Er steht in der Antwort**, damit sichtbar wird, dass ohne Angabe der
  // Anfängermodus gilt.
  modus: string

  // baum ordnet die Bereiche nach Ablauf und Mechanik (AK-02).
  baum: Regelgruppe[]

  // widersprueche sind die Abwahlen, an denen noch etwas hängt.
  //
  // **Sie verhindern nichts** — sie erklären.
  widersprueche: Regelwiderspruch[]
}

// Regelgruppe ist ein Ast des Regelbaums.
export interface Regelgruppe {
  // kennung ist `ablauf` oder `mechanik`.
  kennung: string

  bereiche: string[]
}

// Regelwiderspruch ist eine Abwahl, an der noch etwas hängt.
export interface Regelwiderspruch {
  // bereich ist der abgewählte, haengt der, der ihn braucht.
  bereich: string
  haengt: string

  // quelle belegt die Abhängigkeit — `CR 608 (Fassung …)`.
  quelle: string
}

// RegelfreigabeAntwort ist der Ausgang des Weges.
export interface RegelfreigabeAntwort extends Umschlag {
  freigabe: RegelfreigabeDTO
}

// Regelapi ist der eine Weg.
export interface Regelapi {
  // Regelfreigabe sagt, ob eine Partie so anfangen darf.
  //
  // **Modus und Abwahl sind freiwillig** (PLAY-02,
  // #1710):
  // leer heißt Anfängermodus und nichts abgewählt.
  Regelfreigabe: (
    spiel: string,
    strenge: string,
    sitze: number,
    modus: string,
    abgewaehlt: string[],
  ) => Promise<RegelfreigabeAntwort>
}
