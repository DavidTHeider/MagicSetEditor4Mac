// **Ein Menü hinter einem Knopf** — Werkstück U2 von Phase 5.1
// (#1413),
// nach [Vermerk 0077](../../../docs/entscheidungen/0077-eigene-primitive-statt-radix.md).
//
// # Wozu
//
// Die Kartenhandlungen standen als Knöpfe in einem Kasten, den man erst
// aufklappte; die Paketzeile trug ihr „Entfernen" offen neben jedem Eintrag.
// Ein Menü nimmt beides in eine Tabstelle: der Knopf sagt, was dahinter ist
// (`aria-haspopup="menu"`), die Pfeile wandern, ⏎ wählt, Escape schließt —
// das Muster, das ein Vorleseprogramm bei `role="menu"` erwartet.
//
// # Was es zusagt
//
//   - **Der Fokus kehrt zurück.** Nach Escape, nach einer Wahl und nach
//     einem zweiten Klick auf den Knopf steht er wieder dort.
//   - **Die Tabulatortaste verlässt das Menü** und schließt es — ein Menü
//     ist keine Sackgasse. Der Fokus geht dann weiter, nicht zurück.
//   - **↓ und ↑ auf dem Knopf öffnen** — auf dem ersten und dem letzten
//     Eintrag.
//
// # Was es nicht tut
//
//   - **Keine Untermenüs.** Sieben Einträge brauchen keine Ebene.
//   - **Keine Suche nach Anfangsbuchstaben.** Bei sieben Einträgen ist ↓
//     schneller als das Nachdenken über den Buchstaben.
import {
  useEffect,
  useId,
  useLayoutEffect,
  useRef,
  useState,
  type KeyboardEvent,
  type ReactNode,
} from 'react'

import { useAnkerwacht } from '../haken/useAnkerwacht'
import { nachDemTab } from './fokus'
import { Knopf, type Gewicht } from './Knopf'
import { Menueliste, naechsteStelle, type Menueeintrag } from './Menueliste'

export { naechsteStelle, type Menueeintrag } from './Menueliste'

interface MenueEigenschaften {
  // bezeichnung steht auf dem Knopf und benennt das Menü.
  bezeichnung: string

  eintraege: readonly Menueeintrag[]

  gewicht?: Gewicht

  // ohneTabstelle reicht `Knopf.ohneTabstelle` weiter — für ein Menü in
  // einer Gitterzeile.
  ohneTabstelle?: boolean

  // children ersetzt die Aufschrift des Knopfs, wenn dort etwas anderes stehen
  // soll als die Bezeichnung — ein Zeichen etwa. Die Bezeichnung bleibt der
  // Name für ein Vorleseprogramm.
  children?: ReactNode

  // ausrichtung sagt, an welcher Kante des Knopfs das Menü hängt. **Rechts**
  // für einen Knopf am rechten Rand — das Menü einer Setkarte (UXR-04,
  // #1912) ragte sonst über die Kachel und am Fenster vorbei. Ohne Angabe
  // hängt es links, wie bisher. **Dasselbe brauchte das Menü „Diese Karte"**
  // am rechten Rand der Kartenzeile (UXR-06, #1914) — eine Eigenschaft für
  // beide, nicht zwei.
  ausrichtung?: 'links' | 'rechts'
}

// abstand hält ein Menü vom Fensterrand fern, luft vom Knopf — in Punkten.
const abstand = 8
const luft = 4

