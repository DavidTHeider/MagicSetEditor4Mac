// **Die Kartenliste in vier Gestalten** —
// #891,
// Werkstück 5 von Phase Y.
//
// **Eine eigene Datei**, wie `apibestand.ts`, und aus demselben Grund: `api.ts`
// führt die ganze Fassade und liefe sonst über die Zeilengrenze (§2). Der
// Schnitt folgt der Frage: dort steht, **wie** das Frontend das Backend
// erreicht, hier die Wege zu **den Karten eines Sets**.

import type {
  EinzelkarteAntwort,
  KartenAntwort,
  KartenfeldwerteAntwort,
  KartenzeilenAntwort,
} from './antworten'
import type {
  Auszaehlantwort,
  Auszaehldimensionsantwort,
} from './antwortenauszaehlung'
import type { Autoersetzungsantwort } from './antwortenautoersetzung'
import type { SchreibAntwort } from './api'
import type { Umschlag } from './umschlag'
import type { Parameterhilfenantwort } from './antwortenparameter'
import type {
  Kartenskriptantwort,
  Kartenskriptlaufantwort,
} from './antwortenkartenskripte'
import type { FehlerDTO } from './formen'

// KartenbildAntwort trägt die kleinen Bilder der Liste, je Kartenkennung
// (#1166).
//
// **Eine Tabelle und keine Liste.** Eine Karte ohne Bild steht gar nicht
// darin — der Normalfall, und eine Liste voller leerer Zeichenketten wäre die
// teurere Art, dasselbe zu sagen.
export interface KartenbildAntwort {
  bilder: Record<string, string>
  fehler: FehlerDTO | null
}

// Kartenapi sind die vier Wege zu den Karten eines Sets.
//
// **Vier Gestalten, eine Fähigkeit.** Die Trennung ist eine Frage der Menge:
// an `Black.mse-set` mit 3.788 Karten sind es 4.761.182 gegen 245.641 Bytes
// JSON je Aufruf, gemessen am 2026-08-25.
export interface Kartenapi {
  // Kartenzeilen gibt die Karten eines Sets als **Liste** — ohne ihre
  // Feldwerte (#891).
  //
  // **Der Weg, den das Fenster beim Öffnen geht.** `Karten` daneben gibt
  // alles; bei 3.788 Karten sind das 4.761.182 Bytes JSON je Aufruf.
  Kartenzeilen: (setID: string) => Promise<KartenzeilenAntwort>

  // Kartenskripte nennt die `add cards script:`-Blöcke des Spielpakets
  // (#1337).
  Kartenskripte: (setID: string) => Promise<Kartenskriptantwort>

  // KartenskriptAusfuehren führt eines und hängt seine Karten ans Set — in
  // EINEM Zug: ein ⌘Z holt alle zusammen zurück.
  KartenskriptAusfuehren: (
    setID: string,
    name: string,
  ) => Promise<Kartenskriptlaufantwort>

  // Auszaehldimensionen nennt die `statistics dimension:`-Blöcke des
  // Spielpakets — wonach dieses Spiel auszählen lässt
  // (#1338).
  Auszaehldimensionen: (setID: string) => Promise<Auszaehldimensionsantwort>

  // Auszaehlen zählt das Set nach einer davon aus.
  Auszaehlen: (setID: string, name: string) => Promise<Auszaehlantwort>

  // Parameterhilfen nennt zu einem Schlüsselwort beide Menüs: was sich ins
  // Muster einfügen lässt und wie der Erinnerungstext darauf zeigt
  // (#1342).
  Parameterhilfen: (
    setID: string,
    wortID: string,
  ) => Promise<Parameterhilfenantwort>

  // Autoersetzungen nennt die Regeln, die beim Tippen greifen — die des
  // Spielpakets und die eigenen
  // (#1341).
  Autoersetzungen: (setID: string) => Promise<Autoersetzungsantwort>

  // AutoersetzungSetzen legt eine Regel an oder schreibt sie fort.
  AutoersetzungSetzen: (
    setID: string,
    muster: string,
    ersatz: string,
    ganzesWort: boolean,
    aktiv: boolean,
  ) => Promise<SchreibAntwort>

  // AutoersetzungLoeschen nimmt eine **eigene** Regel heraus.
  AutoersetzungLoeschen: (
    setID: string,
    muster: string,
  ) => Promise<SchreibAntwort>

  // AutoersetzungenZuruecksetzen vergisst alle Abweichungen dieses Spiels.
  AutoersetzungenZuruecksetzen: (setID: string) => Promise<SchreibAntwort>

  // AutoersetzungSchalten legt den globalen Schalter dieses Spiels um.
  AutoersetzungSchalten: (setID: string, an: boolean) => Promise<SchreibAntwort>

  // Karte gibt **eine** Karte samt ihren Feldwerten (#891).
  Karte: (karteID: string) => Promise<EinzelkarteAntwort>

  // Kartenfeldwerte gibt **ein** Feld über alle Karten eines Sets (#891).
  //
  // **Für die Massenänderung**, die neben jeder angehakten Karte den Wert
  // zeigt, um den es geht.
  Kartenfeldwerte: (
    setID: string,
    feld: string,
  ) => Promise<KartenfeldwerteAntwort>

  Karten: (setID: string) => Promise<KartenAntwort>

  // Kartenbilder gibt die kleinen Bilder für die Liste (#1166).
  //
  // **Ein Stapel je Sichtfenster** und kein Ruf je Zeile: bei dreißig
  // sichtbaren Zeilen wären das dreißig Wege über die Naht für eine Bewegung
  // des Rollbalkens.
  Kartenbilder: (
    setID: string,
    karteIDs: string[],
  ) => Promise<KartenbildAntwort>
}

// FeldAntwort ist die Antwort auf ein geschriebenes Kartenfeld (#2535):
// **`gespeichert`** ist der Wert, der danach im Feld steht — seit #2396 der
// gerechnete —, und `null`, wo das Feld nicht selbst im Bestand steht (ein
// Verbund). Ein älterer Stand schickt es gar nicht.
export interface FeldAntwort extends Umschlag {
  gespeichert?: string | null
}
