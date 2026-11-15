// Die Bindungen der Stilauswahl.
//
// **Eine eigene Datei**, wie `apiliste.ts`: `api.ts` ging mit ihnen über 400
// Zeilen (§2). Der Schnitt folgt der Frage — dort steht, **was** die Anwendung
// insgesamt anbietet, hier, was sie über den **Stil eines Sets** anbietet.
import type { SchreibAntwort } from './api'
import type {
  FolgenAntwort,
  StileAntwort,
  StiloptionenAntwort,
  WahlbilderAntwort,
  ZusatzfelderAntwort,
} from './stilwahl'

export interface Stilapi {
  Stile: (setID: string) => Promise<StileAntwort>
  Stilfolgen: (setID: string, paket: string) => Promise<FolgenAntwort>
  StilWechseln: (setID: string, paket: string) => Promise<SchreibAntwort>
  Stiloptionen: (setID: string) => Promise<StiloptionenAntwort>
  StiloptionSetzen: (
    setID: string,
    feld: string,
    wert: string,
  ) => Promise<SchreibAntwort>
  StiloptionZuruecksetzen: (
    setID: string,
    feld: string,
  ) => Promise<SchreibAntwort>

  // Stilwahlbilder gibt zu einer Stiloption je Wahl ihr Bild
  // (#1225).
  //
  // **Je Option und nicht je Stil**: über das ganze Pack wären es 1.097
  // Bildblöcke. Gefragt wird nur dort, wo `hatWahlbilder` es sagt.
  Stilwahlbilder: (setID: string, option: string) => Promise<WahlbilderAntwort>

  // Kartenzusatzfelder gibt die Zusatzfelder des Stils **einer Karte**, die
  // sich von Hand setzen lassen, samt Wert
  // (#2298).
  Kartenzusatzfelder: (
    setID: string,
    karteID: string,
  ) => Promise<ZusatzfelderAntwort>
  ZusatzwertSetzen: (
    setID: string,
    karteID: string,
    feld: string,
    wert: string,
  ) => Promise<SchreibAntwort>
}
