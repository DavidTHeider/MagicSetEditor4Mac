// **Ein Set aufschlagen, speichern, auflisten** — der Zugang zu den Sets.
//
// Herausgelöst aus `api.ts`, als die Datei mit dem Diagnosepaket (Werkstück
// O10, #1434) über die Zeilengrenze aus §2 wuchs. **Der Schnitt folgt der
// Frage**: dort steht, was die Fassade insgesamt kann; hier, wie ein Set in
// dieses Fenster hinein- und wieder herauskommt.
import type {
  BeispielsetsAntwort,
  OeffnenAntwort,
  SetsAntwort,
  SpeicherAntwort,
  StandAntwort,
} from './antworten'
import type { SetvorschauAntwort } from './antwortenbestand'

export interface Setzugangapi {
  // SetVorschauWaehlen fragt im Dateidialog nach einem Set und sagt, was sein
  // Einfahren täte — ohne zu schreiben (FMT-03.AK-01, #1692). Eingelesen
  // wird danach über `SetAusDatei`.
  SetVorschauWaehlen: () => Promise<SetvorschauAntwort>
  Beispielsets: () => Promise<BeispielsetsAntwort>
  BeispielsetOeffnen: (pfad: string) => Promise<OeffnenAntwort>

  // Datenstand sagt, ob ein **anderer** Prozess geschrieben hat (#156).
  Datenstand: () => Promise<StandAntwort>

  SetSpeichern: (setID: string) => Promise<SpeicherAntwort>
  SetSpeichernUnter: (setID: string) => Promise<SpeicherAntwort>
  Sets: () => Promise<SetsAntwort>

  // ScryfallSatzEinlesen legt ein Set aus einem Satzkürzel an
  // (#1033).
  //
  // **Dieselbe Antwort wie beim MTG-JSON-Import**: das Fenster zeigt beides
  // an derselben Stelle, und zwei Gestalten für „ein Set ist entstanden"
  // wären zwei Wege durch dieselbe Oberfläche.
  ScryfallSatzEinlesen: (code: string, spiel: string) => Promise<OeffnenAntwort>
}
