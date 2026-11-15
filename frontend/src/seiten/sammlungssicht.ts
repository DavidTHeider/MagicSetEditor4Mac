// **Die Sicht auf die eigene Sammlung** —
// #1200.
//
// # Der Befund
//
// #1165
// verlangte eine Sicht zum **Suchen, Filtern und Sortieren**. Was dastand, lud
// die ganze Liste und bildete sie unverändert ab: `Sammlung.tsx` hatte außer
// `runde` und `fehler` keinen Zustand.
//
// # Warum hier und nicht im Dienst
//
// **Weil die Liste schon da ist.** `Sammlung()` gibt sie ganz — sie hängt an
// keinem Set und wächst mit dem, was jemand besitzt, nicht mit dem Pack. Ein
// Filter im Dienst wäre ein zweiter Ruf für eine Auswahl, die aus dem
// vorhandenen Stand hervorgeht, und zwei Stände derselben Liste liefen
// auseinander (§7, Regel 2).
//
// **Und weil er dann sofort greift**: wer tippt, sieht die Auswahl im selben
// Zeichenschritt.
//
// # Was das kostet, und wann es zu ändern wäre
//
// Bei zehntausend Exemplaren würde die ganze Liste weiterhin geholt und hier
// gefiltert. Das ist tragbar, solange eine Sammlung in den Arbeitsspeicher
// passt; wer die Grenze verschiebt, verschiebt sie in den Dienst — und dann
// gehört auch die Sortierung dorthin.
import type { ExemplarDTO, Finish } from '../api'
import type { Erlaubt } from '../haken/useAnsicht'

// zeilengrenze ist die Zahl der Zeilen, die höchstens gezeichnet werden.
//
// # Warum es sie gibt
//
// Gemessen am 2026-08-30 an der laufenden Brücke gegen 5.000 Exemplare
// (`scripts/sammlungsmass.mjs`): **1183–1489 ms**, um die Seite aufzubauen.
// Davon liegen 167 ms in der Wails-Naht und 4,5 ms im Go-Weg — der Rest ist
// React, das 5.000 Zeilen zeichnet
// (#1292).
//
// # Warum 200
//
// **Weil es mehr ist, als auf einen Bildschirm passt, und wenig genug, um
// nicht zu kosten.** Bei 1280 × 800 stehen rund zehn Zeilen sichtbar da; 200
// sind zwanzig Bildschirme und damit reichlich Vorlauf für jeden, der scrollt,
// statt zu suchen.
//
// **Sie ist geraten und gemessen**, nicht nur geraten: die Zahl steht in
// `docs/SAMMLUNG.md` neben dem, was sie kostet.
export const zeilengrenze = 200

// Ausfuehrung ist der Filter über die physische Ausführung.
//
// **Drei und nicht zwei, seit T2** (#1467, 0119): `etched` ist eine eigene
// Veredelung und nicht der Gleichstand eines Foils.
//
// **Abgeleitet und nicht abgeschrieben**: eine vierte Ausführung wäre
// damit von selbst eine vierte Wahl, statt an einer zweiten Stelle zu
// fehlen.
export type Ausfuehrung = 'alle' | Finish

// Sortierung ist die Ordnung der Liste.
//
// **Drei und nicht neun.** Nach Zustand oder Währung zu sortieren beantwortet
// keine Frage, die jemand an seine Sammlung stellt; nach Karte, Menge und
// Erwerbsdatum sehr wohl.
export type Sortierung = 'bezeichnung' | 'menge' | 'erwerbsdatum'

// **Ansicht und Gruppierung** — Werkstück U8 von Phase 5.1 (#1419).
//
// **Farbe und Typ warten auf A3** (#1442),
// nicht auf A2: hier stand A2, und dessen Druckzeile führt Name, Setcode,
// Nummer und Sprache — keine Farbe und keinen Typ (Befund 2 des Zweitblicks
// über Phase 5.1). Die MTGJSON-Sammelausgabe aus A3 trägt beides je Druck;
// vorher wäre die Gruppierung geraten (Vermerk 0085).
export type Ansicht = 'tabelle' | 'galerie' | 'gruppen'
export type Gruppierung = 'keine' | 'set' | 'lagerort' | 'zustand' | 'sprache'

