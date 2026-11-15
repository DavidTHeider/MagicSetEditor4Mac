// **Der Tisch im Fenster** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Die Szene lebt länger als ein Render
//
// **#1460 verlangt es wörtlich**: „langlebige BoardScene statt Neuanlage bei
// jedem React-Render". Der Maler hängt deshalb an einer Referenz und nicht am
// Zustand; was React neu rechnet, ist der **Plan**, und der wird gezeichnet,
// nicht nachgebaut.
//
// # Regel 2 aus §7 gilt auch hier
//
// **Ein Maler, der nach dem Aushängen fertig wird, gehört niemandem mehr.**
// Er wird geräumt statt eingetragen — sonst hinge ein WebGL-Kontext an einer
// Leinwand, die es nicht mehr gibt.
import { useCallback, useEffect, useRef, useState } from 'react'

import type { RefObject } from 'react'

import {
  ausDemDokument,
  dunklesBild,
  mitTischbild,
  tischfarben,
} from './tischfarben'
import { eingepasst, gedreht, gezoomt, verschoben } from './kamera'
import { aufAuswahl, aufEigenenSitz } from './kamerazielen'
import { abzugMachen } from './tischabzug'
import { tischbild } from './tischbilder'

import type { Bildquelle, Malhindernis, Tischmalerei } from './maler'
import type { Kamera, Punkt } from './kamera'
import type { Abzug, Abzugsfehler } from './tischabzug'
import type { Ausschnittwahl } from './tischausschnitt'
import type { Tischplan } from './tischplan'
import type { Wirkungsart } from './tischwirkungen'

// Tischanschluss ist, was der Haken braucht.
export interface Tischanschluss {
  plan: Tischplan
  bild: Bildquelle

  // gewaehlt ist die hervorgehobene Karte — oder keine.
  gewaehlt?: string | undefined

  // tischbild ist der Name des gewählten Hintergrunds.
  tischbild?: string | undefined

  // hintergrund ist die Adresse eines eigenen Tischbilds — leer heißt: der
  // gewählte Farbsatz.
  hintergrund?: string | undefined

  // bewegt sagt, ob Karten an ihren neuen Platz fliegen sollen.
  bewegt?: boolean | undefined

  // beschriftung gibt den Namen einer Zone — PLAY-03.AK-06.
  beschriftung?: ((zone: string, kurz?: boolean) => string) | undefined

  // sitzkopf gibt, was am Sitz steht — PLAY-03.
  sitzkopf?: ((sitz: string) => string) | undefined

  // beschriftungsstufe ist der Wert von `tisch.beschriftung`.
  beschriftungsstufe?: string | undefined

  // wirkungen sagt je Kennung, was gerade geschehen ist — PLAY-04.AK-02
  // (#1712).
  wirkungen?: Map<string, Wirkungsart> | undefined

  // leinwand hält der Aufrufer.
  //
  // **Und nicht dieser Haken.** Ein Rückgabewert, der eine Referenz führt,
  // lässt den React-Compiler jedes Feld daneben für eine halten — „Cannot
  // access refs during render", gemeldet über `hindernis`, das ein ganz
  // gewöhnlicher Zustand ist. Wer die Leinwand hält, ist ohnehin der, der sie
  // in den Baum hängt.
  //
  // **Es ist der Ort der Leinwand, nicht sie selbst** — UI-N01 (#2089): die
  // Leinwand legt `tischmalerei` je Anlauf frisch hinein, denn PixiJS nimmt
  // sie beim Abräumen aus dem Dokument und verliert ihren Kontext.
  leinwand: RefObject<HTMLElement | null>

  // malerei ist der Weg zu einem Maler — im Lauf ein Doppel.
  malerei: (ort: HTMLElement) => Promise<Tischmalerei>
}

export interface Tischlauf {
  // hindernis steht, solange nicht gezeichnet werden kann.
  hindernis: Malhindernis | undefined

  kamera: Kamera

  // masse sind die Maße der Zeichenfläche in Punkten.
  //
  // **Sie stehen als Zustand und nicht als Abfrage**, weil die semantische
  // Schicht sie zum Zeichnen braucht: ein Knopf über einer Karte muss dort
  // stehen, wo die Karte gerade ist, und das hängt an der Flächengröße.
  masse: { breite: number; hoehe: number }

  schwenke: (dx: number, dy: number) => void
  zoome: (faktor: number, auf: Punkt) => void
  drehe: (grad: number) => void
  passeEin: () => void

