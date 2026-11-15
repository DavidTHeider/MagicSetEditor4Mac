// Was ein **gewähltes Feld** an Beiwerk mitbringt — Symbolpalette und
// Auswahllisten
// (#1074).
//
// # Warum als eigene Datei
//
// `Karteneditor.tsx` beantwortet die Frage „was tut dieses Fenster": laden,
// abbrechen, schreiben, den Stand halten. **Woher** die Auskünfte zum
// gewählten Feld kommen, ist eine andere Frage (§2) — und mit der zweiten
// stand die Datei an der Zeilengrenze.
import { useCallback } from 'react'

import { api } from '../fassade'
import { useLaden } from '../haken/useLaden'
import type { PalettenEintragDTO, WortlistenbereichDTO } from '../formenworte'
import { auspacken } from '../umschlag'

// leerePalette ist die Antwort für ein Feld, das keine hat.
//
// **Ein fester Wert und kein neues Objekt je Aufruf**: `useLaden` trägt ihn in
// seinen Stand ein, und ein frisches Objekt bei jedem Zeichnen ließe die
// Palette in einer Schleife neu laden.
const leerePalette = { eintraege: [] as PalettenEintragDTO[], roh: false }

// leereWortlisten ist die Antwort, wenn kein Feld gewählt ist — aus demselben
// Grund ein fester Wert.
const leereWortlisten: WortlistenbereichDTO[] = []

// useSymbolpalette lädt, was ein Feld zum Einfügen anbietet.
export function useSymbolpalette(karteID: string, feld: string | undefined) {
  const holen = useCallback(async () => {
    if (feld === undefined) {
      return leerePalette
    }

    const antwort = await api().Symbolpalette(karteID, feld)

    return auspacken(antwort, () => ({
      eintraege: antwort.eintraege,
      roh: antwort.roh,
    }))
  }, [karteID, feld])

  return useLaden(holen, [karteID, feld])
}

// useWortlisten lädt die Auswahlbereiche eines Feldes.
//
// **Der Wert steht in den Abhängigkeiten**, und das ist die halbe Zusage: das
// Menü zeigt, was **jetzt** im Feld steht — welcher Eintrag angekreuzt ist,
// hängt daran. Ohne ihn bliebe nach einer Wahl das alte Kreuz stehen.
//
// **Der geschriebene Wert, nicht der getippte.** Wer den Entwurf nähme, zeigte
// eine Auswahl zu einem Stand, den der Bestand nicht kennt (§7, Regel 3).
export function useWortlisten(
  karteID: string,
  feld: string | undefined,
  wert: string,
) {
  const holen = useCallback(async () => {
    if (feld === undefined) {
      return leereWortlisten
    }

    const antwort = await api().Wortlisten(karteID, feld)

    return auspacken(antwort, () => antwort.bereiche)

    // **`wert` steht absichtlich dabei, obwohl der Rückruf ihn nicht liest.**
    //
    // Er ist der Anlass zum Nachladen und nicht sein Argument: das Menü zeigt,
    // was **jetzt** im Feld steht, und welcher Eintrag angekreuzt ist, hängt
    // daran. Der Grund steht ausgeschrieben über der Funktion.
    //
    // ESLint nennt ihn „unnecessary" — er kennt den Anlass nicht, nur die
    // Lesevorgänge. Seit dem 2026-08-28 sind Warnungen Fehler (#1170),
    // und eine benannte Ausnahme ist besser als eine Warnung, die niemand
    // liest.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [karteID, feld, wert])

  // **Der Stand wird beim Nachladen geräumt** (§7, Regel 1) — die Vorgabe.
  //
  // **Was das kostet, ist gemessen** (2026-08-24, in der laufenden Anwendung):
  // nach einer Wahl schließt sich das Menü, und wer ein zweites Präfix setzen
  // will, klickt dazwischen ins Feld zurück. Die Ursache liegt nicht hier: die
  // Wahl lädt die Karte nach, `Karten.tsx` räumt dabei seine Kartenliste, und
  // der [Karteneditor] wird für einen Augenblick abgebaut — mitsamt dem
  // gewählten Feld. `behalten` hier zu setzen ändert daran nichts; das steht
  // in #1075.
  return useLaden(holen, [karteID, feld, wert])
}
