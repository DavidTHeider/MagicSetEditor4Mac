// **Der Arbeitsplatz: Navigator, Zentrum, Inspector** — Werkstück U2 von
// Phase 5.1 (#1413),
// nach [Vermerk 0076](../../../docs/entscheidungen/0076-die-ordnung-der-arbeitsplaetze.md).
//
// # Die drei Zusagen aus 0076
//
//   1. **Ein Baustein, vier Verbraucher** — Karten (U5), Vorlagen (U7), Sammlung
//      (U8), Decks (U10). Eine vierte Region ändert den Vermerk, nicht die Seite.
//   2. **Dieselbe Lage, dieselben Tasten, dieselben Breiten**: Navigator
//      links mit 240 Punkten, Inspector rechts mit 320, Zentrum dazwischen;
//      ⌥⌘[ klappt den Navigator, ⌥⌘] den Inspector.
//   3. **Unter 1.200 Punkten Inhaltsbreite klappt der Navigator von
//      selbst ein**, damit das Zentrum nie unter 480 fällt — gemessen an
//      diesem Baustein sind das 1.072, siehe `engeGrenze`.
//
// # Was er nicht tut
//
//   - **Er merkt sich nichts — von selbst.** Ob eine Region zu ist, steht im
//     Zustand des Bausteins; das Gedächtnis je Profil und Mensch kommt seit
//     Werkstück U4 (#1415)
//     als Eigenschaft `gedaechtnis` herein (`useArbeitsplatzgedaechtnis`),
//     und dann gilt, was dort steht: der Baustein meldet jedes Klappen
//     dorthin, statt selbst zu klappen, und nimmt die Breiten von dort.
//   - **Er kennt keinen Inhalt.** Was in einer Region steht, sagt die Seite.
import {
  useCallback,
  useEffect,
  useId,
  useRef,
  useState,
  type ReactNode,
} from 'react'

import type {
  Arbeitsplatzgedaechtnis,
  Arbeitsplatzregion,
} from '../haken/useArbeitsplatzgedaechtnis'

import { Arbeitsplatzschalter } from './Arbeitsplatzschalter'
import { Arbeitsspalten } from './Arbeitsspalten'
import { useArbeitsplatzmenue } from '../haken/useArbeitsplatzmenue'
import { useFensterrest } from '../haken/useFensterrest'
import { deuten } from '../rahmen/tastatur'
import { schriftgroessenereignis } from '../erscheinungsbild'
import { useVerdeckterFokus } from '../haken/useVerdeckterFokus'
import { useVerdeckt } from '../verdeckung'
import {
  Spaltenmindesthoehe,
  Spaltenrand,
  navigatorPasst,
  wurzelschrift,
} from './arbeitsplatzmasse'
import { escapeSchliesst, imUeberZentrum } from './ueberzentrum'

// Region ist eine der beiden Seitenregionen samt ihrem Namen.
export interface Region {
  bezeichnung: string
  inhalt: ReactNode
}

interface ArbeitsplatzEigenschaften {
  // bezeichnung benennt den ganzen Arbeitsplatz — das `aria-label` des
  // Rahmens, damit ein Vorleseprogramm ihn von der Seitenleiste unterscheidet.
  bezeichnung: string

  navigator?: Region
  zentrum: ReactNode
  inspector?: Region

  // gedaechtnis macht den Baustein **geführt**: was steht, sagt es, und was
  // jemand klappt, geht als `merke` dorthin.
  gedaechtnis?: Arbeitsplatzgedaechtnis

  // beiEnge sagt, welche Region weicht, wenn das Fenster zu schmal für alle
  // drei ist. Vorgabe ist der Navigator (U5, die Kartenliste ist auch über
  // Kürzel zu ersetzen); die Vorlagenwerkstatt gibt den **Inspector** her,
  // denn ihr Navigator ist der einzige Weg zu einer Datei — bei 1.280 × 800
  // stand sie sonst ohne Baum und mit angehaltenem Knopf da (U7, #1418).
  beiEnge?: 'navigator' | 'inspector' | undefined

  // inspectorAnlass ist ein Zähler der Seite: **ändert er sich, geht der
  // Inspector bei Enge über dem Zentrum auf** — UXR-17 (#1925). Die Pakete
  // zählen jede Wahl und jeden Plan; bei 1280 Punkten stand beides sonst in
  // einer Spalte, die niemand sieht. **Bei Platz gilt der Wunsch** — wer den
  // Inspector geschlossen hat, dem öffnet ihn keine Wahl.
  inspectorAnlass?: number | undefined

  // schalterort stellt die Schalter woandershin — `Arbeitsplatzschalter`.
  schalterort?: HTMLElement | null | undefined
}

