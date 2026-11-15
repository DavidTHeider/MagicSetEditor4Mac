// **Was aus der Sammlung als Stapel kam, legt ein Knopf ins Deck** —
// UX-03.AK-01 (#1675).
//
// # Warum ein Knopf und nicht gleich
//
// **Weil das Deck gewählt sein will.** Der Stapel reist zu den Decks; in
// welches er gehört, entscheidet, wer dort eines aufschlägt. Erst dann steht
// dieser Kasten da, und erst der Knopf legt.
//
// # Und was ein Teilerfolg sagt
//
// **Beides** (S08.NAT-03): was liegt, und warum der Rest nicht liegt — **am
// Stapel**, denn hier wurde gelegt (§6). Die erste Fassung verwies auf „die
// Meldung darüber", und die WebKit-Probe fand dort keine. **Und es bleibt nur
// der Rest**: ein zweiter Klick legte sonst die schon gelegten Drucke noch
// einmal.
//
// **Der Rest ist keine Folge mehr, sondern eine Auswahl** — seit ein
// abgewiesenes Stück die übrigen nicht mehr aufhält (Befund G16, #1666).
// Deshalb trägt jeder Grund den Namen seines Stücks: zwei abgewiesene ohne
// Namen sagten nicht, welches woran scheiterte.
import { useState } from 'react'

import { Knopf } from './Knopf'
import { useAnschluss, useAufgenommen } from '../anschlussraum'
import { Bindungsfehler } from '../bindungsfehler'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Stueck } from '../anschluss'
import type { Stapelergebnis } from '../haken/useDeckaktionen'
import type { Katalog } from '../sprache/kataloge'

// Lage ist, was von einem Stapel schon liegt, was noch nicht, und warum.
interface Lage {
  gelegt: number

  // offen ist, was noch nicht liegt — `undefined`, solange nichts versucht
  // wurde: dann ist es der ganze Stapel.
  offen?: readonly Stueck[] | undefined

  gruende: readonly string[]
}

// gruendeAus macht aus einem Ergebnis die Sätze am Stapel — je abgewiesenem
// Stück einen mit seinem Namen, und den, der den Stapel anhielt.
//
// **Ohne Kennung nennt der Grund den Druck** — die Domäne weist ihn am Feld
// `printing` ab, und ein Feld gibt es hier nicht; der allgemeine Satz
// verwiese auf keines.
function gruendeAus(ergebnis: Stapelergebnis, texte: Katalog): string[] {
  const eigene = texte.seiten.decks
  const einzeln = ergebnis.abgewiesen.map(({ stueck, ursache }) =>
    ursache instanceof Bindungsfehler && ursache.feld === 'printing'
      ? einsetzen(eigene.stapelOhneDruck, { name: stueck.name })
      : einsetzen(eigene.stapelGrund, {
          name: stueck.name,
          grund: meldungAus(ursache, texte),
        }),
  )

  // **Doppelte fallen zusammen**: zwei gleiche Stücke mit demselben Grund
  // sind ein Satz und nicht zwei.
  return [
    ...new Set(
      ergebnis.ursache === undefined
        ? einzeln
        : [...einzeln, meldungAus(ergebnis.ursache, texte)],
    ),
  ]
}

export function Deckstapelaufnahme({
  legeAlle,
  angehalten,
}: {
  // legeAlle legt die Stücke der Reihe nach und sagt, wie viele liegen.
  legeAlle: (stuecke: readonly Stueck[]) => Promise<Stapelergebnis>
  angehalten: boolean
}) {
  const texte = useTexte()
  const eigene = texte.seiten.decks
  const anschluss = useAnschluss()
  const [lage, setzeLage] = useState<Lage>({ gelegt: 0, gruende: [] })

  const stapel =
    anschluss.offen?.art === 'deckstapel' ? anschluss.offen : undefined

  // **Ein neuer Stapel beginnt von vorn.**
  useAufgenommen(stapel, () => {
    setzeLage({ gelegt: 0, gruende: [] })
  })

  if (stapel === undefined) {
    return null
  }

  const rest = lage.offen ?? stapel.stuecke
  const liste = rest
    .map((stueck) => `${stueck.name} × ${String(stueck.menge)}`)
    .join(', ')

  return (
    <div
      role="group"
      aria-label={eigene.stapelTitel}
      className="flex flex-col gap-2 rounded-sm border border-kante p-3"
    >
      <p className="text-sm">
        {einsetzen(eigene.stapelWartet, { name: stapel.name, liste })}
      </p>

      {lage.gruende.length > 0 && (
        <>
          <p className="text-sm">
            {einsetzen(eigene.stapelTeilweise, {
              gelegt: String(lage.gelegt),
              gesamt: String(stapel.stuecke.length),
            })}
          </p>
          <div role="alert" className="flex flex-col gap-1 text-sm text-fehler">
            {lage.gruende.map((grund) => (
              <p key={grund}>{grund}</p>
            ))}
          </div>
        </>
      )}

      <div>
        <Knopf
          gewicht="haupt"
          angehalten={angehalten}
          aufKlick={() => {
            void legeAlle(rest).then((ergebnis) => {
              if (ergebnis.offen.length === 0) {
                anschluss.verbrauche()

                return
              }

              setzeLage({
                gelegt: lage.gelegt + ergebnis.gelegt,
                offen: ergebnis.offen,
                gruende: gruendeAus(ergebnis, texte),
              })
            })
          }}
        >
          {eigene.stapelLegen}
        </Knopf>
      </div>
    </div>
  )
}
