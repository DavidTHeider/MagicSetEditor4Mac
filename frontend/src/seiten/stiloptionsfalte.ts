// **Welche Stiloptionen vorn stehen und welche dahinter** — UXR-07
// (#1915).
//
// # Warum gefaltet wird
//
// **R3 aus `docs/UIREVIEW.md`**: höchstens zwölf gezeichnete Felder, sonst
// gruppiert und eingeklappt bis auf eine Gruppe. 82 der 376 Stile des
// Full-Magic-Pack führen mehr als zwölf Optionen, `magic-m15-dungeons` 66
// (gezählt am 2026-10-02 über `styling field:`). Das Artboard *Eigenschaften*
// zeigt „4 von 11“ und legt den Rest unter *Mehr* — eingeklappt, nicht
// entfernt.
//
// # Warum nach der Reihenfolge des Pakets
//
// **Die Reihenfolge ist die des Stilautors**, und MSE zeigt sie so: der
// `StylingEditor` setzt `styling_style` des Stils
// (`src/gui/control/native_look_editor.cpp`, Zeilen 204–207), und
// `resizeViewers` legt die Felder in dieser Folge untereinander (Zeilen
// 46–83, Fassung f9d9356 vom 2021-07-06). Vorn steht, was der Autor zuerst
// nennt; eine eigene Rangfolge — etwa die gesetzten zuerst — ließe dieselbe
// Option je nach Stand an zwei Orten erscheinen.
//
// # Warum eine Zwischenüberschrift nicht zählt
//
// **Sie ist keine Option** — `InfoField` trägt `editable = false`
// (`src/data/field/information.hpp`, Zeilen 26–28, dieselbe Fassung). Sie
// überschreibt, was ihr folgt; deshalb geht sie mit ihm nach hinten, statt
// vorn über nichts zu stehen.

// sichtbareOptionen ist die Zahl der Optionen vor dem Falz — die des
// Artboards.
export const sichtbareOptionen = 4

const zwischenueberschrift = 'info'

interface Gefaltet<T> {
  vorn: T[]
  weitere: T[]

  // gezeigt und gesamt zählen Optionen, keine Zwischenüberschriften —
  // „4 von 11“.
  gezeigt: number
  gesamt: number
}

// falte teilt die Optionen eines Stils in die vorn und die dahinter.
export function falte<T extends { art: string }>(
  optionen: readonly T[],
): Gefaltet<T> {
  const echt = (option: T) => option.art !== zwischenueberschrift
  const gesamt = optionen.filter(echt).length

  // **Die erste Option hinter dem Falz** — und vor ihr jede
  // Zwischenüberschrift, die unmittelbar davor steht: `kopf` merkt, wo deren
  // Reihe beginnt.
  let falz = optionen.length
  let gezaehlt = 0
  let kopf: number | undefined

  for (const [stelle, option] of optionen.entries()) {
    if (!echt(option)) {
      kopf ??= stelle
      continue
    }

    if (gezaehlt === sichtbareOptionen) {
      falz = kopf ?? stelle
      break
    }

    gezaehlt += 1
    kopf = undefined
  }

  return {
    vorn: optionen.slice(0, falz),
    weitere: optionen.slice(falz),
    gezeigt: Math.min(sichtbareOptionen, gesamt),
    gesamt,
  }
}
