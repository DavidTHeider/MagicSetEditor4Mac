// **Der Editor einer eigenen Übersetzung** — LOC-02.AK-01
// (#1688).
//
// Je Satz stehen **Pfad, deutscher Satz und Basis** über dem Feld: der Pfad
// sagt, wo er erscheint, der deutsche, was er meint, die Basis, was jetzt
// dasteht. Fehlend und veraltet sagt Go; gezeigt wird eine Seite davon
// (`uebersetzungsauswahl.ts`). **Der Befund steht am Satz**, an dem er
// entstand (§6).
import { useMemo, useState } from 'react'

import { flach } from '../sprache/flach'
import { einsetzen } from '../sprache/einsetzen'
import { istSprache } from '../sprache/einstellung'
import { kataloge } from '../sprache/kataloge'
import { useTexte } from '../sprache/kontext'
import { ohneAngaben, useUebersetzung } from '../haken/useUebersetzung'
import { auswahl, type Uebersetzungsfilter } from './uebersetzungsauswahl'
import { Knopf } from './Knopf'
import { Uebersetzungsangaben } from './Uebersetzungsangaben'
import { Uebersetzungszeile } from './Uebersetzungszeile'

// filter sind die vier Auswahlen, in der Reihenfolge der Arbeit.
const filter: readonly Uebersetzungsfilter[] = [
  'fehlend',
  'veraltet',
  'uebersetzt',
  'alle',
]

export function Uebersetzungseditor({
  kennung,
  zurueck,
}: {
  kennung: string
  zurueck: () => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen.eigenuebersetzung
  const uebersetzung = useUebersetzung(kennung)
  const [gewaehlt, setzeGewaehlt] = useState<Uebersetzungsfilter>('fehlend')
  const [suche, setzeSuche] = useState('')
  const [seite, setzeSeite] = useState(1)

  const stand = uebersetzung.stand
  const basis =
    stand !== undefined && istSprache(stand.paket.basis)
      ? stand.paket.basis
      : 'de'
  const vorlage = useMemo(() => flach(kataloge.de), [])
  const basistexte = useMemo(() => flach(kataloge[basis]), [basis])

  const aktuell =
    stand === undefined
      ? undefined
      : auswahl(vorlage, stand, gewaehlt, suche, seite)

  const namen: Record<Uebersetzungsfilter, string> = {
    fehlend: eigene.filterFehlend,
    veraltet: eigene.filterVeraltet,
    uebersetzt: eigene.filterUebersetzt,
    alle: eigene.filterAlle,
  }

  return (
    <section aria-label={kennung} className="flex max-w-4xl flex-col gap-4">
      <div className="flex flex-wrap items-center gap-3">
        <Knopf aufKlick={zurueck}>{eigene.zurueck}</Knopf>
        <h3 className="font-medium">
          {stand === undefined ? kennung : `${stand.paket.name} (${kennung})`}
        </h3>
      </div>

      {uebersetzung.ladefehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {uebersetzung.ladefehler}
        </p>
      )}

      {stand !== undefined && (
        <>
          <Uebersetzungsangaben
            paket={stand.paket}
            angehalten={uebersetzung.laeuft}
            befund={uebersetzung.befund}
            sichern={(angaben) => void uebersetzung.aendere(angaben, {})}
          />

          <div className="flex flex-wrap items-end gap-3">
            <div role="group" aria-label={eigene.filter} className="flex gap-1">
              {filter.map((art) => (
                <Knopf
                  key={art}
                  klein
                  gedrueckt={gewaehlt === art}
                  aufKlick={() => {
                    setzeGewaehlt(art)
                    setzeSeite(1)
                  }}
                >
                  {namen[art]}
                </Knopf>
              ))}
            </div>
            <label className="flex flex-col gap-1 text-sm">
              {eigene.suche}
              <input
                type="search"
                value={suche}
                onChange={(ereignis) => {
                  setzeSuche(ereignis.target.value)
                  setzeSeite(1)
                }}
                className="rounded-md border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              />
            </label>
          </div>

          {aktuell?.pfade.length === 0 && (
            <p className="text-sm text-leise">{eigene.keinTreffer}</p>
          )}

          <ul className="flex flex-col divide-y divide-kante">
            {aktuell?.pfade.map((pfad) => (
              <Uebersetzungszeile
                key={pfad}
                pfad={pfad}
                deutsch={vorlage[pfad] ?? ''}
                basistext={basistexte[pfad] ?? ''}
                text={stand.texte[pfad]?.text}
                veraltet={stand.veraltetePfade.includes(pfad)}
                befund={
                  uebersetzung.befund?.feld === pfad
                    ? uebersetzung.befund.text
                    : undefined
                }
                angehalten={uebersetzung.laeuft}
                uebernehmen={(text) =>
                  void uebersetzung.aendere(ohneAngaben, { [pfad]: text })
                }
              />
            ))}
          </ul>

          {aktuell !== undefined && aktuell.seiten > 1 && (
            <nav aria-label={eigene.filter} className="flex items-center gap-3">
              <Knopf
                angehalten={aktuell.seite === 1}
                aufKlick={() => {
                  setzeSeite(aktuell.seite - 1)
                }}
              >
                {eigene.zurueckBlaettern}
              </Knopf>
              <span className="text-sm text-leise">
                {einsetzen(eigene.seite, {
                  seite: String(aktuell.seite),
                  seiten: String(aktuell.seiten),
                })}
              </span>
              <Knopf
                angehalten={aktuell.seite === aktuell.seiten}
                aufKlick={() => {
                  setzeSeite(aktuell.seite + 1)
                }}
              >
                {eigene.vor}
              </Knopf>
            </nav>
          )}
        </>
      )}
    </section>
  )
}
