// **Wer beim Auslösen den Fokus hatte, bekommt ihn zurück** —
// #1990.
//
// # Der Anlass
//
// Ein `Knopf` ist während seines Vorgangs `disabled` (§7, Regel 4), und
// **ein fokussiertes Element, das `disabled` wird, verliert den Fokus an den
// Seitenanfang** — so will es HTML, und WebKit tut es. Gemessen am
// 2026-10-01 in der Abnahme-VM mit VoiceOver: nach *Add card*, nach
// *Create code* und nach *Choose picture …* mit Escape landete das nächste ⇥
// bei „Skip to content". Wer mit der Tastatur arbeitet, stand danach am
// Anfang der Seite.
//
// # Warum zurückgeben und nicht `aria-disabled`
//
// `aria-disabled` ließe den Fokus stehen, änderte aber, was „angehalten"
// heißt: 169 Zusicherungen in 106 Testdateien lesen es als `disabled`, und
// ein Knopf ohne `disabled` nimmt Klicks an, die jeder Aufrufer selbst
// abweisen müsste. Zurückgegeben wird deshalb nur, was verloren ging.
//
// # Auch für andere Elemente als den Knopf
//
// **Dasselbe gilt für jedes Bedienelement, das für die Dauer eines Schreibens
// `disabled` ist** — UXR-F2
// (#2008):
// der Schalter einer Einstellung und die Optionsfelder einer Segmentreihe.
// Wer mit der Leertaste umlegt oder mit dem Pfeil wählt, soll danach dort
// stehen, wo er war. Eine Reihe von Elementen gibt dafür das Element, das
// den Fokus hatte, beim Merken mit.
//
// # Was es nicht tut
//
// - **Es nimmt keinen Fokus, den das Element beim Auslösen nicht hatte**: ein
//   Klick mit dem Zeiger gibt in WebKit keinen (`bausteine/fokus.ts`).
// - **Es holt ihn nicht von woanders zurück**: wer inzwischen mit ⇥
//   weitergegangen ist, bleibt dort. Zurück geht er nur vom Seitenanfang.
import { useEffect, useRef, type RefObject } from 'react'

export function useFokusUeberAnhalten(
  knopf: RefObject<HTMLElement | null>,
  angehalten: boolean,
): (element?: HTMLElement | null) => void {
  // gemerkt ist das Element, wenn es beim Auslösen den Fokus hatte — sonst
  // nichts.
  const gemerkt = useRef<HTMLElement | null>(null)
  const vorher = useRef(angehalten)

  useEffect(() => {
    const warAngehalten = vorher.current
    vorher.current = angehalten

    if (!warAngehalten || angehalten) {
      return
    }

    // **Sichtbar zurückgegeben** — gemessen am Fenster (`uxrf2`, #2008):
    // WebKit gibt einem per Skript gesetzten Fokus kein `:focus-visible`, und
    // der Ring blieb aus, bis die nächste Taste kam. Gemerkt wird nur, wer den
    // Fokus beim Auslösen hatte — in WebKit fokussiert ein Mausklick keinen
    // Knopf (`fokus.ts`), also kommt der Ring nur nach der Tastatur zurück.
    if (gemerkt.current !== null && document.activeElement === document.body) {
      gemerkt.current.focus({ focusVisible: true })
    }

    gemerkt.current = null
  }, [angehalten])

  // **Gemerkt beim Auslösen**: dann steht der Fokus noch dort, wo der
  // Mensch ihn hingesetzt hat. Ohne Angabe ist es der eine Knopf; eine Reihe
  // von Optionsfeldern nennt das, das sich eben geändert hat.
  return (element = knopf.current) => {
    gemerkt.current = document.activeElement === element ? element : null
  }
}
