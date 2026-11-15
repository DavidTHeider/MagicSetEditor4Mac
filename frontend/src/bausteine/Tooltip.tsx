// **Ein Tooltip, der auch ohne Maus erscheint** — Werkstück U2 von Phase 5.1
// (#1413),
// nach [Vermerk 0077](../../../docs/entscheidungen/0077-eigene-primitive-statt-radix.md).
//
// # Warum nicht `title`
//
// Ein `title` erscheint unter dem Zeiger und nirgends sonst: wer mit der
// Tastatur auf einen Knopf kommt, sieht ihn nicht, und ein Vorleseprogramm
// liest ihn je nach Laune. Hier steht der Text als `role="tooltip"` im Baum,
// der Auslöser nennt ihn über `aria-describedby`, und er erscheint bei Zeiger
// **und** Fokus.
//
// # Warum ein Render-Rückruf und kein `cloneElement`
//
// Der Auslöser braucht drei Dinge: die Kennung des Tooltips, die Griffe für
// Zeiger und Fokus. Sie an ein fremdes Kind zu klonen hieße, dessen
// Eigenschaften zu erraten; der Rückruf gibt sie ausdrücklich, und der
// Aufrufer setzt sie dorthin, wo sie hingehören.
//
// # Warum drei Zustände und kein Zeitgeber in einem Ref
//
// Der erste Anlauf hielt die Frist in einem `useRef`, und die Griffe lasen
// ihn — `react-hooks/refs` weist das ab, weil der Rückruf während des
// Zeichnens gerufen wird und niemand belegen kann, dass die Griffe nicht
// auch dort laufen. Der zweite Anlauf setzte den Zustand in einem Effekt, und
// `react-hooks/set-state-in-effect` weist **das** ab. Also: Zeiger, Fokus und
// „mit Escape verborgen" sind Tatsachen aus den Griffen; was der Zeiger nach
// der Frist bewirkt, setzt der Zeitgeber — und was sichtbar ist, wird daraus
// gerechnet, nicht gespeichert.
//
// # Was er nicht tut
//
//   - **Kein Tooltip an einem Element ohne Namen.** Er beschreibt, er
//     benennt nicht — ein Knopf mit nur einem Zeichen braucht `bezeichnung`,
//     nicht dies.
//   - **Keine Verzögerung beim Fokus.** Wer mit Tab ankommt, wartet nicht.
import { useEffect, useId, useState, type ReactNode } from 'react'

import { useVerdeckt } from '../verdeckung'

// Tooltipgriffe ist, was der Auslöser bekommt.
//
// **Ohne die Zeigergriffe, und das ist die Korrektur aus Befund 4** des
// Zweitblicks über Phase 5.1: der Zeiger ist eine Aussage über die **Fläche**,
// der Fokus über das **Element**. Lagen beide am Auslöser, verschwand der Text
// in dem Augenblick, in dem der Zeiger ihn erreichen wollte — WCAG 1.4.13
// verlangt ausdrücklich das Gegenteil („hoverable").
export interface Tooltipgriffe {
  'aria-describedby': string
  onFocus: () => void
  onBlur: () => void
}

interface TooltipEigenschaften {
  text: string

  // breite sagt, ob der Text umbrechen darf — UX-15.AK-02
  // (#1771).
  //
  // **Schmal ist die Vorgabe und bleibt es**: ein Tooltip an einem Knopf
  // trägt zwei Wörter, und ein Umbruch machte daraus zwei Zeilen ohne Not.
  // **Breit gibt es, seit eine Erklärung dahinter steht** — dreihundert
  // Zeichen ohne Umbruch liefen aus dem Fenster hinaus.
  breite?: 'schmal' | 'breit'

  // verzoegerung ist die Zeit unter dem Zeiger, bevor der Text erscheint —
  // ein Zeiger, der nur vorbeigeht, soll nichts aufreißen.
  verzoegerung?: number

  // druck sagt, was ein Druck auf Auslöser oder Text mit ihm tut (#2068).
  // **Er räumt ihn, wo der Auslöser etwas tut** — dort liegt der Hinweis
  // sonst auf dem nächsten Knopf. **Er bleibt, wo der Druck ihn sehen
  // will**: am Fragezeichen der Feldhilfe ist der Hinweis die Antwort.
  druck?: 'raeumt' | 'bleibt'

  children: (griffe: Tooltipgriffe) => ReactNode
}

