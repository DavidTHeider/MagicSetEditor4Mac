// **Wie ein Feldwert aussieht und wie eine Karte heißt** —
// #1623.
//
// # Warum das neben dem Bauteil steht und nicht darin
//
// **Weil es reine Fragen sind.** „Ist dieser Wert eine Liste?" und „wie
// heißt diese Karte?" haben mit dem Zeichnen nichts zu tun — und ein Zweig
// im JSX lässt sich nur über eine Oberfläche betreten, die ihn gerade
// hergibt. Hier steht er als Funktion und wird einzeln geprüft (§3).

import type { Spielobjekt } from '../apipartie'

// alsText gibt einen Feldwert als Zeichenkette.
//
// **Eine Liste wird wieder eine Zeile** — die Teile sind mit `|` getrennt,
// weil `nutzlastAus` sie an derselben Stelle zerteilt.
export function alsText(wert: string | string[]): string {
  return Array.isArray(wert) ? wert.join('|') : wert
}

// alsListe gibt einen Feldwert als Liste.
//
// **Eine leere Zeichenkette gibt eine leere Liste** und nicht eine mit einem
// leeren Eintrag: „nichts gewählt" ist keine Wahl.
export function alsListe(wert: string | string[]): string[] {
  if (Array.isArray(wert)) {
    return wert
  }

  return wert === '' ? [] : [wert]
}

// kartenwort gibt den Namen, unter dem eine Karte in einer Auswahl steht.
//
// **Drei Quellen, und die Kennung bleibt**: eine Karte trägt ihren Namen in
// den Eigenschaften (M2), ein Spielstein am Objekt, und wo beides fehlt, ist
// die Kennung besser als ein leerer Eintrag (§8).
export function kartenwort(objekt: Spielobjekt): string {
  const eigen = objekt.eigenschaften?.name ?? ''

  if (eigen !== '') {
    return eigen
  }

  if (objekt.name !== '') {
    return objekt.name
  }

  return objekt.karte === '' ? objekt.kennung : objekt.karte
}

// zielVon gibt die zweite Hälfte einer Paarzeile — leer, wo keine steht.
export function zielVon(zeilen: string[], kreatur: string): string {
  const zeile = zeilen.find((eine) => eine.startsWith(`${kreatur}|`))

  if (zeile === undefined) {
    return ''
  }

  // **Abgeschnitten und nicht zerteilt**: `startsWith` hat das Trennzeichen
  // schon belegt, und ein `split(…)[1] ?? ''` trüge einen Rückfall, den kein
  // Lauf betreten kann (§10).
  return zeile.slice(kreatur.length + 1)
}
