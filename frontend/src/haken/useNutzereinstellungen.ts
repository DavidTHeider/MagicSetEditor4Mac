// **Die Einstellungen, die dem Menschen gehören** — Werkstück 8 von
// Phase 4.12
// (#1351),
// Punkte 9 und 11.
//
// # Warum sie ihren eigenen Haken haben und nicht in `useEinstellungen` stehen
//
// **Weil sie einer anderen Sache gehören.** Jene gelten für das **Profil**, in
// dem gearbeitet wird; diese für den **Menschen**, der sie in jedes seiner
// Profile mitnimmt. Sie zusammenzulegen hieße, dem Fenster die Zuordnung zu
// überlassen — und die stünde dann an zwei Stellen.
//
// # Und warum er den Nutzer erst suchen muss
//
// **Weil das Fenster in einem Profil arbeitet und nicht in einem Nutzer.** Wem
// das gezeigte Profil gehört, sagt `GezeigtesProfil`; **in demselben Zug**
// (§7, Regel 2), denn zwei Rufe wären zwei Stände, und im Zwischenraum
// gehörten die Einstellungen jemand anderem.
import { useCallback, useState } from 'react'

import { api, auspacken, type NutzereinstellungDTO } from '../api'
import { useLaden } from './useLaden'
import { useVorgang } from './useVorgang'

// Nutzereinstellungsstand ist, was die Oberfläche über sie weiß.
export interface Nutzereinstellungsstand {
  // eintraege sind die Einstellungen in Anzeigereihenfolge, samt ihren Werten.
  eintraege: readonly NutzereinstellungDTO[]

  // wert gibt den eingestellten — leer, solange nichts geladen ist.
  wert: (schluessel: string) => string

  // nutzerID ist der Mensch, dem sie gehören — leer, solange er nicht feststeht.
  nutzerID: string

  fehler: string | undefined
  setze: (schluessel: string, wert: string) => void
  laeuft: (schluessel: string) => boolean
}

// useNutzereinstellungen lädt sie und schreibt sie fort.
//
// **Ohne eine „darf gefragt werden"-Wahl, anders als [useEinstellungen].**
// Der Grund ist der Ort: dieser Haken läuft **unter** dem Sprachraum und
// damit erst, wenn der Bestand offen ist — ohne ihn zeichnet `App` die
// Wiederherstellungsseite und kommt hier gar nicht an.
export function useNutzereinstellungen(stand = 0): Nutzereinstellungsstand {
  const vorgang = useVorgang()
  const [runde, setzeRunde] = useState(0)

  const holen = useCallback(async () => {
    const profil = await api().GezeigtesProfil()
    const nutzerID = auspacken(profil, () => profil.profil)?.nutzerID ?? ''

    // **Ohne Nutzer keine Einstellungen** — und das ist kein Fehler: ein
    // Bestand ohne gezeigtes Profil ist der Zustand vor der Wanderung, und
    // die Vorgabewerte sind dann die richtige Antwort.
    if (nutzerID === '') {
      return { eintraege: [] as NutzereinstellungDTO[], nutzerID }
    }

    const antwort = await api().Nutzereinstellungen(nutzerID)

    return {
      eintraege: auspacken(antwort, () => antwort.einstellungen),
      nutzerID,
    }
  }, [])

  const { daten, fehler } = useLaden(holen, [stand, runde], true)

  const eintraege = daten?.eintraege ?? []

  // **Einmal ausgerechnet und zweimal benutzt.** Ein zweites `daten?.…` beim
  // Schreiben wäre derselbe Rückfall noch einmal — und einer davon ein Zweig,
  // den kein Klick betreten kann: wer eine Auswahlliste bedient, hat sie
  // gesehen, und dann steht `daten` (§3).
  const nutzerID = daten?.nutzerID ?? ''

  const setze = (schluessel: string, wert: string) => {
    // **Ein Vorgang je Einstellung** (§7, Regel 4): zwei Klicks auf dieselbe
    // Auswahl gingen gegen einen Stand, den der Bestand noch nicht kennt.
    void vorgang.starte(schluessel, async () => {
      const antwort = await api().NutzereinstellungSetzen(
        nutzerID,
        schluessel,
        wert,
      )

      auspacken(antwort, () => antwort.einstellungen)

      // **Nachgeladen wird nur nach einem geglückten Ruf** (§7, Regel 3):
      // `auspacken` wirft vorher, und ein Lesefehler überschriebe sonst die
      // Meldung, die sagt, was zu tun ist.
      setzeRunde((bisher) => bisher + 1)
    })
  }

  return {
    eintraege,
    wert: (schluessel) =>
      eintraege.find((eine) => eine.schluessel === schluessel)?.wert ?? '',
    nutzerID,
    fehler,
    setze,
    laeuft: vorgang.laeuft,
  }
}
