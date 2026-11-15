// Die Antworten des **Probelaufs** — UXR-F7
// (#2015).
//
// Was der Druckbogen und die Exportvorlage ergäben, bevor etwas geschrieben
// wird. **Eine eigene Datei neben `antworten.ts` und `vorlagen.ts`**: beide
// stehen an der Zeilengrenze, und die Antworten dieser zwei Bindungen sind
// keine Antworten eines Schreibens.
import type { Umschlag } from './umschlag'

// BogenprobelaufAntwort sagt, wie ein Druckbogen würde.
export interface BogenprobelaufAntwort extends Umschlag {
  karten: number
  seiten: number

  // spalten × zeilen ist das Raster einer Seite, jeSeite sind ihre Plätze.
  spalten: number
  zeilen: number
  jeSeite: number

  // aufLetzterSeite sagt, wie viele Karten auf der letzten Seite stehen.
  aufLetzterSeite: number

  // bildbreite ist die Breite einer Karte in Bildpunkten.
  bildbreite: number

  // Maße in Millimetern.
  seitenbreite: number
  seitenhoehe: number
  kartenbreite: number
  kartenhoehe: number
  anschnitt: number
  rand: number

  dpi: number
  schnittmarken: boolean
}

// VorlagenprobelaufAntwort sagt, was eine Exportvorlage ergäbe.
export interface VorlagenprobelaufAntwort extends Umschlag {
  // karten sind die Karten, über die die Vorlage lief — nach der Auswahl.
  karten: number

  // dateiname und bytes sind die der Hauptdatei — **genau**: die Vorlage ist
  // ausgewertet.
  dateiname: string
  bytes: number

  // beigaben sind die Dateien daneben, beigabenBytes ihre bekannte Größe und
  // verzeichnis der Ordner dafür — leer, wenn die Vorlage keinen anlegt.
  beigaben: number
  beigabenBytes: number
  verzeichnis: string

  // geschaetzt sagt, dass beigabenBytes eine Untergrenze ist.
  geschaetzt: boolean

  befunde: string[]
}
