// **Die Schubladen des Sets** — UX-08.AK-01
// (#1680),
// nach den Artboards *Eigenschaften* und *Werkzeuge* des Entwurfs.
//
// # Warum es sie gibt
//
// **Aus vierzehn gleichrangigen Einträgen werden ein Arbeitsplatz und vier
// Nebenflächen** — so steht es als Kern der Umordnung in
// `docs/PROTOTYP-0.6.0.md`. Zwei davon sind diese Schubladen: was für das
// ganze Set gilt (Stil, Angaben, Symbol, Schlüsselwörter) und was man
// selten braucht (Statistik, Zufallspack, Konsole). Sie fahren über die
// Kartenfläche, statt sie zu ersetzen — wer den Stil ändert, sieht dahinter
// die Karte, für die er es tut.
//
// **Die Seiten bleiben Seiten.** Jede behält ihren Schlüssel, ihr Kürzel und
// ihren Eintrag in der Befehlspalette; nur der Ort, an dem sie erscheinen,
// ist ein anderer (`docs/inventur.json`, Muster „Drawer").
//
// # Woher die Einteilung kommt
//
// **Aus dem Entwurf, und der aus MSE2**: Stil, Setangaben, Schlüsselwörter,
// Statistik, Zufallspack und Konsole sind dort sechs der sieben Reiter des
// Setfensters (`src/gui/set/window.cpp`, Zeilen 148–154, Fassung `a7336d6`
// vom 2020-06-14), der Symboleditor ein eigenes Fenster. Die Trennung in
// „gilt für das Set" und „Werkzeug" ist **new** — MSE2 kennt sie nicht; sie
// ist die Aussage, die im Entwurf zur Prüfung steht.
import type { Seitenschluessel, Setschluessel } from './navigation'

export const schubladenschluessel = ['eigenschaften', 'werkzeuge'] as const

export type Schubladenschluessel = (typeof schubladenschluessel)[number]

export interface Schublade {
  schluessel: Schubladenschluessel

  // laschen sind die Seiten der Schublade, in der Reihenfolge des Entwurfs —
  // **nie leer**, und der Typ sagt es: die erste ist die, bei der die
  // Schublade aufgeht.
  laschen: readonly [Setschluessel, ...Setschluessel[]]
}

export const schubladen: readonly Schublade[] = [
  {
    schluessel: 'eigenschaften',
    laschen: ['stil', 'setangaben', 'symbol', 'schluesselwoerter'],
  },
  { schluessel: 'werkzeuge', laschen: ['statistik', 'pack', 'konsole'] },
]

// schubladeVon gibt die Schublade, in der eine Seite steht — oder keine: dann
// steht sie auf der Arbeitsfläche selbst oder ist gar keine Setseite.
export function schubladeVon(seite: Seitenschluessel): Schublade | undefined {
  return schubladen.find((schublade) =>
    (schublade.laschen as readonly Seitenschluessel[]).includes(seite),
  )
}
