// Was ein **Kartenfeld** über sich hinaus anbietet
// (#1062).
//
// **Eine eigene Datei, wie `apibilder.ts` und `apiscryfall.ts`** — und aus
// demselben Grund: `api.ts` führt die ganze Fassade und liefe sonst über die
// Zeilengrenze (§2).

import type {
  BefehlsvorschauAntwort,
  Farbantwort,
  KonsolenAntwort,
  Namensantwort,
  PackAntwort,
  PackziehungAntwort,
  Paketdateiantwort,
  Paketdateiinhalt,
  Paketdateistand,
  PaketsymboleAntwort,
  ProbekartenAntwort,
  SchalterAntwort,
  SchnittstellenAntwort,
  SchnittstellenAlleAntwort,
  NetzschnittstellenAntwort,
  Skriptbefundantwort,
  SymbolpaletteAntwort,
  SymbolverwendungenAntwort,
  TechnikAntwort,
  VorschauAntwort,
} from './antworten'
import type { Umschlag } from './api'
import type { Paketdateieinfuegung } from './antwortenvorlagen'

// Feldapi ist der Weg zu dem, was ein Feld zum Einfügen anbietet.
export interface Feldapi {
  // Symbolpalette gibt, was ein Feld zum Einfügen anbietet — Manasymbole,
  // Trenner und Untermenüs, je mit gerastertem Bild
  // (#564).
  //
  // **An der Karte und nicht am Set**: welche Symbolschrift ein Feld benutzt,
  // darf ein Skript sein, und die Bedingungen der Symbolfassungen lesen die
  // Karte.
  Symbolpalette: (
    karteID: string,
    feld: string,
  ) => Promise<SymbolpaletteAntwort>

  // Skript wertet einen MSE-Script-Ausdruck im Kontext einer Karte aus
  // (#1070).
  //
  // **Es ändert nichts.** MSE-Script kann in diesem Programm nicht schreiben —
  // `internal/schichten` weist jeden Zugriff auf das Dateisystem aus
  // `msescript` ab.
  //
  // **Eine Karte ist Pflicht**: ohne sie gibt es kein `card`, und ein
  // erfundenes wäre eine Umgebung, die niemand nachvollziehen kann.
  Skript: (karteID: string, ausdruck: string) => Promise<KonsolenAntwort>

  // SkriptAbbrechen beendet den laufenden Ausdruck an dieser Karte —
  // Werkstück U13 von Phase 5.1 (#1424). Die Antwort des Laufs sagt dann
  // `abgebrochen`; ein Abbruch ohne Lauf ist nichts und kein Fehler.
  SkriptAbbrechen: (karteID: string) => Promise<Umschlag>

  // Skriptnamen gibt, was ein Ausdruck an dieser Karte nennen darf
  // (#1164).
  //
  // **Aus der wirklichen Umgebung und nicht aus einer Liste daneben**: was
  // die Vorlage dazulegt, steht mit darin. Für jemanden, der MSE-Script nicht
  // kennt, ist ein leeres Feld keine Einladung.
  Skriptnamen: (karteID: string) => Promise<Namensantwort>

  // Skriptfarben zerlegt einen Ausdruck in seine Merkmale
  // (#1164).
  //
  // **Aus dem Lexer, der auch zerteilt.** Ein zweiter in TypeScript liefe an
  // dem Tag auseinander, an dem einer von beiden eine Regel lernt — und die
  // Einfärbung zeigte eine Sprache, die es nicht gibt.
  Skriptfarben: (quelltext: string) => Promise<Farbantwort>

  // Paketfarben zerlegt eine **Paketdatei** in ihre Merkmale
  // (#1181).
  //
  // **Eine zweite Bindung und kein Parameter**: es sind zwei Sprachen, und ein
  // Schalter darin wäre eine fachliche Verzweigung in `internal/app` (§2).
  Paketfarben: (quelltext: string) => Promise<Farbantwort>

  // Technikangaben nennen den Ort der Werkzeuge ohne Fenster und fertige
  // Befehlszeilen (#1164).
  //
  // **Zum Kopieren und nicht zum Ausführen.** Die Zeilen entstehen in Go, weil
  // sie Pfade tragen, die nur dort bekannt sind.
  Technikangaben: (setID: string) => Promise<TechnikAntwort>

  // Befehlszeile richtet `mse` und `mse-mcp` in `~/.local/bin` ein — leer
  // liest nur, »installieren« und »entfernen« schreiben (OPS-03, #1720).
  Befehlszeile: (handlung: string) => Promise<BefehlsvorschauAntwort>

  // Schnittstellen liest, ob die Kommandozeile im PATH der Anmeldeshell
  // steht und mse-mcp bei Codex; einrichten und widerrufen schreiben einen
  // markierten Block (OPS-09, #1907).
  Schnittstellen: () => Promise<SchnittstellenAntwort>
  SchnittstelleEinrichten: (name: string) => Promise<SchnittstellenAntwort>
  SchnittstelleWiderrufen: (name: string) => Promise<SchnittstellenAntwort>

