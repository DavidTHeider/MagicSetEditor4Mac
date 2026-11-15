// **Die Fensterliste** — Werkstück U8 von Phase 5.1 (#1419): bei fester
// Zeilenhöhe stehen nur die Zeilen im Sichtfenster im Baum; die Auskunft für
// das Vorleseprogramm nennt trotzdem alle (`aria-rowcount`, `aria-rowindex`),
// und die Tastatur erreicht jede Zeile — auch die zehntausendste.
//
// # Warum keine Bibliothek
//
// Vermerk 0085: TanStack Virtual misst mit `getBoundingClientRect` und einem
// `ResizeObserver`, und jsdom rechnet kein Layout — jede Abdeckung prüfte eine
// Attrappe. Hier kommt die Höhe aus `clientHeight`, und die stellt ein Test.
//
// # Der Halt liegt am Datensatz
//
// Nicht an der Stelle: ändert ein Filter die Liste, bleibt der Halt bei
// derselben Zeile, solange es sie gibt — und fällt sonst auf die erste. Das
// ist §7, Regel 1 für eine Liste: der Bezug ist das Exemplar.
import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  type KeyboardEvent,
  type MouseEvent,
  type ReactNode,
} from 'react'

import { zielzeile } from '../seiten/gitterzeilentaste'
import { inDieZeile } from '../seiten/kartenlistetasten'

// Spaltenkopf ist eine Überschrift — allein oder mit der Breite ihrer
// Spalte. **Die Breite steht am Kopf wie an der Zelle** (AP-12, #2369): gleich
// verteilte Köpfe über ungleichen Zellen stellten „Menge“ über die
// Bezeichnung.
export type Spaltenkopf = string | { titel: string; klasse: string }

interface FensterlisteEigenschaften<T> {
  bezeichnung: string
  eintraege: readonly T[]
  zeilenhoehe: number
  // schluessel benennt eine Zeile — mit ihrem Eintrag **und ihrer Stelle**.
  //
  // **Die Stelle ist da, seit eine Liste ihre Zeilen nachholt** (#1500): wer
  // Platzhalter zeigt, braucht einen Schlüssel, der beim Nachrücken derselbe
  // bleibt — sonst tauscht React den Knoten aus, und der Fokus geht mit.
  schluessel: (eintrag: T, stelle: number) => string | number
  // spalten sind die Überschriften der Kopfzeile — sie zählt als Zeile 1.
  // Ein Kopf mit Klasse ist so breit wie die Zellen seiner Spalte; einer
  // ohne teilt sich den Platz gleich mit den anderen.
  spalten: readonly Spaltenkopf[]
  // zeichne gibt die Zellen einer Zeile (`role="gridcell"`).
  zeichne: (eintrag: T, stelle: number) => ReactNode
  // ueberhang ist, wie viele Zeilen über und unter dem Fenster mitstehen.
  ueberhang?: number

  // aufFenster meldet, welche Zeilen gerade gebraucht werden — Werkstück
  // #1500.
  //
  // **Für Listen, die ihre Zeilen erst holen.** Wer alle hat, braucht es
  // nicht; wer zehntausend hat, holt sie sonst alle über die Naht.
  aufFenster?: ((anfang: number, ende: number) => void) | undefined
  // aufHalt meldet, welche Zeile die Tastatur gerade hält.
  aufHalt?: ((eintrag: T) => void) | undefined
  // aufKlick meldet einen Klick auf eine Zeile.
  aufKlick?: ((eintrag: T) => void) | undefined
  // aufAnstoss bekommt Rechtsklick und Tasten einer Zeile **zuerst** — für
  // ein Kontextmenü über `useAnstoss` (AP-6, Teil 2, #2363). Verbraucht er
  // die Taste (`preventDefault`), wandert der Halt nicht mit.
  aufAnstoss?:
    | ((
        ereignis: MouseEvent<HTMLElement> | KeyboardEvent<HTMLElement>,
        eintrag: T,
      ) => void)
    | undefined
  // ausgewaehlt sagt, welche Zeile `aria-selected` trägt — die Wahl des
  // Aufrufers, nicht der Halt der Tastatur.
  ausgewaehlt?: ((eintrag: T) => boolean) | undefined
}

const mindestzeilen = 10

// **Eine Zeile antwortet dem Zeiger** — AP-4
// (#2361),
// Befund B-11: der Grund `betont` unter dem Zeiger, in `--dauer-kurz`; und die
// Hand nur dort, wo ein Klick die Zeile selbst wählt.
const zeilenklassen =
  'absolute left-0 right-0 flex items-center gap-2 border-b border-kante px-2 text-sm ' +
  'transition-colors duration-(--dauer-kurz) hover:bg-betont ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-ring'
const waehlbar = `${zeilenklassen} cursor-pointer`

