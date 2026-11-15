// **Ein Gitter ist eine Tabstelle** —
// #1290.
//
// # Wozu
//
// Gemessen am 2026-08-30 mit `scripts/serienmass.mjs` bei 1280 × 800: die
// Paketseite hatte **413 Tabstellen** bei 331 installierten Paketen — ein Knopf
// je Zeile. Wer mit der Tastatur an das Ende der Seite wollte, drückte
// vierhundertmal Tabulator.
//
// # Und die Begründung, die vorher dastand
//
// `Paketzeilen.tsx` hat sich ausdrücklich **gegen** einen wandernden Fokus
// entschieden: „am Ende dieselbe Bedienung, die eine Tabelle mit
// Umschaltknöpfen von selbst hat: `Tab` läuft die Knöpfe ab".
//
// **Der Satz stimmt, und sein Preis war nicht bekannt.** „Tab läuft die Knöpfe
// ab" ist bei fünf Zeilen eine Bedienung und bei 331 eine Sperre. Die
// Entscheidung fiel ohne die Zahl; jetzt gibt es sie.
//
// # Was dieser Haken hält
//
// Die **eine** Zeile, die den Tabulator annimmt. Alles andere — welche Taste
// wohin führt — steht bei den Zeilen selbst, denn das weiß nur, wer sie kennt.

import { useCallback, useState } from 'react'

export interface RovierendeZeile {
  // aktiv ist die Stelle der Zeile, die den Tabulator annimmt.
  aktiv: number

  // annehmen sagt, ob **diese** Zeile die Tabstelle ist.
  annehmen: (stelle: number) => boolean

  // rueckenAuf setzt die Tabstelle — beim Fokussieren und beim Blättern.
  rueckenAuf: (stelle: number) => void
}

export function useRovierendeZeile(anzahl: number): RovierendeZeile {
  const [aktiv, setzeAktiv] = useState(0)

  // **Geklemmt und nicht gemerkt.** Wird die Liste kürzer — ein Filter greift
  // —, zeigt eine gemerkte Stelle ins Leere, und **keine** Zeile nähme den
  // Tabulator an. Dann wäre das Gitter gar nicht mehr erreichbar.
  const geklemmt = anzahl === 0 ? 0 : Math.min(aktiv, anzahl - 1)

  return {
    aktiv: geklemmt,
    annehmen: useCallback((stelle: number) => stelle === geklemmt, [geklemmt]),
    rueckenAuf: setzeAktiv,
  }
}
