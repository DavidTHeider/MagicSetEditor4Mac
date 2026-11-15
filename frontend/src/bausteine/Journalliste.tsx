// **Was sich im eigenen Bestand bewegt hat** — Werkstück T6 von Phase 5.6
// (#1471).
//
// # Warum hier der Lagerort steht und im Beleg nicht
//
// **Weil das Journal das eigene Buch ist.** Der Beleg geht hinaus und trägt
// nur Öffentliches; hier steht, aus welcher Kiste das Stück kam — und genau
// das braucht jemand, der es wiederfinden will.
//
// # Warum hier der Name steht und nicht die Kennung des Drucks
//
// **Weil am Fenster „−2 × a1b2c3d4-0000-4000-8000-00000000000a" stand.**
// Gesehen bei der Sichtprüfung am 2026-09-18; ein DOM-Lauf fängt das nicht,
// weil „steht die Menge da" auch dann zutrifft. Nachgeschlagen wird der Name
// beim Lesen — im Journal steht der öffentliche Ausschnitt.
//
// # Und warum eine Korrektur ihren Grund zeigt
//
// **Weil eine Buchung, die niemand erklärt, im Journal schlimmer ist als
// keine.** Sie steht da, sie stimmt nicht mit dem Bestand überein, und
// niemand weiß mehr, warum.
//
// # Und warum zehn Zeilen vorn stehen
//
// **Weil ein Tausch gegen eine Sammelmappe achtzig Bewegungen hat** (UXR-15,
// #1923;
// R4 in `docs/UIREVIEW.md`). Die Liste steht im Kasten „Was sich bewegt
// hat", der sie benennt; eine zweite Überschrift darüber wäre derselbe Name
// zweimal.
import { Faltliste } from './Faltliste'

import type { Journalzeile } from '../apihandelsverbuchung'

// dieZeilenVorn und derSchritt sind, wie viele Bewegungen zu Anfang dastehen
// und wie viele ein Klick dazu zeigt.
const dieZeilenVorn = 10
const derSchritt = 20

interface Journallisteneigenschaften {
  titel: string
  zeilen: Journalzeile[]
  leer: string
}

export function Journalliste({
  titel,
  zeilen,
  leer,
}: Journallisteneigenschaften) {
  if (zeilen.length === 0) {
    return <p className="text-sm text-leise">{leer}</p>
  }

  return (
    <Faltliste
      eintraege={zeilen}
      vorn={dieZeilenVorn}
      schritt={derSchritt}
      bezeichnung={titel}
      klasse="flex flex-col gap-1"
      zeile={(zeile) => (
        <li key={zeile.id} className="text-sm tabular-nums">
          {zeile.art === 'abgang' ? '−' : '+'}
          {zeile.menge} × {zeile.bezeichnung} · {zeile.anlass}
          {zeile.lagerort !== '' && <> · {zeile.lagerort}</>}
          {zeile.grund !== '' && <> · {zeile.grund}</>}
        </li>
      )}
    />
  )
}
