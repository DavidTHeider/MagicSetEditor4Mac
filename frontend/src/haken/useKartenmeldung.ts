// **Die aufgeschlagene Karte gilt im ganzen Fenster** —
// #892, und
// seit UX-08 (#1680)
// auch die, die niemand angeklickt hat.
//
// # Warum jede Wahl und nicht nur der Klick
//
// **Weil die Kartenseite auf mehr als einem Weg wählt**: per Klick, per ⇞/⇟,
// als Nachfolgerin einer gelöschten Karte, als neu angelegte — und beim
// Öffnen eines Sets die erste, ohne dass jemand etwas tut (#668). Gemeldet
// wurde bis zum 2026-09-23 nur der Klick. Die Skriptkonsole, die in der
// Schublade *Werkzeuge* **über** dieser Karte liegt, sagte deshalb „Dafür
// muss eine Karte aufgeschlagen sein", während sie darunter im Editor stand.
// Gefunden in der Sichtprüfung der Schubladen.
//
// **Eine Stelle für alle Wege** — dieselbe Überlegung wie bei MSE2, das eine
// einzige Auswahl des Setfensters führt (`src/gui/set/window.cpp`, Zeilen
// 305–320, Fassung f9d9356 vom 2021-07-06).
//
// # Warum als Effekt
//
// **Weil die Meldung den Zustand des Fensters ändert**, und das darf nicht
// geschehen, während die Kartenseite zeichnet — React weist ein Setzen im
// fremden Zeichnen ab. Die Vorwahl selbst steht beim Zeichnen (`Karten.tsx`);
// das Fenster erfährt sie einen Schritt später.
import { useEffect } from 'react'

// useKartenmeldung meldet `gewaehlt` nach oben, sobald es sich von der
// Vorgabe des Fensters unterscheidet. **Was das Fenster schon kennt, wird
// nicht wiederholt** — sonst zeichnete jede Meldung neu, und jede Zeichnung
// meldete wieder.
export function useKartenmeldung(
  gewaehlt: string | undefined,
  vorgabe: string | undefined,
  melden: (karteID: string) => void,
): void {
  useEffect(() => {
    if (gewaehlt !== undefined && gewaehlt !== vorgabe) {
      melden(gewaehlt)
    }
  }, [gewaehlt, vorgabe, melden])
}
