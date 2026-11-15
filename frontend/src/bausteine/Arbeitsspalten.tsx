// **Die Spalten des Arbeitsplatzes, mit ihren Trennern** — herausgelöst aus
// `Arbeitsplatz.tsx` für AP-12
// (#2369),
// Befund B-25.
//
// Der Arbeitsplatz entscheidet, **welche** Spalte steht und ob sie über dem
// Zentrum liegt; hier wird gezeichnet, wie breit sie ist. Zwischen zwei
// stehenden Spalten liegt ein `Spaltentrenner`, sobald ein Gedächtnis da ist
// — eine Breite, die niemand behält, spränge beim nächsten Öffnen zurück.
//
// # Die obere Grenze
//
// **Dem Zentrum bleiben 480 Punkte** (Vermerk 0076, `zentrumMindestens`): die
// Breite der Zeile, weniger die andere Spalte und die beiden Trenner. Über
// `domain.BreiteHoechstens` (800) geht es nie. Eine Breite, die schon
// darüber liegt — gesetzt in den Einstellungen bei einem breiteren Fenster —,
// bleibt, was sie ist; sie hebt die Grenze, statt ungültig zu werden.
import { useEffect, useState, type ReactNode, type RefObject } from 'react'

import type {
  Arbeitsplatzgedaechtnis,
  Arbeitsplatzregion,
} from '../haken/useArbeitsplatzgedaechtnis'

import { Arbeitsregion } from './Arbeitsregion'
import {
  breite,
  ueberInspector,
  ueberNavigator,
  wurzelschrift,
  zentrumMindestens,
} from './arbeitsplatzmasse'
import { Spaltentrenner } from './Spaltentrenner'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Region } from './Arbeitsplatz'

// Die Grenzen einer Spaltenbreite — `domain.BreiteMindestens` und
// `domain.BreiteHoechstens`; ein Wächter hält sie gleich
// (`internal/werkzeuge/spaltenbreite_test.go`).
export const breiteMindestens = 120
export const breiteHoechstens = 800
const trennerbreite = 16

interface ArbeitsspaltenEigenschaften {
  spalten: RefObject<HTMLDivElement | null>
  hoehe: number | undefined
  navigator: Region | undefined
  inspector: Region | undefined
  zentrum: ReactNode
  navigatorID: string
  inspectorID: string
  navigatorOffen: boolean
  inspectorOffen: boolean
  navigatorUeber: boolean
  inspectorUeber: boolean
  fokusBeimOeffnen: (element: HTMLElement | null) => void
  gedaechtnis: Arbeitsplatzgedaechtnis | undefined
}

