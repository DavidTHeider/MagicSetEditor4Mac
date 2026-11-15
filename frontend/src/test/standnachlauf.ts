// **Der Stand eines Handels, wie der Dienst ihn nachliest** — P13, Befund 1
// des Zweitblicks über Phase 8 (#2329).
//
// Ein Zug liest vor seinem ersten Schritt den Stand nach (`HandelStand`) und
// geht nichts, wenn er nicht mehr der gesehene ist. Im Betrieb lesen
// `HandelStand`, `HandelVorgaenge` und die Antwort eines Schritts denselben
// Bestand. **Diese Vorrichtung hält dieselbe Zusage**: sie nennt den Stand,
// den `HandelVorgaenge` oder ein Schritt zuletzt nannte — und für einen
// Handel, den keiner nannte, einen Fehler mit dem Namen der Methode, statt
// einen zu erfinden.
//
// **Was sie nicht tut** (Befund 4 des Nachlaufs über Phase 8): einen Stand
// nach einem Abschlussschritt fortschreiben — Verbuchen, Abgleich und
// Bestätigung antworten nicht mit einem Stand, und was danach gilt, weiß nur
// der Lauf. Ein Lauf, der einen Zug über die Buchung hinaus prüft, belegt
// `HandelStand` und `HandelVorgaenge` selbst aus seinem Bestand.
import type { Api } from '../api'
import type {
  Handelsvorgang as Vorgangsstand,
  Handelsvorgangantwort,
} from '../apihandelsleitung'

type Belegung = Partial<Api>

// mitStandnachlauf gibt einer Belegung `HandelStand` dazu — wo sie keines
// selbst belegt.
export function mitStandnachlauf(belegung: Belegung): Belegung {
  const zuletzt = new Map<string, Vorgangsstand>()
  const merke = (stand: Vorgangsstand) => {
    zuletzt.set(stand.vorgang, stand)
  }

  const { HandelVorgaenge, HandelSchritt, HandelEroeffnen } = belegung

  return {
    ...belegung,
    ...(HandelVorgaenge && {
      HandelVorgaenge: async (gegenueber: string) => {
        const antwort = await HandelVorgaenge(gegenueber)
        antwort.vorgaenge.forEach(merke)

        return antwort
      },
    }),
    ...(HandelSchritt && {
      HandelSchritt: async (
        vorgang: string,
        schritt: string,
        wer: string,
        gesehen: string,
      ) => {
        const antwort = await HandelSchritt(vorgang, schritt, wer, gesehen)

        if (antwort.fehler === null) {
          merke(antwort.vorgang)
        }

        return antwort
      },
    }),
    ...(HandelEroeffnen && {
      HandelEroeffnen: async (
        wunsch: Parameters<Api['HandelEroeffnen']>[0],
      ) => {
        const antwort = await HandelEroeffnen(wunsch)

        if (antwort.fehler === null) {
          merke(antwort.vorgang)
        }

        return antwort
      },
    }),
    HandelStand:
      belegung.HandelStand ??
      ((vorgang: string): Promise<Handelsvorgangantwort> => {
        const stand = zuletzt.get(vorgang)

        return Promise.resolve(
          stand === undefined
            ? ({
                vorgang: {},
                fehler: {
                  code: 'nichtGefunden',
                  meldung: `HandelStand: ${vorgang} nannte bisher niemand`,
                  schluessel: '',
                  feld: '',
                },
              } as unknown as Handelsvorgangantwort)
            : { vorgang: stand, fehler: null },
        )
      }),
  }
}
