// **Wie es um das Schreiben steht** — UX-05.AK-03
// (#1677):
// *„… einen eindeutigen Speicherstatus."*
//
// # Warum es das gibt
//
// **Geschrieben wird beim Verlassen eines Feldes**, nicht mit einem Knopf —
// und bis hierher sagte nichts, ob es geklappt hat. Ein Fehler stand am Feld,
// ein Erfolg nirgends; wer das Feld verließ, ohne hinzusehen, wusste nicht,
// ob seine Eingabe im Bestand steht.
//
// # Warum ein Speicher außerhalb von React
//
// **Weil Schreiber und Anzeige nicht nebeneinander stehen**: geschrieben wird
// im Karteneditor, gezeigt über ihm. Ein Zustand dort müsste nach oben
// gereicht werden, und der Editor steht an der Zeilengrenze aus §2.
//
// # Warum je Set
//
// **Weil ein Fehler zu dem Set gehört, in dem er entstand** (§7, Regel 1):
// wer ein anderes Set öffnet, sieht nicht den Fehlschlag von drüben — und
// kommt er zurück, steht er wieder da.
//
// # Was nicht meldet
//
// **Der Verweis auf einen Druck** (`Kartenverweis.tsx`) schreibt am Stand
// vorbei: er kennt sein Set nicht, und ein Schreiben ohne Set hätte hier
// keinen Platz. Sein Erfolg und sein Fehler stehen an ihm selbst — der
// Satz „Jede Änderung steht im Bestand." sagt über ihn nichts (Nachlauf über
// Phase 6.3, N17).
import { useSyncExternalStore } from 'react'

export type Speicherlage =
  // still: seit dem Öffnen wurde hier nichts geschrieben.
  | { art: 'still' }
  | { art: 'schreibt' }
  | { art: 'geschrieben' }
  // fehler: das Feld, dessen Wert nicht im Bestand steht.
  | { art: 'fehler'; feld: string }

// Stand ist, was ein Set über sein Schreiben weiß.
//
// **Fehler je Karte und Feld, Zähler je Set** — angestrichen vom Zweitblick
// über Phase 6.3 (M8): hier stand **ein** Fehlerplatz je Set, der nur den
// Feldnamen kannte, und **ein** Zähler für alle Sets. Ein geglücktes Feld
// räumte den Fehler eines anderen, ein geglücktes `name` auf Karte 2 den von
// Karte 1, und ein Schreiben, das in A endete, ersetzte B's Zähler — und
// „Jede Änderung steht im Bestand." stand da, obwohl ein Wert fehlte.
interface Stand {
  offen: number

  // fehler führt je `Karte/Feld` das Feld, dessen Wert nicht im Bestand
  // steht — in der Reihenfolge, in der sie scheiterten.
  fehler: ReadonlyMap<string, string>
  geschrieben: boolean
}

const leer: Stand = { offen: 0, fehler: new Map(), geschrieben: false }

let staende: ReadonlyMap<string, Stand> = new Map()
const hoerer = new Set<() => void>()

// setze tauscht den Stand eines Sets und sagt es jedem Hörer.
function setze(setID: string, neu: Stand): void {
  staende = new Map(staende).set(setID, neu)
  for (const hoere of hoerer) {
    hoere()
  }
}

// imSet gibt den Stand dieses Sets — ein anderes fängt leer an.
function imSet(setID: string): Stand {
  return staende.get(setID) ?? leer
}

// schreibenBeginnt merkt, dass ein Wert unterwegs ist.
export function schreibenBeginnt(setID: string): void {
  const jetzt = imSet(setID)

  setze(setID, { ...jetzt, offen: jetzt.offen + 1 })
}

// schreibenEndet merkt, wie es ausging — ein geglücktes Schreiben räumt den
// Fehler **desselben Feldes derselben Karte**, und keinen anderen.
export function schreibenEndet(
  setID: string,
  karteID: string,
  feld: string,
  geglueckt: boolean,
): void {
  const jetzt = imSet(setID)
  const fehler = new Map(jetzt.fehler)
  const schluessel = `${karteID}/${feld}`

  if (geglueckt) {
    fehler.delete(schluessel)
  } else {
    fehler.set(schluessel, feld)
  }

  setze(setID, {
    offen: Math.max(0, jetzt.offen - 1),
    geschrieben: true,
    fehler,
  })
}

// lageAus liest den Stand für ein Set — **ein Fehler nennt das Feld, das
// zuerst scheiterte** und noch nicht geschrieben ist.
export function lageAus(setID: string): Speicherlage {
  const jetzt = imSet(setID)

  if (jetzt.offen > 0) {
    return { art: 'schreibt' }
  }

  const [erstes] = jetzt.fehler.values()
  if (erstes !== undefined) {
    return { art: 'fehler', feld: erstes }
  }

  return jetzt.geschrieben ? { art: 'geschrieben' } : { art: 'still' }
}

// abonniere meldet einen Hörer an — und gibt, wie er wieder geht.
function abonniere(hoere: () => void): () => void {
  hoerer.add(hoere)

  return () => {
    hoerer.delete(hoere)
  }
}

// useSpeicherlage folgt dem Stand eines Sets.
//
// **Der Schnappschuss ist der ganze Stand**, und er ändert sich nur mit
// `setze` — so zeichnet React neu, wenn geschrieben wird, und sonst nicht.
export function useSpeicherlage(setID: string): Speicherlage {
  useSyncExternalStore(abonniere, () => staende)

  return lageAus(setID)
}

// vergessen räumt den Stand — für Läufe, die unabhängig sein müssen.
export function vergessen(): void {
  staende = new Map()
  for (const hoere of hoerer) {
    hoere()
  }
}
