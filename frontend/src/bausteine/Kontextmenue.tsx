// **Ein Kontextmenü, das auch ohne rechte Maustaste aufgeht** — Werkstück
// U2 von Phase 5.1
// (#1413),
// nach [Vermerk 0077](../../../docs/entscheidungen/0077-eigene-primitive-statt-radix.md).
//
// # Die Regel, für die es dieses Bauteil gibt
//
// **Keine Aktion, die es nur im Kontextmenü gibt** — das ist die Abnahme von
// U2. Ein Kontextmenü ist eine Abkürzung für die, die es kennen; wer es nicht
// kennt, findet dieselben Einträge in einem sichtbaren `Menue`. Deshalb nimmt
// dieses Bauteil **dieselbe** Liste entgegen wie jenes, und die Seite gibt
// beiden dieselbe.
//
// # Wie es aufgeht
//
//   - mit der rechten Maustaste (`contextmenu`) — am Zeiger,
//   - mit ⇧F10 oder der Menütaste — unter dem Element, das den Fokus hat.
//
// Danach ist es eine `Menueliste`: Pfeile, ⏎, Escape, Fokus zurück. Der Kasten
// darunter bleibt, was er ist — eine Zeile der Kartenliste, eine Zeile der
// Paketliste —, und dieses Bauteil legt nur den Horcher darüber.
import {
  useEffect,
  useId,
  useRef,
  useState,
  type KeyboardEvent,
  type MouseEvent,
  type ReactNode,
} from 'react'

import { fokusZurueck, merkeRueckkehr, nachDemTab } from './fokus'
import {
  useKontextmenuelage,
  type Anstoss,
  type Lage,
} from './kontextmenuelage'
import { Menueliste, naechsteStelle, type Menueeintrag } from './Menueliste'

interface KontextmenueEigenschaften {
  bezeichnung: string
  eintraege: readonly Menueeintrag[]

  // element sagt, als **was** der Kasten im Baum steht — Vorgabe `div`.
  //
  // # Warum das nötig ist
  //
  // **Weil ein `div` in einer Liste nichts zu suchen hat.** Der erste
  // Aufrufer war die Platzliste des Decks: `ul > Kontextmenue > li` ergab
  // `ul > div > li`, und damit ist die Liste für ein Vorleseprogramm keine
  // mehr (axe, Regel `list`). Zugleich verlor das `li` sein `grid-cols-subgrid`
  // — es war kein direktes Kind des Gitters mehr, und die Spalten der Zeilen
  // liefen auseinander. Angestrichen vom Nachlauf zum Zweitblick über Phase
  // 5.1 (Befund 3).
  //
  // **Nur Elemente ohne eigene Bedeutung**: `div` und `li`. Wer ein `button`
  // daraus machte, bekäme ein Bedienelement mit zwei Bedeutungen.
  element?: 'div' | 'li'

  // klasse reicht die Stilklassen an dasselbe Element durch — sonst bräuchte
  // der Aufrufer wieder einen Kasten darum.
  klasse?: string | undefined

  // zeile macht den Kasten selbst zur Fokusstelle einer Liste — UXR-11
  // (#1919).
  //
  // **Weil ⇧F10 dort ankommen muss, wo der Fokus steht.** Die Plätze eines
  // Decks nehmen den Tabulator an **einer** Zeile an und wandern mit den
  // Pfeilen; stünde die Fokusstelle in einem Kasten um dieses Bauteil, hörte
  // es die Taste nie — ein Ereignis steigt nach oben, nicht nach unten.
  //
  // **Seine Taste kommt nach der des Menüs**, und was das Menü schon
  // verbraucht hat (`defaultPrevented`), sieht die Zeile nicht mehr: sonst
  // wanderte sie mit, während die Pfeile im offenen Menü wandern.
  zeile?: Zeilengriffe

  // objekt nennt Art und Kennung des Objekts, dessen Liste hier aufgeht —
  // als `data-objekt` am Rahmen, damit eine Probe am Fenster die Stelle
  // findet, ohne den Aufbau der Seite zu kennen (GESTE, #2358).
  objekt?: string | undefined

