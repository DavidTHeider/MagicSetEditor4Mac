// Die Lasche „Stil“: welches Stilpaket ein Set benutzt und wie es eingestellt ist.
//
// # Warum das nicht mehr über den Karten steht
//
// Bis #288
// stand die Stilwahl **über** der Kartenliste — samt einem Auswahlfeld mit
// rund zweihundert Stilpaketen und der Liste aller Stiloptionen darunter. Wer
// die Kartenseite aufschlug, sah zuerst Einstellungen und musste an ihnen
// vorbeirollen, um zur ersten Karte zu kommen.
//
// **Die Karte ist das Hauptobjekt** (AGENTS.md §6). Der Stil gilt für das
// ganze Set, wird selten gewechselt und gehört deshalb neben die Karten, nicht
// vor sie.
import type { Seitenschluessel } from '../navigation'
import { Stilwahl } from './Stilwahl'

interface StilEigenschaften {
  // setID ist **gesetzt** — der Arbeitsplatz rendert diese Lasche nur mit
  // offenem Set und zeigt sonst seinen eigenen leeren Zustand
  // (#288).
  // Ein Wächter „kein Set gewählt" wäre hier ein Zweig, den kein Aufruf
  // betreten kann (§3).
  setID: string

  // geaendert meldet dem Fenster, dass der Bestand sich geändert hat — die
  // Kartenlasche lädt dann beim Zurückwechseln neu.
  geaendert: () => void

  // wechseln führt ohne installierten Stil zu den Paketen (G02, #2249).
  wechseln?: (seite: Seitenschluessel) => void
}

export function Stil({ setID, geaendert, wechseln }: StilEigenschaften) {
  return (
    <Stilwahl
      setID={setID}
      geaendert={geaendert}
      {...(wechseln === undefined
        ? {}
        : {
            zuDenPaketen: () => {
              wechseln('pakete')
            },
          })}
    />
  )
}
