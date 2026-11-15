// **Die Seitenleiste als eine Tabstelle** — #1488: ↑ und ↓ wandern zwischen
// den Einträgen, Home und End springen, und nur ein Eintrag trägt
// `tabIndex 0`. Derselbe Vertrag wie die Paketliste (#1290) und der
// Vorlagenbaum (U7); die Zielrechnung kommt von dort (`zielzeile`).
//
// **Eine Taste mit ⌘, ⌥, ⌃ oder ⇧ ist kein Wandern**: ⌘↓ gehört dem Fenster.
import type { KeyboardEvent } from 'react'

import { zielzeile } from '../seiten/gitterzeilentaste'

export function leistentaste(
  ereignis: KeyboardEvent<HTMLButtonElement>,
  stelle: number,
  anzahl: number,
  rueckenAuf: (stelle: number) => void,
  waagerecht = false,
): void {
  if (
    ereignis.metaKey ||
    ereignis.altKey ||
    ereignis.ctrlKey ||
    ereignis.shiftKey
  ) {
    return
  }

  const ziel = zielzeile(
    waagerecht ? quer(ereignis.key) : ereignis.key,
    stelle,
    anzahl,
  )
  if (ziel === undefined) {
    return
  }

  ereignis.preventDefault()
  rueckenAuf(ziel)

  // **Der Fokus geht mit** — der Halt allein ist nur für das nächste Tab.
  //
  // **Gesucht wird in der ganzen Leiste und nicht in der nächsten Liste** —
  // UX-02 (#1674).
  // Jeder Bereich hat seine eigene Liste, `ziel` aber zählt über die
  // **ganze** Leiste. Mit `closest('ul')` suchte ein ↓ im ersten Bereich
  // („Start", ein Eintrag) den zweiten in einer Liste, die nur einen hat, und
  // der Fokus bliebe stehen.
  const knoepfe = ereignis.currentTarget
    .closest('nav')
    ?.querySelectorAll<HTMLButtonElement>('button')
  knoepfe?.[ziel]?.focus()
}

// quer übersetzt die Pfeile einer Zeile in die einer Spalte — UX-14.AK-02
// (#1686):
// oben wandert → wie ↓ und ← wie ↑. **↑ und ↓ tun dort nichts**, wie in
// jeder waagerechten Leiste; Home und End bleiben.
function quer(taste: string): string {
  const uebersetzt: Record<string, string> = {
    ArrowRight: 'ArrowDown',
    ArrowLeft: 'ArrowUp',
    ArrowDown: '',
    ArrowUp: '',
  }

  return uebersetzt[taste] ?? taste
}