  // anstoss öffnet das Menü an einer Stelle des Fensters, sobald es Einträge
  // hat — AP-5 (#2362). **Für den Rechtsklick auf eine andere Zeile**: sie
  // wird erst gewählt, und das Menü gehört dann ihr. Jeder Anstoß öffnet
  // einmal; danach meldet [anstossVerbraucht], damit er nicht wiederkommt.
  anstoss?: Anstoss | undefined
  anstossVerbraucht?: (() => void) | undefined

  // nurAngestossen lässt das Menü **nur** über [anstoss] aufgehen — AP-5
  // (#2362). Für eine Liste, die ihre Zeilen selbst anstößt: ein Rechtsklick
  // auf eine Gruppe, einen Abstandhalter oder die leere Fläche zeigte sonst
  // das Menü des gewählten Objekts, also eines anderen.
  nurAngestossen?: boolean | undefined

  // anstossen bekommt Rechtsklick und Tasten des Rahmens **zuerst** — AP-6
  // (#2363). Für einen Kasten, der selbst keine Zeile mit Griffen ist, etwa
  // eine Deckzeile: verbraucht er das Ereignis (`preventDefault`), geht das
  // eigene Menü nicht auf.
  anstossen?:
    | ((ereignis: MouseEvent<HTMLElement> | KeyboardEvent<HTMLElement>) => void)
    | undefined

  children: ReactNode
}

export type { Anstoss } from './kontextmenuelage'

// Zeilengriffe sind, was eine Zeile braucht, die selbst den Fokus nimmt.
export interface Zeilengriffe {
  tabIndex: number
  onFocus: () => void
  onKeyDown: (ereignis: KeyboardEvent<HTMLElement>) => void
  onPointerEnter: () => void
  onPointerLeave: () => void
}

// imFeld sagt, ob ein Rechtsklick in einem Feld landet — **dann gehört er
// macOS** (GESTE, #2358): Kopieren, Einsetzen und die Rechtschreibung stehen
// dort, und ein Menü um eine ganze Karte nähme sie ihren Feldern.
//
// **Nur der Rechtsklick**: ⇧F10 hat in einem Feld unter macOS keine eigene
// Bedeutung, und dort ist es der Weg ohne Maus zu den Handlungen.
export function imFeld(ziel: EventTarget): boolean {
  return (
    ziel instanceof Element &&
    ziel.closest(
      'input, textarea, [contenteditable]:not([contenteditable="false"])',
    ) !== null
  )
}

// oeffnetKontextmenue sagt, ob eine Taste das Kontextmenü meint — ⇧F10 oder
// die Menütaste, die manche Tastaturen tragen.
export function oeffnetKontextmenue(taste: string, umschalt: boolean): boolean {
  return taste === 'ContextMenu' || (taste === 'F10' && umschalt)
}

