// **Geld als Text** — herausgezogen für
// #1539,
// auf den Skalenvertrag gestellt mit LOC-03.AK-02
// (#1689).
//
// # Warum eine eigene Datei
//
// **Weil zwei Stellen dieselbe Zahl schreiben.** Sie stand in `Kosten.tsx`
// als private Funktion; die Freigabe eines KI-Auftrags nennt denselben
// Betrag, und eine zweite Fassung daneben liefe beim ersten Sonderfall
// auseinander.
//
// # Der Vertrag
//
// **Was hereinkommt, ist eine ganze Zahl in der kleinsten Einheit der
// Währung** — Cent beim Euro, **Yen beim Yen**, Fils beim Dinar. Dieselbe
// Zusage wie in `internal/domain/geld.go`, und die Stellen kommen aus
// derselben Quelle wie dort: aus der Währung.
//
// # Und woher die Stellen kommen
//
// **Aus derselben Liste wie in Go** (`waehrungsstellen.ts`) — und nicht aus
// `Intl`: `Intl` antwortet nach CLDR, Go nach ISO 4217, und bei 17 von 162
// Codes ist das eine andere Antwort (gemessen am 2026-09-22).
//
// **`Intl` zeigt nur noch** — Zeichen, Trenner und Stellung, **mit den
// Stellen des Bestands**. Bis zum Zweitblick über Phase 6.3 zeigte es den
// Forint ohne Nachkommastellen, und 1.999 Fillér standen gerundet als
// »20 HUF« da (G8): eine Anzeige, die rundet, verschweigt einen Betrag.

import { waehrungsstellen } from './waehrungsstellen'

// alsGeld schreibt einen Betrag in der Sprache des Fensters.
export function alsGeld(
  wert: number,
  waehrung: string,
  sprache: string,
): string {
  // **Ohne Währung bleibt die Zahl eine Zahl.** `Intl` verlangt einen
  // ISO-Code und wirft ohne einen; ein leeres Feld heißt hier aber „noch
  // nichts freigegeben" und nicht „kaputt".
  const formatierer = formatiererFuer(waehrung, sprache)

  if (formatierer === undefined) {
    return String(wert)
  }

  return formatierer.format(wert / 10 ** waehrungsstellen(waehrung))
}

// geldAusText schreibt einen Betrag, den Go als **kanonischen Text** schickt
// (»19.99«, `domain.GeldText`) — Markt, Portfolio und Verlauf
// (`internal/app/preisdto.go`). **Was es nicht schreiben kann, lässt es
// stehen**: kein Betrag, keine Währung, kein Zahltext. Hier standen die
// Werte bis zum Zweitblick über Phase 6.3 roh, und »1.234 BHD« las sich im
// Deutschen als tausendzweihundertvierunddreißig Dinar (M10).
export function geldAusText(
  text: string,
  waehrung: string,
  sprache: string,
): string {
  // **Kein Betrag ist kein Text** — auch keine Währung allein.
  if (text.trim() === '') {
    return ''
  }

  const zahl = Number(text)
  const formatierer = formatiererFuer(waehrung, sprache)

  if (!Number.isFinite(zahl) || formatierer === undefined) {
    return `${text} ${waehrung}`.trim()
  }

  return formatierer.format(zahl)
}

// formatiererFuer baut den Formatierer — oder gibt auf.
//
// **Ein ungültiger Code wirft**, und ein geworfener Fehler beim Zeichnen
// nimmt die ganze Seite mit: `Intl` weist alles ab, was nicht drei
// Buchstaben sind (gemessen am 2026-09-22: »EURO« gibt einen `RangeError`).
// Was hier hereinkommt, stammt aus einem Feld, in das jemand tippt.
function formatiererFuer(
  waehrung: string,
  sprache: string,
): Intl.NumberFormat | undefined {
  if (waehrung === '') {
    return undefined
  }

  const stellen = waehrungsstellen(waehrung)

  try {
    return new Intl.NumberFormat(sprache, {
      style: 'currency',
      currency: waehrung,
      minimumFractionDigits: stellen,
      maximumFractionDigits: stellen,
    })
  } catch {
    return undefined
  }
}
