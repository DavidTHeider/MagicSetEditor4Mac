import { ohneAuszeichnung } from './ohneauszeichnung'

// Die Zusatzspalten einer Zeile der Kartenliste.
//
// **Eigene Datei, weil zwei Bauteile sie brauchen** (§2): `Kartenliste` nimmt
// sie entgegen, `Kartenzeile` zeichnet sie. Ein Import über Kreuz zwischen
// beiden wäre ein Kreis
// (#1290).

// Listenzusatz ist eine Spalte samt ihren Werten — je Feld einmal und nicht je
// Karte, damit eine Liste von 3.788 Zeilen die Aufschrift nicht 3.788-mal
// trägt.
export interface Listenzusatz {
  feld: string
  aufschrift: string
  ausrichtung: string

  // werte ist Kartenkennung → Wert, ohne Auszeichnungen.
  werte: Record<string, string>
}

// gefuellte gibt die Spalten, die für diese Karte einen Wert haben.
//
// **Ein leerer Wert steht nicht da**: »Name:« ohne etwas dahinter ist keine
// Auskunft. Der Vorschlag von `magic.mse-game` enthält Felder wie `full name`,
// die eine Karte oft nicht führt.
export function gefuellte(
  spalten: readonly Listenzusatz[],
  karteID: string,
  name: string,
): Listenzusatz[] {
  return spalten.filter((spalte) => {
    const wert = spalte.werte[karteID] ?? ''

    if (wert === '') {
      return false
    }

    // **Was schon in der Überschrift steht, steht nicht noch einmal darunter**
    // (#1271).
    //
    // `magic.mse-game` schlägt elf sichtbare Spalten vor, und die erste ist
    // `name` — derselbe Text, den die Zeile schon als Überschrift trägt.
    // Gemessen am 2026-08-28 bei 1280 × 800: der Streifen ist **161 px**
    // breit und trägt **618 px** Inhalt; „Name: Fireball" nahm davon das
    // erste Fünftel und sagte nichts, was nicht zwei Millimeter darüber
    // stünde.
    //
    // **Über den Wert und nicht über den Feldnamen.** Ein fremdes Spiel
    // nennt sein Namensfeld anders, und die Überschrift kommt aus
    // `zeilenname` — verglichen wird deshalb das, was dasteht.
    return ohneAuszeichnung(wert) !== name
  })
}
