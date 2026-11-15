// Eine Bestandspflege, die nichts tut — für Tests, die sie nur brauchen, weil
// die Seite sie verlangt (#940).
//
// **Ruhende Vorgaben und keine Spione.** Wer die Pflege selbst prüft, baut sich
// ihre Teile im Test; wer eine Seite prüft, auf der sie nur mitläuft, soll
// nicht an ihr vorbeikommen müssen.
import { vi } from 'vitest'

import type { Bestandspflege } from '../haken/useBestandspflege'
import type { Pflegelauf } from '../haken/useDatenpflege'
import type { Ruecksprungstand } from '../haken/useRuecksprung'

export function stillePflege(
  teile: Partial<Bestandspflege> = {},
): Bestandspflege {
  return {
    sicherung: {
      sichern: vi.fn(),
      ausMenue: vi.fn(),
      laeuft: false,
      pfad: undefined,
      fehler: undefined,
      raeumen: vi.fn(),
    },
    rueckspiel: {
      waehlen: vi.fn(),
      gesperrt: undefined,
      entsperren: vi.fn(),
      zurueckspielen: vi.fn(),
      laeuft: false,
      auswahl: undefined,
      lage: undefined,
      beiseite: undefined,
      fehler: undefined,
      raeumen: vi.fn(),
    },
    chiffre: {
      verschluesselt: false,
      umschalten: vi.fn(),
      laeuft: false,
      fehler: undefined,
    },
    alter: {
      stand: undefined,
      planfehler: undefined,
      fehler: undefined,
      einrichten: vi.fn(),
      beenden: vi.fn(),
      laeuft: false,
    },
    ersetzt: 0,
    gewechselt: vi.fn(),
    veraendert: vi.fn(),
    ...teile,
  }
}

// stillerRuecksprung ist die Rückkehrseite, die nichts tut — Werkstück O7 von
// Phase 5.2 (#1431).
//
// **Aus demselben Grund wie [stillePflege]**: sie läuft auf der
// Einstellungsseite mit, und wer die Anbieterkennung prüft, soll nicht an ihr
// vorbeikommen müssen.
export function stillerRuecksprung(
  teile: Partial<Ruecksprungstand> = {},
): Ruecksprungstand {
  return {
    arten: [],
    rueckkehren: [],
    laedt: false,
    fehler: undefined,
    quittieren: vi.fn(),
    laeuft: () => false,
    ...teile,
  }
}

// stilleDatenpflege ist das Aufräumen, das nichts tut — Werkstück O8 von
// Phase 5.2 (#1432).
//
// **Aus demselben Grund wie [stillePflege]**: es läuft auf der
// Einstellungsseite mit.
export function stilleDatenpflege(teile: Partial<Pflegelauf> = {}): Pflegelauf {
  return {
    nachsehen: vi.fn(),
    pflegen: vi.fn(),
    laeuft: false,
    vorschau: undefined,
    bilanz: undefined,
    fehler: undefined,
    raeumen: vi.fn(),
    ...teile,
  }
}