export function Menue({
  bezeichnung,
  eintraege,
  gewicht = 'neben',
  ohneTabstelle = false,
  children,
  ausrichtung = 'links',
}: MenueEigenschaften) {
  const [offen, setzeOffen] = useState(false)
  const [stelle, setzeStelle] = useState(0)
  const menueID = useId()
  const wurzel = useRef<HTMLDivElement>(null)
  const knopf = useRef<HTMLButtonElement>(null)

  // **Das Menü steht fest am Fenster, gemessen am Knopf** — AP-6 (#2363):
  // `absolute` unter dem Knopf schnitt jeder Vorfahr mit `overflow` ab, an
  // der Deckzeile die Spalte des Navigators. Unten am Fensterrand klappt es
  // über den Knopf, wie ein Menü unter macOS; seitlich rückt es ein.
  const [lage, setzeLage] = useState({ left: 0, top: 0 })

  // messung zählt Rollen und Größenänderungen, während es offen ist — dann
  // misst es seinen Knopf neu und folgt ihm (Befund 5).
  const [messung, setzeMessung] = useState(0)

  // **Gemessen vor dem Zeichnen** — ein Layout-Effekt läuft, bevor das Menü
  // an der vorigen Stelle zu sehen wäre.
  useLayoutEffect(() => {
    const menue = wurzel.current?.querySelector('[role="menu"]')
    const am = knopf.current?.getBoundingClientRect()
    if (!offen || menue == null || am === undefined) {
      return
    }

    const rand = menue.getBoundingClientRect()
    const breite = rand.right - rand.left
    const hoehe = rand.bottom - rand.top
    let left = ausrichtung === 'rechts' ? am.right - breite : am.left
    left = Math.max(
      abstand,
      Math.min(left, window.innerWidth - abstand - breite),
    )
    const top =
      am.bottom + luft + hoehe > window.innerHeight - abstand
        ? Math.max(abstand, am.top - luft - hoehe)
        : am.bottom + luft

    if (left !== lage.left || top !== lage.top) {
      setzeLage({ left, top })
    }
  }, [offen, lage, ausrichtung, messung])

  // **Ein Klick daneben schließt** — und gibt den Fokus nicht zurück: wer
  // daneben klickt, will dorthin.
  useEffect(() => {
    if (!offen) {
      return
    }

    const aufDruck = (ereignis: MouseEvent) => {
      if (!wurzel.current?.contains(ereignis.target as Node)) {
        setzeOffen(false)
      }
    }

    // **Rollen und Verkleinern messen neu** — es steht fest am Fenster und
    // folgt seinem Knopf (Zweitblick über Phase 6, Befund 5). Nur das Rollen
    // im Menü selbst nicht.
    const neuMessen = () => {
      setzeMessung((bisher) => bisher + 1)
    }
    const aufRollen = (ereignis: Event) => {
      if (!wurzel.current?.contains(ereignis.target as Node)) {
        neuMessen()
      }
    }
    const aufGroesse = neuMessen

    document.addEventListener('mousedown', aufDruck)
    document.addEventListener('scroll', aufRollen, true)
    window.addEventListener('resize', aufGroesse)

    return () => {
      document.removeEventListener('mousedown', aufDruck)
      document.removeEventListener('scroll', aufRollen, true)
      window.removeEventListener('resize', aufGroesse)
    }
  }, [offen])

  // **Und ohne Rollen** — eine Liste, die sich umordnet, rückt den Knopf
  // weiter, ohne etwas zu melden (Nachlauf, R2).
  useAnkerwacht(
    offen,
    () => knopf.current,
    () => {
      setzeMessung((bisher) => bisher + 1)
    },
  )

  const schliessen = (zurueck: boolean) => {
    setzeOffen(false)

    if (zurueck) {
      knopf.current?.focus()
    }
  }

  const oeffnen = (von: 0 | -1) => {
    setzeStelle(naechsteStelle(eintraege, von, von === 0 ? 'Home' : 'End'))
    setzeOffen(true)
  }

  const aufTaste = (ereignis: KeyboardEvent<HTMLUListElement>) => {
    switch (ereignis.key) {
      case 'Escape':
        // **Verbraucht** — eine Schublade darum bleibt offen (UX-08, #1680).
        ereignis.preventDefault()
        ereignis.stopPropagation()
        schliessen(true)

        return
      case 'Tab':
        nachDemTab(() => {
          schliessen(false)
        })

        return
      case 'ArrowDown':
      case 'ArrowUp':
      case 'Home':
      case 'End':
        ereignis.preventDefault()
        setzeStelle(naechsteStelle(eintraege, stelle, ereignis.key))

        return
      default:
        return
    }
  }

  return (
    <div ref={wurzel} className="relative inline-block">
      <Knopf
        griff={knopf}
        gewicht={gewicht}
        {...(children === undefined ? {} : { bezeichnung })}
        hatMenue
        ohneTabstelle={ohneTabstelle}
        ausgeklappt={offen}
        steuert={menueID}
        aufKlick={() => {
          if (offen) {
            schliessen(true)
          } else {
            oeffnen(0)
          }
        }}
        aufPfeil={oeffnen}
      >
        {children ?? bezeichnung}
      </Knopf>

      {offen && (
        <Menueliste
          id={menueID}
          bezeichnung={bezeichnung}
          eintraege={eintraege}
          stelle={stelle}
          aufTaste={aufTaste}
          aufWahl={() => {
            schliessen(true)
          }}
          lage={{ position: 'fixed', ...lage }}
        />
      )}
    </div>
  )
}
