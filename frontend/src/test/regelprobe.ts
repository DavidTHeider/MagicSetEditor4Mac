// **Die Vorrichtung der Regelläufe** — eine Antwort der Regelfreigabe, wie
// die Naht sie gibt.
//
// **Eine eigene Datei**, seit Hindernisse und Widersprüche aus der
// Regelschau in die Regeltafel wanderten (das Artboard *Eine Partie
// anfangen*): zwei Läufe brauchen dieselbe Antwort, und eine Vorrichtung,
// die zweimal dastünde, liefe beim ersten neuen Feld auseinander.
import type { RegelfreigabeDTO } from '../apiregeln'

// eineFreigabe gibt eine Antwort, wie die Naht sie gibt.
export function eineFreigabe(
  teile: Partial<RegelfreigabeDTO> = {},
): RegelfreigabeDTO {
  return {
    erlaubt: true,
    regelwerk: 'sandkasten',
    fassung: 1,
    wirksam: '2026-08-19',
    hoechste: 'begleitet',
    modus: 'anfaenger',
    baum: [
      { kennung: 'ablauf', bereiche: ['zonen'] },
      { kennung: 'mechanik', bereiche: ['marken'] },
    ],
    widersprueche: [],
    matrix: [
      {
        bereich: 'zonen',
        grad: 'automatisch',
        quelle: 'CR 400 (Fassung 2026-08-19)',
        grenzen: ['kein Auslöser, wenn eine Karte die Zone wechselt'],
      },
      { bereich: 'combat', grad: 'manuell', quelle: '', grenzen: [] },
    ],
    luecken: ['combat'],
    hindernisse: [],
    ...teile,
  }
}
