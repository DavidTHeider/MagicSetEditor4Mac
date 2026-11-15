// **Wie eine Karte als Knopf über der Leinwand steht** — Werkstück G3 von
// Phase 5.5
// (#1460).
//
// **Eine eigene Datei**, seit `Tischflaeche.tsx` über die Zeilengrenze aus §2
// lief. Der Schnitt folgt der Frage: dort steht, **was die Fläche ist** —
// Leinwand, Kamera, Tischbild, Ziehen; hier, **wie eine einzelne Karte
// darüber erreichbar wird**.
//
// # Warum es den Knopf überhaupt gibt
//
// **Weil eine Leinwand nichts ist, worauf man mit der Tastatur zeigt.** Der
// Tisch ist gezeichnet; bedienbar wird er durch einen Knopf je Lage, der
// dort steht, wo die Karte liegt — und der einem Vorleseprogramm sagt, was
// sie ist und in welchem Zustand.
import type { KeyboardEvent, ReactNode } from 'react'
import { Kontextmenue } from './Kontextmenue'
import type { Menueeintrag } from './Menueliste'
import type { Klick } from '../spiel/kartengriff'
import { nachPunkten } from '../tisch/kamera'
import { useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { zonenwort } from '../seiten/partiewoerter'

import type { Kamera } from '../tisch/kamera'
import type { Tischlage } from '../tisch/tischplan'
import { schluesselvergleich } from '../ordnung'

type Spieltexte = ReturnType<typeof useTexte>['seiten']['spielen']

// stelle rechnet eine Lage auf ihren Platz über der Leinwand — **ohne
// Drehung**: sie gehört an den Knopf ([drehung]). Ein Vorfahr mit `transform`
// machte sich zum Bezug des Kartenmenüs (`position: fixed`), und die Fläche
// mit `overflow: hidden` schnitt es ab (P4, #2372, am nativen Fenster).
export function stelle(
  lage: Tischlage,
  kamera: Kamera,
  masse: { breite: number; hoehe: number },
): React.CSSProperties {
  const punkt = nachPunkten(
    kamera,
    { x: lage.x + lage.breite / 2, y: lage.y + lage.hoehe / 2 },
    masse,
  )
  const breite = lage.breite * kamera.massstab
  const hoehe = lage.hoehe * kamera.massstab

  return {
    left: `${String(punkt.x - breite / 2)}px`,
    top: `${String(punkt.y - hoehe / 2)}px`,
    width: `${String(breite)}px`,
    height: `${String(hoehe)}px`,
    zIndex: lage.hoehenrang,
  }
}

// drehung ist, wie weit der Knopf einer Lage gedreht steht — Lage und Kamera.
export function drehung(lage: Tischlage, kamera: Kamera): string {
  return `rotate(${String(lage.drehung + kamera.drehung)}deg)`
}

export interface Knopfeigenschaften {
  lage: Tischlage
  eigene: Spieltexte
  gezeichnet: boolean
  gewaehlt: boolean
  // aufDruck bekommt, wie geklickt wurde — Anzahl und ⌘ (P4, #2372).
  aufDruck: (klick: Klick) => void

  // tasten sind die Einzeltasten dieser Karte (`aria-keyshortcuts`), aufTaste
  // stellt, was eine meint, und sagt, ob sie galt (C-09).
  tasten?: string
  aufTaste?: (taste: string) => boolean

  // gedreht ist die Drehung am gezeichneten Tisch ([drehung]).
  gedreht?: string
}

export function Kartenknopf({
  lage,
  eigene,
  gezeichnet,
  gewaehlt,
  aufDruck,
  tasten = '',
  aufTaste,
  gedreht,
}: Knopfeigenschaften) {
  // **Verdeckt steht einmal da und nicht zweimal.** Wo kein Name ist, *ist*
  // „verdeckt" der Name; ihn daneben noch als Zustand zu führen las sich als
  // „verdeckt · Bibliothek · verdeckt" — gesehen in der Sichtprüfung am
  // 2026-09-16.
  const unkenntlich = lage.verdeckt || lage.geheim
  const name = lage.name || (unkenntlich ? eigene.tischVerdeckt : lage.kennung)
  // **Marken und Anhängsel gehören in die Beschriftung** — CR 400.5 verlangt,
  // dass für **alle** klar bleibt, was woran hängt und was getappt ist, und
  // „alle" schließt ein, wer die Karte nicht sieht, sondern hört. Eine Linie
  // auf der Leinwand erreicht ein Vorleseprogramm nicht.
  // **Marken heißen Marken** — PLAY-03.AK-04
  // (#1711).
  // `${zahl} ${art}` — also „3 +1/+1", eine Zahl und ein roher Schlüssel —
  // klänge neben einer Haufenzahl gleich, und das Kriterium verlangt, sie zu
  // trennen.
  const marken = Object.entries(lage.zaehler)
    .filter(([, zahl]) => zahl > 0)
    .sort(([eine], [andere]) => schluesselvergleich(eine, andere))
    .map(([art, zahl]) =>
      einsetzen(zahl === 1 ? eigene.tischMarkeEine : eigene.tischMarkeViele, {
        zahl: String(zahl),
        art,
      }),
    )

  // **Und die Art, wo sie nicht »Karte« ist.** Eine Karte sagt nicht, dass
  // sie eine ist — das ist der Normalfall, und wer ihn ansagte, machte aus
  // jeder Zeile eine Wiederholung.
  const arten: Record<string, string | undefined> = eigene.tischArt
  const art = arten[lage.art] ?? ''

  const zustand = [
    unkenntlich && lage.name ? eigene.tischVerdeckt : '',
    art,
    lage.drehung === 0 ? '' : eigene.tischGetappt,

    // **Die Zahl auf dem Haufen, und nur wo einer ist.** „1 auf einem
    // Haufen" wäre eine Zahl ohne Sache.
    lage.haufen === ''
      ? ''
      : einsetzen(eigene.tischAufHaufen, {
          zahl: String(lage.haufenzahl),
        }),
    ...marken,
    lage.haengtAn === '' ? '' : `${eigene.tischHaengtAn} ${lage.haengtAn}`,
  ].filter(Boolean)

  return (
    <button
      type="button"
      aria-pressed={gewaehlt}
      {...(gedreht === undefined ? {} : { style: { transform: gedreht } })}
      onClick={(ereignis) => {
        aufDruck({ detail: ereignis.detail, metaKey: ereignis.metaKey })
      }}
      {...(tasten === '' ? {} : { 'aria-keyshortcuts': tasten })}
      onKeyDown={(ereignis) => {
        if (einzeltaste(ereignis, aufTaste)) {
          ereignis.preventDefault()
          ereignis.stopPropagation()
        }
      }}
      className={
        gezeichnet
          ? 'h-full w-full rounded-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'
          : 'rounded-sm border border-kante bg-flaeche px-2 py-1 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'
      }
    >
      <span className={gezeichnet ? 'sr-only' : undefined}>
        {[name, zonenwort(eigene, lage.zone)].concat(zustand).join(' · ')}
      </span>
    </button>
  )
}

// Kartenmenue legt das Kontextmenü um einen Kartenknopf — P4 (#2372).
//
// **Am gezeichneten Tisch füllt es den Platz der Karte**, damit Rechtsklick
// und ⇧F10 dieselbe Fläche treffen wie der Klick. Ohne Menü steht der Knopf
// allein, wie vorher.
export function Kartenmenue({
  menue,
  gezeichnet,
  children,
}: {
  menue: { bezeichnung: string; eintraege: Menueeintrag[] } | undefined
  gezeichnet: boolean
  children: ReactNode
}) {
  if (menue === undefined) {
    return children
  }

  return (
    <Kontextmenue
      bezeichnung={menue.bezeichnung}
      eintraege={menue.eintraege}
      klasse={gezeichnet ? 'block h-full w-full' : ''}
    >
      {children}
    </Kontextmenue>
  )
}

// einzeltaste stellt eine Einzeltaste — **nur ohne Zusatztaste**: ⌘T, ⌥E und
// ⌃⌫ gehören dem Fenster und dem System.
export function einzeltaste(
  ereignis: KeyboardEvent<HTMLElement>,
  aufTaste: ((taste: string) => boolean) | undefined,
): boolean {
  if (ereignis.metaKey || ereignis.ctrlKey || ereignis.altKey) {
    return false
  }

  return aufTaste?.(ereignis.key) ?? false
}