  // **Alle mit einem Ruf** — eingerichtet, was fehlt; zurückgenommen, was
  // eingerichtet ist; ein fremder Eintrag bleibt (G08, #2272).
  SchnittstellenAlleEinrichten: () => Promise<SchnittstellenAlleAntwort>
  SchnittstellenAlleWiderrufen: () => Promise<SchnittstellenAlleAntwort>

  // Netzschnittstellen nennt die Schnittstellen dieses Rechners — für die
  // Wahl in `netz.lauschen` (G08).
  Netzschnittstellen: () => Promise<NetzschnittstellenAntwort>

  // Schalter nennt die Bereiche und Funktionen samt Reife und Stand (OPS-10,
  // #1908) — geschaltet wird über `EinstellungSetzen`.
  Schalter: () => Promise<SchalterAntwort>

  // Vorlagenvorschau zeichnet eine Karte mit einer **ungespeicherten** Vorlage
  // (#1164).
  //
  // **Welche Karte, entscheidet Go** — die erste des ersten Sets, das diesen
  // Stil benutzt —, **oder die Werkstatt nennt eine** (Werkstück U7,
  // #1418):
  // `karteID` ist leer oder die Kennung einer Testkarte aus `Probekarten`.
  Vorlagenvorschau: (
    paket: string,
    datei: string,
    inhalt: string,
    vergroesserung: number,
    karteID: string,
  ) => Promise<VorschauAntwort>

  // Probekarten nennt die Testkarten zu einem Paket — die Karten des ersten
  // Sets, das seinen Stil benutzt (U7, #1418).
  Probekarten: (paket: string) => Promise<ProbekartenAntwort>

  // Paketsymbole nennt jede Definition in den Dateien eines Pakets — für
  // Vervollständigung, Hover und den Sprung zur Definition (U7, #1418).
  Paketsymbole: (paket: string) => Promise<PaketsymboleAntwort>

  // Symbolverwendungen nennt jede Zeile, die einen Namen nennt (U7, #1418).
  Symbolverwendungen: (
    paket: string,
    name: string,
  ) => Promise<SymbolverwendungenAntwort>

  // Paketdateien nennt die bearbeitbaren Dateien eines installierten Pakets
  // (#1164).
  Paketdateien: (paket: string) => Promise<Paketdateiantwort>

  // PaketdateiLesen gibt eine Paketdatei samt ihrem **Stand**.
  //
  // Der Stand geht beim Speichern zurück; stimmt er dann nicht mehr, hat
  // jemand anderes geschrieben — und das Fenster erfährt es, statt die fremde
  // Änderung zu überschreiben (§7, Regel 2).
  PaketdateiLesen: (paket: string, datei: string) => Promise<Paketdateiinhalt>

  // PaketdateiPruefen gibt den ersten Syntaxfehler in den Skripten eines
  // Paketdateitextes — Werkstück B von Phase 4.9
  // (#1324).
  //
  // **Geprüft wird der übergebene Text**, nicht die installierte Datei: eine
  // Änderung soll sich prüfen lassen, bevor sie geschrieben wird.
  PaketdateiPruefen: (inhalt: string) => Promise<Skriptbefundantwort>

  // PaketdateiSchreiben setzt eine Paketdatei zurück, wenn der Stand stimmt.
  PaketdateiSchreiben: (
    paket: string,
    datei: string,
    inhalt: string,
    stand: string,
  ) => Promise<Paketdateistand>

  // **Der Dateibaum** (FMT-02, #1691): eine Vorlage anlegen, eine Datei aus
  // dem Dialog einfügen — das Fenster reicht keinen Pfad, nur den Ordner im
  // Paket —, umbenennen und entfernen. Eine Vorlage nur mit ihrem Stand.
  PaketdateiAnlegen: (
    paket: string,
    datei: string,
    inhalt: string,
  ) => Promise<Paketdateistand>
  PaketdateiEinfuegen: (
    paket: string,
    ordner: string,
  ) => Promise<Paketdateieinfuegung>
  PaketdateiUmbenennen: (
    paket: string,
    alt: string,
    neu: string,
    stand: string,
  ) => Promise<Umschlag>
  PaketdateiEntfernen: (
    paket: string,
    datei: string,
    stand: string,
  ) => Promise<Umschlag>

  // Packarten gibt die Packarten des Spiels hinter einem Set
  // (#1072).
  Packarten: (setID: string) => Promise<PackAntwort>

  // PackZiehen zieht einen Booster.
  //
  // **Der Same als Text und nicht als Zahl**: eine JavaScript-Zahl verliert
  // oberhalb von 2⁵³ die letzten Stellen, und ein Same, der beim Hinweg schon
  // ein anderer ist, wäre keine Wiederholbarkeit.
  PackZiehen: (
    setID: string,
    art: string,
    same: string,
  ) => Promise<PackziehungAntwort>
}
