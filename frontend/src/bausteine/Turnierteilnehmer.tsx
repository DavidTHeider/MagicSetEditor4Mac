// **Die Teilnehmer eines Turniers** — Werkstück 5 von Phase 4.7
// (#1294).
//
// # Der Knopf steht an einer Zeile, nicht an jeder
//
// **Seit UXR-12** (#1920).
// Sechzehn Teilnehmende trugen sechzehn Knöpfe *Abmelden* — die Werkbank
// damit 25 Bedienelemente (R2 in `docs/UIREVIEW.md`), und zum Anmelden
// darunter sechzehn Tabstopps. Jetzt ist die Liste **eine** Tabstelle, die
// Pfeile wandern, und der Knopf steht an der Zeile mit der Tabstelle und an
// der unter dem Zeiger — dasselbe Muster wie die Plätze eines Decks (UXR-11)
// und die Paketliste (UXR-17). → führt in die Zeile, ← heraus.
import { useId, useState } from 'react'

import { api } from '../api'
import { auspacken } from '../umschlag'
import { Auswahlfeld } from './Auswahlfeld'
import { Knopf } from './Knopf'
import { useRovierendeZeile } from '../haken/useRovierendeZeile'
import { gitterzeilentaste } from '../seiten/gitterzeilentaste'
import { useTexte } from '../sprache/kontext'
import { useLaden } from '../haken/useLaden'
import type { DeckDTO } from '../apidecks'
import type { TurnierDTO } from '../apiturniere'
import type { Turnierstand } from '../haken/useTurniere'
import { Pflichtfeld } from './Pflichtfeld'

