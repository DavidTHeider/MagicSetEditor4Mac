// **Eine kurze Eingabe dort, wo sie hingehört** — UX-08.AK-01
// (#1680):
// *„Popover für kurze kontextuelle Eingaben"*, nach dem Artboard
// *BelegMuster*, auf Mantine wie `Dialog` und `Schublade`
// ([0125](../../../docs/entscheidungen/0125-mantine-als-bauteilbibliothek.md)).
//
// # Was es zusagt
//
//   - **Es steht an seinem Auslöser** und über dem, was darunter liegt —
//     nicht in einer eigenen Zeile, die die Liste verschiebt.
//   - **Es wird von keinem Behälter abgeschnitten, und am Rand weicht es
//     aus** — **Mantine**: es steht in einem Portal und nimmt die Seite, auf
//     der Platz ist (`flip`, `shift`). Im Behälter seines Auslösers endete
//     etwa das Namensfeld im Navigator der Decks an dessen Kante — sichtbar
//     in der WebKit-Probe, nicht in einem Lauf: jsdom rechnet kein Layout
//     (§3).
//   - **Der Fokus geht hinein** (auf das erste Feld) und kehrt beim
//     Schließen zum Auslöser zurück — über `merkeRueckkehr`, auch in WebKit,
//     wo ein Klick keinen Fokus gibt. **Eigen**, aus demselben Grund wie beim
//     `Dialog`: Mantines Rückgabe merkte sich `document.activeElement`.
//   - **Escape schließt es** — **eigen**, denn Mantines Escape läuft weiter
//     (`close-on-escape.mjs`, Mantine 9.6.1, hält die Taste nicht an), und
//     dann schlösse derselbe Druck auch, was darunter liegt.
//   - **Ein Klick daneben schließt es** — **eigen**, in der Capture-Phase:
//     Mantine horcht auf `mousedown` in der Bubble-Phase, und ein Element,
//     das die Taste für sich nimmt, ließe es offen.
//   - **Es ist nicht modal**: `role="group"` mit Namen, wie die
//     `Entscheidung`. Was daneben steht, bleibt bedienbar; `modalregel.test.ts`
//     hält, dass es keine vierte Unterbrechung wird.
//   - **Kante und kein Schatten** (Vermerk 0065, Festlegung 9) — deshalb
//     `unstyled`: Mantines Kasten brächte Rand und Fläche als Klasse ohne
//     Schicht mit, und gegen die verlören die Tailwind-Klassen.
//   - **Die Textfarbe steht am Kasten**, wie beim `Tooltip`: im Portal erbt
//     er sie nicht vom Rahmen, sondern von `body` — und dort steht Mantines
//     helles Schwarz. Im dunklen Erscheinungsbild war der Titel darum kaum
//     zu lesen, gesehen in der WebKit-Probe am 2026-09-25.
import { Popover as Mantinepopover } from '@mantine/core'
import { useEffect, useId, useRef, type ReactNode, type RefObject } from 'react'

import { useBewegungsdauer } from '../haken/useBewegungsdauer'
import { useRueckkehr } from '../haken/useRueckkehr'

interface PopoverEigenschaften {
  // ausloeser ist, woran es steht — meist der Knopf, der es öffnet.
  ausloeser: ReactNode

  offen: boolean

  // titel benennt die Gruppe — ein Vorleseprogramm sagt ihn beim Eintritt.
  titel: string

  schliessen: () => void

  children: ReactNode
}

