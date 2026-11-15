// **Die Antworten der Kartenskripte** —
// #1337.
//
// **Eine eigene Datei**, seit `antworten.ts` an der Zeilengrenze aus §2 stand.
// Der Schnitt folgt der Frage: dort stehen die Antworten des Bestands, hier
// die des **Spielpakets** — was es an fertigen Zügen mitbringt.
import type { Umschlag } from './api'

// KartenskriptDTO ist ein `add cards script:`-Block des Spielpakets
// (#1337).
//
// **Ohne den Skripttext.** Er ist bis zu neunzehntausend Zeichen lang und für
// niemanden lesbar, der ihn nicht ausführt.
export interface KartenskriptDTO {
  name: string

  // kuerzel ist der Buchstabe, den MSEs `&` markiert — leer, wenn keiner
  // dasteht.
  kuerzel: string

  beschreibung: string
}

// Kartenskriptantwort trägt die Skripte des Spielpakets.
export interface Kartenskriptantwort extends Umschlag {
  skripte: KartenskriptDTO[]
}

// Kartenskriptlaufantwort trägt, was ein Lauf angelegt hat.
//
// **Flach und nicht um ein geschachteltes DTO herum.** Ein Zeiger auf der
// Go-Seite käme hier als `null` an, sobald ein Fehler dasteht — und diese
// Datei müsste einen Zweig anbieten, den `auspacken` nie erreicht, weil es
// vorher wirft. Beide Listen sind auf jedem Weg da (§3).
export interface Kartenskriptlaufantwort extends Umschlag {
  karten: string[]
  befunde: string[]
}
