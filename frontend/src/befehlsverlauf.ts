// **Der Verlauf der Kommandozentrale — je Sitzung, ohne Geheimnisse** —
// Werkstück U12 von Phase 5.1 (#1423).
//
// Er lebt außerhalb des Dialogs, denn der wird abgebaut, wenn er zugeht —
// und wer ihn wieder öffnet, will sehen, was er vorhin getan hat. Nicht im
// Bestand: ein Verlauf mit Argumenten ist Bedienung, keine Fachlichkeit, und
// er stirbt mit dem Fenster.
import { useSyncExternalStore } from 'react'

export type Ausgang = 'gelungen' | 'gescheitert' | 'abgebrochen'

export interface Verlaufseintrag {
  nummer: number
  name: string
  argumente: Record<string, unknown>
  ausgang: Ausgang
  dauerMs: number
  zeitpunkt: string
}

// hoechstens ist, wie viele Einträge stehen bleiben — die jüngsten.
export const hoechstens = 20

let eintraege: Verlaufseintrag[] = []
let naechste = 1
const horcher = new Set<() => void>()

function melden() {
  for (const einer of horcher) {
    einer()
  }
}

// merke hängt einen Eintrag vorn an — die Argumente kommen **bereinigt** an,
// das ist Sache des Aufrufers (`bereinigt` aus `befehle.ts`).
export function merke(eintrag: Omit<Verlaufseintrag, 'nummer'>): void {
  eintraege = [{ ...eintrag, nummer: naechste++ }, ...eintraege].slice(
    0,
    hoechstens,
  )
  melden()
}

export function leere(): void {
  eintraege = []
  melden()
}

function abonniere(rueckruf: () => void): () => void {
  horcher.add(rueckruf)

  return () => {
    horcher.delete(rueckruf)
  }
}

function stand(): Verlaufseintrag[] {
  return eintraege
}

export function useBefehlsverlauf(): Verlaufseintrag[] {
  return useSyncExternalStore(abonniere, stand, stand)
}