export function Popover({
  ausloeser,
  offen,
  titel,
  schliessen,
  children,
}: PopoverEigenschaften) {
  const titelID = useId()
  const dauer = useBewegungsdauer('mittel')
  const kasten = useRef<HTMLDivElement>(null)
  const anker = useRef<HTMLSpanElement>(null)

  return (
    <Mantinepopover
      opened={offen}
      position="bottom-start"
      offset={4}
      unstyled
      withRoles={false}
      closeOnEscape={false}
      closeOnClickOutside={false}
      // **Tab bleibt drin** (Zweitblick über Phase 6.3, M4): das Popover
      // steht im Portal hinter `#root`, und Tab vom letzten Feld ging sonst
      // ans Ende des Dokuments, während es offen blieb. Mantines Falle
      // (`FocusTrap`, aktiv, solange es offen ist) nimmt den Fokus beim
      // Aufgehen auf das erste Feld — dasselbe, was `Inhalt` tut; den Weg
      // zurück hält weiter `Inhalt`, deshalb bleibt `returnFocus` aus.
      trapFocus
      // **Es erscheint in `--dauer-mittel`** (AP-4, #2361) — die Zahl kommt
      // aus dem Stilblatt, und unter wenig Bewegung ist sie 0.
      transitionProps={{ duration: dauer }}
    >
      <Mantinepopover.Target>
        <span ref={anker} className="inline-flex">
          {ausloeser}
        </span>
      </Mantinepopover.Target>

      <Mantinepopover.Dropdown
        ref={kasten}
        role="group"
        aria-labelledby={titelID}
        // **Die Gruppe hört Escape von ihren Feldern** — dieselbe Ausnahme
        // wie in `Entscheidung.tsx`, aus demselben Grund.
        onKeyDown={(ereignis) => {
          // **Nicht, wenn ein Hinweis darin sie schon genommen hat** —
          // das Innerste zuerst (G12, WCAG 1.4.13).
          if (ereignis.key === 'Escape' && !ereignis.defaultPrevented) {
            ereignis.preventDefault()
            ereignis.stopPropagation()
            schliessen()
          }
        }}
        // **Eine Ebene über der Seite** — AP-4 (#2361), Vermerk 0065
        // Festlegung 11: die Kante bleibt, eine Schattenstufe kommt hinzu,
        // mehr Kontrast nimmt sie weg.
        className="absolute flex min-w-64 flex-col gap-2 rounded-md border-2 border-kante bg-flaeche p-3 text-text shadow-ueberlage contrast-more:shadow-none"
      >
        <Inhalt kasten={kasten} anker={anker} schliessen={schliessen}>
          <p id={titelID} className="text-sm font-medium">
            {titel}
          </p>
          {children}
        </Inhalt>
      </Mantinepopover.Dropdown>
    </Mantinepopover>
  )
}

// Inhalt steht, solange das Popover offen ist — und nimmt beim Aufgehen den
// Fokus, beim Schließen gibt es ihn zurück.
function Inhalt({
  kasten,
  anker,
  schliessen,
  children,
}: {
  kasten: RefObject<HTMLDivElement | null>

  // anker umschließt den Auslöser — das Ziel, wenn der gemerkte Fokus
  // nichts taugt (`useRueckkehr`, `ersatz`).
  anker: RefObject<HTMLSpanElement | null>
  schliessen: () => void
  children: ReactNode
}) {
  // **Nur beim echten Aufgehen und Schließen** (`useRueckkehr`, G11) — und
  // zurück nur, wenn der Fokus sonst verloren wäre: mit dem Popover ging
  // auch das Feld, in dem er stand. Wer inzwischen woanders arbeitet, dem
  // nimmt ein spätes Schließen ihn nicht — ein Sichern kann fertig werden,
  // wenn der Fokus längst in der nächsten Zeile steht.
  //
  // **Ohne zu rollen**: in dem Augenblick steht das Popover womöglich noch
  // nicht an seinem Platz, und ein Feld oben links zöge die Seite dorthin.
  useRueckkehr(
    kasten,
    () =>
      kasten.current
        ?.querySelector<HTMLElement>('input, select, textarea, button')
        ?.focus({ preventScroll: true }),
    true,
    () => anker.current?.querySelector('button, a[href], [tabindex]'),
  )

  // **Ein Klick daneben schließt es** — in der Capture-Phase, damit ein Knopf
  // daneben, der etwas anderes öffnet, nicht erst nach dem Popover dran ist.
  useEffect(() => {
    const aufDruck = (ereignis: PointerEvent) => {
      const drinnen = ereignis
        .composedPath()
        .some((knoten) => knoten === kasten.current)

      if (!drinnen) {
        schliessen()
      }
    }

    document.addEventListener('pointerdown', aufDruck, true)

    return () => {
      document.removeEventListener('pointerdown', aufDruck, true)
    }
  }, [kasten, schliessen])

  return children
}