  // zeigeAuswahl und zeigeEigenenSitz holen ein Ziel heran — PLAY-05.AK-05
  // (#1713).
  //
  // **Beide tun nichts, wo es das Ziel nicht gibt**: eine weggeräumte Auswahl
  // und ein Zuschauer ohne eigenen Sitz sind gewöhnliche Lagen, und ein
  // Sprung auf den Ursprung wäre eine Bewegung, die niemand angeordnet hat.
  zeigeAuswahl: () => void
  zeigeEigenenSitz: () => void

  // abzug gibt ein Bild des gewählten Ausschnitts — PLAY-05.AK-03 und AK-04
  // (#1713).
  //
  // **Der Rückgabewert trägt den Grund**, wenn es keines gibt: kein Maler,
  // kein Ausschnitt, keine gültigen Maße. Drei Gründe, drei Sätze — einer für
  // alle nähme dem Lesenden die Auskunft, ob ein zweiter Versuch lohnt.
  abzug: (
    wahl: Ausschnittwahl,
    dpi: number,
  ) => Promise<Abzug | { fehler: Abzugsfehler }>
}

// Einpassrand ist die Luft, die beim Einpassen um den Tisch bleibt.
const Einpassrand = 16

export function useTisch({
  plan,
  bild,
  malerei,
  gewaehlt,
  tischbild: bildname = 'holz',
  hintergrund,
  bewegt,
  beschriftung,
  sitzkopf,
  beschriftungsstufe,
  wirkungen,
  leinwand: knoten,
}: Tischanschluss): Tischlauf {
  const maler = useRef<Awaited<Tischmalerei> | undefined>(undefined)
  const [hindernis, setzeHindernis] = useState<Malhindernis | undefined>(
    undefined,
  )
  // **Die Kamera ist erst gewählt, wenn jemand sie bewegt hat.** Bis dahin
  // ist sie eine **Ableitung** aus Plan und Fläche — und keine Zuweisung in
  // einem Effekt, die einen zweiten Render nach sich zöge.
  const [eingestellt, setzeKamera] = useState<Kamera | undefined>(undefined)

  // bereit sagt, dass ein Maler da ist.
  //
  // **Ohne diesen Zustand zeichnete nie jemand.** Der Maler hängt an einer
  // Referenz — sie zu setzen löst keinen Render aus, und der Zeicheneffekt
  // liefe erst, wenn sich zufällig der Plan änderte. Gefunden am roten Lauf:
  // `zeichne` wurde kein einziges Mal gerufen.
  const [bereit, setzeBereit] = useState(false)
  const [masse, setzeMasse] = useState({ breite: 0, hoehe: 0 })

  useEffect(() => {
    const leinwand = knoten.current
    let gilt = true

    // **Auch die fehlende Leinwand geht durch dasselbe Versprechen.** Sie
    // synchron im Effekt zu melden hieße, im Effekt Zustand zu setzen — ein
    // zweiter Render, bevor der erste steht, und der React-Compiler streicht
    // es zu Recht an.
    const anlauf: Promise<Tischmalerei> = leinwand
      ? malerei(leinwand)
      : Promise.resolve({ art: 'ohne', grund: 'kein-renderer' })

    void anlauf.then((ergebnis) => {
      if (!gilt) {
        if (ergebnis.art === 'da') ergebnis.maler.raeume()

        return
      }

      maler.current = ergebnis
      setzeHindernis(ergebnis.art === 'ohne' ? ergebnis.grund : undefined)
      setzeBereit(ergebnis.art === 'da')
      if (ergebnis.art === 'da') {
        setzeMasse(ergebnis.maler.masse())

        // **Auch der Horcher gehört zu diesem Stand** (§7, Regel 2).
        //
        // Gefunden in der Sichtprüfung am 2026-09-16: React hängt Effekte im
        // Entwicklungslauf doppelt ein — an, ab, an. Das Abräumen zerstört
        // den Renderer, und PixiJS feuert dabei `webglcontextlost`. Ohne
        // diese Frage meldete der Horcher des **alten** Malers „verloren" auf
        // den **neuen** Zustand, und der Tisch stand mit einer Fehlermeldung
        // da, ohne je gezeichnet zu haben.
        ergebnis.maler.beiVerlust(() => {
          if (gilt) setzeHindernis('verloren')
        })
      }
    })

    return () => {
      gilt = false
      if (maler.current?.art === 'da') maler.current.maler.raeume()
      maler.current = undefined
      setzeBereit(false)
    }
  }, [malerei, knoten])

  // **Ein Wechsel des Tischbilds färbt um und baut nicht neu auf.** Der
  // Renderer bleibt stehen; sein erstes Bild kostet das vierundvierzigfache
  // eines gewöhnlichen (`docs/TISCHLEISTUNG.md`).
  useEffect(() => {
    const laufender = maler.current
    const leinwand = knoten.current
    if (!bereit || laufender?.art !== 'da' || !leinwand) return

    const lesen = ausDemDokument(leinwand.ownerDocument)
    laufender.maler.faerbe(
      mitTischbild(tischfarben(lesen), tischbild(bildname), dunklesBild(lesen)),
    )
  }, [bereit, bildname, knoten])

  const kamera = eingestellt ?? eingepasst(plan, masse, Einpassrand)

  useEffect(() => {
    const laufender = maler.current
    if (!bereit || laufender?.art !== 'da' || hindernis) return

    laufender.maler.zeichne({
      plan,
      kamera,
      bild,
      gewaehlt,
      hintergrund,
      bewegt,
      beschriftung,
      sitzkopf,
      beschriftungsstufe,
      wirkungen,
    })
  }, [
    bereit,
    plan,
    kamera,
    bild,
    gewaehlt,
    hintergrund,
    bewegt,
    beschriftung,
    sitzkopf,
    beschriftungsstufe,
    wirkungen,
    hindernis,
  ])

  const flaeche = useCallback(() => {
    const laufender = maler.current

    return laufender?.art === 'da'
      ? laufender.maler.masse()
      : { breite: 0, hoehe: 0 }
  }, [])

  // **Die Fläche folgt ihrem Behälter** — dem Knoten, dem PixiJS über
  // `resizeTo` folgt. Ohne das zeichnete die semantische Schicht über einer
  // Fläche, die es nicht mehr gibt.
  //
  // **Hier horchte bis zum 2026-09-24 das Fenster auf `resize`**, und zwei
  // Dinge gingen daran vorbei: PixiJS passt die Leinwand erst im nächsten
  // Bild an, und die Maße im selben Augenblick waren die alten — der Tisch
  // stand abgeschnitten da; und wenn die Seitenleiste auf- oder zuging,
  // änderte sich der Behälter, ohne dass das Fenster etwas meldete. Gesehen
  // in der WebKit-Probe (#1828). **Deshalb passt der Maler an, bevor
  // gemessen wird**, und beobachtet wird der Behälter.
  useEffect(() => {
    const behaelter = knoten.current?.parentElement
    if (!behaelter) return undefined

    const beobachter = new ResizeObserver(() => {
      const laufender = maler.current
      if (laufender?.art === 'da') setzeMasse(laufender.maler.passeAn())
    })
    beobachter.observe(behaelter)

    return () => {
      beobachter.disconnect()
    }
  }, [bereit, knoten])

  return {
    hindernis,
    kamera,
    masse,
    schwenke: useCallback(
      (dx, dy) => {
        setzeKamera(verschoben(kamera, dx, dy))
      },
      [kamera],
    ),
    zoome: useCallback(
      (faktor, auf) => {
        setzeKamera(gezoomt(kamera, faktor, auf, flaeche()))
      },
      [kamera, flaeche],
    ),
    drehe: useCallback(
      (grad) => {
        setzeKamera(gedreht(kamera, grad))
      },
      [kamera],
    ),
    passeEin: useCallback(() => {
      setzeKamera(eingepasst(plan, flaeche(), Einpassrand, kamera.drehung))
    }, [plan, flaeche, kamera.drehung]),

    zeigeAuswahl: useCallback(() => {
      const ziel = aufAuswahl(plan, gewaehlt, flaeche(), kamera.drehung)

      if (ziel) setzeKamera(ziel)
    }, [plan, gewaehlt, flaeche, kamera.drehung]),

    zeigeEigenenSitz: useCallback(() => {
      const ziel = aufEigenenSitz(plan, flaeche(), kamera.drehung)

      if (ziel) setzeKamera(ziel)
    }, [plan, flaeche, kamera.drehung]),

    // **Der Abzug greift auf den laufenden Maler und nicht auf den Zustand.**
    // Ein Bild entsteht aus dem Szenenbaum, wie er **jetzt** steht — genau
    // das meint AK-04 mit „konsistenter Snapshot".
    //
    // **Noch kein Maler heißt: er kommt** — und nicht, dass es keine
    // Grafikkarte gibt. Der Tisch wird bei jedem Sichtwechsel neu aufgebaut.
    abzug: useCallback(
      async (wahl: Ausschnittwahl, dpi: number) =>
        maler.current === undefined
          ? { fehler: 'aufbau' as const }
          : abzugMachen(maler.current, plan, kamera, flaeche(), wahl, dpi),
      [plan, kamera, flaeche],
    ),
  }
}
