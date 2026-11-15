// **Die eingereichten Listen eines Turniers** — DOM-04.AK-02
// (#1699).
//
// **Eine eigene Lasche, wie im Artboard *Turniere*** (Runde, Rangliste,
// Teilnehmende, Decklisten). Bis hierher stand der Abzug in der Zeile jedes
// Teilnehmers; wer Listen prüft, prüft sie nacheinander, und wer jemanden
// aussteigen lässt, braucht sie nicht.
//
// **Die Liste ist ein Abzug und das Deck ihre Herkunft** (Vermerk 0062,
// Entscheidung 3): *Deck aufschlagen* führt zu dem Deck, wie es **heute** ist
// — die Liste darüber bleibt, wie sie eingereicht wurde.
//
// **Die Knöpfe stehen an einer Zeile** — seit UXR-12
// (#1920),
// wie in der Lasche *Teilnehmende*: an der Zeile mit der Tabstelle und an
// der unter dem Zeiger. Acht eingereichte Listen trugen sechzehn Knöpfe
// (R2 in `docs/UIREVIEW.md`). An den übrigen Zeilen steht die Kartenzahl
// als Text — eine Angabe, und ein Knopf erst dort, wo man ist.
import { useId, useState } from 'react'

import { Knopf } from './Knopf'
import { Turnierabzug, kartenzahl } from './Turnierabzug'
import { useRovierendeZeile } from '../haken/useRovierendeZeile'
import { gitterzeilentaste } from '../seiten/gitterzeilentaste'
import { useAnschluss } from '../anschlussraum'
import { useTexte } from '../sprache/kontext'
import type { TurnierDTO } from '../apiturniere'

export function Turnierdecklisten({ turnier }: { turnier: TurnierDTO }) {
  const eigene = useTexte().seiten.turniere
  const anschluss = useAnschluss()
  const ueberschrift = useId()

  const mitListe = turnier.teilnehmer.filter(
    (teilnehmer) => teilnehmer.deckabzug.length > 0 || teilnehmer.deckID !== '',
  )

  const rovierend = useRovierendeZeile(mitListe.length)
  const [zeiger, setzeZeiger] = useState<number>()

  return (
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-3">
      <h3 id={ueberschrift} className="font-semibold">
        {eigene.gebietDecklisten}
      </h3>

      {mitListe.length === 0 ? (
        <p>{eigene.decklistenLeer}</p>
      ) : (
        <ul
          aria-label={eigene.gebietDecklisten}
          className="flex flex-col gap-1"
        >
          {mitListe.map((teilnehmer, stelle) => {
            const tabstelle = rovierend.annehmen(stelle)
            const bedienbar = tabstelle || zeiger === stelle

            return (
              // **Die Zeile ist die Tabstelle einer wandernden Liste** —
              // dieselbe Ausnahme und derselbe Grund wie in
              // `Turnierteilnehmer`.
              // eslint-disable-next-line jsx-a11y/no-noninteractive-element-interactions
              <li
                key={teilnehmer.id}
                // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex -- die Tabstelle der wandernden Liste, wie oben begründet
                tabIndex={tabstelle ? 0 : -1}
                onFocus={() => {
                  rovierend.rueckenAuf(stelle)
                }}
                onKeyDown={(ereignis) => {
                  gitterzeilentaste(
                    ereignis,
                    stelle,
                    mitListe.length,
                    rovierend.rueckenAuf,
                  )
                }}
                onPointerEnter={() => {
                  setzeZeiger(stelle)
                }}
                onPointerLeave={() => {
                  setzeZeiger(undefined)
                }}
                className="flex min-h-8 flex-wrap items-center gap-3 rounded-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              >
                <span className="min-w-24">{teilnehmer.name}</span>

                {bedienbar ? (
                  <Turnierabzug
                    abzug={teilnehmer.deckabzug}
                    wer={teilnehmer.name}
                  />
                ) : (
                  teilnehmer.deckabzug.length > 0 && (
                    <span className="tabular-nums text-leise">
                      {kartenzahl(teilnehmer.deckabzug)}
                    </span>
                  )
                )}

                {bedienbar && teilnehmer.deckID !== '' && (
                  <Knopf
                    gewicht="leise"
                    bezeichnung={`${eigene.deckAufschlagen}: ${teilnehmer.name}`}
                    aufKlick={() => {
                      anschluss.uebergib({
                        art: 'deckansicht',
                        name: teilnehmer.name,
                        deckID: teilnehmer.deckID,
                      })
                    }}
                  >
                    {eigene.deckAufschlagen}
                  </Knopf>
                )}
              </li>
            )
          })}
        </ul>
      )}
    </section>
  )
}
