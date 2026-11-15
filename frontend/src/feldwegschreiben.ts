// **Ein offenes Feld wird verlassen, bevor das Bezugsobjekt wechselt** —
// #1290.
//
// # Warum es das gibt
//
// Ein Feldwert wird beim **Verlassen** geschrieben (`Feldeingabe.tsx`,
// `onBlur`). Wer das Bezugsobjekt wechselt, während der Fokus in einem Feld
// steht, räumt das Feld ab — und **ein programmgesteuerter Abbau erzeugt kein
// `blur`**. Der Wert ist dann weg.
//
// Gemessen an der laufenden Brücke am 2026-08-30:
//
//	A · tippen, dann blur:   ["MitBlur",""]
//	B · tippen, dann ⇟:      ["MitBlur",""]
//
// Im Fall B stand der getippte Text im Feld und **nichts** im Bestand.
//
// # Und deshalb steht es hier und nicht in einem Haken
//
// Der erste Anlauf hat es in `useKartenblaettern` behoben — für ⇞ und ⇟. Der
// Nachlauf über Phase 4.3 hat angestrichen, dass die **Klasse** damit offen
// blieb: *Zuletzt verwendetes Set* aus der Menüleiste, das Öffnen einer Datei
// und eine abgelegte Datei rufen denselben Kern (`useSetwahl`) unmittelbar auf,
// und keiner von ihnen verließ das Feld.
//
// **Eine Stelle, und jeder Wechsel geht durch sie.**

// wegschreiben verlässt das offene Feld, damit sein Wert hinausgeht.
//
// **`select` gehört dazu**, obwohl ein Auswahlfeld sofort schreibt: der Fokus
// soll auch von dort mitwandern (siehe `useFeldfokus`). Angestrichen vom
// Nachlauf über Phase 4.3, Befund 5.
export function wegschreiben(): void {
  const stand = document.activeElement

  if (
    stand instanceof HTMLInputElement ||
    stand instanceof HTMLTextAreaElement ||
    stand instanceof HTMLSelectElement
  ) {
    stand.blur()
  }
}

// offenesFeld gibt den Namen des Feldes, in dem der Fokus steht.
//
// **`aria-label` und nicht `id`**: der Editor beschriftet seine Felder so, und
// die Kennung wechselt mit jedem Neuaufbau.
export function offenesFeld(): string | undefined {
  const stand = document.activeElement

  if (
    stand instanceof HTMLInputElement ||
    stand instanceof HTMLTextAreaElement ||
    stand instanceof HTMLSelectElement
  ) {
    return stand.getAttribute('aria-label') ?? undefined
  }

  return undefined
}
