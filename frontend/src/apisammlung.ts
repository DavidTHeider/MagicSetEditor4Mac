// Die Fassade der **Sammlung** — Werkstück 2 und 3 von Phase 3.4
// (#1165).
//
// **Eine eigene Schnittstelle**, wie `apifeld.ts` und `apibilder.ts`: `api.ts`
// steht an der Zeilengrenze aus §2, und der Schnitt folgt der Frage — dort das
// Set, hier der **Besitz**.
import type {
  Ausfuhrtextantwort,
  Einfuhrantwort,
  Listendeutungsantwort,
  Listeneinfuhrzeile,
  Sammelvorschauantwort,
  Zusammenfuehrung,
} from './antwortenlisteneinfuhr'
import type {
  Besitzantwort,
  Dublettenantwort,
  ExemplarDTO,
  Exemplarantwort,
  Lagerortantwort,
  Massenantwort,
  MassenaenderungDTO,
  MassenzeileDTO,
  Portfolioantwort,
  Preisschubantwort,
  Bulkbezugsantwort,
  Preisantwort,
  Preisreihe,
  Sammlungsantwort,
  Sammlungsfensterantwort,
} from './antwortensammlung'

import type {
  Anbieterstandantwort,
  Fremdpreisantwort,
  KIGeraetSitzungAntwort,
  KIKatalogantwort,
  KIProbeAntwort,
  KIVorschlagAntwort,
  KIVorschlagsbezug,
  KIUebernahmeAntwort,
  KIFeldherkunftAntwort,
  Marktantwort,
} from './antwortenmarkt'
import type {
  KIErkennungAntwort,
  KIErkennungsbezug,
} from './antwortenerkennung'
import type {
  KIBatchAntwort,
  KIGruppenantwort,
  KIGruppenwahl,
} from './antwortenbatch'

import type { FehlerDTO } from './formen'
import type {
  KIEinschaetzungAntwort,
  KIEinschaetzungsbezug,
} from './antworteneinschaetzung'

export interface Sammlungsapi {
  // **Die Sammlung** (#1165). Sie hängt an keinem Set.
  Sammlung: () => Promise<Sammlungsantwort>

  // **Der Ausschnitt** (#1500): filtern und ordnen geschehen in Go, und das
  // Fenster holt nur, was es zeichnet. `anzahl` von null heißt alles.
  Sammlungsfenster: (
    suche: string,
    ausfuehrung: string,
    sortierung: string,
    ab: number,
    anzahl: number,
  ) => Promise<Sammlungsfensterantwort>
  ExemplarAnlegen: (exemplar: ExemplarDTO) => Promise<Exemplarantwort>
  ExemplarAendern: (exemplar: ExemplarDTO) => Promise<FehlerDTO | null>
  // **Die Fassung geht mit** — sonst nähme ein Löschen eine fremde Änderung
  // mit, die dieses Fenster nie gesehen hat (§7, Regel 2).
  ExemplarLoeschen: (id: number, fassung: number) => Promise<FehlerDTO | null>

  // **Die Preise** (#1165). `PreiseHolen` ist der einzige, der ins Netz geht.
  PreiseHolen: (printing: string) => Promise<Preisantwort>
  Preise: (printing: string) => Promise<Preisantwort>
  // **Die ganze Reihe und keine Vorgabe** (A3, #1442): welche Reihe gemeint
  // ist, sagt der Aufrufer.
  Preisverlauf: (reihe: Preisreihe) => Promise<Preisantwort>
  // **Ein Ruf, eine große Datei** (A3, #1442): MTGJSON antwortet nicht je
  // Karte, sondern mit allem. Quelle und Umfang werden **genannt** — eine
  // Vorgabe wäre eine Preisstrategie, die aussieht wie keine.
  Preisbulk: (quelle: string, umfang: string) => Promise<Bulkbezugsantwort>
  // **Die Angebote eines Marktplatzes** (A4, #1443). Der Abruf geht ins Netz,
  // und gekauft wird nichts. Sprache, Ausführung und Zustände engen ein — leer
  // heißt: alle.
  Marktangebote: (
    quelle: string,
    printing: string,
    sprache: string,
    finish: string,
    zustaende: string[],
  ) => Promise<Marktantwort>
  // **Das Kontingent an einem Anbieterkonto** (A5, #1444). Die Abfrage geht
  // ins Netz und kostet selbst einen Kredit.
  Anbieterstand: (dienst: string) => Promise<Anbieterstandantwort>

