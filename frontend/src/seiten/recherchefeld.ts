// **Was die Recherche und die Ähnlichkeit teilen** — Werkstück A16 von Phase
// 5.3 (#1515).
//
// **Eine eigene Datei, weil beide sie brauchen** und keiner der beiden
// Seitenbausteine dem anderen gehört.
import type { RecherchefilterDTO } from '../apirecherche'

// derLeerfilter ist die Einschränkung ohne jede Einschränkung.
//
// **Ausgeschrieben und nicht `{}`**: die Go-Seite erwartet jedes Feld, und
// ein fehlendes wäre dort ein Nullwert, den niemand geschrieben hat.
export const derLeerfilter: RecherchefilterDTO = {
  sprache: '',
  korpusstand: '',

  // **Leer, und das ist die Vorgabe** (A13.4, #1452): ohne ausdrückliche
  // Auswahl geht kein privater Korpus hinaus.
  freigegeben: [],
}

// derEigeneBestand ist der Korpus der eigenen Karten.
//
// **Der Name steht einmal hier** — er ist derselbe wie in
// `internal/domain/retrieval.go`, und zwei Schreibweisen davon liefen beim
// ersten Tippfehler auseinander.
export const derEigeneBestand = 'eigenekarte'

// feldklassen sind die Klassen eines freien Eingabefelds.
//
// **Mit sichtbarem Fokusring** (§6): eine Fläche, die nur mit der Maus
// bedienbar aussieht, ist für jemanden, der mit der Tastatur arbeitet, nicht
// bedienbar — und ein Wächter zählt es nach.
export const feldklassen =
  'rounded-sm border border-kante bg-grund px-2 py-1 ' +
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'
