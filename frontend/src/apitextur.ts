// **Die Spieltexturen an der Naht** — Werkstück G2 von Phase 5.5
// (#1459).
//
// **Eine Adresse und keine Bytes.** 182 KB kosten als Base64 über die
// Wails-Brücke **554 ms** — gemessen für die Vorschau
// (#1270) —,
// und ein Spieltisch zeigt nicht eine Karte, sondern sechzig.
import type { Umschlag } from './umschlag'

// Textur ist ein gerastertes Kartenbild samt seiner Adresse.
export interface Textur {
  // schluessel ist, woran zwei gleiche Texturen sich erkennen.
  //
  // **Er steht in der Antwort**, damit das Fenster teilen kann: sieben Wälder
  // am Tisch sind ein Bild und sieben Darstellungen davon.
  schluessel: string

  adresse: string
  guete: string
  seite: string

  breite: number
  hoehe: number

  // befunde sagen, **warum** eine Karte behilflich aussieht.
  befunde: string[]

  // zweiseitig sagt, ob diese Karte eine zweite öffentliche Seite führt.
  zweiseitig: boolean
}

// TexturAntwort ist der Ausgang eines Texturwegs.
export interface TexturAntwort extends Umschlag {
  textur: Textur
}

// Texturwunsch sagt, welches Bild gemeint ist.
//
// # Warum zwei Gestalten und nicht eine mit Lücken
//
// **Weil es zwei Fragen sind.** Eine Kartenseite gehört einer **Karte**, ein
// Rücken einem **Spiel**. Eine gemeinsame Gestalt hätte an jeder Stelle ein
// Feld, das die Hälfte der Zeit nichts bedeutet — und der Aufrufer müsste
// raten, welche Hälfte gerade gilt.
export type Texturwunsch =
  | {
      art: 'karte'
      karteID: string
      seite: string
      guete: string
      // bild und objekt sind die Bildkennung und das Objekt am Tisch — P9
      // (#2377): ohne eigene Zeile fragt das Fenster damit den, dem das Deck
      // gehört.
      bild?: string
      objekt?: string
    }
  | { art: 'ruecken'; spiel: string; guete: string }

// Texturapi sind die zwei Wege.
//
// # Warum es keine Liste der Güten gibt
//
// **Weil sie keinen Aufrufer hätte.** Der erste Anlauf band eine
// `Texturgueten`, die dem Fenster die fünf Stufen nennt — und niemand rief
// sie: das Fenster benennt die Stufe, die es will. Eine Bindung ohne
// Aufrufer ist keine Anwendungsfunktion, sondern eine Zusage ohne Einlösung
// (§2), und der Wächter dazu hat es angestrichen.
//
// **Dass die Wörter trotzdem zusammenpassen**, hält
// `internal/werkzeuge/texturwoerter_test.go`: was `frontend/src` an Güten und
// Seiten nennt, muss es in Go geben.
export interface Texturapi {
  Spieltextur: (
    karteID: string,
    seite: string,
    guete: string,
  ) => Promise<TexturAntwort>

  // Rueckentextur gibt den Rücken eines **Spiels**.
  //
  // **Ohne Kartenkennung, und das ist der Punkt**: ein verdecktes Objekt
  // trägt keine — die Projektion hält sie zurück. Und es darf auch keine
  // geben: der Rücken ist für jede Karte derselbe, und wer ihn aus einer
  // Karte baute, verriete sie über seine Maße.
  Rueckentextur: (spiel: string, guete: string) => Promise<TexturAntwort>

  // Fremdkartenbild gibt das Bild eines Objekts, das dieses Fenster nicht
  // selbst zeichnet — P9 (#2377), Vermerk 0198: am fremden Tisch erfragt
  // beim Wirt. **Nach dem Objekt und der Bildkennung**, nie nach einer Zeile
  // eines fremden Bestands; ein Grund des Wirts steht in `befunde`. **Mit
  // der Partie**, zu deren Tisch das Bild gehört — gefragt wird nur dort.
  Fremdkartenbild: (
    partie: string,
    objekt: string,
    bild: string,
    seite: string,
    guete: string,
  ) => Promise<TexturAntwort>
}
