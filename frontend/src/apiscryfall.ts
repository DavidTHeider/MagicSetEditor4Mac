// Der Zugang zu Scryfall
// (#1005).
//
// **Eine eigene Datei, wie `apibilder.ts`** — und aus demselben Grund: `api.ts`
// führt die ganze Fassade und liefe sonst über die Zeilengrenze (§2).

import type { Umschlag } from './umschlag'
import type { AuswahlDTO, FehlerDTO, SchreibAntwort } from './api'
import type { ScryfallvorschlagAntwort } from './antworten'

// Scryfallapi ist der Weg zu den Kartendaten von Scryfall.
//
// **Zwei Schritte und nicht einer** (§16, Regel 6): „Nichts wird still
// übernommen. Was von außen in ein Kartenfeld will, zeigt vorher, was sich
// ändert, und trägt danach seine Herkunft."
//
// `ScryfallVorschlag` schreibt **nichts**; erst `ScryfallUebernehmen` mit den
// ausgewählten Feldern schreibt.
export interface Scryfallapi {
  ScryfallVorschlag: (
    setID: string,
    karteID: string,
    name: string,
  ) => Promise<ScryfallvorschlagAntwort>

  // felder sind die Feldnamen aus dem Vorschlag, die übernommen werden sollen.
  //
  // **Ein Feld, das nicht im Vorschlag steht, wird abgewiesen** — sonst wäre
  // der Vorschlag eine Höflichkeit statt einer Grenze.
  // **`stand` bindet den Vorschlag an die Kartenwerte**, gegen die verglichen
  // wurde (#1241).
  //
  // Dieser Weg liest die Karte neu und rechnet den Vorschlag ein zweites Mal;
  // zwischen Zeigen und Übernehmen kann sie sich geändert haben. **Leer heißt:
  // ohne Prüfung** — auf der Kommandozeile, über MCP und über den Dienst gibt
  // es keinen Bestätigungsschritt, den etwas binden könnte.
  ScryfallUebernehmen: (
    setID: string,
    karteID: string,
    name: string,
    felder: string[],
    stand: string,
  ) => Promise<SchreibAntwort>

  // **Der Bildweg, und er hat dieselben zwei Schritte.**
  //
  // `ScryfallBild` holt das Bild und schlägt einen Zuschnitt vor — es
  // schreibt nichts. Erst `ScryfallBildEinsetzen` legt ab.
  ScryfallBild: (
    karteID: string,
    feld: string,
    name: string,
    art: string,
  ) => Promise<ScryfallbildAntwort>

  // kuenstler wird als `illustrator` in die Karte geschrieben.
  //
  // **Das ist die Nennungspflicht, nicht eine Bequemlichkeit.** Scryfall
  // verlangt Künstler und Copyright in derselben Oberfläche wie das Bild —
  // und die `.mse-set`-Ausfuhr ist auch eine. Ein leerer Name schreibt
  // nichts, damit eine Nennungspflicht keine vorhandene Nennung löscht.
  ScryfallBildEinsetzen: (
    setID: string,
    karteID: string,
    feld: string,
    pfad: string,
    kuenstler: string,

    // herkunft ist die Kartenseite bei Scryfall.
    //
    // **Sie landet im Verlauf**, an beiden Zeilen — dem Bildfeld und dem
    // Künstler. Wer den Verlauf liest, soll eine Übernahme von einer Eingabe
    // unterscheiden können, ohne den Wert zu erraten.
    herkunft: string,

    auswahl: AuswahlDTO,
  ) => Promise<SchreibAntwort>

