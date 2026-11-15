// **Wohin ein Profilwechsel meldet** — Werkstück 3 von Phase 4.12
// (#1351).
//
// # Warum es diese Stelle gibt
//
// **§7, Regel 1: ein Wechsel des Bezugsobjekts räumt den offenen Kontext.**
// Beim Set ist das Bezugsobjekt eine Zeile im Bestand; beim Profil ist es
// **alles** — Sets, Decks, Sammlung, Turniere, Einstellungen. Ein offenes
// Kartenformular nach einem Wechsel zeigte eine Karte, die dieses Profil nicht
// hat, und sein „Speichern" schriebe ins Leere.
//
// **Deshalb wird die ganze Anwendung neu aufgebaut**, über einen `key` in
// `App.tsx`. Das ist die dritte Gestalt aus `docs/SCHREIBWEGE.md` — der `key`,
// der Regel 1 und Regel 2 zugleich einlöst: was noch unterwegs ist, kommt in
// einem Bauteil an, das es nicht mehr gibt, und wird verworfen.
//
// # Warum ein Kontext und keine Eigenschaft
//
// Dieselbe Begründung wie bei `ablage.tsx` und `ruecknahme.tsx`: der Weg von
// `App` bis zur Nutzerseite führt über vier Bauteile, die mit Profilen nichts
// zu tun haben. Eine Eigenschaft durch alle hindurchzureichen hieße, sie alle
// zu ändern, sobald jemand eine zweite Stelle baut, an der gewechselt wird.

import { createContext, useContext, type ReactNode } from 'react'

// Profilwechsel ist, was eine Stelle meldet, die gewechselt hat.
export type Profilwechsel = () => void

// **Ohne Raum passiert nichts.** Ein Bauteil außerhalb — im Test, in einer
// Vorschau — soll nicht scheitern; es hat schlicht keine Anwendung, die sich
// neu aufbauen ließe.
// **Der Vorgabewert tut ausdrücklich nichts** — er ist keine vergessene
// Belegung. `no-empty-function` sieht das nicht; deshalb steht es hier.
// eslint-disable-next-line @typescript-eslint/no-empty-function
const Raum = createContext<Profilwechsel>(() => {})

export function Profilraum({
  gewechselt,
  children,
}: {
  gewechselt: Profilwechsel
  children: ReactNode
}) {
  return <Raum.Provider value={gewechselt}>{children}</Raum.Provider>
}

// useProfilwechsel gibt die Meldestelle.
export function useProfilwechsel(): Profilwechsel {
  return useContext(Raum)
}
