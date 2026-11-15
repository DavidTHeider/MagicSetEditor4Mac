// **Wie ein Exemplar beginnt** — herausgelöst aus `Sammlungsformular.tsx`,
// als es mit UX-06 über die Zeilengrenze wuchs (§2). Der Schnitt folgt der
// Frage: dort steht, wie das Formular aussieht und was es prüft, hier, womit
// es anfängt — nach einem geglückten Aufnehmen wieder.
import { keineHerkunft } from '../apiherkunft'

import type { ExemplarDTO } from '../api'

// leeresExemplar ist der Stand eines frischen Formulars.
//
// **Die Sprache bleibt leer.** Ihre Vorgabe setzt die Domäne — `en` —, und
// ein Formular, das sie vorwegnähme, hätte zwei Stellen, an denen sie steht.
//
// **Der Zustand kommt seit Werkstück 9 von Phase 4.12 aus dem Profil**
// (#1351,
// Punkt 3): wer immer dasselbe einträgt, sagt es einmal. **Ohne Vorgabe
// bleibt er leer, wie er es immer war** — und die Domäne macht daraus
// `nearMint`, wie sie es immer tat.
export function leeresExemplar(
  zustand: string,
  lagerort: string,
  waehrung: string,
): ExemplarDTO {
  return {
    id: 0,

    // **Fassung 0 für etwas, das es noch nicht gibt.** Sie kommt vom Bestand,
    // sobald die Zeile dort steht.
    fassung: 0,

    printing: '',
    druckID: '',
    vorlaeufig: false,
    menge: 1,
    zustand,
    sprache: '',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',

    // **Die Währung aus dem Profil** — LOC-03
    // (#1689).
    // Sie steht vor und rechnet nicht (AK-02); ohne Vorgabe bleibt sie leer.
    erwerbswaehrung: waehrung,
    notiz: '',

    // **Der lesbare Bezug** (#1200) — leer, bis jemand einen Druck aus der
    // Suche wählt; ohne ihn steht die Kennung da.
    name: '',
    setcode: '',
    nummer: '',

    // **Der Lagerort** (#1292, Vermerk 0057) — Freitext, leer erlaubt.
    //
    // **Mit der Vorgabe des Profils, wenn es eine gibt** — Werkstück 15
    // (#1351), Punkt 7.
    lagerort,

    // **Die Absicht.** Falsch heißt Besitz; wahr heißt Wunsch — und ein
    // Wunsch zählt weder ins Portfolio noch in »besitze ich das?«.
    wunsch: false,

    // **Die Bezeichnung kommt aus Go** und wird hier nie gesetzt: sie ist
    // gerechnet, und ein Wert von hier wäre eine zweite Wahrheit (§16).
    bezeichnung: '',
    herkunft: keineHerkunft,
  }
}
