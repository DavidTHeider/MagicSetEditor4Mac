// **Was an einer Karte sofort geht** — PLAY-04.AK-01
// (#1712).
//
// # Wofür es das gibt
//
// **Weil das Kriterium drei Wege nennt** — „Drag/Drop, Kontextaktionen und
// Tastaturwege" — und der zweite am Tisch fehlte. Gemessen am 2026-09-22:
// `bausteine/Kontextmenue.tsx` hatte genau einen Aufrufer, und das war die
// **Platzliste eines Decks**.
//
// # Und warum nur die ohne Feld
//
// **Weil ein Menüeintrag eine Handlung ist und kein Formular.** „Schieben"
// braucht eine Zone, „Zähler" eine Zahl, „Anhängen" ein zweites Objekt; wer
// das in ein Kontextmenü legte, baute darin ein zweites Fenster nach. Die
// Handlungsleiste stellt sie weiterhin, und das Menü führt mit einem Eintrag
// dorthin.
//
// **Damit gilt die Abnahme von U2 weiter**
// (#1413):
// *keine Aktion, die es nur im Kontextmenü gibt.* Ein Lauf daneben hält es —
// alles, was hier herauskommt, steht auch in der Handlungsleiste.
import { formenDerGruppe } from './formenwahl'

// Objektzustand ist, was ein Befehl am Objekt selbst nachliest.
interface Objektzustand {
  getappt: boolean
  haengtAn: string
  haufen: string
}

// sofortigeHandlungen gibt die Befehlsarten, die an dieser Karte ohne ein
// weiteres Feld gestellt werden können — in der Reihenfolge des Tisches.
export function sofortigeHandlungen(
  erlaubt: readonly string[],
  objekt: Objektzustand,
): string[] {
  return formenDerGruppe('karte', erlaubt, objekt)
    .filter((form) => form.felder.length === 0)
    .map((form) => form.art)
}