const feldklassen =
  'rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Turnierteilnehmer({
  stand,
  turnier,
  nurLesen = false,
}: {
  stand: Turnierstand
  turnier: TurnierDTO

  // nurLesen heißt: das Turnier ist abgeschlossen — niemand meldet an oder
  // ab (DOM-04.AK-04, #1699).
  nurLesen?: boolean
}) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere

  const [name, setzeName] = useState('')
  const [deck, setzeDeck] = useState('')

  // **Die Decks für die Wahl** — dieselbe Bindung, die die Deckseite nimmt.
  // Ein Turnier gehört zu einem Spiel, und ein Deck auch; gezeigt werden
  // deshalb nur die des Spiels.
  const decks = useLaden<DeckDTO[]>(async () => {
    const antwort = await api().Decks(turnier.spiel)

    return auspacken(antwort, () => antwort.decks)
  }, [turnier.spiel])

  const anmelden = async () => {
    await stand.fuehre(`turnier:${turnier.id}`, async () => {
      const antwort = await api().TurnierAnmelden(turnier.id, name, deck)

      auspacken(antwort, () => antwort.teilnehmer)

      setzeName('')
      setzeDeck('')
    })
  }

  const umschalten = async (teilnehmerID: string, ausgestiegen: boolean) => {
    await stand.fuehre(`turnier:${turnier.id}`, async () => {
      auspacken(
        await api().TurnierAussteigen(turnier.id, teilnehmerID, ausgestiegen),
        () => undefined,
      )
    })
  }

  const laeuft = stand.vorgang.laeuft(`turnier:${turnier.id}`)
  const ueberschrift = useId()
  const deckhinweis = useId()
  const deckwahl = useId()

  const anzahl = turnier.teilnehmer.length
  const rovierend = useRovierendeZeile(anzahl)
  const [zeiger, setzeZeiger] = useState<number>()

  return (
    // **Eine `section` mit Namen und kein `div`** — Werkstück I von Phase
    // 4.10. Auf der Turnierseite steht „Name" zweimal: im Kasten, mit dem ein
    // Turnier angelegt wird, und hier. Ohne Landmarke muss jede Bedienhilfe —
    // und jeder Test — raten, welches gemeint ist.
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-3">
      <h3 id={ueberschrift} className="font-semibold">
        {eigene.teilnehmer}
      </h3>

      {turnier.teilnehmer.length === 0 ? (
        <p>{eigene.teilnehmerLeer}</p>
      ) : (
        <ul aria-label={eigene.teilnehmer} className="flex flex-col gap-1">
          {turnier.teilnehmer.map((teilnehmer, stelle) => {
            const tabstelle = rovierend.annehmen(stelle)

            return (
              // **Die Zeile ist die Tabstelle einer wandernden Liste**
              // (`useRovierendeZeile`) — `jsx-a11y` sieht ein `li` mit
              // Horchern und weiß nicht, dass es fokussierbar ist und die
              // Pfeile hört. Die Plätze eines Decks bekommen dasselbe über
              // `Kontextmenue`; hier steht die Ausnahme, damit sie begründet
              // ist.
              // eslint-disable-next-line jsx-a11y/no-noninteractive-element-interactions
              <li
                key={teilnehmer.id}
                // **Ein abgeschlossenes Turnier hat nichts an der Zeile** —
                // dann ist sie auch keine Tabstelle.
                tabIndex={nurLesen ? undefined : tabstelle ? 0 : -1}
                onFocus={() => {
                  rovierend.rueckenAuf(stelle)
                }}
                onKeyDown={(ereignis) => {
                  gitterzeilentaste(
                    ereignis,
                    stelle,
                    anzahl,
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
                {/* **Eine Mindestbreite, damit die Knöpfe fluchten.** Ohne
                    sie verschiebt jeder Name den Knopf dahinter, und acht
                    Teilnehmer lesen sich als Treppe — angestrichen in der
                    Sichtprüfung zu #1294. */}
                <span className="min-w-24">{teilnehmer.name}</span>

                {teilnehmer.ausgestiegen && (
                  <span className="text-sm text-leise">
                    {eigene.ausgestiegen}
                  </span>
                )}

                {!nurLesen && (tabstelle || zeiger === stelle) && (
                  <Knopf
                    klein
                    bezeichnung={`${
                      teilnehmer.ausgestiegen
                        ? eigene.einsteigen
                        : eigene.aussteigen
                    }: ${teilnehmer.name}`}
                    angehalten={laeuft}
                    aufKlick={() => {
                      void umschalten(teilnehmer.id, !teilnehmer.ausgestiegen)
                    }}
                  >
                    {teilnehmer.ausgestiegen
                      ? eigene.einsteigen
                      : eigene.aussteigen}
                  </Knopf>
                )}
              </li>
            )
          })}
        </ul>
      )}

      {!nurLesen && (
        <>
          <div className="flex flex-wrap items-end gap-3">
            {/* **Pflichtfeld mit Satz am Feld** — Werkstück F von Phase 4.10 (§6). */}
            <Pflichtfeld
              beschriftung={eigene.name}
              wert={name}
              geaendert={setzeName}
            />

            {/* **Das Deck wird gewählt, nicht getippt** — Werkstück I von Phase
            4.10
            (#1325).
            Die Kennung eines Decks steht nirgends in der Oberfläche. */}
            <div className="flex flex-col gap-1">
              <label htmlFor={deckwahl} className="text-sm">
                {eigene.deck}
              </label>

              {/* **Ab fünfzehn Decks mit Suche** — UXR-F1
                  (#2007),
                  R7. Die Beschriftung steht neben dem Feld und nicht um es. */}
              <Auswahlfeld
                id={deckwahl}
                bezeichnung={eigene.deck}
                eintraege={[
                  { wert: '', name: eigene.ohneDeck },
                  ...(decks.daten ?? []).map((eintrag) => ({
                    wert: eintrag.id,
                    name: eintrag.name,
                  })),
                ]}
                wert={deck}
                aendern={setzeDeck}
                beschriebenVon={deckhinweis}
                klasse={feldklassen}
              />
            </div>

            <Knopf
              angehalten={laeuft}
              aufKlick={() => {
                void anmelden()
              }}
            >
              {eigene.anmelden}
            </Knopf>
          </div>

          <p id={deckhinweis} className="text-sm text-leise">
            {eigene.deckHinweis}
          </p>
        </>
      )}
    </section>
  )
}
