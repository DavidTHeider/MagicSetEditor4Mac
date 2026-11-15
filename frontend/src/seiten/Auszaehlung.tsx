// **Auszählen nach dem, was das Spielpaket erklärt** —
// #1338.
//
// # Was MSE2 hier tut
//
// Es zeigt links eine Liste der Dimensionen und rechts ein Diagramm
// (`src/gui/set/stats_panel.cpp`, Fassung f9d9356 vom 2021-07-06). Hier
// stehen die Dimensionen in der Wahl „Worüber" der Statistik
// (`Statistikwahl`), und dieses Bauteil zeichnet das Diagramm der gewählten:
// `magic.mse-game` erklärt **fünfzig** Dimensionen, und fünfzig Zeilen neben
// einem Diagramm wären die halbe Seite.
//
// **Bis UXR-08 trug es ein eigenes Auswahlfeld** unter eigener Überschrift
// (#1916) —
// ein zweites neben dem für die Felder, für dieselbe Frage.
//
// # Und warum sie erst auf Verlangen rechnet
//
// Weil eine Auszählung jede Karte durch ein Skript schickt. Gemessen am
// 2026-08-31 an `White.mse-set`: 3.761 Karten in 25 Sekunden. Ungefragt beim
// Öffnen der Lasche wäre das eine Sanduhr für jemanden, der nur die
// Kennzahlen sehen wollte.
import { Gruppenbalken } from './Gruppenbalken'
import { Ladefeld } from './Ladefeld'
import { alsDimension } from './Statistikwahl'
import { Statistikbild } from './Statistikbild'
import type { Auszaehlgriff } from '../haken/useAuszaehlung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface AuszaehlungEigenschaften {
  // dimension ist die gewählte — nie leer, sonst stünde dieses Bild nicht da.
  dimension: string
  zaehlung: Auszaehlgriff
}

export function Auszaehlung({ dimension, zaehlung }: AuszaehlungEigenschaften) {
  const eigene = useTexte().seiten.statistik

  const ganzes = zaehlung.gruppen.reduce(
    (summe, gruppe) => summe + gruppe.karten,
    0,
  )

  return (
    <Statistikbild
      ueberschrift={einsetzen(eigene.nachDimension, { dimension })}
    >
      <Erklaerungsabsatz
        text={eigene.dimensionenErklaerung}
        klasse="mt-1 max-w-prose text-sm text-leise"
      />

      <Ladefeld
        laedt={zaehlung.laedt}
        fehler={zaehlung.fehler}
        leer={zaehlung.gruppen.length === 0}
      >
        <p className="mt-3 max-w-prose text-leise">{eigene.dimensionLeer}</p>
      </Ladefeld>

      {zaehlung.gruppen.length > 0 && (
        <Gruppenbalken
          eintraege={zaehlung.gruppen}
          ganzes={ganzes}
          wahl={alsDimension(dimension)}
          quelle={dimension}
        />
      )}

      {zaehlung.befunde.map((befund) => (
        <p key={befund} className="mt-2 text-sm text-hinweis">
          {befund}
        </p>
      ))}
    </Statistikbild>
  )
}
