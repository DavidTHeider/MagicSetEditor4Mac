// **Die Zeitzone als Auswahl mit Suche** — UXR-F1
// (#2007),
// R7 aus `docs/UIREVIEW.md`: bis dahin war sie ein Freitext, und »Europe/Berlin«
// aus dem Kopf zu tippen ist, was die Anwendung dem Menschen aufbürdete.
//
// # Woher die Werte kommen
//
// Aus `sprache/zeitzonen.json` — **und jede Zeile darin nimmt Go an**
// (`internal/werkzeuge/zeitzonenliste_test.go` hält es gegen
// `domain.PruefeZeitzone`). Eine Liste aus `Intl.supportedValuesOf` zur
// Laufzeit könnte das nicht: was das Fenster zählt, entscheidet ICU, was
// gespeichert wird, die Zonendatenbank des Rechners.
//
// # Was sie zusagt
//
//   - **Leer heißt „wie das System"** — die erste Zeile, mit leerem Wert,
//     wie die Einstellung es sagt (`domain.PruefeZeitzone`). Ohne sie gäbe es
//     keinen Weg zurück.
//   - **Ein Wert, der nicht in der Liste steht**, bleibt sichtbar gewählt
//     (`US/Pacific`, von Hand gesetzt oder aus einem Austausch): das Feld
//     zeigt ihn, wie er ist. Eine Auswahl, die ihn verschluckt, schriebe beim
//     nächsten Klick still etwas anderes.
import { useMemo } from 'react'

import { useTexte } from '../sprache/kontext'
import zonen from '../sprache/zeitzonen.json'
import { Suchauswahl } from './Suchauswahl'
import type { Suchwahl } from './suchwahl'

// zonenwahlen sind die Zeilen der Liste — einmal gebaut, denn sie ändern sich
// mit keiner Sprache.
const zonenwahlen: readonly Suchwahl[] = zonen.map((zone) => ({
  wert: zone,
  name: zone,
}))

interface ZeitzonenwahlEigenschaften {
  id: string

  // bezeichnung benennt die Liste — der Name der Einstellung.
  bezeichnung: string

  wert: string
  aendern: (wert: string) => void
  angehalten: boolean
  einstellung: string
}

export function Zeitzonenwahl({
  id,
  bezeichnung,
  wert,
  aendern,
  angehalten,
  einstellung,
}: ZeitzonenwahlEigenschaften) {
  const system = useTexte().bausteine.zeitzonenwahl.system

  const eintraege = useMemo(
    () => [{ wert: '', name: system }, ...zonenwahlen],
    [system],
  )

  return (
    <Suchauswahl
      id={id}
      bezeichnung={bezeichnung}
      eintraege={eintraege}
      wert={wert}
      aendern={aendern}
      angehalten={angehalten}
      einstellung={einstellung}
    />
  )
}
