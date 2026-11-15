// **Die Gruppen der Setlaschen** — Werkstück 1 von Phase 3.5
// (#1166).
//
// # Warum eine Gruppierung und kein bloßes Umsortieren
//
// Zwölf gleichrangige Laschen sind keine Gliederung, sondern eine Aufzählung:
// *Konsole* steht neben *Karten* mit demselben Gewicht, und wer zum ersten Mal
// hier ist, hat keinen Anhalt, welche der zwölf sein Weg ist.
//
// **MSE2 hat das Problem nicht, weil es weniger Laschen hat** — sieben, und
// das ist nachgelesen (`src/gui/set/window.cpp:148–154`, Fassung `a7336d6` vom
// 2020-06-14): Karten, Stil, Setangaben, Schlüsselwörter, Statistik,
// Zufallspack, Konsole. Was hier zusätzlich als Lasche steht, liegt dort
// woanders: *Ausgeben* und *Druckbogen* im Ablagemenü (`window.cpp:63, 73–74`),
// *Suche* im Bearbeitenmenü (`window.cpp:91–93`), der *Symboleditor* in einem
// **eigenen Fenster** (`src/gui/symbol/window.cpp`). *Bilder* gibt es dort gar
// nicht.
//
// **Die Gruppen sind damit `adapted` und nicht `new`** (§15): die Zuordnung
// „was ist eine Ansicht des Sets, was ist eine Handlung daran" stammt aus
// MSE2s eigener Trennung zwischen Reitern und Menüs. Neu ist nur, dass sie
// hier **sichtbar** wird, statt in zwei Bedienformen zu zerfallen — ein Fenster
// ohne Menüleiste für den Browservorschau-Fall braucht beides an einer Stelle.
import { describe, expect, it } from 'vitest'

import { ausgabeseiten } from './ausgabe'
import { setschluessel } from './navigation'
import { laschengruppen, setlaschen } from './laschengruppen'
import { schubladen, schubladeVon } from './schubladen'

describe('die Gruppen der Setlaschen', () => {
  // **Jede Lasche steht in genau einer Gruppe.**
  //
  // Eine zu kurze Aufzählung ist grün, und niemand sieht es (§3) — dieselbe
  // Sorge wie bei der Seitenliste selbst. Eine Lasche, die in keiner Gruppe
  // steht, verschwände aus der Leiste; eine in zweien stünde doppelt.
  //
  // **Seit UX-08 steht eine Setseite entweder in der Leiste oder in einer
  // Schublade** (#1680) — und zusammen sind es wieder alle, jede genau
  // einmal. Eine Seite in beiden stünde doppelt; eine in keinem wäre nur
  // noch über die Palette erreichbar, und das sähe niemand.
  //
  // **Seit UXR-09 ist der Ausgabeassistent der dritte Ort** (#1917):
  // *Ausgeben*, *Bilder* und *Druckbogen* schlägt der Knopf im Kopf auf.
  it('führt jede Lasche genau einmal — in der Leiste, in einer Schublade oder im Ausgabeassistenten', () => {
    const gruppiert = laschengruppen.flatMap((gruppe) => gruppe.laschen)
    const verstaut = schubladen.flatMap((schublade) => schublade.laschen)

    expect([...gruppiert, ...verstaut, ...ausgabeseiten].sort()).toEqual(
      [...setschluessel].sort(),
    )
  })

  // **Und die Leiste zeigt sie in der Reihenfolge der Gruppen.**
  it('gibt die Laschenreihe aus den Gruppen', () => {
    expect(setlaschen.map((lasche) => lasche.schluessel)).toEqual(
      laschengruppen.flatMap((gruppe) => gruppe.laschen),
    )
  })

  // **Der Weg des Einsteigers steht vor dem Werkzeugkasten.**
  //
  // Das ist die Abnahme des Werkstücks, als Reihenfolge ausgedrückt: wer von
  // links liest, kommt an *Karten* vorbei, bevor er an *Konsole* vorbeikommt —
  // und die Konsole steht in der **letzten** Gruppe, nicht neben dem
  // Karteneditor.
  //
  // **Seit UX-08 liegt die Technik ganz aus dem Weg**: die Konsole steht in
  // der Schublade *Werkzeuge*, nicht in der Leiste neben dem Karteneditor.
  it('stellt das Bauen vor alles andere und die Technik in eine Schublade', () => {
    const reihe = setlaschen.map((lasche) => lasche.schluessel)

    expect(reihe.indexOf('karten')).toBe(0)
    expect(reihe).not.toContain('konsole')
    expect(schubladeVon('konsole')?.schluessel).toBe('werkzeuge')
    expect(schubladeVon('stil')?.schluessel).toBe('eigenschaften')
    expect(schubladeVon('karten')).toBeUndefined()
  })

  // **Und jede Gruppe hat eine Aufschrift in beiden Katalogen.**
  //
  // Ohne diese Zeile wäre eine Gruppe ohne Namen ein `undefined` in der
  // Leiste — sichtbar erst im Fenster.
  it('nennt für jede Gruppe einen Katalogschlüssel', () => {
    for (const gruppe of laschengruppen) {
      expect(gruppe.schluessel).not.toBe('')
    }
  })
})
