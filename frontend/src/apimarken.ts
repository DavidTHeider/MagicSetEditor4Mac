// Die Bindungen der Marken an Karten
// (#1208).
//
// **Eine eigene Datei**, wie `apiliste.ts` und `apistil.ts`: `api.ts` steht an
// der Zeilengrenze aus §2, und der Schnitt folgt der Frage — dort steht,
// **was** die Anwendung insgesamt anbietet, hier, was sie über **Marken**
// anbietet.
//
// # Warum es hier nur zwei Wege gibt und nicht vier
//
// **Weil das Fenster die Marken schon hat**: sie reisen mit den Zeilen der
// Kartenliste (`ListenzeileDTO.marken`). Eine Bindung, die sie ein zweites Mal
// holte, wäre ein zweiter Stand derselben Sache (§7, Regel 2). Die drei
// Flächen ohne Fenster lesen sie sehr wohl — dort gibt es keine Liste, aus der
// sie mitkämen.
import type { SchreibAntwort } from './api'

export interface Markenapi {
  // MarkeSetzen hängt eine Marke an eine Karte — zweimal dieselbe ist kein
  // Fehler, sondern nichts.
  MarkeSetzen: (
    setID: string,
    karteID: string,
    marke: string,
  ) => Promise<SchreibAntwort>

  // MarkeEntfernen nimmt eine Marke wieder ab — eine, die gar nicht dasteht,
  // wird abgewiesen.
  MarkeEntfernen: (
    setID: string,
    karteID: string,
    marke: string,
  ) => Promise<SchreibAntwort>
}
