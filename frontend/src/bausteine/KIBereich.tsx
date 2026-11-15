// **Der KI-Teil des Karteneditors** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Ein eigenes Bauteil
//
// **Weil `Karteneditor.tsx` mit ihm über die Zeilengrenze aus §2 ginge** —
// und weil der Schnitt der Frage folgt: dort steht das **Formular**, hier,
// was die KI dazu beiträgt.
//
// **Und weil beides zusammengehört**: der Vorschlag schreibt, die Herkunft
// liest, was geschrieben wurde. Wer den Zähler dazwischen an zwei Stellen
// pflegte, hätte eine Anzeige, die nach der Übernahme den Stand von vorhin
// zeigt.
import { useState } from 'react'

import { KIBild } from './KIBild'
import { AehnlicheKarten } from './AehnlicheKarten'
import { KIEinschaetzung } from './KIEinschaetzung'
import { KIHerkunft } from './KIHerkunft'
import { KIBatchkasten } from './KIBatchkasten'
import { KIVorschlag } from './KIVorschlag'

interface KIBereichEigenschaften {
  karteID: string

  // setID ist das Set dieser Karte — der Batchkasten liest daraus seine
  // Auswahl (A13.1b, #1452).
  setID: string

  // bildfelder sind die Bildfelder dieser Karte (A12, #1451) — leer heißt:
  // die Vorlage hat keine Illustration, und dann gibt es den Bildkasten
  // nicht.
  bildfelder: string[]

  // zuschneiden öffnet das vorhandene Zuschnittfenster für einen wartenden
  // Bildvorschlag ([useBildfeld.vonKI]).
  zuschneiden: (feld: string, kennung: string) => Promise<boolean>

  // bildstand zählt, was über den Bildweg geschrieben wurde — danach liest
  // der Bildnachweis neu.
  bildstand: number

  // geaendert meldet dem Editor, dass er neu laden soll.
  geaendert: () => void
}

// KIBereich zeigt Vorschlag und Herkunft zu einer Karte.
export function KIBereich({
  karteID,
  setID,
  bildfelder,
  zuschneiden,
  bildstand,
  geaendert,
}: KIBereichEigenschaften) {
  // uebernahmen zählt die Übernahmen dieser Sitzung — nach jeder wird die
  // Herkunft neu gelesen.
  const [uebernahmen, setzeUebernahmen] = useState(0)

  return (
    <>
      <KIVorschlag
        karteID={karteID}
        geaendert={() => {
          setzeUebernahmen((bisher) => bisher + 1)
          geaendert()
        }}
      />

      {/* **Der Bildkasten steht zwischen beiden** (A12, #1451): er schreibt
          wie der Vorschlag, und was er schreibt, liest die Herkunft. */}
      <KIBild
        karteID={karteID}
        bildfelder={bildfelder}
        zuschneiden={zuschneiden}
        stand={bildstand}
      />

      {/* **Die Einschätzung schreibt nie** (A13.2a, #1452) — sie steht
          deshalb unter dem, was schreibt, und über der Herkunft: sie ist
          eine Auskunft über die Karte und keine Änderung an ihr. */}
      <KIEinschaetzung karteID={karteID} />

      {/* **Der Batch steht unter der Einschätzung** — er meint nicht diese
          Karte, sondern mehrere, und gehört deshalb ans Ende des Bereichs. */}
      <KIBatchkasten setID={setID} geaendert={geaendert} />

      <KIHerkunft karteID={karteID} stand={uebernahmen} />

      {/* **Und die Ähnlichkeit** (A16, #1515) — sie fragt den lokalen
          Wissensbestand und ruft **kein** Modell. Sie steht hier, weil der
          Bereich der Ort ist, an dem eine Karte über sich hinaus fragt; das
          Issue verbietet einen eigenen Hauptbereich dafür. */}
      <AehnlicheKarten karteID={karteID} />
    </>
  )
}
