// **Eine ganze Zahl ab null, streng gelesen** — F03
// (#2245).
//
// `Number.parseInt` liest, so weit es kommt, und schweigt über den Rest:
// „12a" ist 12, „1.5" ist 1, „1e3" ist 1. Vier Stellen im Fenster gaben so
// eine Zahl weiter, die niemand getippt hat — ein anderes Exemplar in einem
// Handel, ein anderes Ergebnis in einem Turnier, ein anderer Port.
//
// **Hier gilt nur, was ganz aus Ziffern besteht**; Leerraum davor und danach
// zählt nicht. Ein Vorzeichen, ein Komma, ein Exponent oder eine Ziffer aus
// einer anderen Schrift ist keine. Was dann geschieht, entscheidet die Stelle:
// ein Befund am Feld oder die Abweisung, die die Domäne gibt.

const nurZiffern = /^[0-9]+$/

// ganzzahlAus gibt die Zahl oder `undefined`, wenn der Text keine ist.
export function ganzzahlAus(roh: string): number | undefined {
  const text = roh.trim()

  if (!nurZiffern.test(text)) {
    return undefined
  }

  const zahl = Number(text)

  return Number.isSafeInteger(zahl) ? zahl : undefined
}