export function Fensterliste<T>({
  bezeichnung,
  eintraege,
  zeilenhoehe,
  schluessel,
  spalten,
  zeichne,
  ueberhang = 8,
  aufFenster,
  aufHalt,
  aufKlick,
  aufAnstoss,
  ausgewaehlt,
}: FensterlisteEigenschaften<T>) {
  const fenster = useRef<HTMLDivElement>(null)
  const [scrollTop, setzeScrollTop] = useState(0)
  const [hoehe, setzeHoehe] = useState(0)
  const [halt, setzeHalt] = useState<string | number>()
  // fokusGewuenscht merkt sich einen Sprung bis zum nächsten Zeichnen — als
  // Referenz, nicht als Zustand: der Effekt unten liest und löscht sie, ohne
  // ein weiteres Zeichnen anzustoßen.
  const fokusGewuenscht = useRef(false)

  // **Die Höhe kommt aus dem Fenster selbst** — beim Zeichnen und bei jeder
  // Größenänderung des Fensters. Ohne Layout (jsdom, ein noch leeres
  // Fenster) gelten zehn Zeilen, damit immer etwas dasteht.
  useLayoutEffect(() => {
    const messen = () => {
      /* v8 ignore next -- die Referenz steht ab dem Zeichnen, und der Horcher
         geht vor dem Abbau; `null` ist nur für den Typ da. */
      setzeHoehe(fenster.current?.clientHeight ?? 0)
    }
    messen()
    window.addEventListener('resize', messen)

    return () => {
      window.removeEventListener('resize', messen)
    }
  }, [])

  // **Wird die Zeile höher, bleibt dieselbe oben** — Nachlauf zum Zweitblick
  // über Phase 6 von 0.9.0, zweiter Teil, R5: die Rollstelle stand sonst in
  // alten Pixeln, und eine gehaltene Zeile fiel aus dem Ausschnitt. Vor dem
  // Malen, damit kein Bild mit der falschen Stelle entsteht.
  const vorigeZeilenhoehe = useRef(zeilenhoehe)
  useLayoutEffect(() => {
    const vorher = vorigeZeilenhoehe.current
    if (vorher === zeilenhoehe) {
      return
    }
    vorigeZeilenhoehe.current = zeilenhoehe

    const element = fenster.current
    /* v8 ignore next 3 -- die Referenz steht ab dem Zeichnen; `null` ist nur
       für den Typ da. */
    if (element === null) {
      return
    }

    const neu = Math.round((element.scrollTop * zeilenhoehe) / vorher)
    element.scrollTop = neu
    setzeScrollTop(neu)
  }, [zeilenhoehe])

  const sichtbareHoehe = hoehe > 0 ? hoehe : zeilenhoehe * mindestzeilen
  const anzahl = eintraege.length
  const anfang = Math.max(0, Math.floor(scrollTop / zeilenhoehe) - ueberhang)
  const ende = Math.min(
    anzahl,
    Math.ceil((scrollTop + sichtbareHoehe) / zeilenhoehe) + ueberhang,
  )

  // **Das gebrauchte Fenster wird gemeldet** — nach dem Zeichnen, nicht
  // währenddessen: ein Ruf im Rumpf setzte den Zustand des Aufrufers während
  // des eigenen Zeichnens.
  useEffect(() => {
    aufFenster?.(anfang, ende)
  }, [aufFenster, anfang, ende])

  // **Ohne Halt wird nicht gesucht.** Solange niemand eine Zeile angefasst
  // hat, ist `halt` leer — und eine Suche danach liefe über die ganze Liste,
  // bei jedem Zeichnen, um die erste Zeile zu finden, die ohnehin gilt.
  const haltstelle =
    halt === undefined
      ? 0
      : Math.max(
          0,
          eintraege.findIndex(
            (eintrag, stelle) => schluessel(eintrag, stelle) === halt,
          ),
        )

  const springeZu = useCallback(
    (ziel: number) => {
      const eintrag = eintraege[ziel]
      /* v8 ignore next 3 -- unerreichbar: `zielzeile` klemmt das Ziel in die
         Liste, und ohne Zeilen gibt es keine Taste; der Typ verlangt die
         Prüfung trotzdem. */
      if (eintrag === undefined) {
        return
      }

      setzeHalt(schluessel(eintrag, ziel))
      fokusGewuenscht.current = true
      aufHalt?.(eintrag)

      // **Die Zeile kommt ins Fenster**, bevor sie den Fokus bekommt — sonst
      // steht sie nicht im Baum.
      const oben = ziel * zeilenhoehe
      const unten = oben + zeilenhoehe
      const element = fenster.current
      /* v8 ignore next 3 -- unerreichbar: gerufen wird aus der Taste einer
         gezeichneten Zeile, und die steht im Fenster, dessen Referenz seit dem
         Zeichnen gesetzt ist. */
      if (element === null) {
        return
      }
      if (oben < element.scrollTop) {
        element.scrollTop = oben
        setzeScrollTop(oben)
      } else if (unten > element.scrollTop + sichtbareHoehe) {
        element.scrollTop = unten - sichtbareHoehe
        setzeScrollTop(unten - sichtbareHoehe)
      }
    },
    [aufHalt, eintraege, schluessel, sichtbareHoehe, zeilenhoehe],
  )

  // Nach dem Zeichnen bekommt die gehaltene Zeile den Fokus — einmal je Sprung.
  useEffect(() => {
    if (!fokusGewuenscht.current) {
      return
    }
    fokusGewuenscht.current = false
    const zeile =
      fenster.current?.querySelector<HTMLElement>(`[data-halt="ja"]`)
    zeile?.focus()
  }, [scrollTop, halt])

  const aufTaste = (
    ereignis: KeyboardEvent<HTMLDivElement>,
    stelle: number,
  ) => {
    if (
      ereignis.metaKey ||
      ereignis.altKey ||
      ereignis.ctrlKey ||
      ereignis.shiftKey
    ) {
      return
    }

    // **→ und ← führen in die Zeile und zurück** (#2249) — wie in jedem
    // Gitter dieser Anwendung. Ohne das stand jedes Bedienelement einer
    // Zeile einzeln im Tabweg.
    if (ereignis.key === 'ArrowRight' || ereignis.key === 'ArrowLeft') {
      inDieZeile(ereignis)

      return
    }

    const ziel = zielzeile(ereignis.key, stelle, anzahl)
    if (ziel === undefined) {
      return
    }

    ereignis.preventDefault()
    springeZu(ziel)
  }

  return (
    <div
      ref={fenster}
      data-testid="fenster"
      onScroll={(ereignis) => {
        setzeScrollTop(ereignis.currentTarget.scrollTop)
      }}
      // **Ein Behälter**: eine Zeile fragt nach der Breite der Liste, nicht
      // nach der des Fensters — zwischen Navigator und Inspector ist sie
      // schmal, auch wenn das Fenster breit ist (AP-12, #2369).
      className="@container relative h-full min-h-0 overflow-auto"
    >
      <div
        role="grid"
        aria-label={bezeichnung}
        aria-rowcount={anzahl + 1}
        className="relative"
        style={{ height: `${String((anzahl + 1) * zeilenhoehe)}px` }}
      >
        <div
          role="row"
          aria-rowindex={1}
          className="sticky top-0 z-10 flex items-center gap-2 border-b border-kante bg-flaeche px-2 text-xs font-medium text-leise"
          style={{ height: `${String(zeilenhoehe)}px` }}
        >
          {spalten.map((spalte) => {
            const titel = typeof spalte === 'string' ? spalte : spalte.titel
            const klasse =
              (typeof spalte === 'string' ? 'flex-1' : spalte.klasse) +
              ' truncate'

            return (
              <div key={titel} role="columnheader" className={klasse}>
                {titel}
              </div>
            )
          })}
        </div>
        {eintraege.slice(anfang, ende).map((eintrag, versatz) => {
          const stelle = anfang + versatz
          const gehalten = stelle === haltstelle

          return (
            <div
              key={schluessel(eintrag, stelle)}
              role="row"
              aria-rowindex={stelle + 2}
              tabIndex={gehalten ? 0 : -1}
              data-halt={gehalten ? 'ja' : undefined}
              aria-selected={ausgewaehlt?.(eintrag)}
              onClick={() => {
                aufKlick?.(eintrag)
              }}
              onContextMenu={(ereignis) => {
                aufAnstoss?.(ereignis, eintrag)
              }}
              onKeyDown={(ereignis) => {
                aufAnstoss?.(ereignis, eintrag)
                if (!ereignis.defaultPrevented) {
                  aufTaste(ereignis, stelle)
                }
              }}
              onFocus={(ereignis) => {
                // **Nur die Zeile selbst**, nicht ein Haken darin: `focus`
                // steigt in React auf, und ein Klick in ein Bedienelement der
                // Zeile ist keine Wahl der Zeile.
                if (ereignis.target === ereignis.currentTarget && !gehalten) {
                  setzeHalt(schluessel(eintrag, stelle))
                  aufHalt?.(eintrag)
                }
              }}
              className={aufKlick === undefined ? zeilenklassen : waehlbar}
              style={{
                top: `${String((stelle + 1) * zeilenhoehe)}px`,
                height: `${String(zeilenhoehe)}px`,
              }}
            >
              {zeichne(eintrag, stelle)}
            </div>
          )
        })}
      </div>
    </div>
  )
}
