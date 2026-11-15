// **Die Wurzelschrift, die gerade gilt** — Nachlauf zum Zweitblick über
// Phase 6 von 0.9.0, zweiter Teil, R5
// (#2369).
//
// `wurzelschrift()` liest sie beim Aufruf; beim Zeichnen gelesen, stand nach
// einem Wechsel der Schriftgröße die alte da, denn das Fenster setzt die neue
// erst in einem Effekt. Dieser Haken zeichnet neu, sobald
// `setzeSchriftgroesse` es meldet.
//
// **Was er nicht hört**: eine größere Schrift in den Systemeinstellungen,
// solange die Anwendung läuft — sie meldet sich nicht am Wurzelelement.
import { useSyncExternalStore } from 'react'

import { wurzelschrift } from '../bausteine/arbeitsplatzmasse'
import { schriftgroessenereignis } from '../erscheinungsbild'

function horchen(geaendert: () => void): () => void {
  const wurzel = document.documentElement
  wurzel.addEventListener(schriftgroessenereignis, geaendert)

  return () => {
    wurzel.removeEventListener(schriftgroessenereignis, geaendert)
  }
}

export function useWurzelschrift(): number {
  return useSyncExternalStore(horchen, wurzelschrift)
}
