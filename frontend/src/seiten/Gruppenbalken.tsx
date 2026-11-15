// **Balken, die man wählen kann** — UXR-F6
// (#2013).
//
// # Was MSE2 hier tut (§15)
//
// Ein Klick ins Diagramm wählt die Gruppe darunter, und die Liste daneben
// zeigt nur ihre Karten (`GraphControl::onMouseDown` → `EVENT_GRAPH_SELECT`
// → `StatsPanel::onGraphSelect` → `filterCards`, `src/gui/control/graph.cpp`
// Zeilen 1116–1124 und `src/gui/set/stats_panel.cpp` Zeilen 526–547, Fassung
// f9d9356 vom 2021-07-06). Gewählt wird **eine** Gruppe je Achse; ein Klick
// ersetzt die Wahl, und die Pfeiltasten verschieben sie
// (`GraphControl::onChar`, Zeilen 1126–1163). Eine Wahl mehrerer Gruppen zugleich
// gibt es nicht — „mehrere" heißt dort mehrere **Achsen** (zwei oder drei
// Dimensionen zugleich), und die Statistik hier zählt nach einer.
//
// # Was hier anders ist, mit Grund
//
//   - **Ein Klick auf den gewählten Balken hebt die Wahl auf.** MSE2 kennt das
//     nicht — dort kommt man mit der Pfeiltaste zurück, nach links über die
//     erste Gruppe hinaus —, und eine Maus allein käme nie zurück.
//   - **Ein Balken ohne Karten ist angehalten.** MSE2 wählt auch ihn und zeigt
//     eine leere Liste.
//   - **Die Liste steht auf den Karten und nicht unter dem Diagramm**: die
//     Statistik ist hier eine Schublade (Artboard *Werkzeuge*).
//   - **Wechselt das Bild, endet die Wahl** — wie dort, wo `setData` die Wahl
//     leert; beim Neuladen folgt sie ihrem Balken (`useGruppenabgleich`, in
//     `Statistik`: dort steht auch das Bild, das nichts mehr zu wählen hat).
import type { WertzahlDTO } from '../api'
import { useStatistikfilter } from '../statistikfilter'
import { Gruppenwahlzeile } from './Gruppenwahlzeile'
import { Balkenliste } from './Statistikbild'

interface BalkenEigenschaften {
  eintraege: readonly WertzahlDTO[]
  ganzes: number

  // wahl ist der Eintrag „Worüber", zu dem das Bild gehört, quelle der Name
  // der Dimension oder des Felds darin.
  wahl: string
  quelle: string
}

export function Gruppenbalken({
  eintraege,
  ganzes,
  wahl,
  quelle,
}: BalkenEigenschaften) {
  const filter = useStatistikfilter()

  const dieseWahl = filter.wahl

  const gewaehlt =
    dieseWahl?.wahl === wahl
      ? eintraege.find((eintrag) => eintrag.name === dieseWahl.gruppe)
      : undefined

  return (
    <>
      <Gruppenwahlzeile gewaehlt={gewaehlt} />

      <Balkenliste
        eintraege={eintraege}
        ganzes={ganzes}
        feldwert
        wahl={{
          gewaehlt: gewaehlt?.name,
          waehlen: (eintrag) => {
            if (eintrag.name === gewaehlt?.name) {
              filter.aufheben()

              return
            }

            filter.waehlen({
              wahl,
              quelle,
              gruppe: eintrag.name,
              kartenIDs: eintrag.kartenIDs,
            })
          },
        }}
      />
    </>
  )
}
