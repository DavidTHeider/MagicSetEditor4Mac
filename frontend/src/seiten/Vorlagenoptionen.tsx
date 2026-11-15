// **Die Optionen einer Exportvorlage — vier vorn, der Rest dahinter** —
// UXR-09 (#1917),
// R3 aus `docs/UIREVIEW.md`.
//
// # Warum es sie gibt
//
// **Gezählt am 2026-10-02** in `full-magic-pack`: `magic-simple-exporter`
// und `magic-symbol-extractor` erklären je **zwölf** Optionen. Mit der Wahl
// der Vorlage und der Kartenwahl darüber waren das fünfzehn Felder auf einer
// Fläche, und R3 erlaubt zwölf, danach gruppiert und **eingeklappt bis auf
// eine Gruppe**.
//
// **Welche vorn stehen, entscheidet die Vorlage**: ihre Reihenfolge. Ein
// Paket stellt, was man meistens braucht, an den Anfang; eine eigene
// Rangfolge hier wäre eine Entscheidung über fremde Vorlagen, die niemand
// nachprüfen kann.
//
// **Drei und nicht vier.** Die Stiloptionen zeigen vier (UXR-07, nach dem
// Artboard *Eigenschaften*: „4 von 11"); diese Fläche steht im
// Ausgabeassistenten aber neben fünf Zielen, und mit der Kartenwahl (zwei),
// der Suche und der Wahl der Vorlage sind drei Optionen genau die zwölf
// Felder, die R3 erlaubt. Gezählt mit vier waren es dreizehn.
//
// **Eingeklappt heißt nicht entfernt**: was dahinter steht, geht mit seinem
// Wert in jede Ausgabe, und der Kasten hält seinen Inhalt, wenn er zugeht.
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import type { ExportoptionDTO } from '../vorlagen'
import { Wahlfeld } from './Wahlfeld'

// vorn ist die Zahl der Optionen, die ohne Aufklappen dastehen.
const vorn = 3

interface VorlagenoptionenEigenschaften {
  optionen: readonly ExportoptionDTO[]
  wahl: Record<string, string>
  setzen: (option: ExportoptionDTO, wert: string) => void
}

export function Vorlagenoptionen({
  optionen,
  wahl,
  setzen,
}: VorlagenoptionenEigenschaften) {
  const texte = useTexte()

  const feld = (option: ExportoptionDTO) => (
    <Wahlfeld
      key={option.name}
      option={option}
      /* v8 ignore next -- `wahl` trägt jede Option, `vorgaben` setzt sie */
      wert={wahl[option.name] ?? option.wert}
      setzen={(wert) => {
        setzen(option, wert)
      }}
    />
  )

  const weitere = optionen.slice(vorn)

  return (
    <>
      {optionen.slice(0, vorn).map(feld)}

      {weitere.length > 0 && (
        <div className="mt-3">
          <Aufklappkasten
            titel={einsetzen(texte.seiten.ausgeben.weitereOptionen, {
              zahl: String(weitere.length),
            })}
          >
            {weitere.map(feld)}
          </Aufklappkasten>
        </div>
      )}
    </>
  )
}
