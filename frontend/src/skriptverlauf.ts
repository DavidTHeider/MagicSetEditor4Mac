// **Der Verlauf der Skriptkonsole — je Sitzung, ohne Geheimnisse** —
// Werkstück U13 von Phase 5.1 (#1424).
//
// Er lebt außerhalb der Seite, so wie der Verlauf der Kommandozentrale
// (`befehlsverlauf.ts`): die Konsole wird beim Kartenwechsel abgebaut, und
// wer danach nachsehen will, was er vorhin gefragt hat, soll es noch finden.
// Nicht im Bestand: ein Verlauf ist Bedienung, keine Fachlichkeit, und er
// stirbt mit dem Fenster.
//
// **MSE2 hat ihn auch** — `HistoryTextCtrl` in `src/gui/set/console_panel.cpp`
// (Fassung f9d9356 vom 2021-07-06) hängt jeden Befehl an `history` und
// blättert mit ↑/↓ nach dem Präfix vor dem Zeiger. **adapted**: das Feld hier
// ist mehrzeilig, ↑/↓ wechseln die Zeile — der Verlauf steht als Liste unter
// dem Ergebnis, und ein Klick lädt zurück.
import { useSyncExternalStore } from 'react'

export type Skriptausgang = 'gelungen' | 'gescheitert' | 'abgebrochen'

export interface Skriptverlaufseintrag {
  nummer: number
  ausdruck: string
  ausgang: Skriptausgang
  dauerMs: number
  zeitpunkt: string
}

// hoechstens ist, wie viele Einträge stehen bleiben — die jüngsten.
export const hoechstens = 20

// redigiertesZeichen steht für einen Ausdruck, der ein Geheimnis tragen
// könnte.
export const redigiertesZeichen = '•••'

// geheimnisverdacht trifft, was nach Token, Kennwort oder Geheimnis klingt —
// in beiden Sprachen, denn ein Ausdruck ist keiner davon zugeordnet.
const geheimnisverdacht = /token|kennwort|passwor|geheim|secret/i

let eintraege: Skriptverlaufseintrag[] = []
let naechste = 1
const horcher = new Set<() => void>()

function melden() {
  for (const einer of horcher) {
    einer()
  }
}

// redigiert gibt den Ausdruck zurück — oder das Zeichen, wenn er nach einem
// Geheimnis klingt. **Als Ganzes**: die Konsole weiß nicht, welcher Teil das
// Geheimnis ist.
export function redigiert(ausdruck: string): string {
  return geheimnisverdacht.test(ausdruck) ? redigiertesZeichen : ausdruck
}

// merke hängt einen Lauf vorn an. Der Ausdruck kommt **roh** an und wird
// hier redigiert — damit kein Aufrufer es vergessen kann.
export function merke(eintrag: Omit<Skriptverlaufseintrag, 'nummer'>): void {
  eintraege = [
    { ...eintrag, ausdruck: redigiert(eintrag.ausdruck), nummer: naechste++ },
    ...eintraege,
  ].slice(0, hoechstens)
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

function stand(): Skriptverlaufseintrag[] {
  return eintraege
}

// useSkriptverlauf liefert den Verlauf und zeichnet neu, wenn er sich ändert.
export function useSkriptverlauf(): Skriptverlaufseintrag[] {
  return useSyncExternalStore(abonniere, stand, stand)
}