  // **Der Preisverlauf bei einer fremden Quelle** (A5, #1444). Er geht ins
  // Netz und legt die Stände ab.
  Fremdpreisverlauf: (
    quelle: string,
    printing: string,
    finish: string,
    zustaende: string[],
  ) => Promise<Fremdpreisantwort>
  // **Was über KI-Modelle feststeht** (A6, #1445). Liest nur; eine leere
  // Liste heißt: es steht keines fest.
  KIModelle: () => Promise<KIKatalogantwort>

  // **Apples Gerätemodell für diese Sitzung ab- oder anschalten** (AI-05,
  // #2042) — nach einem Neustart gilt wieder `ki.geraetemodell`.
  KIGeraetemodellFuerSitzung: (an: boolean) => Promise<KIGeraetSitzungAntwort>

  // **Ein Vorschlag ändert die Karte nicht** (A7, #1446) — dafür gibt es die
  // Bindung darunter.
  // **Ein Verbund statt fünf Zeichenketten** (A13.4, #1452): mit Belegfrage
  // und Freigabe wären es sieben, und wer zwei davon vertauscht, merkt es an
  // keiner Stelle — der Typ ist derselbe.
  KIVorschlagen: (
    karteID: string,
    bezug: KIVorschlagsbezug,
  ) => Promise<KIVorschlagAntwort>
  // **Die Auswahl steht im Ruf** (A11, #1450): übernommen wird, was
  // angehakt ist — und nicht alles, was das Modell vorgeschlagen hat.
  KIVorschlagUebernehmen: (
    kennung: string,
    felder: string[],
  ) => Promise<KIUebernahmeAntwort>

  // **Ein Modell für MEHRERE Karten fragen** (A13.1b, #1452).
  //
  // **Es schreibt an keiner Karte** — je Ziel entsteht ein Vorschlag oder ein
  // Grund; geschrieben wird erst über `KIGruppeUebernehmen`.
  KIBatchVorschlagen: (
    karteIDs: string[],
    anbieter: string,
    alias: string,
    auftrag: string,
    nachweis: string,
  ) => Promise<KIBatchAntwort>

  // **Mehrere Vorschläge als EINE Gruppe übernehmen** (A13.1b, #1452).
  //
  // **Ganz oder gar nicht**: scheitert ein Schritt, ist keine Karte
  // geschrieben und kein Vorschlag verbraucht. Wer stattdessen
  // `KIVorschlagUebernehmen` in einer Schleife ruft, hat eine Transaktion je
  // Karte — und ein Abbruch in der Mitte lässt die halbe Gruppe stehen.
  KIGruppeUebernehmen: (wahl: KIGruppenwahl[]) => Promise<KIGruppenantwort>

  // **Eine Karte beurteilen lassen** (A13.2a, #1452). Die Belege kommen aus
  // dem lokalen Suchindex; was das Modell darüber hinaus behauptet, steht
  // unter `befunde` und **nicht** als Quelle.
  //
  // **Sie ändert nichts** — und wird nicht übernommen.
  //
  // **Ein Verbund und keine sieben Stellungsparameter** (A13.2b, #1452): mit
  // dem Format wären es acht, und wer zwei davon vertauscht, merkt es an
  // keiner Stelle — der Typ ist derselbe.
  KIEinschaetzen: (
    bezug: KIEinschaetzungsbezug,
  ) => Promise<KIEinschaetzungAntwort>

  // **Ein Kartenfoto lesen lassen** (A13.3, #1452). Das Modell nennt
  // **gelesene Stellen**, die Kandidaten kommen aus derselben Drucksuche wie
  // bei einer getippten Liste.
  //
  // **Sie legt nichts an** — weder Druckbindung noch Sammlungsmenge. Ein
  // leerer `pfad` öffnet am Fenster den Dateidialog.
  KIKarteErkennen: (bezug: KIErkennungsbezug) => Promise<KIErkennungAntwort>

