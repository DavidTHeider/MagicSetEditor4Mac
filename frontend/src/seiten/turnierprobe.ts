// **Die Doppel der Turnierseite** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, wie `deckprobe.ts` und aus demselben Grund: die
// Läufe teilen sie, und eine Vorrichtung in einem von ihnen wäre für die
// anderen nicht da.
import type {
  PaarungDTO,
  RangDTO,
  RundeDTO,
  TeilnehmerDTO,
  TurnierDTO,
} from '../apiturniere'

export function turnierprobe(teil: Partial<TurnierDTO> = {}): TurnierDTO {
  return {
    id: '7',
    fassung: '1',
    name: 'Küchentisch',
    spiel: 'magic',
    notiz: '',
    einzelspiel: false,
    formatID: '',
    abgeschlossen: '',
    teilnehmer: [],
    runden: [],
    empfohleneRunden: 0,
    ...teil,
  }
}

export function teilnehmerprobe(
  teil: Partial<TeilnehmerDTO> = {},
): TeilnehmerDTO {
  return {
    id: '3',
    name: 'Ada',
    deckID: '',
    deckabzug: [],
    ausgestiegen: false,
    stelle: 1,
    ...teil,
  }
}

export function paarungsprobe(teil: Partial<PaarungDTO> = {}): PaarungDTO {
  return { id: '12', a: '3', b: '4', ergebnis: null, ...teil }
}

export function rundenprobe(teil: Partial<RundeDTO> = {}): RundeDTO {
  return { id: '9', nummer: 1, paarungen: [paarungsprobe()], ...teil }
}

// ohneFolgen antwortet auf die Frage nach den Folgen einer Berichtigung:
// **nichts verschiebt sich** — DOM-04.AK-02 (#1699). Für Läufe, die ein
// vorhandenes Ergebnis ändern und dabei nicht über die Frage gehen wollen.
export const ohneFolgen = () =>
  Promise.resolve({
    folgen: { runde: 1, spaetereRunden: [], verschoben: [] },
    fehler: null,
  })

export function rangprobe(teil: Partial<RangDTO> = {}): RangDTO {
  return {
    teilnehmerID: '3',
    name: 'Ada',
    platz: 1,
    siege: 1,
    niederlagen: 0,
    unentschieden: 0,
    byes: 0,
    matchpunkte: 3,
    matchquote: 1,
    gegnermatchquote: 0.33,
    spielquote: 1,
    gegnerspielquote: 0.33,
    ausgestiegen: false,
    ...teil,
  }
}
