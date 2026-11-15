// **Wann die Einführung steht** —
// #1325.
//
// # Warum das ein eigenes Bauteil ist
//
// Aus `Anwendung.tsx` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// Der Schnitt folgt der Frage: [Einfuehrung] daneben sagt, **was** im Fenster
// steht; hier steht, **ob** es kommt.
//
// # Die Bedingung, und warum sie abgeleitet ist
//
// **Kein Effekt.** Einer, der beim Laden `setzeOffen(true)` riefe, wäre eine
// zweite Zeichnung, bevor irgendetwas geschehen ist — der Linter weist das ab,
// und der Grund steht bei `gewaehltesSpiel` in `SetNeu.tsx`.
//
// **Und erst, wenn die Einstellungen da sind.** Vor der Antwort gälte die
// Vorgabe „an", und wer sie abgeschaltet hat, sähe das Fenster für einen
// Wimpernschlag doch — genau der Sprung, gegen den es gebaut ist.
import { Einfuehrung } from './Einfuehrung'
import type { Einstellungsstand } from '../haken/useEinstellungen'
import type { Seitenschluessel } from '../navigation'

interface StarteinfuehrungEigenschaften {
  einstellungen: Einstellungsstand

  // zugeklappt heißt: in dieser Sitzung schon geschlossen.
  zugeklappt: boolean

  // erneut heißt: aus der Hilfe verlangt — dann zählt die Einstellung nicht (#1422).
  erneut: boolean

  setID: string | undefined
  wechseln: (seite: Seitenschluessel) => void
  schliessen: () => void
  gewaehlt: () => void
}

export function Starteinfuehrung({
  einstellungen,
  zugeklappt,
  erneut,
  setID,
  wechseln,
  schliessen,
  gewaehlt,
}: StarteinfuehrungEigenschaften) {
  if (
    zugeklappt ||
    !einstellungen.geladen ||
    (!erneut && einstellungen.wert('ersteschritte') === 'aus')
  ) {
    return null
  }

  return (
    <Einfuehrung
      setID={setID}
      wechseln={wechseln}
      wiederZeigen
      gewaehlt={gewaehlt}
      schliessen={schliessen}
    />
  )
}