export interface Sicht {
  suche: string
  ausfuehrung: Ausfuehrung
  sortierung: Sortierung
  ansicht: Ansicht
  gruppierung: Gruppierung
}

// **Die gemerkte Ansicht** — UX-11.AK-04 (#1683). Der Schlüssel steht in
// `internal/domain/ansicht.go`, und die erste Form ist die Vorgabe.
export const sammlungsansichtsschluessel = 'arbeitsplatz.ansicht.sammlung'
export const sammlungsformen: Erlaubt<Ansicht> = [
  'tabelle',
  'galerie',
  'gruppen',
]

// leereSicht ist der Stand, mit dem die Seite aufgeht.
//
// **Nach der Karte geordnet**, weil das die Frage ist, mit der jemand seine
// Sammlung ansieht: *was habe ich?* Die Reihenfolge des Bestands ist nach
// Kennung sortiert und damit für einen Menschen zufällig.
export const leereSicht: Sicht = {
  suche: '',
  ausfuehrung: 'alle',
  sortierung: 'bezeichnung',
  ansicht: 'tabelle',
  gruppierung: 'keine',
}

// **Gesucht, gefiltert und geordnet wird im Bestand**, nicht hier: die
// Sammlung kommt als Fenster aus `Sammlungsfenster`, geordnet nach der
// Sprache, die das Fenster mitschickt (`internal/domain/sammlungssicht.go`).
// Hier stand bis F01 (#2243) eine zweite Ordnung samt Suche — ohne Aufrufer
// außer ihrem eigenen Test, und mit einer Ordnung in der Sprache des
// Systems.

// Abschnitt ist eine Gruppe der Sicht — mit Schlüssel, Exemplaren und der
// Summe ihrer Mengen. Ohne Gruppierung gibt es genau einen, ohne Schlüssel.
export interface Abschnitt {
  schluessel: string
  exemplare: readonly ExemplarDTO[]
  menge: number
}

// gruppiert teilt die **gesehene** Liste in Abschnitte — in der Reihenfolge
// ihres ersten Vorkommens, damit die Sortierung der Sicht die Gruppen ordnet.
// **Ein leerer Schlüssel ist eine eigene Gruppe am Ende**: „ohne Lagerort" ist
// eine Antwort, kein Loch.
export function gruppiert(
  exemplare: readonly ExemplarDTO[],
  gruppierung: Gruppierung,
): Abschnitt[] {
  if (exemplare.length === 0) {
    return []
  }

  const abschnitte = new Map<string, ExemplarDTO[]>()
  for (const eines of exemplare) {
    const schluessel = schluesselVon(eines, gruppierung)
    const gruppe = abschnitte.get(schluessel)
    if (gruppe === undefined) {
      abschnitte.set(schluessel, [eines])
    } else {
      gruppe.push(eines)
    }
  }

  const fertig = [...abschnitte.entries()].map(([schluessel, gruppe]) => ({
    schluessel,
    exemplare: gruppe,
    menge: gruppe.reduce((summe, eines) => summe + eines.menge, 0),
  }))

  return [
    ...fertig.filter((a) => a.schluessel !== ''),
    ...fertig.filter((a) => a.schluessel === ''),
  ]
}

function schluesselVon(
  exemplar: ExemplarDTO,
  gruppierung: Gruppierung,
): string {
  switch (gruppierung) {
    case 'set':
      return exemplar.setcode
    case 'lagerort':
      return exemplar.lagerort
    case 'zustand':
      return exemplar.zustand
    case 'sprache':
      return exemplar.sprache
    default:
      return ''
  }
}