export function Tooltip({
  text,
  breite = 'schmal',
  verzoegerung = 300,
  druck = 'raeumt',
  children,
}: TooltipEigenschaften) {
  const kennung = useId()
  const [zeiger, setzeZeiger] = useState(false)
  const [zeigerLange, setzeZeigerLange] = useState(false)
  const [fokus, setzeFokus] = useState(false)
  const [verborgen, setzeVerborgen] = useState(false)

  // **Ein Druck räumt ihn** (#2068). Er liegt unter seinem Auslöser und
  // fängt den Zeiger — in einer Reihe von Knöpfen also auf dem nächsten, und
  // wer klickte und gleich darunter weiterklickte, traf ihn. **Bis der Zeiger
  // neu kommt oder der Fokus geht**: der Fokus, den der Klick selbst setzt,
  // holt ihn nicht zurück.
  const [gedrueckt, setzeGedrueckt] = useState(false)

  // **Der Zeiger zählt erst nach der Frist.** Der Zeitgeber setzt den
  // Zustand; geht der Zeiger vorher, räumt der Effekt die Frist mit.
  useEffect(() => {
    if (!zeiger) {
      return undefined
    }

    const frist = setTimeout(() => {
      setzeZeigerLange(true)
    }, verzoegerung)

    return () => {
      clearTimeout(frist)
    }
  }, [zeiger, verzoegerung])

  const sichtbar = !verborgen && !gedrueckt && (fokus || zeigerLange)

  // **Unter einer Schublade horcht er nicht** (UX-08, #1680): ein Hinweis,
  // der beim Öffnen noch stand, nähme ihr sonst das erste Escape weg.
  const verdeckt = useVerdeckt()

  // **Escape verbirgt ihn** — WCAG 1.4.13: was bei Zeiger oder Fokus
  // erscheint, muss sich ohne Bewegung wieder schließen lassen. Bis der
  // Zeiger oder der Fokus neu kommt, bleibt er verborgen.
  useEffect(() => {
    if (!sichtbar || verdeckt) {
      return undefined
    }

    const aufTaste = (ereignis: KeyboardEvent) => {
      if (ereignis.key === 'Escape') {
        // **Verbraucht, weil er zu sehen war** — der Druck gilt dem Hinweis
        // und nicht der Schublade, in der er steht (UX-08, #1680). Ein
        // verborgener horcht gar nicht erst.
        ereignis.preventDefault()
        setzeVerborgen(true)
      }
    }

    // **In der Capture-Phase**, also vor jedem Behälter (G12 aus dem
    // Zweitblick über Phase 6.3): `Entscheidung` und `Popover` hielten die
    // Taste an ihrer Gruppe an, der `Dialog` schloss vorher — der Hinweis
    // sah sie nie. `preventDefault` sagt jedem danach, dass sie genommen ist.
    window.addEventListener('keydown', aufTaste, true)

    return () => {
      window.removeEventListener('keydown', aufTaste, true)
    }
  }, [sichtbar, verdeckt])

  return (
    // **Der Zeiger hängt am Rahmen und nicht am Auslöser** — er umschließt
    // beide, und der Weg vom einen zum anderen führt damit nicht durch
    // „nirgends" (WCAG 1.4.13).
    <span
      className="relative inline-flex"
      onMouseEnter={() => {
        setzeVerborgen(false)
        setzeGedrueckt(false)
        setzeZeiger(true)
      }}
      onPointerDown={() => {
        setzeGedrueckt(druck === 'raeumt')
      }}
      onMouseLeave={() => {
        setzeZeiger(false)
        setzeZeigerLange(false)
      }}
    >
      {children({
        'aria-describedby': kennung,
        onFocus: () => {
          setzeVerborgen(false)
          setzeFokus(true)
        },
        onBlur: () => {
          setzeFokus(false)
          setzeGedrueckt(false)
        },
      })}

      {/* **Der Text steht immer im Baum**, sichtbar oder nicht — sonst zeigte
          `aria-describedby` ins Leere, solange niemand darüber ist, und ein
          Vorleseprogramm läse die Beschreibung nie. */}
      <span
        id={kennung}
        role="tooltip"
        hidden={!sichtbar}
        className={
          'absolute left-1/2 top-full z-40 mt-1 -translate-x-1/2 rounded-md ' +
          'border-2 border-kante bg-flaeche px-2 py-1 text-sm text-text ' +
          (breite === 'breit' ? 'w-64 max-w-[80vw]' : 'whitespace-nowrap')
        }
      >
        {text}
      </span>
    </span>
  )
}
