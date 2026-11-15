// **Ein Betrag in der kleinsten Einheit ist eine Folge von Ziffern** —
// Nachlauf über Phase 6.3, N5; LOC-03.AK-03 verlangt, dass eine mehrdeutige
// Eingabe nicht still falsch gelesen wird.
//
// `Number('1.000')` ist eins, und im Deutschen hieße es tausend;
// `Number('1,000')` ist keine Zahl, und im Englischen hieße es tausend.
// **Statt zu raten, gilt nur, was eindeutig ist**: Ziffern, ohne Punkt,
// Komma, Vorzeichen oder Leerraum dazwischen — und nicht mehr, als eine
// Zahl in JavaScript genau fasst.
export function ziffernbetrag(text: string): number | undefined {
  const bereinigt = text.trim()
  if (!/^\d+$/.test(bereinigt)) {
    return undefined
  }

  const zahl = Number(bereinigt)

  return Number.isSafeInteger(zahl) ? zahl : undefined
}
