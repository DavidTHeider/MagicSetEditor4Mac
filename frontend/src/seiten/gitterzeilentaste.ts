// **Der Tastenvertrag einer Gitterzeile** —
// #1290.
//
// **Eigene Datei, weil zwei Gitter ihn brauchen** und ihre Zeilen sonst nichts
// gemeinsam haben: die Kartenliste wählt beim Bewegen eine Karte aus, die
// Paketliste bewegt nur den Fokus. Was gleich ist, sind die Tasten.
//
// # Woher die Tasten kommen
//
// Aus der Plattform, nicht aus MSE2: ↑ und ↓ zwischen den Zeilen, Pos1 und
// Ende an die Enden, → und ← in die Zeile hinein und zurück. So wird ein
// Gitter überall bedient, und MSE2 bekommt es von `wxListCtrl` geschenkt
// (`src/gui/control/card_list.cpp`, Fassung f9d9356 vom 2021-07-06).

import type { KeyboardEvent } from 'react'

import { inDieZeile } from './kartenlistetasten'

// gitterzeilentaste bewegt den Fokus zwischen und in den Zeilen.
//
// **Sie fokussiert selbst und meldet die neue Stelle** — anders als in der
// Kartenliste, wo der Fokus der **Auswahl** folgt und die Zeile später kommt.
// Hier steht sie schon da: die Paketliste zeichnet alle Zeilen.
export function gitterzeilentaste(
  ereignis: KeyboardEvent<HTMLElement>,
  stelle: number,
  anzahl: number,
  rueckenAuf: (stelle: number) => void,
): void {
  if (
    ereignis.metaKey ||
    ereignis.altKey ||
    ereignis.ctrlKey ||
    ereignis.shiftKey
  ) {
    return
  }

  if (ereignis.key === 'ArrowRight' || ereignis.key === 'ArrowLeft') {
    inDieZeile(ereignis)

    return
  }

  const ziel = zielzeile(ereignis.key, stelle, anzahl)
  if (ziel === undefined) {
    return
  }

  ereignis.preventDefault()
  rueckenAuf(ziel)

  // **Der Fokus wandert über den Nachbarn im DOM**, nicht über eine Kennung:
  // die Zeilen dieser Tabelle haben keine, und eine hinzuzufügen wäre ein
  // Attribut für einen einzigen Leser.
  const zeilen = [
    ...(ereignis.currentTarget.parentElement?.children ?? []),
  ].filter((knoten): knoten is HTMLElement => knoten instanceof HTMLElement)

  zeilen[ziel]?.focus()
}

// zielzeile sagt, wohin eine Taste führt — geklemmt, ohne Umlauf.
// **Exportiert seit #1488**: die Seitenleiste wandert mit denselben Tasten.
export function zielzeile(
  taste: string,
  stelle: number,
  anzahl: number,
): number | undefined {
  switch (taste) {
    case 'ArrowDown':
      return Math.min(stelle + 1, anzahl - 1)

    case 'ArrowUp':
      return Math.max(stelle - 1, 0)

    case 'Home':
      return 0

    case 'End':
      return anzahl - 1

    default:
      return undefined
  }
}
