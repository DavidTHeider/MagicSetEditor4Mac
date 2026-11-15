// **Welche Seiten stehen bleiben, wenn man sie verlässt** — UX-03.AK-02
// (#1675):
// *„Zurück/Vorwärts, Direktaufruf und Wiederaufnahme funktionieren ohne
// Verlust ungespeicherter Eingaben."*
//
// # Warum aufheben und nicht zwischenspeichern
//
// **Weil eine Seite mehr hält als ihre Felder.** Ein halb getippter Deckname,
// ein aufgeklappter Kasten, eine gewählte Zeile, eine Rückfrage: das alles ist
// Zustand der Seite, und bis hierher entstand sie beim Zurückkommen neu. Einen
// Entwurfsspeicher je Feld hätte man an über sechzig freien Eingaben
// nachziehen müssen (`docs/FELDPRUEFUNG.md`) — und an jeder künftigen, sonst
// wäre er die halbe Zusage. React hält verborgene Bäume selbst (`<Activity>`,
// seit 19.2): der Zustand bleibt, die Effekte ruhen, bis die Seite wieder
// dasteht — und laden dann neu, wie beim ersten Mal.
//
// # Was nicht aufgehoben wird
//
// **Die Seiten eines Sets, wenn das Set wechselt** (AGENTS.md §7, Regel 1):
// ein Wechsel des Bezugsobjekts räumt den offenen Kontext. **Und nichts über
// einen Profilwechsel hinweg** — `App` baut die Anwendung dann neu (`key`).
//
// **Und nichts über ein Zurückspielen hinweg** — Befund M9 (a) des
// Zweitblicks über Phase 6.3
// (#1666).
// Es tauscht den ganzen Bestand, und geräumt wurden Set und Karte, aber nicht
// die verlassenen Seiten: ein halb getippter Deckname aus dem ersetzten
// Bestand stand danach wieder da. **Stehen bleibt die Seite, auf der es
// geschah** — sie zeigt, wohin der alte Bestand ging, und lädt neu wie jede.
import { useState } from 'react'

import {
  istSetseite,
  type Anwendungsschluessel,
  type Seitenschluessel,
  type Setschluessel,
} from '../navigation'
import { ausgabeplatz } from '../ausgabe'
import { schubladeVon } from '../schubladen'

export interface Aufgehoben {
  // bestand zählt, wie oft der Bestand ersetzt war, als diese Seiten
  // aufgehoben wurden — ein neuer Stand räumt sie (§7, Regel 1).
  bestand: number

  // setID ist das Set, zu dem die aufgehobenen Setseiten gehören.
  setID: string | undefined

  // anwendung sind die besuchten Anwendungsseiten, in der Reihenfolge des
  // ersten Besuchs.
  anwendung: readonly Anwendungsschluessel[]

  // setseite ist die zuletzt besuchte Seite des Sets — sie bestimmt, was der
  // Arbeitsplatz zeigt, während er verborgen steht.
  setseite: Setschluessel | undefined

  // laschen sind die besuchten Laschen unter den Schubladen; unter einer
  // Schublade stehen die Karten.
  laschen: readonly Setschluessel[]

  // schubladen sind die besuchten Laschen in den Schubladen.
  schubladen: readonly Setschluessel[]
}

export const nichtsAufgehoben: Aufgehoben = {
  bestand: 0,
  setID: undefined,
  anwendung: [],
  setseite: undefined,
  laschen: [],
  schubladen: [],
}

// aufheben merkt sich die Seite, auf der man steht. **Gibt es nichts Neues,
// kommt derselbe Stand zurück** — der Haken erkennt daran, dass er fertig ist.
//
// **Ohne `bestand` bleibt der Bestand derselbe**; ein anderer räumt alles,
// auch die Anwendungsseiten.
export function aufheben(
  vorher: Aufgehoben,
  seite: Seitenschluessel,
  setID: string | undefined,
  bestand = vorher.bestand,
): Aufgehoben {
  const stand =
    vorher.bestand !== bestand
      ? { ...nichtsAufgehoben, bestand, setID }
      : vorher.setID === setID
        ? vorher
        : { ...nichtsAufgehoben, bestand, anwendung: vorher.anwendung, setID }

  if (!istSetseite(seite)) {
    return stand.anwendung.includes(seite)
      ? stand
      : { ...stand, anwendung: [...stand.anwendung, seite] }
  }

  // **Unter einer Schublade stehen die Karten, und die drei Seiten des
  // Ausgabeassistenten stehen auf einem Platz** (UXR-09, #1917).
  const inSchublade = schubladeVon(seite) !== undefined
  const lasche: Setschluessel = inSchublade ? 'karten' : ausgabeplatz(seite)
  const laschen = mit(stand.laschen, lasche)
  const schubladen = inSchublade
    ? mit(stand.schubladen, seite)
    : stand.schubladen

  return stand.setseite === seite &&
    laschen === stand.laschen &&
    schubladen === stand.schubladen
    ? stand
    : { ...stand, setseite: seite, laschen, schubladen }
}

// mit hängt an, was fehlt — und gibt sonst dieselbe Liste zurück.
function mit<T>(liste: readonly T[], eintrag: T): readonly T[] {
  return liste.includes(eintrag) ? liste : [...liste, eintrag]
}

// useAufgehobeneSeiten führt, welche Seiten stehen bleiben — `bestand` zählt,
// wie oft der Bestand ersetzt wurde ([Bestandspflege.ersetzt]).
export function useAufgehobeneSeiten(
  seite: Seitenschluessel,
  setID: string | undefined,
  bestand = 0,
): Aufgehoben {
  const [stand, setzeStand] = useState(nichtsAufgehoben)
  const neu = aufheben(stand, seite, setID, bestand)

  // **Der Vergleich steht neben dem Zustand** — die Form, die React für
  // „Zustand an eine geänderte Eigenschaft anpassen" vorsieht.
  if (neu !== stand) {
    setzeStand(neu)
  }

  return neu
}
