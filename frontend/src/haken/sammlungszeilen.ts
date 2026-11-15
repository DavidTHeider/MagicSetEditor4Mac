// **Die Zeilen einer Sammlungsansicht, geholt und noch nicht geholt** —
// Werkstück #1500.
//
// # Warum Platzhalter und keine Lücken
//
// **Weil die Liste ihre volle Länge behalten muss.** `aria-rowcount` nennt
// alle Zeilen, die Tastatur erreicht jede, und ein Sprung ans Ende holt nach —
// die Zusagen aus U8. Eine kurze Liste bräche sie alle drei.
//
// **Ein Platzhalter trägt eine negative Kennung**: die Liste braucht einen
// Schlüssel je Zeile, und eine echte Kennung wäre eine Zeile, die es nicht
// gibt — anhaken ließe sich dann etwas, das niemand sieht.
import type { ExemplarDTO } from '../api'
import { keineHerkunft } from '../apiherkunft'

export interface Nachschlag {
  // schluessel benennt die Sicht, zu der die Zeilen gehören.
  schluessel: string

  // zeilen liegt nach dem Anfang des Fensters, das sie brachte.
  zeilen: Map<number, ExemplarDTO[]>
}

export const leererNachschlag: Nachschlag = {
  schluessel: '',
  zeilen: new Map(),
}

// erstesFenster ist, wie viele Zeilen der erste Ruf holt.
//
// # Warum zweihundert
//
// **Weil es mehr ist, als auf einen Bildschirm passt, und wenig genug, um
// nicht zu kosten** — dieselbe Zahl und derselbe Grund wie bei der früheren
// Zeilengrenze: bei 1280 × 800 stehen rund zehn Zeilen sichtbar da, und 200
// sind zwanzig Bildschirme Vorlauf für jeden, der scrollt statt zu suchen.
export const erstesFenster = 200

// platzhalter gibt eine Zeile, die noch nicht geholt ist.
export function platzhalter(stelle: number): ExemplarDTO {
  return {
    id: -(stelle + 1),
    fassung: 0,
    printing: '',
    druckID: '',
    vorlaeufig: false,
    menge: 0,
    zustand: '',
    sprache: '',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: '',
    setcode: '',
    nummer: '',
    lagerort: '',
    wunsch: false,
    bezeichnung: '',
    herkunft: keineHerkunft,
  }
}

// zeilenAus legt das erste Fenster, die nachgeholten und die Platzhalter
// übereinander.
export function zeilenAus(
  erste: readonly ExemplarDTO[],
  gesamt: number,
  nachgeholt: Nachschlag,
): ExemplarDTO[] {
  // **Ohne Gesamtzahl steht, was da ist.** Das ist der Zustand vor der ersten
  // Antwort — und eine Liste aus Platzhaltern wäre dort eine Behauptung über
  // eine Länge, die niemand kennt.
  if (gesamt === 0) {
    return [...erste]
  }

  const zeilen: ExemplarDTO[] = Array.from({ length: gesamt }, (_, stelle) =>
    platzhalter(stelle),
  )

  // **Was über das Ende hinausreicht, fällt weg.** Ein Fenster von vorhin
  // gehört zu einer Liste, die inzwischen kürzer sein kann — dazwischen hat
  // jemand gelöscht. Ohne diese Prüfung wüchse das Feld über `gesamt` hinaus,
  // und `aria-rowcount` nennte eine andere Zahl als die Liste hat.
  const einsetzen = (anfang: number, geholte: readonly ExemplarDTO[]) => {
    geholte.forEach((eines, versatz) => {
      if (anfang + versatz < gesamt) {
        zeilen[anfang + versatz] = eines
      }
    })
  }

  einsetzen(0, erste)
  for (const [anfang, geholte] of nachgeholt.zeilen) {
    einsetzen(anfang, geholte)
  }

  return zeilen
}
