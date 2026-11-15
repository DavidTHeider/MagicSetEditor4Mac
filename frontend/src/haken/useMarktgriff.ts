// **Greifen, wählen und legen am Markttisch** — P12
// (#2380).
//
// # Ein Vertrag für Maus und Tastatur
//
// **Leertaste greift, Pfeile wählen, Eingabe legt** — der Gestenvertrag aus
// GESTE (#2358) und keine zweite Tastenfolge. **Das Ziehen mit der Maus
// endet in derselben Absicht** (`absichtAusZug`): losgelassen über einer
// Hälfte mit `data-ablage`. Gezogen wird über Zeigerereignisse und nicht
// über HTML-Ziehen — dieselbe Mausgeste wie am Spieltisch.
//
// # Mehrfachwahl
//
// **⌘ und ein Klick nehmen eine Karte dazu**; gegriffen wird die ganze Wahl,
// wenn die Karte darin liegt. Eine gemischte Wahl hat kein Ziel
// (`zieleFuer`). **Gelegt ist gelegt**: danach ist die Wahl leer.
//
// # Nebenläufig
//
// **Ein Griff gilt nur für das Gegenüber, aus dem er kam** (§7, Regel 1 und
// 2): wechselt es, verwirft `useGeste` ihn.
import {
  useEffect,
  useRef,
  useState,
  type PointerEvent as Zeigerereignis,
} from 'react'

import { absichtAusZug, type Gestenziel } from '../bausteine/gestenvertrag'
import {
  legen,
  zieleFuer,
  type Ablage,
  type Fachkarte,
} from '../bausteine/marktlegen'
import { useTexte } from '../sprache/kontext'
import { useGeste } from './useGeste'

import type { Handelsposition } from '../apihandelsleitung'

interface Eigenschaften {
  ich: string
  gegenueber: string
  karten: readonly Fachkarte[]
  positionen: Handelsposition[]
  setzePositionen: (positionen: Handelsposition[]) => void
}

export function useMarktgriff({
  ich,
  gegenueber,
  karten,
  positionen,
  setzePositionen,
}: Eigenschaften) {
  const texte = useTexte().seiten.handeln
  const [auswahl, setzeAuswahl] = useState<readonly string[]>([])

  const stand = { profil: `${ich}/${gegenueber}`, epoche: gegenueber }

  const geste = useGeste({
    stand,
    ausfuehren: (absicht) => {
      const gewaehlt = absicht.objekt
        .split('\n')
        .map((kennung) => karten.find((karte) => karte.kennung === kennung))
        .filter((karte) => karte !== undefined)

      // **Gemessen am ganzen Fach** (#2497): eine Zeile wächst bis zu allem,
      // was dort von ihrer Variante liegt, nicht bis zur gelegten Karte.
      setzePositionen(
        legen(
          positionen,
          gewaehlt,
          absicht.ziel as Ablage,
          ich,
          gegenueber,
          karten,
        ),
      )
      setzeAuswahl([])

      return Promise.resolve()
    },
  })

  const ziel = (ablage: Ablage): Gestenziel => ({
    kennung: ablage,
    name:
      ablage === 'gebe' ? texte.positionVonIch : texte.positionVonGegenueber,
  })

  // wahlMit gibt, was mit dieser Karte gegriffen wird — die Wahl, wenn sie
  // darin liegt, sonst sie allein.
  const wahlMit = (karte: Fachkarte) =>
    auswahl.includes(karte.kennung)
      ? karten.filter((eine) => auswahl.includes(eine.kennung))
      : [karte]

  const klicken = (karte: Fachkarte, dazu: boolean) => {
    setzeAuswahl((bisher) => {
      if (!dazu) {
        return bisher.length === 1 && bisher[0] === karte.kennung
          ? []
          : [karte.kennung]
      }

      return bisher.includes(karte.kennung)
        ? bisher.filter((kennung) => kennung !== karte.kennung)
        : [...bisher, karte.kennung]
    })
  }

  // taste deutet eine Taste an einer Karte — und sagt, ob sie verbraucht ist.
  const taste = (karte: Fachkarte, name: string) => {
    if (geste.taste(name)) {
      return true
    }

    if (name !== ' ') {
      return false
    }

    const gewaehlt = wahlMit(karte)
    geste.greifen({
      objekt: gewaehlt.map((eine) => eine.kennung).join('\n'),
      quelle: karte.quelle,
      ziele: zieleFuer(gewaehlt).map(ziel),
    })

    return true
  }

  // **Der neueste Stand beim Loslassen** — Gesamtreview über 0.9.0, Befund 4:
  // der Horcher hielt `geste` vom Zugreifen fest, und mit ihr die Positionen
  // von damals; was sich während des Griffs änderte, überschrieb das
  // Loslassen. Die **Epoche** bleibt die vom Zugreifen — sie soll einen
  // Wechsel des Gegenübers erkennen.
  const aktuell = useRef(geste)
  useEffect(() => {
    aktuell.current = geste
  })

  // offen räumt den laufenden Griff ab — **einer zur Zeit**, und beim Abbau
  // der Seite keiner mehr.
  const offen = useRef<() => void>(undefined)
  useEffect(
    () => () => {
      offen.current?.()
    },
    [],
  )

  // ziehen beginnt einen Zug mit der Maus; er endet beim Loslassen **dieses**
  // Zeigers oder bei seinem Abbruch (`pointercancel`).
  const ziehen = (karte: Fachkarte, ereignis: Zeigerereignis) => {
    if (ereignis.button !== 0) {
      return
    }

    offen.current?.()

    const gewaehlt = wahlMit(karte)
    const objekt = gewaehlt.map((eine) => eine.kennung).join('\n')
    const erlaubt = zieleFuer(gewaehlt)
    const zeiger = ereignis.pointerId
    const griffstand = stand

    function abraeumen() {
      document.removeEventListener('pointerup', loslassen)
      document.removeEventListener('pointercancel', abbrechen)
      offen.current = undefined
    }

    function abbrechen(abgebrochen: PointerEvent) {
      if (abgebrochen.pointerId === zeiger) {
        abraeumen()
      }
    }

    function loslassen(losgelassen: PointerEvent) {
      if (losgelassen.pointerId !== zeiger) {
        return
      }

      abraeumen()

      const ablage = (losgelassen.target as Element | null)
        ?.closest('[data-ablage]')
        ?.getAttribute('data-ablage') as Ablage | undefined

      if (ablage !== undefined && erlaubt.includes(ablage)) {
        aktuell.current.ablegen(
          absichtAusZug(
            { objekt, quelle: karte.quelle, ziel: ziel(ablage) },
            griffstand,
            'maus',
          ),
        )
      }
    }

    document.addEventListener('pointerup', loslassen)
    document.addEventListener('pointercancel', abbrechen)
    offen.current = abraeumen
  }

  const gegriffen = (karte: Fachkarte) =>
    geste.zustand.art === 'gegriffen' &&
    geste.zustand.objekt.split('\n').includes(karte.kennung)

  return {
    auswahl,
    ansage: geste.ansage,
    klicken,
    taste,
    ziehen,
    gegriffen,
  }
}
