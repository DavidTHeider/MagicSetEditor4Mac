// **Der Auftrag eines Feldvorschlags** — AI-05.AK-05 und AK-06
// (#2042).
//
// # Warum die Karte mitgeht
//
// **Das Modell sieht nur, was im Auftrag steht** — das Schema sagt ihm, welche
// Felder es füllen darf, nicht, was auf der Karte schon steht. Ein Name für
// eine Karte, von der es nichts weiß, wäre geraten. Deshalb hängt der Auftrag
// die Textfelder an, **gekürzt** auf die Grenze einer kleinen Aufgabe: darüber
// käme das Gerätemodell nicht dran.

import type { WertDTO } from '../formen'

// kleineFelder sind dieselben wie in `domain.IstKleineTextaufgabe` — eine
// zweite Liste mit anderem Inhalt fragte das Gerät nach etwas, das der Dienst
// ihm nicht gibt.
const kleineFelder = new Set(['name', 'flavor text'])

// auftragsgrenze ist `domain.KIKleineAuftragsgrenze` — gemessen, nicht
// geschätzt (Vermerk 0166).
export const auftragsgrenze = 2000

// wertgrenze ist, wie viel eines einzelnen Feldes mitgeht — damit ein langer
// Regeltext die anderen nicht verdrängt.
const wertgrenze = 300

// istKleinesFeld sagt, ob ein Feld einen Vorschlag am Feld bekommt.
export function istKleinesFeld(feld: string): boolean {
  return kleineFelder.has(feld)
}

// kiFeldauftrag baut den Auftrag aus Satz, Überschrift und Kartenwerten.
export function kiFeldauftrag(
  satz: string,
  ueberschrift: string,
  werte: readonly WertDTO[],
  formen: ReadonlyMap<string, { art: string }>,
): string {
  const zeilen = werte
    .filter((wert) => wert.wert !== '' && formen.get(wert.feld)?.art === 'text')
    .map((wert) => `${wert.feld}: ${gekuerzt(wert.wert, wertgrenze)}`)

  if (zeilen.length === 0) {
    return satz
  }

  return gekuerzt(
    `${satz}\n\n${ueberschrift}\n${zeilen.join('\n')}`,
    auftragsgrenze,
  )
}

// gekuerzt schneidet nach Zeichen, nicht nach UTF-16-Einheiten — die Grenze
// zählt Zeichen wie `utf8.RuneCountInString`.
function gekuerzt(text: string, grenze: number): string {
  const zeichen = Array.from(text)

  return zeichen.length <= grenze ? text : zeichen.slice(0, grenze).join('')
}
