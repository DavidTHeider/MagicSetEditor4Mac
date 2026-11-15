// **Die Gruppen der Setlaschen** — abgeteilt von `navigation.ts`, damit beide
// unter der Zeilengrenze aus §2 bleiben. Der Schnitt folgt der Frage: dort
// steht, **welche** Seiten es gibt und welche Taste sie tragen, hier, **wie**
// die Laschenleiste sie ordnet.
import { seiten, type Seite, type Setschluessel } from './navigation'

// **Die Gruppen der Setlaschen** —
// #1166.
//
// # Zwölf gleichrangige Laschen sind keine Gliederung
//
// Sie sind eine Aufzählung: *Konsole* stand neben *Karten* mit demselben
// Gewicht, und wer zum ersten Mal hier war, hatte keinen Anhalt, welche der
// zwölf sein Weg ist. Bei 1.440 Punkten — dem Maß der Bildschirmabzüge — war
// die letzte zusätzlich abgeschnitten.
//
// # Woher die Einteilung kommt (§15)
//
// **Nicht erfunden, sondern aus MSE2 abgelesen.** Nachgelesen in
// `src/gui/set/window.cpp`, Fassung `a7336d6` vom 2020-06-14:
//
//   - Zeilen 148–154 registrieren **sieben** Reiter: Karten, Stil,
//     Setangaben, Schlüsselwörter, Statistik, Zufallspack, Konsole.
//   - *Ausgeben* und *Drucken* stehen im **Ablagemenü** (Zeilen 63, 73–74),
//     *Suchen und Ersetzen* im **Bearbeitenmenü** (Zeilen 91–93).
//   - Der **Symboleditor ist ein eigenes Fenster** (`src/gui/symbol/window.cpp`).
//   - *Bilder* gibt es dort überhaupt nicht.
//
// MSE2 trennt also sehr wohl — nur trennt es in **Bedienformen** statt in
// Gruppen: Ansicht des Sets gegen Handlung daran. Diese Anwendung hat beides
// als Lasche, weil sie auch ohne Menüleiste bedienbar sein muss (die
// Browservorschau hat keine). Die Trennung bleibt, sie wird nur **sichtbar**.
//
// **Damit ist die Gruppierung `adapted` und nicht `new`.**
export interface Laschengruppe {
  // schluessel benennt die Gruppe im Katalog — nicht ihre Aufschrift (§9).
  schluessel: Gruppenschluessel

  laschen: readonly Setschluessel[]
}

export const gruppenschluessel = ['bauen'] as const

export type Gruppenschluessel = (typeof gruppenschluessel)[number]

// **Seit UX-08 stehen hier nur noch zwei Gruppen** (#1680), **seit UXR-09
// eine** (#1917). *Gestalten*, *Das Set* und *Technik* sind in zwei
// Schubladen gewandert — *Eigenschaften* und *Werkzeuge* (`schubladen.ts`) —,
// weil sie für das ganze Set gelten oder selten gebraucht werden und nicht
// neben der Arbeit stehen sollen. *Bilder*, *Ausgeben* und *Druckbogen* sind
// in den Ausgabeassistenten gewandert, den der Knopf *Ausgeben* im Kopf
// öffnet (`ausgabe.ts`) — in MSE2 liegen Ausgeben und Drucken ebenfalls nicht
// bei den Reitern, sondern im Ablagemenü. Was hier bleibt, wandert, wenn
// seine neue Fläche steht: *Suche* und *Ersetzen* in den Navigator
// (`docs/PROTOTYP-0.6.0.md`, Inventurdifferenz).
export const laschengruppen: readonly Laschengruppe[] = [
  // **Woran jemand arbeitet.** Die Karte ist das Hauptobjekt (§6). Die Suche
  // steht hier, weil sie in MSE2 im Bearbeitenmenü liegt — also bei der
  // Arbeit an den Karten und nicht bei den Angaben über das Set.
  { schluessel: 'bauen', laschen: ['karten', 'suche', 'ersetzen'] },
]

// setlaschen ist die Laschenreihe — **aus den Gruppen abgeleitet und nicht
// daneben aufgeschrieben**.
//
// **Die Kürzel folgen der Reihe nicht, und das ist Absicht.** ⌘6 heißt „Stil"
// und nicht „die dritte Lasche"; wer eine Taste gelernt hat, hat ein Ziel
// gelernt.
export const setlaschen: readonly Seite[] = laschengruppen.flatMap((gruppe) =>
  gruppe.laschen.map((schluessel) => {
    const seite = seiten.find((eine) => eine.schluessel === schluessel)

    // **Ohne Ausweichwert.** Eine Gruppe, die eine Lasche nennt, die es nicht
    // gibt, ist ein Fehler beim Bauen und kein leerer Knopf im Fenster; der
    // Typ lässt es gar nicht zu, und `laschengruppen.test.ts` hält beide
    // Listen gegeneinander.
    /* v8 ignore next 3 -- der Typ schließt es aus; die Zeile ist der Beleg,
       dass hier nichts still ausfällt */
    if (seite === undefined) {
      throw new Error(schluessel)
    }

    return seite
  }),
)
