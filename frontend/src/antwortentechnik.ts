// Die Antworten der Technikseite und der Diagnose.
//
// **Aus `antworten.ts` herausgelöst**, als die Datei mit der
// Vorlagenwerkstatt (#1418)
// über die Zeilengrenze aus §2 ging. `./antworten` reicht sie weiter.
import type { FehlerDTO } from './formen'
import type { Umschlag } from './umschlag'

export interface Werkzeugzeile {
  schluessel: string
  befehl: string
}

// TechnikAntwort trägt den Ort der Werkzeuge und die Zeilen dazu (#1164).
export interface TechnikAntwort extends Umschlag {
  ort: string

  // vorhanden nennt die Werkzeuge, die dort **wirklich** liegen — gezählt,
  // nicht behauptet.
  vorhanden: string[]

  zeilen: Werkzeugzeile[]

  // agenten sind fertige Konfigurationsbruchstücke für Claude Code und Codex
  // (A14, #1453).
  agenten: AgentenbruchstueckDTO[]
}

// AgentenbruchstueckDTO ist ein fertiges Stück Konfiguration für einen
// Agenten — kopierbar, ausführbar und **ohne Geheimnis**.
//
// **Geschrieben wird nichts**: `mse-mcp` liest seine Zugangsdaten selbst aus
// dem Schlüsselbund, und ein Schlüssel in einer Konfigurationsdatei läge dort
// im Klartext.
export interface AgentenbruchstueckDTO {
  // agent ist `claude` oder `codex` — die Marke, unter der das Fenster seinen
  // Namen nachschlägt (§9).
  agent: string

  // datei ist der Ort, an den das Bruchstück gehört.
  datei: string

  // form ist `json` oder `toml`.
  form: string

  inhalt: string

  // befehl ist der Weg über die Kommandozeile des Agenten — dieselbe Wirkung
  // ohne eine Datei von Hand.
  befehl: string

  // nachgelesen nennt Quelle und Abrufdatum. Eine Konvention aus dem
  // Gedächtnis ist kein Prüfergebnis (§10).
  nachgelesen: string
}

// DiagnoseAntwort trägt den Text für einen Fehlerbericht (#1166).
export interface DiagnoseAntwort extends Umschlag {
  text: string
}

// Befehlseintrag ist ein Befehl der Kommandozeile im Zielverzeichnis —
// OPS-03 (#1720). `stand` ist frei, eigen, veraltet oder fremd.
export interface Befehlseintrag {
  name: string
  ziel: string
  quelle: string
  stand: string

  // verdeckt nennt einen gleichnamigen Befehl, der im PATH vorn steht.
  verdeckt: string
}

// BefehlsvorschauAntwort sagt, wie die Kommandozeile eingerichtet ist.
export interface BefehlsvorschauAntwort extends Umschlag {
  verzeichnis: string
  imPfad: boolean
  pfadzeile: string
  pruefbefehl: string
  eintraege: Befehlseintrag[]
}

// SchnittstelleDTO ist ein Ziel samt Stand und den Zeilen, die einrichten
// hinzufügt (OPS-09, #1907).
export interface SchnittstelleDTO {
  name: string

  // art ist `''` (ein Block), `json`, `befehl` oder `agent` (OPS-09b, #1961).
  art: string
  stand: string
  datei: string
  zeilen: string[]
}

// SchnittstellenAntwort ist der Stand aller Ziele.
export interface SchnittstellenAntwort extends Umschlag {
  schnittstellen: SchnittstelleDTO[]

  // nurEintrag nennt das Ziel, dessen Datei nach dem Widerrufen nicht wieder
  // ist wie vorher — nur der eigene Eintrag ging (OPS-09b, #1961).
  nurEintrag: string
}

// SchnittstellenausgangDTO ist, was einem Ziel beim Einrichten oder
// Zurücknehmen aller geschah (G08, #2272): `eingerichtet`, `widerrufen`,
// `fremd`, `unveraendert`, `gescheitert` (dann mit `fehler`), `abgebrochen`.
export interface SchnittstellenausgangDTO {
  name: string
  ausgang: string
  fehler: FehlerDTO | null
}

// SchnittstellenAlleAntwort ist der Stand aller Ziele danach samt Ausgängen.
export interface SchnittstellenAlleAntwort extends Umschlag {
  schnittstellen: SchnittstelleDTO[]
  ausgaenge: SchnittstellenausgangDTO[]
}

// NetzschnittstelleDTO ist eine Schnittstelle dieses Rechners — zum Wählen,
// wo die eigenen Dienste lauschen (G08). `art` ist `schleife`, `netz`,
// `tunnel`, `bruecke`, `direkt` oder leer.
export interface NetzschnittstelleDTO {
  name: string
  art: string
  aktiv: boolean
  adressen: string[]
  befund: string
}

// NetzschnittstellenAntwort trägt die Liste.
export interface NetzschnittstellenAntwort extends Umschlag {
  schnittstellen: NetzschnittstelleDTO[]
}

// SchalterDTO ist ein Schalter mit Reife und Stand (OPS-10, #1908).
export interface SchalterDTO {
  kennung: string
  art: string
  reife: string
  vorgabe: string
  an: boolean
}

// SchalterAntwort ist das Verzeichnis der Schalter.
export interface SchalterAntwort extends Umschlag {
  schalter: SchalterDTO[]
}
