// **Wie die Partie ausgegangen ist** — Werkstück M17 der Regelmaschine
// (#1618).
//
// # Warum das Ergebnis oben steht und nicht unten
//
// **Weil es die Antwort auf die einzige Frage ist, die dann noch zählt.** Eine
// beendete Partie endet **sofort** (CR 104.1); was auf dem Tisch liegt, ist
// danach Geschichte, und wer erst scrollen muss, um zu erfahren, dass er
// gewonnen hat, liest ein Protokoll und kein Ergebnis.
//
// # Und warum der Grund dabeisteht
//
// **Weil `beendet` nicht sagt, warum.** Aufgegeben, bei null Leben, aus einer
// leeren Bibliothek gezogen oder vergiftet sind vier verschiedene Ausgänge —
// und für das Turnier aus G7 sind es vier verschiedene Auskünfte.
import { useTexte } from '../sprache/kontext'

import type { Katalog } from '../sprache/kataloge'
import type { Partie } from '../apipartie'

export function Partieausgang({ partie }: { partie: Partie }) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const { ausgang } = partie

  // **Solange niemand gewonnen hat und es nicht unentschieden ist, läuft
  // sie** — und dann gibt es nichts zu melden.
  if (ausgang.sieger === '' && !ausgang.unentschieden) {
    return undefined
  }

  return (
    <section
      role="status"
      className="flex flex-col gap-2 rounded-sm border border-kante px-3 py-2"
      aria-label={eigene.ausgangTitel}
    >
      <p className="text-base font-semibold">
        {ausgang.unentschieden
          ? eigene.ausgangUnentschieden
          : `${eigene.ausgangSieger}: ${namenwort(partie, ausgang.sieger)}`}
      </p>

      <ul className="flex flex-col gap-1 text-sm text-leise">
        {ausgang.ausgeschieden.map((einer) => (
          <li key={einer.sitz}>
            {namenwort(partie, einer.sitz)} — {grundwort(einer.grund, eigene)}
          </li>
        ))}
      </ul>
    </section>
  )
}

// namenwort gibt den Namen eines Sitzes — oder seine Kennung.
//
// **Die Kennung als Rückfall**: ein Sitz, den die Sicht nicht führt, ist kein
// Grund für eine leere Stelle (§8).
function namenwort(partie: Partie, kennung: string): string {
  const sitz = partie.sitze.find((einer) => einer.kennung === kennung)

  return sitz === undefined || sitz.name === '' ? kennung : sitz.name
}

// grundwort nennt den Ausscheidungsgrund in Worten — CR 104.3.
//
// **Aus dem Katalog und nicht aus der Kennung**: `bibliothek` ist ein
// Bezeichner und kein Satz, und eine englische Sitzung bekäme ihn sonst
// unübersetzt (§9).
function grundwort(
  grund: string,
  eigene: Katalog['seiten']['spielen'],
): string {
  const gruende: Record<string, string> = {
    aufgabe: eigene.grundAufgabe,
    leben: eigene.grundLeben,
    bibliothek: eigene.grundBibliothek,
    gift: eigene.grundGift,
  }

  return gruende[grund] ?? grund
}
