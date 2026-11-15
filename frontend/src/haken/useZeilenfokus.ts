// **Der Fokus folgt der Auswahl in die Kartenliste** —
// #1290.
//
// # Warum das ein eigener Haken ist
//
// Weil die Zeile beim Springen ans Ende **noch nicht da ist**. Die Liste
// zeichnet nur ihr Sichtfenster (#891); wer mit Ende an das Ende springt,
// wählt eine Karte, deren Zeile erst nach dem Rollen entsteht — und der Fokus
// kann ihr nicht folgen, solange sie fehlt.
//
// Gemessen am 2026-08-30 an der laufenden Brücke, vor diesem Haken: ↓ wählte
// die zweite Karte, und **Ende tat danach nichts** — der Fokus lag am `body`,
// weil die Zeile darunter neu gebaut worden war.
//
// # Warum kein Beobachter wie bei [useFeldfokus]
//
// Weil die Abhängigkeiten hier reichen: `erste` und `letzte` **sind** das
// Sichtfenster, und Rollen ändert genau sie. Der Effekt läuft damit noch
// einmal, sobald die Zeile gezeichnet ist — ohne einen Beobachter, der auf den
// ganzen Baum hört.
//
// # Nur bei einer Taste
//
// Wer mit der Maus klickt, hat seinen Fokus schon dort, wo er ihn haben will.
// Ihn dann zu verschieben nähme ihm die Wahl — deshalb kommt der Wunsch von
// außen und wird vom Tastenweg gesetzt.
//
// **Und er wird als Wert übergeben, nicht als Referenz.** Eine fremde Referenz
// zu beschreiben ist eine Wirkung an einem Ort, den der Aufrufer besitzt; der
// Rückruf sagt stattdessen, was geschehen ist, und der Aufrufer entscheidet.

import { useEffect, type RefObject } from 'react'

export function useZeilenfokus(
  gitter: RefObject<HTMLElement | null>,
  ziel: string | undefined,
  erledigt: () => void,
  gewaehlt: string | undefined,
  sichtfenster: readonly [number, number],
): void {
  const [erste, letzte] = sichtfenster

  useEffect(() => {
    if (ziel === undefined || ziel !== gewaehlt) {
      return
    }

    const zeile = gitter.current?.querySelector<HTMLElement>(
      `[data-karte="${gewaehlt}"]`,
    )

    if (zeile === null || zeile === undefined) {
      return
    }

    // **Nur, wenn niemand den Fokus inzwischen selbst gesetzt hat** — Befund 6
    // des Zweitblicks über Phase 4.3. Warum, steht bei [useFeldfokus].
    //
    // **Und die eigene Zeile zählt als „niemand"**: sie ist der Ort, an dem
    // der Tastenweg begonnen hat.
    const steht = document.activeElement
    if (
      steht !== null &&
      steht !== document.body &&
      steht.closest('[role="grid"]') === null
    ) {
      erledigt()

      return
    }

    zeile.focus()
    erledigt()
    // **`erste` und `letzte` gehören in die Liste**, auch wenn der Rumpf sie
    // nicht liest: sie sind das Sichtfenster, und ihre Änderung ist genau das
    // Ereignis, auf das dieser Effekt wartet.
  }, [gitter, ziel, erledigt, gewaehlt, erste, letzte])
}