export function Kontextmenue({
  bezeichnung,
  eintraege,
  element = 'div',
  klasse = '',
  zeile,
  objekt,
  anstoss,
  anstossVerbraucht,
  nurAngestossen = false,
  anstossen,
  children,
}: KontextmenueEigenschaften) {
  const [stelle, setzeStelle] = useState(0)
  const rueckkehr = useRef<Element | null>(null)
  const { kasten, lage, setzeLage, verankern, festmachen } =
    useKontextmenuelage()

  const menueID = useId()

  // **Ein anderes Objekt schließt das offene Menü** — AP-5 (#2362). Seine
  // Einträge gehörten dem vorigen; stehen blieben sie unter fremdem Namen.
  const [fuerObjekt, setzeFuerObjekt] = useState(objekt)
  if (fuerObjekt !== objekt) {
    setzeFuerObjekt(objekt)
    setzeLage(null)
  }

  const oeffnen = (bei: Lage) => {
    rueckkehr.current = merkeRueckkehr()
    setzeStelle(naechsteStelle(eintraege, -1, 'Home'))
    setzeLage(bei)
  }

  const schliessen = (zurueck: boolean) => {
    setzeLage(null)

    if (zurueck) {
      fokusZurueck(rueckkehr.current)
    }
  }

  // **Der Anstoß wird geöffnet, sobald Einträge dastehen** — die des neu
  // gewählten Objekts kommen erst mit einem späteren Zeichnen.
  // **Der Anstoß von außen kommt über einen Verweis** — seine Nummer löst
  // aus, die übrigen Angaben liest der Effekt beim Aufgehen.
  const anstossNummer = anstoss?.nummer
  const anstossJetzt = useRef(anstoss)
  useEffect(() => {
    anstossJetzt.current = anstoss
  })
  const hatEintraege = eintraege.length > 0
  useEffect(() => {
    const jetzt = anstossJetzt.current
    if (anstossNummer === undefined || jetzt === undefined || !hatEintraege) {
      return
    }

    rueckkehr.current = merkeRueckkehr()
    setzeStelle(0)
    setzeLage(festmachen(jetzt))
    anstossVerbraucht?.()
  }, [anstossNummer, hatEintraege, anstossVerbraucht, festmachen, setzeLage])

  const aufRechtsklick = (ereignis: MouseEvent<HTMLElement>) => {
    if (nurAngestossen || eintraege.length === 0 || imFeld(ereignis.target)) {
      return
    }

    ereignis.preventDefault()

    oeffnen(
      verankern(ereignis.currentTarget, ereignis.clientX, ereignis.clientY),
    )
  }

  const aufTasteImMenue = (ereignis: KeyboardEvent<HTMLUListElement>) => {
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

  // **⇧F10 und die Menütaste** — unter dem Element, das den Fokus hat. Nur,
  // solange das Menü zu ist: offen hört die Liste selbst.
  const aufTasteImKasten = (ereignis: KeyboardEvent<HTMLElement>) => {
    if (
      nurAngestossen ||
      lage !== null ||
      eintraege.length === 0 ||
      !oeffnetKontextmenue(ereignis.key, ereignis.shiftKey)
    ) {
      return
    }

    ereignis.preventDefault()

    const ziel = ereignis.target as HTMLElement
    const rand = ziel.getBoundingClientRect()
    oeffnen(verankern(ziel, rand.left, rand.bottom))
  }

  // **Der Rahmen ist ein Element, kein Bauteil** — deshalb die Zuweisung an
  // einen großgeschriebenen Namen; JSX liest ein kleines Wort als Tag. Der
  // Typ ist der weiteste, den beide erfüllen: `div` und `li` unterscheiden
  // sich in ihren Eigenschaften, nicht in dem, was hier gebraucht wird.
  const Rahmen = element as unknown as React.ElementType<{
    ref: React.Ref<HTMLElement>
    className: string
    tabIndex: number | undefined
    onFocus: (() => void) | undefined
    onPointerEnter: (() => void) | undefined
    onPointerLeave: (() => void) | undefined
    onContextMenu: (ereignis: MouseEvent<HTMLElement>) => void
    onKeyDown: (ereignis: KeyboardEvent<HTMLElement>) => void
    'data-objekt': string | undefined
    children: ReactNode
  }>

  return (
    // **Ein Element mit zwei Horchern und keiner Rolle.** Es ist kein
    // Bedienelement — die Zeilen darin sind es. Eine Ausnahme von
    // `jsx-a11y/no-static-element-interactions` braucht es nicht: der Rahmen
    // kommt über eine Variable, die Regel sieht ihn nicht und meldete die
    // Ausnahme als überflüssig.
    <Rahmen
      ref={kasten}
      className={`relative ${klasse}`.trim()}
      data-objekt={objekt}
      tabIndex={zeile?.tabIndex}
      onFocus={zeile?.onFocus}
      onPointerEnter={zeile?.onPointerEnter}
      onPointerLeave={zeile?.onPointerLeave}
      onContextMenu={(ereignis: MouseEvent<HTMLElement>) => {
        anstossen?.(ereignis)
        if (!ereignis.defaultPrevented) {
          aufRechtsklick(ereignis)
        }
      }}
      onKeyDown={(ereignis: KeyboardEvent<HTMLElement>) => {
        anstossen?.(ereignis)
        if (!ereignis.defaultPrevented) {
          aufTasteImKasten(ereignis)
        }

        if (!ereignis.defaultPrevented) {
          zeile?.onKeyDown(ereignis)
        }
      }}
    >
      {children}

      {lage !== null && (
        <Menueliste
          id={menueID}
          bezeichnung={bezeichnung}
          eintraege={eintraege}
          stelle={stelle}
          aufTaste={aufTasteImMenue}
          aufWahl={() => {
            schliessen(true)
          }}
          lage={{ position: 'fixed', ...lage }}
        />
      )}
    </Rahmen>
  )
}