export function Arbeitsspalten({
  spalten,
  hoehe,
  navigator,
  inspector,
  zentrum,
  navigatorID,
  inspectorID,
  navigatorOffen,
  inspectorOffen,
  navigatorUeber,
  inspectorUeber,
  fokusBeimOeffnen,
  gedaechtnis,
}: ArbeitsspaltenEigenschaften) {
  const texte = useTexte().bausteine.arbeitsplatz

  // **Was gerade gezogen wird**, als Vorschau — geschrieben wird beim
  // Loslassen, und bis dahin wächst die Spalte hier mit.
  const [vorschau, setzeVorschau] = useState<
    Partial<Record<Arbeitsplatzregion, number>>
  >({})

  // **Null ist kein Maß** — wie im Arbeitsplatz: ohne Breite gilt die
  // Grenze aus der Domäne.
  const [zeilenbreite, setzeZeilenbreite] = useState(0)
  useEffect(() => {
    const element = spalten.current

    /* v8 ignore next 3 -- die Zeile steht, wenn der Effekt läuft; ohne
       Element gäbe es nichts zu messen. TypeScript sieht das nicht. */
    if (element === null) {
      return undefined
    }

    const messen = () => {
      setzeZeilenbreite(element.clientWidth)
    }
    messen()

    const beobachter = new ResizeObserver(messen)
    beobachter.observe(element)

    return () => {
      beobachter.disconnect()
    }
  }, [spalten])

  const navigatorbreite = vorschau.navigator ?? gedaechtnis?.navigatorbreite
  const inspectorbreite = vorschau.inspector ?? gedaechtnis?.inspectorbreite

  // hoechstens ist die obere Grenze einer Spalte neben der anderen.
  const hoechstens = (eigene: number, andere: number) => {
    if (zeilenbreite <= 0) {
      return Math.max(eigene, breiteHoechstens)
    }

    const zentrum = (zentrumMindestens * wurzelschrift()) / 16
    const frei = zeilenbreite - andere - 2 * trennerbreite - zentrum

    return Math.max(
      eigene,
      breiteMindestens,
      Math.min(breiteHoechstens, Math.floor(frei)),
    )
  }

  // trenner zeichnet den Trenner einer Spalte — nur mit Gedächtnis.
  const trenner = (
    region: Arbeitsplatzregion,
    bezeichnung: string,
    steuert: string,
    andere: number,
  ) => {
    if (gedaechtnis === undefined) {
      return undefined
    }

    const jetzt =
      region === 'navigator'
        ? gedaechtnis.navigatorbreite
        : gedaechtnis.inspectorbreite

    return (
      <Spaltentrenner
        bezeichnung={einsetzen(texte.breiteVon, { region: bezeichnung })}
        wertetext={einsetzen(texte.breiteWert, { punkte: String(jetzt) })}
        steuert={steuert}
        breite={jetzt}
        mindestens={breiteMindestens}
        hoechstens={hoechstens(jetzt, andere)}
        richtung={region === 'navigator' ? 1 : -1}
        aendern={(punkte) => {
          gedaechtnis.breiteWuenschen(region, punkte)
        }}
        vorschau={(punkte) => {
          setzeVorschau((war) => ({ ...war, [region]: punkte }))
        }}
      />
    )
  }

  const andereFuerNavigator =
    inspector !== undefined && inspectorOffen ? (inspectorbreite ?? 0) : 0
  const andereFuerInspector =
    navigator !== undefined && navigatorOffen ? (navigatorbreite ?? 0) : 0

  return (
    // **Gedehnt, nicht `items-start`**: jede Spalte hat die Höhe der Zeile,
    // und was länger ist, rollt in ihr. **Ohne Abstand** — den halten die
    // Trenner, und ohne Gedächtnis ein leerer Streifen derselben Breite.
    <div ref={spalten} className="relative flex" style={{ height: hoehe }}>
      {/* **Was über dem Zentrum liegt, steht vorn** — so führt Tab vom
          Knopf zuerst hinein, gleich an welcher Kante es liegt. */}
      {navigator !== undefined && navigatorUeber && (
        <Arbeitsregion
          id={navigatorID}
          bezeichnung={navigator.bezeichnung}
          inhalt={navigator.inhalt}
          ueber
          spalte={fokusBeimOeffnen}
          className={ueberNavigator}
        />
      )}
      {inspector !== undefined && inspectorUeber && (
        <Arbeitsregion
          id={inspectorID}
          bezeichnung={inspector.bezeichnung}
          inhalt={inspector.inhalt}
          ueber
          spalte={fokusBeimOeffnen}
          className={ueberInspector}
        />
      )}

      {navigator !== undefined && navigatorOffen && (
        <>
          <Arbeitsregion
            id={navigatorID}
            bezeichnung={navigator.bezeichnung}
            inhalt={navigator.inhalt}
            className="flex min-h-0 w-60 shrink-0 flex-col overflow-y-auto"
            style={breite(navigatorbreite)}
          />
          {trenner(
            'navigator',
            navigator.bezeichnung,
            navigatorID,
            andereFuerNavigator,
          ) ?? <div aria-hidden="true" className="w-4 shrink-0" />}
        </>
      )}

      {/* **`min-w-0`, nicht 480.** Die 480 aus 0076 hält `engeGrenze`:
          solange der Navigator steht, hat das Zentrum sie. Im Mindestfenster
          (1.024 × 700) misst der Arbeitsplatz 670, und 480 + 320 passten
          nicht hinein — die Seite rollte waagerecht. Dort nimmt das Zentrum,
          was bleibt, und die Vorschau passt sich an. */}
      <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-y-auto">
        {zentrum}
      </div>

      {inspector !== undefined && inspectorOffen && (
        <>
          {trenner(
            'inspector',
            inspector.bezeichnung,
            inspectorID,
            andereFuerInspector,
          ) ?? <div aria-hidden="true" className="w-4 shrink-0" />}
          <Arbeitsregion
            id={inspectorID}
            bezeichnung={inspector.bezeichnung}
            inhalt={inspector.inhalt}
            className="flex min-h-0 w-80 shrink-0 flex-col overflow-y-auto"
            style={breite(inspectorbreite)}
          />
        </>
      )}
    </div>
  )
}
