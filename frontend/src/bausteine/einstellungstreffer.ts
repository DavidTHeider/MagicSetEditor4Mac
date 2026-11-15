// **Wonach die Einstellungssuche vergleicht** — UX-04.AK-03
// (#1676).
//
// # Warum das ein eigenes Modul ist
//
// **Weil es seit dem 2026-09-21 zwei Verbraucher hat.** Die Regel stand im
// Trefferkasten, und der fiel weg: der Entwurf zeigt keine zweite Liste,
// sondern markiert das Feld **an seinem Platz**. Seitdem fragt die
// Gruppenliste nach ihren Zahlen und die Fläche nach ihren Marken — zwei
// Stellen, eine Regel (§10).
import { dienstnamen, schalternamen } from './dienstnamen'

import type { EinstellungDTO } from '../api'
import type { Katalog } from '../sprache/kataloge'
import { falte } from '../sprache/vergleich'

// namenAus legt die drei Namensquellen übereinander.
//
// **Dieselben wie in den Gebieten**: die Anwendungseinstellungen aus dem
// Katalog, die Schalter aus `dienstnamen`, und der Katalog selbst trägt
// oben noch seine eigenen Schlüssel.
function namenAus(texte: Katalog): Record<string, string | undefined> {
  const eigene = texte.seiten.einstellungen

  return {
    ...(eigene as unknown as Record<string, string | undefined>),
    ...eigene.namen,
    ...schalternamen,
    ...dienstnamen,
  }
}

// passtZurFrage sagt, ob eine Einstellung zu einer Suchfrage gehört.
//
// **Eine leere Frage passt auf nichts.** „Ich suche nichts" ist nicht
// „alles passt" — sonst stünde beim ersten Zeichnen die ganze Seite als
// Treffer da.
export function passtZurFrage(
  eintrag: EinstellungDTO,
  frage: string,
  texte: Katalog,
  sprache: string,
): boolean {
  const gesucht = falte(frage.trim(), sprache)

  if (gesucht === '') {
    return false
  }

  const eigene = texte.seiten.einstellungen
  const namen = namenAus(texte)
  const hinweise: Record<string, string | undefined> = eigene.hinweise
  const werte: Record<string, string | undefined> = eigene.werte

  const texteDerZeile = [
    eintrag.schluessel,
    namen[eintrag.schluessel] ?? '',
    hinweise[eintrag.schluessel] ?? '',
    ...eintrag.werte.map(
      (wert) => werte[`${eintrag.schluessel}.${wert}`] ?? wert,
    ),
  ]

  return texteDerZeile.some((text) => falte(text, sprache).includes(gesucht))
}