  // **Eine Karte aus einem Namen** (#1032).
  //
  // Bis dahin arbeitete der Abgleich nur an einer **vorhandenen** Karte: erst
  // eine leere anlegen, dann abgleichen — zwei Schritte für einen Vorgang.
  //
  // **Das Bild ist wahlweise dabei**: ohne `art` wird keines geholt. Mit `art`
  // muss `bildfeld` dastehen — welches Feld ein Bild aufnimmt, weiß nur die
  // Vorlage.
  //
  // **Das Spiel steht am Set** und wird nicht mitgegeben: eine Fläche, die es
  // mitgeben müsste, könnte es falsch mitgeben.
  // ScryfallSuche sucht Karten bei Scryfall
  // (#1167).
  //
  // **Die Frage ist Scryfalls eigene Syntax** — ein Kartenname genügt, aber
  // `t:creature c:r cmc<=3` geht auch. Sie geht unverändert hinaus.
  ScryfallSuche: (frage: string) => Promise<ScryfallsuchAntwort>

  ScryfallKarteAnlegen: (
    setID: string,
    name: string,

    // bildfeld nimmt das Bild auf — nötig nur, wenn `art` gesetzt ist.
    bildfeld: string,

    // art ist leer, wenn kein Bild geholt werden soll.
    art: string,
  ) => Promise<ScryfallKarteAntwort>

  // **Zwei lesende Wege** (#888). Sie schreiben nichts und hängen an keinem
  // Set — eine Regelauskunft und eine Manakostenprüfung lesen nichts aus dem
  // Bestand.
  ScryfallRegeln: (name: string) => Promise<RegelauskunftAntwort>

  ScryfallManakosten: (kosten: string) => Promise<ManakostenAntwort>

  // **Die Sammeldatei ist der Betrieb ohne Netz** (#888): fällt Scryfall aus,
  // kommen die Kartenangaben aus ihr — mit einem Hinweis, dass sie bis zu
  // einen Tag alt sind.
  ScryfallSammeldaten: () => Promise<SammeldatenAntwort>

  // **Lädt rund 25 MB.** Deshalb ein eigener Aufruf und kein Schalter am
  // ersten — wer nachsieht, soll nicht laden.
  ScryfallSammeldatenHolen: () => Promise<SammeldatenAntwort>

  // **Der Verweis auf einen echten Druck** — Werkstück 2 von Phase 3.6
  // (#1167).
  //
  // Ohne Verweis holt jeder Abgleich **irgendeinen** Druck: `/cards/named`
  // wählt selbst aus. Mit Verweis holt er **den gewählten** — und der
  // Vorschlag nennt ihn.
  KarteVerweis: (karteID: string) => Promise<VerweisAntwort>

  // printing ist die Kennung aus der Trefferliste.
  //
  // **Der Druck wird geholt, bevor der Verweis entsteht**: eine Kennung, die
  // es nicht gibt, wird abgewiesen — sonst stünde ein Verweis da, den niemand
  // lesen kann.
  KarteVerweisSetzen: (
    karteID: string,
    printing: string,
  ) => Promise<VerweisAntwort>

  KarteVerweisLoesen: (karteID: string) => Promise<SchreibAntwort>
}

// KartenverweisDTO ist der Bezug einer Karte auf einen echten Druck (#1167).
export interface KartenverweisDTO {
  karteID: string

  // printing ist die Scryfall-Kennung des Drucks.
  printing: string

  // druckID ist die Zeile des Drucks in diesem Bestand (A2, #1441);
  // vorlaeufig sagt, dass der Druck keinen Alias hat.
  druckID: string
  vorlaeufig: boolean

  // name, setcode und nummer sind der **letzte gute Stand** der Anzeige —
  // ohne sie wäre der Verweis ohne Netz eine UUID (§16, Regel 5).
  name: string
  setcode: string
  nummer: string

  // adresse ist die Kartenseite bei Scryfall — anklickbar.
  adresse: string

  // gesetzt sagt, wie alt die Anzeige ist. RFC 3339.
  gesetzt: string
}

// VerweisAntwort trägt einen Verweis oder keinen.
export interface VerweisAntwort {
  // verweis ist `null`, wenn die Karte keinen hat — der Normalfall, kein
  // Fehler.
  verweis: KartenverweisDTO | null
  befunde: string[]
  fehler?: FehlerDTO | null
}

