// **Wo ein Kontextmenü steht** — aus `Kontextmenue.tsx` herausgelöst, als
// es nach dem Nachlauf zum Zweitblick über Phase 6 von 0.9.0 über die
// Zeilengrenze wuchs (§2). Dort steht, **wann** es aufgeht und was die Tasten
// darin tun; hier, **wo** es steht und wann es ihm folgt.
//
// # Fest am Fenster, gehalten an einem Anker
//
// **Das Menü steht fest am Fenster** — AP-5 (#2362): `absolute` im Kasten
// schnitt jeder Vorfahr mit `overflow` ab. Es hängt an einem Anker — dem
// Kasten beim Rechtsklick, dem Element mit dem Fokus bei ⇧F10, der Zeile
// eines Anstoßes — mit der Stelle in ihm, und **folgt ihm**: beim Rollen,
// beim Verkleinern (Zweitblick, Befund 5) und, wenn eine Liste sich umordnet,
// je Bild (`useAnkerwacht`, Nachlauf R2). Es schließt erst, wenn es ihn nicht
// mehr gibt: ein Rechtsklick auf eine andere Karte rollt die Liste zu ihr,
// und ein Menü, das dabei schloss, sah niemand.
import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  type RefObject,
} from 'react'

import { useAnkerwacht } from '../haken/useAnkerwacht'

// Lage ist die Stelle des Menüs im Fenster.
export interface Lage {
  left: number
  top: number
}

// abstand hält ein eingerücktes Menü vom Fensterrand fern, in Punkten.
const abstand = 8

// Anstoss ist eine Stelle im Fenster, an der das Menü aufgehen soll — und
// seit dem Zweitblick über Phase 6 von 0.9.0 (Befund 5) die Zeile, an der
// geklickt wurde, mit der Stelle in ihr: steht sie beim Aufgehen woanders
// (die Deckliste sortiert „zuletzt benutzt“), folgt das Menü ihr.
export interface Anstoss {
  x: number
  y: number
  nummer: number
  anker?: Element | undefined
  versatz?: { x: number; y: number } | undefined
}

interface Anker {
  element: Element
  versatz: { x: number; y: number }
}

export interface Kontextmenuelage {
  // kasten gehört an den Rahmen — in ihm sucht der Haken das Menü.
  kasten: RefObject<HTMLElement | null>
  lage: Lage | null
  setzeLage: (lage: Lage | null) => void

  // verankern hängt das Menü an `element`, an der Stelle (x, y) im Fenster.
  verankern: (element: Element, x: number, y: number) => Lage

  // festmachen hängt es an die Zeile eines Anstoßes — wo sie jetzt steht,
  // oder an die Stelle des Klicks, wenn es sie nicht mehr gibt.
  festmachen: (anstoss: Anstoss) => Lage
}

export function useKontextmenuelage(): Kontextmenuelage {
  // **Der Rahmen nimmt das Menü auf, aber es steht fest am Fenster** — im
  // Baum bleibt es, damit Tab nach dem Auslöser weitergeht. **`HTMLElement`
  // und nicht `HTMLDivElement`**: der Rahmen ist seit dem Nachlauf zum
  // Zweitblick über Phase 5.1 auch ein `li` (Befund 3).
  const kasten = useRef<HTMLElement>(null)
  const [lage, setzeLage] = useState<Lage | null>(null)
  const anker = useRef<Anker | null>(null)

  const verankern = (element: Element, x: number, y: number): Lage => {
    const rand = element.getBoundingClientRect()
    anker.current = { element, versatz: { x: x - rand.left, y: y - rand.top } }

    return { left: x, top: y }
  }

  const festmachen = useCallback((anstoss: Anstoss): Lage => {
    const { anker: zeile, versatz } = anstoss
    if (zeile?.isConnected !== true || versatz === undefined) {
      anker.current = null

      return { left: anstoss.x, top: anstoss.y }
    }

    anker.current = { element: zeile, versatz }
    const rand = zeile.getBoundingClientRect()

    return { left: rand.left + versatz.x, top: rand.top + versatz.y }
  }, [])

  // **Das Menü bleibt im Fenster** — AP-5 (#2362). Am Zeiger geöffnet, ragte
  // es an einer unteren Zeile über den Rand, und was dort stand, war mit der
  // Maus nicht zu erreichen. Es rückt deshalb nach oben und nach links, mit
  // einem Rand, wie ein Menü unter macOS — nie über den oberen Rand hinaus.
  useLayoutEffect(() => {
    const menue = kasten.current?.querySelector('[role="menu"]')
    if (lage === null || menue == null) {
      return
    }

    // **Aus der Größe und der gesetzten Lage, nicht aus der gezeichneten
    // Stelle** — unter wenig Bewegung gleiten `top` und `left` 0,01 ms
    // lang (jedes Element trägt dann eine Übergangsdauer, und
    // `transition-property` ist von Haus aus `all`). Die Messung im selben
    // Bild lieferte die alte Stelle, das Einrücken rechnete von ihr weiter,
    // und nach fünfzig Schritten brach React ab (#185, gefunden von
    // `ap4zustaende`). So steht es nach einem Schritt.
    const rand = menue.getBoundingClientRect()
    const hoch = Math.min(
      lage.top + rand.height - window.innerHeight + abstand,
      lage.top - abstand,
    )
    const links = Math.min(
      lage.left + rand.width - window.innerWidth + abstand,
      lage.left - abstand,
    )
    if (hoch > 0 || links > 0) {
      setzeLage({
        top: lage.top - Math.max(hoch, 0),
        left: lage.left - Math.max(links, 0),
      })
    }
  }, [lage])

  // folgen stellt das Menü an seinen Anker — oder schließt es, wenn es ihn
  // nicht mehr gibt.
  const folgen = useCallback(() => {
    const am = anker.current
    if (am?.element.isConnected !== true) {
      setzeLage(null)

      return
    }

    const rand = am.element.getBoundingClientRect()
    setzeLage({
      left: rand.left + am.versatz.x,
      top: rand.top + am.versatz.y,
    })
  }, [])

  // **Ein Klick daneben schließt** — wie beim Menü hinter einem Knopf.
  const offen = lage !== null
  useEffect(() => {
    if (!offen) {
      return
    }

    const imMenue = (ziel: EventTarget | null) =>
      kasten.current?.querySelector('[role="menu"]')?.contains(ziel as Node) ===
      true

    const aufDruck = (ereignis: MouseEvent) => {
      if (!imMenue(ereignis.target)) {
        setzeLage(null)
      }
    }
    // **Nur das Rollen im Menü selbst bewegt es nicht.**
    const aufRollen = (ereignis: Event) => {
      if (!imMenue(ereignis.target)) {
        folgen()
      }
    }

    document.addEventListener('mousedown', aufDruck)
    document.addEventListener('scroll', aufRollen, true)
    window.addEventListener('resize', folgen)

    return () => {
      document.removeEventListener('mousedown', aufDruck)
      document.removeEventListener('scroll', aufRollen, true)
      window.removeEventListener('resize', folgen)
    }
  }, [offen, folgen])

  // **Und ohne Rollen** — eine Liste, die sich umordnet (Nachlauf, R2).
  useAnkerwacht(offen, () => anker.current?.element, folgen)

  return { kasten, lage, setzeLage, verankern, festmachen }
}
