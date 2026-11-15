// **Die Tasten in der Kartenliste** —
// #1290.
//
// # ported (§15)
//
// MSE2 zeigt seine Karten in einem `wxListCtrl` (`src/gui/control/card_list.cpp`,
// Fassung f9d9356 vom 2021-07-06). Wie eine solche Liste sich bewegt, steht
// nicht in MSEs Quelltext, sondern in der Plattform — und dort überall gleich:
// ↑ und ↓ bewegen die Auswahl, Pos1 und Ende springen an die Enden, und der
// Tabulator verlässt die Liste als **Ganzes**.
//
// Was MSE2 dazu selbst schreibt, ist die **Grenze**:
//
//	bool ItemList::canSelectPrevious() const { return selected_item_pos - 1 >= 0; }
//	bool ItemList::canSelectNext() const {
//	  return selected_item_pos >= 0 && (size_t)(selected_item_pos + 1) < sorted_list.size();
//	}
//
// `src/gui/control/item_list.cpp`, Zeilen 37–42 — **kein Umlauf**, und
// dieselbe Regel, die [nachbarkarte] trägt.
//
// # Warum ein `grid` und keine `listbox`
//
// Weil eine Zeile **zwei** Dinge trägt: das Ankreuzfeld für die Massenwahl und
// den Namen. Ein `option` darf nichts Bedienbares enthalten, und `aria-checked`
// am `option` vermischte „aufschlagen" mit „anhaken" — zwei verschiedene
// Fragen an dieselbe Zeile.
//
// **Was es einbringt**, gemessen am 2026-08-30 bei 1280 × 800: die Liste war
// **54** Tabulatorschritte — 27 gezeichnete Zeilen mal zwei Bedienelemente —
// und ist jetzt **einer**. Der Weg bis zum ersten Kartenfeld fiel damit von 85
// auf 20.
//
// # Warum eine eigene Datei
//
// `Kartenliste.tsx` läge mit ihr über der Zeilengrenze (§2), und der Schnitt
// folgt der Frage: dort steht, **wie** eine Zeile aussieht, hier, **was eine
// Taste in ihr bedeutet**.

import type { KeyboardEvent } from 'react'

// blaettertaste deutet eine Taste in einer Zeile der Kartenliste.
//
// **Auch für die Liste der Schlüsselwörter** (UXR-F3,
// #2009):
// „Tastatur wie in der Kartenliste" heißt hier dieselbe Funktion und keine
// zweite Fassung. Sie braucht von einer Zeile nur ihre Kennung — deshalb
// nimmt sie diesen Ausschnitt und nicht `KartenzeileDTO`.
//
// **Ohne Zusatztaste.** ⌘↓ ist bei macOS etwas anderes, und ⇧↓ wäre eine
// Bereichswahl — die gibt es hier nicht, und sie zu verschlucken hieße, sie zu
// versprechen.
export function blaettertaste(
  ereignis: KeyboardEvent<HTMLElement>,
  karten: readonly { readonly id: string }[],
  stelle: number,
  waehlen: (karteID: string) => void,
): void {
  if (
    ereignis.metaKey ||
    ereignis.altKey ||
    ereignis.ctrlKey ||
    ereignis.shiftKey
  ) {
    return
  }

  // **→ und ← führen in die Zeile und zurück** — die Gestalt eines Gitters.
  if (ereignis.key === 'ArrowRight' || ereignis.key === 'ArrowLeft') {
    inDieZeile(ereignis)

    return
  }

  const ziel = zielstelle(ereignis.key, stelle, karten.length)
  if (ziel === undefined) {
    return
  }

  // **Die Taste ist verbraucht, auch am Rand** — sonst rollte die Liste, und
  // das sähe aus wie ein halber Sprung.
  ereignis.preventDefault()

  const karte = karten[ziel]
  if (karte !== undefined && ziel !== stelle) {
    waehlen(karte.id)
  }
}

// zielstelle sagt, wohin eine Taste führt — `undefined`, wenn sie nichts tut.
//
// **Geklemmt und nicht umgelaufen.** MSE2 prüft die Grenze und tut dann
// nichts; das Ergebnis ist dasselbe wie Klemmen, und Klemmen ist die kürzere
// Schreibweise.
function zielstelle(
  taste: string,
  stelle: number,
  anzahl: number,
): number | undefined {
  switch (taste) {
    case 'ArrowDown':
      return Math.min(stelle + 1, anzahl - 1)

    case 'ArrowUp':
      return Math.max(stelle - 1, 0)

    case 'Home':
      return 0

    case 'End':
      return anzahl - 1

    default:
      return undefined
  }
}

// inDieZeile bewegt den Fokus zwischen den Bedienelementen einer Zeile.
//
// **Ausgeführt, weil zwei Gitter sie brauchen** (#1290): die Kartenliste und
// die Paketliste. Zwei Fassungen davon liefen an dem Tag auseinander, an dem
// eine von ihnen einen dritten Griff je Zeile bekam.
//
// **Über den DOM und nicht über einen Zustand.** Welche Zelle bedienbar ist,
// entscheidet die Zeile beim Zeichnen; eine zweite Liste daneben liefe mit ihr
// auseinander.
export function inDieZeile(ereignis: KeyboardEvent<HTMLElement>): void {
  const zeile = ereignis.currentTarget
  const bedienbar = [
    ...zeile.querySelectorAll<HTMLElement>(
      'input, button, a[href], select, textarea',
    ),
  ]

  // **Kein Wächter auf die leere Zeile.** Der Zugriff unten beantwortet sie
  // schon: `bedienbar[…]?.focus()` tut dann nichts, und ein `if` davor wäre
  // ein Zweig, den kein Lauf betritt (§3).
  ereignis.preventDefault()

  const stand = bedienbar.indexOf(document.activeElement as HTMLElement)
  const weiter = ereignis.key === 'ArrowRight' ? stand + 1 : stand - 1

  // **Am linken Rand zurück auf die Zeile.** Sonst käme man nicht mehr heraus,
  // ohne die Liste zu verlassen.
  if (weiter < 0) {
    zeile.focus()

    return
  }

  bedienbar[Math.min(weiter, bedienbar.length - 1)]?.focus()
}

// zurZeile rollt die Liste so weit, dass eine Stelle sichtbar wird.
//
// **Ohne das gibt es die Zeile nicht.** Die Liste zeichnet nur ihr Sichtfenster
// (#891); wer mit Ende an das Ende springt, wählt eine Karte, deren Zeile
// **nicht im DOM steht** — und der Fokus kann ihr nicht folgen.
//
// **Gerollt wird nur, wenn nötig**, und nur so weit: ein Sprung, der die Zeile
// immer in die Mitte holte, verschöbe bei jedem ↓ die ganze Liste.
export function zurZeile(
  gitter: HTMLElement | null,
  stelle: number,
  zeilenhoehe: number,
): void {
  if (gitter === null || zeilenhoehe === 0) {
    return
  }

  const oben = stelle * zeilenhoehe
  const unten = oben + zeilenhoehe

  if (oben < gitter.scrollTop) {
    gitter.scrollTop = oben

    return
  }

  if (unten > gitter.scrollTop + gitter.clientHeight) {
    gitter.scrollTop = unten - gitter.clientHeight
  }
}

// zielstelleVon gibt die Stelle, auf die eine Taste führt — für den Aufrufer,
// der davor rollen muss.
export function zielstelleVon(
  taste: string,
  stelle: number,
  anzahl: number,
): number | undefined {
  return zielstelle(taste, stelle, anzahl)
}
