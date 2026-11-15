// **Was ein Modell über eine vorhandene Karte sagt** — Werkstück A13.2a/b
// von Phase 5.3
// (#1452),
// herausgezogen für A13.4, als `antwortenmarkt.ts` an die Zeilengrenze aus §2
// kam.
//
// **Der Schnitt folgt der Frage**: dort steht, was über **Modelle und
// Anbieter** feststeht — Katalog, Belege, Herkunft, Vorschlag. Hier steht,
// was ein Modell über **eine Karte** sagt: die Einschätzung samt ihren
// Vergleichskarten.
import type { BelegDTO } from './apirecherche'
import type { KIBelegbefundDTO } from './antwortenmarkt'
import type { KorpusstandDTO } from './apiindex'
import type { Umschlag } from './umschlag'

// KIEinschaetzungAntwort ist die Beurteilung einer Karte (A13.2a, #1452).
//
// **Sie ändert nichts** — weder Karte noch Sammlung noch Spielstand. Und sie
// wird nicht übernommen: es gibt keine Bindung, die sie in eine Karte trüge.
export interface KIEinschaetzungAntwort extends Umschlag {
  art: string

  // satz ist die Einschätzung — mit den Fußnoten darin, wie das Modell sie
  // gesetzt hat.
  satz: string

  // belegt sagt, ob sie **nur** vergebene Belege nennt.
  //
  // **`false` heißt nicht „falsch"**: es heißt, dass an ihrem
  // Quellenanspruch etwas nicht stimmt — und `befunde` sagt, was.
  belegt: boolean

  // vorbehalt steht an jeder: keine amtliche Regelauskunft.
  vorbehalt: string

  befunde: KIBelegbefundDTO[]

  // einschraenkungen sagt, was das Modell **nicht** betrachtet hat.
  einschraenkungen: string[]

  // belege sind **alle** mitgegebenen Ausschnitte, nicht nur die genannten.
  belege: BelegDTO[]

  korpusstaende: KorpusstandDTO[]

  anbieter: string
  modell: string

  // balance ist `null` bei `regel` — eine Regeleinschätzung hat keine
  // Vergleichskarten, und ein leerer Verbund sähe aus, als hätte die Suche
  // keine gefunden.
  //
  // **An den drei Flächen ohne Fenster fehlt der Schlüssel stattdessen.**
  // Dort liest ein Sprachmodell, für das `null` und „leer" dasselbe Bild
  // sind; hier liest TypeScript, und der Typ unterscheidet es.
  balance: KIBalanceDTO | null
}

// KIEinschaetzungsbezug ist, was für eine Einschätzung gebraucht wird.
export interface KIEinschaetzungsbezug {
  // art ist `regel` oder `balance`.
  art: string

  karteID: string

  anbieter: string
  alias: string

  // frage geht **auch** an den Suchindex.
  frage: string

  // format ist, worauf sich ein Vergleich bezieht — nur bei `balance`.
  format: string

  // nachweis ist die Freigabe eines Menschen — leer heißt: keine.
  nachweis: string

  // freigegeben nennt die **privaten** Korpora, die mitgesucht werden dürfen
  // — A13.4 (#1452).
  //
  // **Ein Vergleich braucht sie**: er sucht über `aehnlich`, und der liest
  // die eigenen Karten. Leer heißt: keine, und die Antwort sagt unter
  // `einschraenkungen`, dass sie ausgelassen wurden.
  freigegeben: string[]
}

// KIVergleichDTO ist eine Karte, gegen die verglichen wurde (A13.2b, #1452).
export interface KIVergleichDTO {
  // beleg ist die Kennung, unter der sie im Kontextpaket steht.
  //
  // **Eine Kennung und kein Name**: zwei Karten dürfen denselben Namen
  // tragen, und ein Name ließe sich erfinden.
  beleg: string

  name: string

  // unterschiede sind die konkreten — Manawert, Effekt, Zusatzfähigkeit.
  unterschiede: string[]
}

// KIBalanceDTO trägt, was nur eine Balanceaussage hat.
//
// **Ohne Zahlenfeld, und das ist die Zusage**: „kein objektiver Balancewert
// aus Ähnlichkeitsscore" steht nicht als Prüfung da, sondern als fehlendes
// Feld. Ein Rang ordnet, er bewertet nicht.
export interface KIBalanceDTO {
  // strategie sagt, **wie** die Vergleichskarten gefunden wurden — vom
  // Programm und nicht vom Modell: ein Modell, das seine eigene Suche
  // beschriebe, beschriebe eine erfundene.
  strategie: string

  // format ist leer, wenn keines genannt wurde — dann steht das in den
  // Einschränkungen.
  format: string

  staende: KorpusstandDTO[]

  vergleiche: KIVergleichDTO[]

  // nichtBetrachtet sind die Synergien, die außen vor blieben.
  nichtBetrachtet: string[]
}
