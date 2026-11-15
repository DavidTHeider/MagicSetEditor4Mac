// **Wie Zahlen aussehen und wo die Uhr steht** — LOC-03.AK-03
// (#1689):
// „Die volle Locale bleibt erhalten."
//
// # Warum neben der Sprache
//
// **Weil die Sprache nur den Sprachanteil braucht und die Darstellung die
// Gegend.** Die Kataloge gibt es auf Deutsch und Englisch; eine Zahl gibt es
// in jeder Gegend anders — `de-CH` schreibt »1'234.50«, `de-DE` »1.234,50«,
// `en-GB` den Tag vor dem Monat (gemessen am 2026-09-22).

// darstellungAus wählt die Locale der Darstellung: die Wahl, sonst die des
// Systems — **ganz**, mit Gegend —, sonst die Sprache.
//
// **Was `Intl` nicht versteht, fällt zurück.** Go prüft die Wahl, aber das
// System meldet, was es will; ein `RangeError` beim Zeichnen nähme jede
// Seite mit einer Zahl mit.
export function darstellungAus(
  gewaehlt: string,
  system: string,
  sprache: string,
): string {
  return [gewaehlt, system].find(verstanden) ?? sprache
}

// zeitzoneAus gibt die gewählte Zone — oder keine, dann gilt die des
// Rechners.
export function zeitzoneAus(gewaehlt: string): string | undefined {
  if (gewaehlt === '') {
    return undefined
  }

  try {
    new Intl.DateTimeFormat(undefined, { timeZone: gewaehlt })

    return gewaehlt
  } catch {
    return undefined
  }
}

// verstanden sagt, ob `Intl` eine Locale annimmt.
function verstanden(locale: string): boolean {
  if (locale === '') {
    return false
  }

  try {
    return Intl.getCanonicalLocales(locale).length > 0
  } catch {
    return false
  }
}