export function Arbeitsplatz({
  bezeichnung,
  navigator,
  zentrum,
  inspector,
  gedaechtnis,
  beiEnge = 'navigator',
  inspectorAnlass,
  schalterort,
}: ArbeitsplatzEigenschaften) {
  const kennung = useId()
  const rahmen = useRef<HTMLDivElement>(null)
  const spalten = useRef<HTMLDivElement>(null)
  const hoehe = useFensterrest(spalten, {
    unten: Spaltenrand,
    mindestens: Spaltenmindesthoehe,
  })

  // **Zwei Wünsche und ein Maß.** Der Wunsch kommt von der Person (Taste,
  // Knopf); das Maß vom Fenster. Ein Navigator steht, wenn beides ja sagt.
  const [navigatorLokal, setzeNavigatorLokal] = useState(true)
  const [inspectorLokal, setzeInspectorLokal] = useState(true)
  const [eng, setzeEng] = useState(false)

  // **Was über dem Zentrum liegt** — nur bei Enge. Ins Gedächtnis geht es
  // nicht, denn gemerkt wird, wie die Spalten stehen, wenn Platz ist. **Es
  // ruht, solange Platz ist**, und ist beim Zurückkehren wieder da
  // (S06.NAT-03: Maximieren und Rückkehr erhalten den Zustand der Spalten).
  const [ueber, setzeUeber] = useState<Arbeitsplatzregion>()
  const schliesseUeber = useCallback(() => {
    setzeUeber(undefined)
  }, [])
  const navigatorKnopf = useRef<HTMLButtonElement>(null)
  const inspectorKnopf = useRef<HTMLButtonElement>(null)
  const knopfVon = (region: Arbeitsplatzregion) =>
    region === 'navigator' ? navigatorKnopf : inspectorKnopf

  const navigatorGewuenscht = gedaechtnis?.navigator ?? navigatorLokal
  const inspectorGewuenscht = gedaechtnis?.inspector ?? inspectorLokal
  const navigatorEng = eng && beiEnge === 'navigator'
  const inspectorEng = eng && beiEnge === 'inspector'
  const inspectorOffen =
    inspector !== undefined && inspectorGewuenscht && !inspectorEng
  const navigatorUeber = navigatorEng && ueber === 'navigator'
  const inspectorUeber = inspectorEng && ueber === 'inspector'

  // **Beim Zeichnen nachgezogen und nicht in einem Effekt** — wie das
  // `gezeichnet` im `Aufklappkasten`: der Inspector steht im selben Bild wie
  // die Wahl, die ihn verlangt. Der erste Stand ist kein Anlass.
  const [letzterAnlass, setzeLetzterAnlass] = useState(inspectorAnlass)
  if (letzterAnlass !== inspectorAnlass) {
    setzeLetzterAnlass(inspectorAnlass)

    if (inspectorEng) {
      setzeUeber('inspector')
    }
  }

  // klappen ist, was Knopf und Taste tun: bei Enge über dem Zentrum, sonst
  // über den Wunsch.
  const klappen = (region: Arbeitsplatzregion, offen: boolean) => {
    if (region === 'navigator' ? navigatorEng : inspectorEng) {
      // **Wer sie schließt, während der Fokus in ihr steht, bekommt ihn am
      // Knopf zurück** (M1) — sonst fiele er auf `body`.
      if (!offen && imUeberZentrum(document.activeElement)) {
        knopfVon(region).current?.focus()
      }
      setzeUeber(offen ? region : undefined)

      return
    }

    // **Wer bei Platz klappt, meint den Wunsch** — eine ruhende Überlagerung
    // von vorhin gilt dann nicht mehr.
    setzeUeber(undefined)
    wuenschen(region, offen)
  }

  // wuenschen klappt — im Gedächtnis, wenn es eines gibt, sonst hier.
  const wuenschen = (region: Arbeitsplatzregion, offen: boolean) => {
    if (gedaechtnis !== undefined) {
      gedaechtnis.merke(region, offen)

      return
    }

    if (region === 'navigator') {
      setzeNavigatorLokal(offen)
    } else {
      setzeInspectorLokal(offen)
    }
  }

  useEffect(() => {
    const element = rahmen.current

    /* v8 ignore next 3 -- der Rahmen steht, wenn der Effekt läuft: das `ref`
       ist beim ersten Zeichnen gesetzt, und ohne Element gäbe es nichts zu
       messen. TypeScript sieht das nicht. */
    if (element === null) {
      return undefined
    }

    // **Null ist kein Maß.** jsdom misst jeden Kasten mit null Punkten, und
    // ein verborgener Kasten im Browser ebenso; wer daraus „eng" schlösse,
    // klappte den Navigator weg, bevor jemand etwas gesehen hat. Ohne Maß
    // gilt die Vorgabe: er steht.
    const messen = () => {
      const breite = element.clientWidth

      setzeEng(breite > 0 && !navigatorPasst(breite, wurzelschrift()))
    }

    messen()

    const beobachter = new ResizeObserver(messen)
    beobachter.observe(element)

    // **Und nach einem Wechsel der Schriftgröße** (R5): die Grenze rechnet
    // in der Wurzelschrift, die Breite bleibt dieselbe.
    const wurzel = document.documentElement
    wurzel.addEventListener(schriftgroessenereignis, messen)

    return () => {
      beobachter.disconnect()
      wurzel.removeEventListener(schriftgroessenereignis, messen)
    }
  }, [])

  // **Unter einer Schublade klappt nichts** (UX-08, #1680) — ⌥⌘[ gehörte
  // dann einem Navigator, den niemand sieht. Warum, steht in
  // `verdeckung.ts`.
  const verdeckt = useVerdeckt()

  // **Geht der Fokus unter die Spalte, geht sie** (N12) — wie, sagt der Haken.
  useVerdeckterFokus(rahmen, ueber !== undefined, schliesseUeber)

  // **Die Tasten kommen aus `tastatur.ts`**, damit die Kürzeltafel sie
  // kennt — eine dritte Liste hier wäre die, die niemand nachführt.
  useEffect(() => {
    if (verdeckt) {
      return undefined
    }

    const aufTaste = (ereignis: KeyboardEvent) => {
      const treffer = deuten({
        taste: ereignis.key,
        code: ereignis.code,
        meta: ereignis.metaKey,
        ctrl: ereignis.ctrlKey,
        alt: ereignis.altKey,
        umschalt: ereignis.shiftKey,
        imEingabefeld: false,
      })

      if (treffer?.art === 'navigator' && navigator !== undefined) {
        ereignis.preventDefault()
        klappen(
          'navigator',
          navigatorEng ? !navigatorUeber : !navigatorGewuenscht,
        )
      }

      if (treffer?.art === 'inspector' && inspector !== undefined) {
        ereignis.preventDefault()
        klappen('inspector', !(inspectorOffen || inspectorUeber))
      }

      // **Escape schließt, was über dem Zentrum liegt**, und der Fokus geht
      // zu dem Knopf, der es geöffnet hat — wann, sagt `ueberzentrum.ts`.
      if (
        ueber !== undefined &&
        escapeSchliesst(ereignis, navigatorUeber || inspectorUeber)
      ) {
        setzeUeber(undefined)
        knopfVon(ueber).current?.focus()
      }
    }

    window.addEventListener('keydown', aufTaste)

    return () => {
      window.removeEventListener('keydown', aufTaste)
    }
  })

  const navigatorOffen =
    navigator !== undefined && navigatorGewuenscht && !navigatorEng
  const navigatorID = `${kennung}-navigator`
  const inspectorID = `${kennung}-inspector`

  // **Die Menüleiste führt die Spalten mit** — UXR-F14 (#2029), der Haken sagt wie.
  const fokusBeimOeffnen = useArbeitsplatzmenue(
    {
      navigator: {
        vorhanden: navigator !== undefined,
        offen: navigatorOffen || navigatorUeber,
      },
      inspector: {
        vorhanden: inspector !== undefined,
        offen: inspectorOffen || inspectorUeber,
      },
    },
    verdeckt,
    klappen,
  )

  return (
    <div
      ref={rahmen}
      role="region"
      aria-label={bezeichnung}
      className="flex min-h-0 flex-col gap-2"
    >
      <Arbeitsplatzschalter
        ort={schalterort}
        navigator={
          navigator === undefined
            ? undefined
            : {
                offen: navigatorOffen || navigatorUeber,
                eng: navigatorEng,
                griff: navigatorKnopf,
                steuert: navigatorID,
                angehalten: gedaechtnis?.laeuft('navigator') ?? false,
                klappen: () => {
                  klappen('navigator', !(navigatorOffen || navigatorUeber))
                },
              }
        }
        inspector={
          inspector === undefined
            ? undefined
            : {
                offen: inspectorOffen || inspectorUeber,
                eng: inspectorEng,
                griff: inspectorKnopf,
                steuert: inspectorID,
                angehalten: gedaechtnis?.laeuft('inspector') ?? false,
                klappen: () => {
                  klappen('inspector', !(inspectorOffen || inspectorUeber))
                },
              }
        }
      />

      <Arbeitsspalten
        spalten={spalten}
        hoehe={hoehe}
        navigator={navigator}
        inspector={inspector}
        zentrum={zentrum}
        navigatorID={navigatorID}
        inspectorID={inspectorID}
        navigatorOffen={navigatorOffen}
        inspectorOffen={inspectorOffen}
        navigatorUeber={navigatorUeber}
        inspectorUeber={inspectorUeber}
        fokusBeimOeffnen={fokusBeimOeffnen}
        gedaechtnis={gedaechtnis}
      />
    </div>
  )
}
