// **Die gemerkte Ansicht eines Bereichs** — UX-11.AK-04
// (#1683):
// *„Ansichtspräferenzen sind je Fachbereich und Profil speicherbar."*
//
// **Dieselbe Gestalt wie `useFeldtiefe`**: gelesen wird, was gilt — beim
// Profil oder geerbt vom Menschen —, und eine Wahl gilt sofort und wird im
// Hintergrund geschrieben. Die Schlüssel stehen in `internal/domain/ansicht.go`.
import { useState } from 'react'

import { useEinstellungen } from './useEinstellungen'

// Ansichtswahl ist, was ein Bereich mit seiner Ansicht tun kann.
export interface Ansichtswahl<F extends string> {
  form: F
  setzeForm: (form: F) => void
}

// Erlaubt ist, was ein Bereich zeigen kann — **nie leer**: die erste ist die
// Vorgabe.
export type Erlaubt<F extends string> = readonly [F, ...F[]]

export function useAnsicht<F extends string>(
  schluessel: string,
  erlaubt: Erlaubt<F>,
): Ansichtswahl<F> {
  const einstellungen = useEinstellungen()
  const [gewaehlt, setzeGewaehlt] = useState<F>()
  const gelesen = einstellungen.wert(schluessel)

  // **Die erste ist die Vorgabe** — dieselbe wie in der Domäne. Ein Wert,
  // den dieser Bereich nicht kennt, kann nur aus einem von Hand geänderten
  // Bestand kommen.
  const [vorgabe] = erlaubt

  return {
    form:
      gewaehlt ??
      ((erlaubt as readonly string[]).includes(gelesen)
        ? (gelesen as F)
        : vorgabe),

    setzeForm: (form) => {
      setzeGewaehlt(form)
      einstellungen.setze(schluessel, form)
    },
  }
}
