// **Die Zeilen der Aufmerksamkeit** — AP-8
// (#2365),
// Befund B-14. Sets und Decks stehen je für sich; **alle Pakete teilen eine
// Zeile**, an der Stelle des ersten. 274 Paketbefunde mit ihren rohen
// Skriptsätzen beherrschten die Liste, und wer einem Set oder Deck nachging,
// fand es unter ihnen nicht.
import type { HinweisDTO } from '../apiaufmerksamkeit'
import type { Setbefund } from './problemlage'
import { befundsatz } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import type { Katalog } from '../sprache/kataloge'

// **Die Paketzeile trägt die Hinweise und nicht nur die Namen** — ihre
// Gründe stehen auf Wunsch darunter. Mit einer Zahl allein ginge verloren,
// was die Startseite bis dahin je Paket sagte: der erste Skriptfehler, in der
// Sprache des Fensters (#1963). Die Paketliste nennt Skriptfehler nicht.
export type Aufmerksamkeitszeile =
  | { art: 'hinweis'; hinweis: HinweisDTO }
  | { art: 'pakete'; hinweise: readonly HinweisDTO[] }
  | { art: 'setbefund'; befund: Setbefund }

// **Die Sets mit Befunden zuerst** (AP-8, Teil 2b) — sie kommen aus den
// Prüfständen der Kacheln, nicht aus der Antwort der Aufmerksamkeit, und nur
// aus aktuell geprüften.
export function aufmerksamkeitszeilen(
  hinweise: readonly HinweisDTO[],
  setbefunde: readonly Setbefund[] = [],
): Aufmerksamkeitszeile[] {
  const pakete = hinweise.filter((hinweis) => hinweis.objekt.art === 'paket')
  const zeilen: Aufmerksamkeitszeile[] = setbefunde.map((befund) => ({
    art: 'setbefund',
    befund,
  }))

  for (const hinweis of hinweise) {
    if (hinweis.objekt.art !== 'paket') {
      zeilen.push({ art: 'hinweis', hinweis })
    } else if (hinweis === pakete[0]) {
      zeilen.push({ art: 'pakete', hinweise: pakete })
    }
  }

  return zeilen
}

// grundsatz gibt den Grund eines Hinweises in der Sprache des Fensters.
//
// **Zwei Kataloge**: der eigene unter `gruende` — der Grund an einem Set —
// und der der Paketprüfung über [befundsatz]. Kennt keiner den Schlüssel,
// oder gibt es keinen, bleibt der Satz aus Go stehen — eine Karte ist ein
// Name und kein Satz, und ein unbekannter Schlüssel ist besser unübersetzt
// als leer (§9).
export function grundsatz(hinweis: HinweisDTO, texte: Katalog): string {
  const eigene = texte.seiten.start.aufmerksamkeit
  const eigener = new Map(Object.entries(eigene.gruende)).get(
    hinweis.schluessel,
  )
  const satz =
    eigener === undefined
      ? befundsatz(hinweis.schluessel, hinweis.werte, texte)
      : einsetzen(
          eigener,
          Object.fromEntries(
            hinweis.werte.map((wert, stelle) => [stelle, wert]),
          ),
        )

  if (satz === undefined) {
    return hinweis.grund
  }

  return hinweis.ort === ''
    ? satz
    : einsetzen(eigene.amOrt, { ort: hinweis.ort, satz })
}
