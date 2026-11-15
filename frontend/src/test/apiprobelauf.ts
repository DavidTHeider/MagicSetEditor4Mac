// Die stillen Antworten des Probelaufs — UXR-F7
// (#2015).
//
// **Aus demselben Grund wie [ohnePlatz]**: wer ein Ziel des Ausgabeassistenten
// zeichnet, fragt „Wie es wird" beim Aufschlagen und bei jeder Wahl — nicht,
// weil ein Test es meint, sondern weil die Fläche es tut. Ohne Beleg stünde in
// jedem Lauf über den Druckbogen oder die Exportvorlage ein zweiter
// Fehlerkasten neben dem, um den es geht.
//
// **Die häufigste Antwort und keine erfundene Besonderheit**: ein Set mit
// zehn Karten auf A4 und eine Vorlage, die eine kleine Datei ergibt. Wer den
// Probelauf **meint**, belegt ihn selbst und überschreibt diese hier.
import type {
  BogenprobelaufAntwort,
  VorlagenprobelaufAntwort,
} from '../antwortenprobelauf'

export const stillerBogenprobelauf: BogenprobelaufAntwort = {
  karten: 10,
  seiten: 2,
  spalten: 3,
  zeilen: 3,
  jeSeite: 9,
  aufLetzterSeite: 1,
  bildbreite: 744,
  seitenbreite: 210,
  seitenhoehe: 297,
  kartenbreite: 63,
  kartenhoehe: 88,
  anschnitt: 0,
  rand: 10,
  dpi: 300,
  schnittmarken: true,
  fehler: null,
}

export const stillerVorlagenprobelauf: VorlagenprobelaufAntwort = {
  karten: 2,
  dateiname: 'ausfuhr.txt',
  bytes: 120,
  beigaben: 0,
  beigabenBytes: 0,
  verzeichnis: '',
  geschaetzt: false,
  befunde: [],
  fehler: null,
}

export const ohneBogenprobelauf = () => Promise.resolve(stillerBogenprobelauf)

export const ohneVorlagenprobelauf = () =>
  Promise.resolve(stillerVorlagenprobelauf)
