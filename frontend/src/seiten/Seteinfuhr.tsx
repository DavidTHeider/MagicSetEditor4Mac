// **Ein Set von woanders** — die Wege, die nicht über eine Datei gehen.
//
// # Wo sie stehen
//
// **Eingeklappt unter dem neuen Set**, in der Spalte neben der Liste — nach
// dem Artboard *Sets* („Von woanders holen"), UXR-04
// (#1912).
// Sie sind der seltenere Fall: wer ein Set von woanders holt, tut es einmal
// je Set; wer die Seite aufschlägt, sucht meistens ein vorhandenes. Offen
// ist das Neue, nicht beides (R3 aus `docs/UIREVIEW.md`): zusammen stünden
// sonst MTG JSON, Scryfall und der Assistent mit bis zu vierzehn Feldern da.
//
// Bis Werkstück C von Phase 4.10
// (#1325)
// standen sie **über** der Liste und schoben das erste Set unter die Falz;
// danach unter ihr, bis UXR-04.
//
// # Warum eine eigene Datei
//
// `Sets.tsx` beantwortet, *welche Sets es gibt* und wie eines aufgeht; hier
// steht, *wie eines von woanders hereinkommt* (§2). Das Einlesen selbst
// läuft über den Öffnungshaken der Seite — einer für jede Quelle, seit
// FMT-03.
import { api, type OeffnenAntwort } from '../api'
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { type Umschlaglauf } from '../haken/useUmschlag'
import { useTexte } from '../sprache/kontext'
import { Beispielsets } from './Beispielsets'
import { MTGJSONOeffnen } from './MTGJSONOeffnen'
import { ScryfallSatz } from './ScryfallSatz'
import { Umschlagannahme } from './Umschlagannahme'

interface SeteinfuhrEigenschaften {
  laeuft: boolean

  // oeffnen liest über den Haken der Seite ein — der kennt Teilerfolg,
  // Befunde und das Nachladen der Liste.
  oeffnen: (lesen: () => Promise<OeffnenAntwort>) => void

  // umschlag ist der Haken der **Seite** — derselbe, mit dem die Setkarten
  // ablegen (§7, Regel 4).
  umschlag: Umschlaglauf
}

export function Seteinfuhr({
  laeuft,
  oeffnen,
  umschlag,
}: SeteinfuhrEigenschaften) {
  const texte = useTexte()

  return (
    <Aufklappkasten titel={texte.seiten.sets.einlesenTitel} schreibt>
      {/* **Was mitkommt, steht vor dem, was man erst besorgen muss**
          (#702). */}
      <Beispielsets
        laeuft={laeuft}
        oeffnen={(pfad) => {
          oeffnen(() => api().BeispielsetOeffnen(pfad))
        }}
      />

      <MTGJSONOeffnen
        laeuft={laeuft}
        oeffnen={(verzeichnis, spiel) => {
          oeffnen(() =>
            verzeichnis
              ? api().MTGJSONVerzeichnisOeffnen(spiel)
              : api().MTGJSONOeffnen(spiel),
          )
        }}
      />

      {/* **Der dritte Weg zu einem Set** (#1033): kein Dateidialog, nur ein
          Satzkürzel. Er steht nach MTG JSON, weil er ein Netzgang ist. */}
      <ScryfallSatz
        einlesen={(code, spiel) => {
          oeffnen(() => api().ScryfallSatzEinlesen(code, spiel))
        }}
        laeuft={laeuft}
      />

      {/* **Der vierte Weg** (#1457): ein Umschlag ist ein Set, das jemand
          anderes abgelegt hat. Er steht zuletzt, weil er den seltensten
          Fall bedient — und weil er als einziger eine ausdrückliche
          Übernahme verlangt. */}
      <Umschlagannahme umschlag={umschlag} />
    </Aufklappkasten>
  )
}
