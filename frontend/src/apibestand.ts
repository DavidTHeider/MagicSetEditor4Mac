// **Der Bestand als Gegenstand** — sichern, zurückspielen, und was beim Start
// geschah
// (#940,
// #962).
//
// **Eine eigene Datei**, wie `apisymbol.ts`, und aus demselben Grund: `api.ts`
// führt die ganze Fassade und liefe sonst über die Zeilengrenze (§2).

import type { OeffnenAntwort } from './antworten'
import type { Umschlag } from './umschlag'
import type {
  Rueckspielantwort,
  SetAnlegeAntwort,
  Sicherungsantwort,
  Sicherungsauswahl,
  SicherungsstandAntwort,
  SetLoeschAntwort,
  Startbefund,
  Verschluesselungsantwort,
} from './antwortenbestand'

// Bestandsapi ist der Weg zur Bestandsdatei selbst.
export interface Bestandsapi {
  // SetAnlegen legt ein **leeres** Set an und gibt seine Kennung (#891).
  //
  // **Nicht `SetOeffnen`**: jener liest eine Datei. Hier entsteht ein Set aus
  // nichts — der Weg, den MSE2 „Neu" nennt.
  //
  // `stil` ist wahlfrei; ohne ihn zeichnet das Set mit dem, den das Spiel
  // vorgibt.
  SetAnlegen: (
    name: string,
    spiel: string,
    stil: string,
  ) => Promise<SetAnlegeAntwort>

  // SetLoeschen nimmt ein Set aus dem Bestand — **endgültig** (#1218).
  //
  // **`name` ist die Bestätigung** und muss der heutige Name des Sets sein.
  // Stimmt er nicht, wird abgewiesen und nichts gelöscht — das fängt auch den
  // Fall ab, in dem jemand anderes das Set inzwischen umbenannt hat (§7,
  // Regel 2).
  SetLoeschen: (setID: string, name: string) => Promise<SetLoeschAntwort>

  // SetAusDatei liest ein Setpaket an einem **genannten** Pfad ein (#891).
  //
  // **Nicht `SetOeffnen`**: jene fragt im Dateidialog. Hier kommt der Pfad aus
  // einem Ablegen — dieselbe Handlung, ein anderer Weg. Ein zweiter Dialog
  // wäre die falsche Antwort auf eine Geste, die die Datei schon genannt hat.
  SetAusDatei: (pfad: string) => Promise<OeffnenAntwort>

  // SetBenutzt merkt sich, dass ein Set aufgeschlagen wurde (#891).
  //
  // **Daraus entsteht „Zuletzt benutzt"** — im Menü wie auf der Startseite.
  // Der Bestand ist die Liste: `sets.geoeffnet` trägt den Zeitpunkt, und die
  // Setliste sortiert danach. Eine zweite Liste wäre eine zweite Wahrheit.
  SetBenutzt: (setID: string) => Promise<Umschlag>

  // **Ohne Ziel**: gefragt wird im nativen Speichern-Dialog, und was dort
  // gewählt wird, weiß das Fenster nicht besser als macOS. Ein Abbruch gibt
  // einen leeren Pfad ohne Fehler — das unterscheidet „nichts getan" von
  // „getan". **Mit Kennwort und Klartextwunsch** seit OPS-06 (#1723): leer und
  // `false` heißt wie bisher, ein verschlüsselter Bestand verlangt eines von
  // beiden.
  BestandSichern: (
    kennwort: string,
    klartext: boolean,
  ) => Promise<Sicherungsantwort>

  // Verschluesselungsstand sagt, ob der Bestand verschlüsselt auf der Platte
  // liegt (#1351, Werkstück 14).
  Verschluesselungsstand: () => Promise<Verschluesselungsantwort>

  // VerschluesselungUmschalten schaltet sie an oder aus.
  //
  // **Ein Ruf und nicht zwei**: es ist ein Schalter, und zwei Knöpfe, von
  // denen immer einer nichts tut, sind einer zu viel. Die Abweisung bleibt
  // trotzdem — wer einschalten verlangt, während es schon an ist, geht gegen
  // einen Stand, den die Fläche noch nicht kennt (§7, Regel 4).
  VerschluesselungUmschalten: (an: boolean) => Promise<Verschluesselungsantwort>

  // SicherungWaehlen fragt nach einer Sicherung und liest, was darin steht —
  // **ohne** sie zu übernehmen
  // (#940).
  SicherungWaehlen: () => Promise<Sicherungsauswahl>

  // SicherungBeschauen liest eine schon gewählte, verschlüsselte Sicherung mit
  // ihrem Kennwort — der zweite Schritt nach `SicherungWaehlen` (OPS-06).
  SicherungBeschauen: (
    pfad: string,
    kennwort: string,
  ) => Promise<Sicherungsauswahl>

  // SicherungZurueckspielen ersetzt den Bestand durch die gewählte Sicherung.
  //
  // **Der Pfad kommt vom Aufrufer und nicht aus einem zweiten Dialog**: geprüft
  // wird derselbe, den `SicherungWaehlen` gezeigt hat — und noch einmal, denn
  // zwischen Auswahl und Bestätigung liegt die Entscheidung eines Menschen.
  SicherungZurueckspielen: (
    pfad: string,
    kennwort: string,
  ) => Promise<Rueckspielantwort>

  // Startbefund sagt, ob der Bestand offen ist — und wenn nicht, warum.
  //
  // **Sie wird gerade dann gerufen, wenn nichts aufging.** Bis zum 2026-08-25
  // beendete sich die Anwendung bei einem Datenbankfehler wortlos; es gab kein
  // Fenster, keine Bindung und keinen Knopf, und der Weg zur Sicherung war im
  // gescheiterten Fall nicht erreichbar
  // (#962).
  Startbefund: () => Promise<Startbefund>

  // StartWiederholen wiederholt einen gescheiterten Start und gibt den neuen
  // Befund — F05 (#2344): für belegt, voll und gesperrt, die von allein oder
  // durch einen Handgriff vergehen.
  StartWiederholen: () => Promise<Startbefund>

  // Sicherungsstand nennt das Alter der letzten Sicherung und den Plan —
  // OPS01 (#2347), dieselbe Antwort wie an den drei anderen Flächen.
  Sicherungsstand: () => Promise<SicherungsstandAntwort>

  // SicherungsplanEinrichten fragt im nativen Dialog nach einem Ordner und
  // richtet den Plan dort ein — **nur nach einer Ordnerwahl**; ein Abbruch
  // richtet nichts ein.
  SicherungsplanEinrichten: () => Promise<SicherungsstandAntwort>

  // SicherungsplanBeenden nimmt den Ordner weg; die Sicherungen darin bleiben.
  SicherungsplanBeenden: () => Promise<SicherungsstandAntwort>
}
