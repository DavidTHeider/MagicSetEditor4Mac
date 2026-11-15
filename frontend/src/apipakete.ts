// **Die Bindungen der Paketverwaltung** — herausgelöst, als `api.ts` mit dem
// Anlegen über die Zeilengrenze aus §2 ging
// (#1350).
//
// Der Schnitt folgt der Frage: hier steht, was mit **Vorlagenpaketen** geht —
// beziehen, einbinden, anlegen, entfernen und alles, was sie beschreibt.

import type {
  AbhaengigkeitenAntwort,
  Anlegbarkeit,
  BezugsquellenAntwort,
  OrteAntwort,
  PaketAntwort,
  PaketeAntwort,
  PaketplanAntwort,
  Paketweitergabe,
  UebernahmeAntwort,
} from './pakete'
import type { BezugAntwort, HerkunftAntwort } from './antworten'
import type { BezugsvorschauAntwort, PaketbezuegeAntwort } from './paketbezuege'
import type { Umschlag } from './umschlag'

// SchreibAntwort steht in `api.ts` und würde von hier aus einen Ring bilden.
type SchreibAntwort = Umschlag

export interface Paketapi {
  Bezugsquellen: () => Promise<BezugsquellenAntwort>
  PaketHerkunft: (adresse: string, stand: string) => Promise<HerkunftAntwort>
  PaketBeziehen: (
    adresse: string,
    stand: string,
    unterverzeichnis: string,
    name: string,
  ) => Promise<BezugAntwort>
  Abhaengigkeiten: () => Promise<AbhaengigkeitenAntwort>

  // **Die Stände der bezogenen Bäume** — G04 (#2251). Gefragt wird die
  // Quelle nur mit `fragen`; aktualisiert wird auf einen Commit, und der
  // bisherige Baum bleibt zum Zurückholen.
  PaketBezuege: (fragen: boolean) => Promise<PaketbezuegeAntwort>
  PaketBezugVorschau: (
    name: string,
    stand: string,
  ) => Promise<BezugsvorschauAntwort>
  PaketBezugAktualisieren: (
    name: string,
    stand: string,
    bestaetigt: boolean,
  ) => Promise<BezugAntwort>
  PaketBezugZurueck: (name: string) => Promise<BezugAntwort>
  MSE2Vorschlaege: () => Promise<OrteAntwort>
  VorlagenUebernehmen: (vorschlag: string) => Promise<UebernahmeAntwort>

  // PaketWeitergeben und InstallerEinlesen gehen über `.mse-installer` —
  // FMT-02.AK-04 (#1691). **Beide fragen selbst nach dem Ort**; das Paket
  // steht mit Endung da, etwa `probe.mse-game`.
  PaketWeitergeben: (paket: string) => Promise<Paketweitergabe>
  InstallerEinlesen: () => Promise<UebernahmeAntwort>
  Pakete: () => Promise<PaketeAntwort>

  // **Erst der Plan, dann die Änderung** — DOM-01.AK-02 (#1696). Der
  // Verzeichnisdialog gibt einen Plan und bindet noch nichts ein;
  // eingebunden oder ersetzt wird mit der Quelle aus dem Plan.
  PaketPlanWaehlen: () => Promise<PaketplanAntwort>
  PaketPlan: (
    vorgang: string,
    art: string,
    name: string,
    quelle: string,
  ) => Promise<PaketplanAntwort>
  PaketEinbindenVon: (quelle: string) => Promise<PaketAntwort>

  // PaketErsetzen legt eine neue Fassung ein; die bisherige bleibt als
  // vorige. Ein lokal geändertes Paket braucht `bestaetigt` (AK-03).
  PaketErsetzen: (quelle: string, bestaetigt: boolean) => Promise<PaketAntwort>
  PaketZurueck: (art: string, name: string) => Promise<PaketAntwort>

  // Paketbedarf nennt, was einem Set an Paketen fehlt — DOM-01.AK-04.
  Paketbedarf: (setID: string) => Promise<AbhaengigkeitenAntwort>

  // PaketAnlegen legt ein Paket aus einem Gerüst an (#1350).
  //
  // **Ein Stil braucht sein Spiel**: er legt dessen Felder aus, und ohne sie
  // zeichnete er eine leere Karte. Der Vollname darf leer bleiben — dann gilt
  // der Name.
  PaketAnlegen: (
    art: string,
    name: string,
    vollname: string,
    spiel: string,
  ) => Promise<PaketAntwort>

  // PaketKopieren legt eine eigene Kopie an — FMT-02.AK-04 (#1691). Ohne
  // `neu` heißt sie wie das Original mit »-kopie«.
  PaketKopieren: (
    art: string,
    name: string,
    neu: string,
  ) => Promise<PaketAntwort>

  // PaketArten nennt, was sich anlegen lässt — aus der Anwendung und nicht
  // aus einer Liste hier, die eine zweite Wahrheit wäre.
  PaketArten: () => Promise<Anlegbarkeit[]>

  PaketEntfernen: (art: string, name: string) => Promise<SchreibAntwort>
}
