// Die Bildhälfte der Fassade — ausgeben, einsetzen, zuschneiden.
//
// **Eigene Datei, weil `api.ts` mit ihr über 400 Zeilen ginge** (§2). Der
// Schnitt folgt der Frage: dort steht, was die Anwendung mit **Daten** tut,
// hier, was sie mit **Bildern** tut — die Ausgabe eines Sets als Bilder
// (#491) und
// der Weg eines Bildes in ein Kartenfeld
// (#246,
// #576).
import type { AuswahlDTO, ZuschnittAntwort } from './formen'
import type { BildAntwort, BildnamenAntwort, OrdnerAntwort } from './antworten'
import type { SchreibAntwort } from './api'
import type {
  KIBildvorschlagAntwort,
  KIBildherkunftAntwort,
} from './antwortenbild'

export interface Bilderapi {
  Bildordner: () => Promise<OrdnerAntwort>
  Bildnamen: (
    setID: string,
    vorlage: string,
    format: string,
    verhalten: string,

    // auswahl sind die Kartenkennungen — **leer heißt: das ganze Set**
    // (#1167).
    // MSE2 stellt die Wahl vor der Bildausgabe genauso
    // (`SetWindow::onFileExportImages`).
    auswahl: string[],
  ) => Promise<BildnamenAntwort>
  BildSchreiben: (
    karteID: string,
    verzeichnis: string,
    datei: string,
    format: string,
    breite: number,
    verhalten: string,
    guete: string,
  ) => Promise<BildAntwort>
  // BildSetzen fragt nach einer Bilddatei und setzt sie in ein Feld
  // (#246).
  //
  // **Der Dialog steht in Go**: ein Fenster kann keine Datei vom Dateisystem
  // lesen — es bekäme einen Namen und keinen Pfad.
  //
  // **Es schreibt nicht.** Seit
  // #576 liegt
  // das Zuschnittfenster dazwischen: hier wird gewählt und gelesen, abgelegt
  // wird über `BildEinsetzen`.
  BildSetzen: (karteID: string, feld: string) => Promise<ZuschnittAntwort>

  // BildAusDatei setzt ein Bild ein, dessen Pfad schon feststeht — beim Ziehen
  // und Ablegen (#246).
  //
  // **Der Pfad kommt aus Wails**, nicht aus dem Browser: der sieht bei einem
  // Ablegen nur einen Namen.
  BildAusDatei: (
    karteID: string,
    feld: string,
    pfad: string,
  ) => Promise<ZuschnittAntwort>

  // BildEinsetzen schneidet auf die gewählte Fläche zu und legt das Bild ab
  // (#576).
  BildEinsetzen: (
    setID: string,
    karteID: string,
    feld: string,
    pfad: string,
    auswahl: AuswahlDTO,
  ) => Promise<SchreibAntwort>

  // **Der Bildvorschlag** (A12, #1451). Fünf Wege, weil ein Bild vier
  // Zustände durchläuft: erzeugen, zuschneiden, einsetzen — oder verwerfen.
  // Dazu die Auskunft, was über ein übernommenes feststeht.
  KIBildVorschlagen: (
    karteID: string,
    feld: string,
    anbieter: string,
    alias: string,
    aufgabe: string,
    groesse: string,
    guete: string,
    bearbeiten: boolean,

    // nachweis ist die Freigabe eines Menschen — leer heißt: keine (#1539).
    // **Ein Bildauftrag braucht eine eigene**: die Freigabe für Text gilt
    // dafür nicht, und ein Bild kostet mehr.
    nachweis: string,
  ) => Promise<KIBildvorschlagAntwort>
  KIBildVerwerfen: (kennung: string) => Promise<SchreibAntwort>

  // KIBildAbbrechen hält den laufenden Bildauftrag einer Karte an — die
  // Rückgabe sagt, ob einer lief, und ist kein Fehler, wenn keiner lief
  // (A17, #1532).
  KIBildAbbrechen: (karteID: number) => Promise<boolean>
  KIBildZuschnitt: (
    karteID: string,
    feld: string,
    kennung: string,
  ) => Promise<ZuschnittAntwort>
  KIBildEinsetzen: (
    setID: string,
    karteID: string,
    feld: string,
    kennung: string,
    auswahl: AuswahlDTO,
  ) => Promise<SchreibAntwort>
  KIBildHerkunft: (
    karteID: string,
    feld: string,
  ) => Promise<KIBildherkunftAntwort>

  // Zuschnittgrenze gibt die Auswahl, die MSEs Regeln aus einer gezogenen
  // machen (#576).
  //
  // **Sie liegt in Go, weil die Regel portiert ist** (`ImageSlice::constrain`).
  // Eine zweite Fassung hier wäre eine zweite Antwort auf dieselbe Frage.
  Zuschnittgrenze: (
    quelleBreite: number,
    quelleHoehe: number,
    zielBreite: number,
    zielHoehe: number,
    auswahl: AuswahlDTO,
    frei: boolean,
    bevorzugt: string,
  ) => Promise<AuswahlDTO>

  // Zuschnittziehen gibt die Auswahl, nachdem der Rahmen gezogen wurde
  // (#576).
  //
  // **Die Strecke in Punkten der Quelle**, nicht des Bildschirms: den Maßstab
  // kennt nur das Fenster, die Regel steht in Go
  // (`ImageSliceSelector::onMotion`).
  //
  // **kanteX und kanteY sind −1, 0 oder 1** — beide null heißt verschieben.
  Zuschnittziehen: (
    quelleBreite: number,
    quelleHoehe: number,
    zielBreite: number,
    zielHoehe: number,
    auswahl: AuswahlDTO,
    frei: boolean,
    x: number,
    y: number,
    kanteX: number,
    kanteY: number,
  ) => Promise<AuswahlDTO>
}