// SammeldatenAntwort sagt, ob die Sammeldatei bereitliegt.
export interface SammeldatenAntwort {
  art: string
  pfad: string
  vorhanden: boolean

  // mb und angebotenMB sind die Größen in Megabyte — die Go-Seite rechnet,
  // damit das Fenster es nicht zweimal tut.
  mb: number
  angebotenMB: number

  stand: string

  // geholt ist, wann die liegende Datei geschrieben wurde — leer, wenn sie
  // fehlt. Ohne Netz die einzige Auskunft über ihr Alter (F04, #2246).
  geholt: string
  fehler?: { code: string; schluessel: string; meldung: string } | null
}

// RegelauskunftAntwort sind die Regelhinweise zu einer Karte.
export interface RegelauskunftAntwort {
  // kartenname ist der Name, unter dem Scryfall die Karte führt — **nicht**
  // der eingegebene. Wer sich vertippt hat, sieht es hier.
  kartenname: string

  hinweise: {
    // offiziell sagt, ob der Hinweis von Wizards of the Coast stammt.
    //
    // **Die Go-Seite rechnet es aus.** Ein `=== 'wotc'` hier wäre eine zweite
    // Fassung derselben Regel, und sie liefe auseinander, sobald Scryfall
    // einen dritten Wert einführt.
    offiziell: boolean

    quelle: string
    datum: string
    text: string
  }[]

  adresse: string
  befunde: string[]
  alterStunden: number
  fehler?: { code: string; schluessel: string; meldung: string } | null
}

// ManakostenAntwort nennt die Symbole, die es in einem Manakosten nicht gibt.
export interface ManakostenAntwort {
  // unbekannt ist leer, wenn der Kosten in Ordnung ist.
  unbekannt: string[]

  fehler?: { code: string; schluessel: string; meldung: string } | null
}

// ScryfallKarteAntwort ist eine angelegte Karte samt dem, was in sie kam
// (#1032).
export interface ScryfallKarteAntwort {
  // karteID ist die angelegte Karte — **auch wenn das Füllen scheitert**.
  //
  // Eine Karte, die entstanden ist, verschwindet nicht; wer sie nicht nennt,
  // lässt sie im Bestand zurück, ohne dass jemand davon weiß.
  karteID: string

  felder: string[]
  befunde: string[]
  herkunft: string
  alterStunden: number

  // nennung steht dabei, wenn ein Bild mitgeholt wurde.
  nennung: string

  fehler?: FehlerDTO | null
}

// ScryfallbildAntwort ist ein geholtes Bild samt Zuschnittvorschlag.
export interface ScryfallbildAntwort {
  pfad: string
  auswahl: AuswahlDTO
  kuenstler: string

  // nennung ist der Satz, der in derselben Oberfläche stehen muss.
  nennung: string

  // adresse ist die Kartenseite bei Scryfall — die „source" der Bedingung.
  adresse: string

  alterStunden: number

  // befunde ist, was die Kartenauskunft tolerant hingenommen hat — der Grund
  // steht bei `service.Bilduebernahme`.
  befunde: string[]

  fehler?: { code: string; schluessel: string; meldung: string } | null
}

// SuchtrefferDTO ist eine bei Scryfall gefundene Karte (#1167).
export interface SuchtrefferDTO {
  // kennung ist die Scryfall-Kennung des **Drucks** — der Bezug, an dem
  // Preise, Sammlung und Übernahme hängen.
  kennung: string
  name: string
  setcode: string
  setname: string
  nummer: string
  seltenheit: string
  manakosten: string
  typzeile: string

  // bild ist die Adresse eines kleinen Kartenbilds — leer, wenn es keines gibt.
  bild: string
}

// ScryfallsuchAntwort trägt die Treffer samt ihrer Herkunft (#1167).
export interface ScryfallsuchAntwort extends Umschlag {
  treffer: SuchtrefferDTO[]

  // weitere sagt, ob Scryfall mehr hat — **nachgeladen wird nicht**.
  weitere: boolean

  alterStunden: number
  befunde: string[]
}