  // **Woher die Werte einer Karte stammen** (A11, #1450). Eine Auskunft aus
  // dem Bestand — hier geht kein Aufruf nach draußen.
  KIFeldherkunft: (karteID: string) => Promise<KIFeldherkunftAntwort>

  // **Messen, was ein Modell kann** (A9, #1448) — nicht jeder Anbieter sagt
  // es. Kostet Geld, wo er anderswo rechnet.
  KIModellPruefen: (
    anbieter: string,
    alias: string,
    nachweis: string,
  ) => Promise<KIProbeAntwort>
  KIModellVergessen: (
    anbieter: string,
    alias: string,
  ) => Promise<KISchreibAntwort>
  Portfolio: () => Promise<Portfolioantwort>

  // **Die Listeneinfuhr** (#1292). Zwei Schritte, nicht einer: `ListeDeuten`
  // schreibt nichts und sagt, welche Zeile welchen Druck meint; „Counterspell“
  // ohne Setangabe passt auf 72 Drucke, und welchen jemand besitzt, weiß nur er.
  ListeDeuten: (text: string) => Promise<Listendeutungsantwort>
  ListeEinlesen: (
    zeilen: Listeneinfuhrzeile[],
    zusammenfuehrung: Zusammenfuehrung,
  ) => Promise<Einfuhrantwort>

  // **Die Vorschauen** — DOM-03.AK-01 (#1698). Sie schreiben nichts und
  // sagen, was das Einlesen zusammenführte.
  ListeVorschauen: (
    zeilen: Listeneinfuhrzeile[],
  ) => Promise<Sammelvorschauantwort>
  SammlungCSVVorschauen: (text: string) => Promise<Sammelvorschauantwort>

  // **Die Ausfuhr** (#1292). »csv« trägt jede Spalte, »liste« ist die
  // einfügbare Gestalt und verlustbehaftet.
  SammlungAusgeben: (form: string) => Promise<Ausfuhrtextantwort>

  // **`quelle` sagt, woher das CSV kam** — ein Dateiname, ein Pfad, eine
  // Bezeichnung; leer heißt „niemand hat es gesagt".
  //
  // **Die Bindung verlangt sie** (Befund F03,
  // #1747),
  // und `tsc` merkt es nicht, wenn sie fehlt: dieser Vertrag und sein
  // Aufrufer stimmen überein, über Go könnte der Vertrag trotzdem lügen.
  // `internal/werkzeuge/bindungsstelligkeit_test.go` hält beide Seiten auf
  // derselben Stelligkeit.
  SammlungCSVEinlesen: (
    text: string,
    quelle: string,
    zusammenfuehrung: Zusammenfuehrung,
  ) => Promise<Einfuhrantwort>

  // **Die vier Fragen an eine geordnete Sammlung** (#1292, Vermerk 0057).
  SammlungBesitz: (printing: string) => Promise<Besitzantwort>
  SammlungLagerorte: () => Promise<Lagerortantwort>
  SammlungDubletten: (behalten: number) => Promise<Dublettenantwort>
  SammlungWuensche: () => Promise<Sammlungsantwort>

  // **Die Massenvorgänge** (#1292). Jede Zeile trägt ihre Fassung.
  MasseAendern: (aenderung: MassenaenderungDTO) => Promise<Massenantwort>
  MasseLoeschen: (zeilen: MassenzeileDTO[]) => Promise<Massenantwort>

  // **Ein Schub je Ruf** — der Grund ist die Drossel bei Scryfall. Wer den
  // Fortschritt zeigen will, ruft wieder, bis »fertig« wahr ist.
  MassePreise: (hoechstensStunden: number) => Promise<Preisschubantwort>
}

// KISchreibAntwort ist eine Antwort ohne Nutzlast (A9, #1448).
//
// **Eine eigene Angabe und kein Import**: die Kante zeigte sonst zurück, und
// zwei Dateien, die einander kennen, sind eine — dieselbe Begründung wie bei
// `SchreibAntwortZugang`.
export interface KISchreibAntwort {
  fehler: FehlerDTO | null
}
