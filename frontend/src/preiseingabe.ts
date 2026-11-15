// **Ein Preis, wie ein Mensch ihn tippt, und wie er über die Naht geht** —
// LOC-03.AK-03 (#1689).
//
// **Geteilt von Anlegen und Bearbeiten.** Bis zum Zweitblick über Phase 6.3
// las nur das Anlegen (`Sammlungsformular`); der Inspector schickte den Text
// roh an Go, »1.234« Dinar wurde still zu 1,234 und »19,99« Euro abgewiesen
// (S6). Die Regel steht jetzt einmal, und beide Wege nehmen sie.
import { alsZahl, lesarten } from './lokal'
import { einsetzen } from './sprache/einsetzen'
import { waehrungsstellen } from './waehrungsstellen'

// Preisauslegung ist, was aus einer Eingabe wird: ein Wert, der gilt, oder
// eine Frage an den, der ihn tippte.
export type Preisauslegung =
  { art: 'gilt'; kanonisch: string } | { art: 'frage'; befund: string }

// preisAuslegen liest eine Eingabe in der Schreibweise des Fensters.
//
// **Was keine Lesart hat, geht unverändert** — und Go erklärt es. **Zwei
// Lesarten werden genannt**, in der Schreibweise dessen, der getippt hat;
// `satz` ist der Befund mit `{erste}` und `{zweite}`.
export function preisAuslegen(
  eingabe: string,
  waehrung: string,
  darstellung: string,
  satz: string,
): Preisauslegung {
  const stellen = waehrungsstellen(waehrung)
  const lesart = lesarten(eingabe, darstellung, stellen)

  if (lesart.art === 'mehrdeutig') {
    return {
      art: 'frage',
      befund: einsetzen(satz, {
        erste: alsZahl(lesart.lesarten[0], darstellung, stellen),
        zweite: alsZahl(lesart.lesarten[1], darstellung, stellen),
      }),
    }
  }

  return {
    art: 'gilt',
    kanonisch: lesart.art === 'eindeutig' ? lesart.kanonisch : eingabe,
  }
}

// kanonischerBetrag ist die Form über die Naht (`internal/domain/geld.go`).
const kanonischerBetrag = /^-?\d+(?:\.(\d+))?$/u

// preisImFeld schreibt einen kanonischen Preis so, wie das Fenster ihn
// schriebe — damit ein Bearbeiten in der Schreibweise beginnt, in der man
// tippt.
//
// **Was sich so nicht verlustfrei schreiben lässt, bleibt stehen**: ein Text,
// der keine kanonische Zahl ist, und einer mit mehr Nachkommastellen, als die
// Währung hat — gerundet stünde im Feld ein anderer Betrag, und ein Sichern
// schriebe ihn.
export function preisImFeld(
  kanonisch: string,
  waehrung: string,
  darstellung: string,
): string {
  const stellen = waehrungsstellen(waehrung)
  const treffer = kanonischerBetrag.exec(kanonisch)

  if (treffer === null || (treffer[1]?.length ?? 0) > stellen) {
    return kanonisch
  }

  return alsZahl(kanonisch, darstellung, stellen)
}
