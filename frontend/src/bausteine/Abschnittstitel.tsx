// **Die Überschrift eines Abschnitts in der Stufe, die ihr Ort verlangt** —
// #1992.
//
// # Der Anlass
//
// Am Tisch sprangen die Überschriften von h1 („At the table") auf h3 („What
// lies where") — gemessen am 2026-10-01 in der Abnahme-VM mit VoiceOver. Der
// Tisch steht aber an zwei Orten: als eigene Seite unter einer h1 und beim
// Gast unter einer h2 (`Partiekopf`, `ebene`). Eine feste Stufe ist an einem
// der beiden falsch.
//
// # Wie
//
// **Der Ort sagt die Stufe, der Abschnitt fragt sie ab.** Ohne Angabe bleibt
// es h3 — so stand es bisher überall, und wer einen Abschnitt allein
// zeichnet, sieht keine Änderung.
import { createContext, useContext, type ReactNode } from 'react'

type Stufe = 2 | 3

const Abschnittsstufe = createContext<Stufe>(3)

// AbschnitteUnter setzt die Stufe der Abschnitte unter einer Überschrift
// der Stufe `ueber`.
export function AbschnitteUnter({
  ueber,
  children,
}: {
  ueber: 1 | 2
  children: ReactNode
}) {
  return (
    <Abschnittsstufe.Provider value={ueber === 1 ? 2 : 3}>
      {children}
    </Abschnittsstufe.Provider>
  )
}

interface Titeleigenschaften {
  id?: string
  className?: string
  children: ReactNode
}

// Abschnittstitel ist die Überschrift eines Abschnitts.
export function Abschnittstitel({
  id,
  className,
  children,
}: Titeleigenschaften) {
  const Ueberschrift = useContext(Abschnittsstufe) === 2 ? 'h2' : 'h3'

  return (
    <Ueberschrift id={id} className={className}>
      {children}
    </Ueberschrift>
  )
}

// Unterabschnittstitel ist eine Stufe darunter.
export function Unterabschnittstitel({
  id,
  className,
  children,
}: Titeleigenschaften) {
  const Ueberschrift = useContext(Abschnittsstufe) === 2 ? 'h3' : 'h4'

  return (
    <Ueberschrift id={id} className={className}>
      {children}
    </Ueberschrift>
  )
}
