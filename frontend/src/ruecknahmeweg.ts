// **Welcher Verlauf den Menübefehl bekommt** —
// #1340,
// Werkstück F von Phase 4.11.
//
// # Warum eine eigene Datei
//
// `Anwendung.tsx` steht an der Zeilengrenze aus §2, und der Schnitt folgt der
// Frage: *wer* den Befehl bekommt, steht hier; *dass* es ihn gibt, dort.
//
// # Die Regel in einem Satz
//
// **Die offene Zeichnung geht vor, solange es sie gibt.** Der Verlauf des
// Bestands nimmt zurück, was **geschrieben** ist; eine Zeichnung, an der
// gerade gearbeitet wird, ist es noch nicht.

import type { Ruecknahmeweg, Ruecknahmeziel } from './ruecknahme'

// ruecknahmeweg legt den Menübefehl auf den Verlauf, der ihn meint.
//
// **Die Entscheidung fällt beim Aufruf und nicht beim Zeichnen.** Ein
// `ziel.gemeldetes()` im Rumpf einer Komponente läse den Stand von vorhin: die
// Seite meldet sich beim Zeichnen an, und der Rahmen über ihr zeichnet zuerst.
// Deshalb sind die beiden Rückgaben Funktionen, die **dann** nachsehen.
export function ruecknahmeweg(
  ziel: Ruecknahmeziel,
  amBestand: Ruecknahmeweg,
): Ruecknahmeweg {
  return {
    zuruecknehmen: () => {
      ;(ziel.gemeldetes() ?? amBestand).zuruecknehmen()
    },
    wiederherstellen: () => {
      ;(ziel.gemeldetes() ?? amBestand).wiederherstellen()
    },
  }
}
