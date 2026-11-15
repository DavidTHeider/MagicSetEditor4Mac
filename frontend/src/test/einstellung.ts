// **Eine Einstellung für Vorrichtungen** — Werkstück 9 von Phase 4.12
// (#1351).
//
// # Warum es diesen Helfer gibt
//
// **Weil ein `EinstellungDTO` seit Werkstück 9 fünf Felder hat.** Einundzwanzig
// Stellen bauten es von Hand, und jedes neue Feld hieß einundzwanzig
// Änderungen — von denen jede vergessen werden kann. Der Helfer ist die eine
// Stelle.
//
// # Und warum `wirksam` dem `wert` folgt
//
// **Weil das der Normalfall ist**: nur wo `erben` gewählt ist, weichen die
// beiden voneinander ab. Ein Doppel, das sie immer gleichsetzte, wäre
// großzügiger als der Betrieb — deshalb lässt sich `wirksam` überschreiben,
// und genau das tut der Lauf über die Übersteuerung.
import type { EinstellungDTO } from '../api'

export function probeeinstellung(
  teile: Partial<EinstellungDTO> & Pick<EinstellungDTO, 'schluessel'>,
): EinstellungDTO {
  const wert = teile.wert ?? ''

  return {
    wert,
    wirksam: wert,
    werte: [],
    bereich: 'anwendung',
    ...teile,
  }
}
