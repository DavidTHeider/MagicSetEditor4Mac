// **Welcher Handel** — UX-08.AK-04
// (#1680):
// *„Normalnutzende müssen keine Netzwerk- oder Transaktionskennungen
// eingeben, soweit diese technisch ableitbar sind; Expertendetails bleiben mit
// verständlicher Erläuterung erreichbar."*
//
// # Was vorher war
//
// **Vier Felder „Welcher Handel"** — im Vorgang, bei den Zusagen, im Ausgang
// und bei der Buchung —, und in jedes gehörte dieselbe Kennung von Hand.
// **Wer einen Handel empfing, erfuhr sie nirgends im Fenster.**
//
// # Was jetzt ist
//
// **Eine Wahl über allen vier**: die Handel mit diesem Gegenüber, je mit
// Zustand, Revision und Zahl der Positionen, und „ein neuer Handel", dessen
// Kennung die Leitung beim Eröffnen vergibt. **Die Kennung selbst steht unter
// „Für Fachleute"** — mit dem Satz, was sie ist, und einem Feld, über das
// sich ein Handel öffnen lässt, der hier noch nicht steht.
import { useId, useState } from 'react'

import { Aufklappkasten } from './Aufklappkasten'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Knopf } from './Knopf'
import { Ladefehler, Wirdgeladen } from './Ladehinweise'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'
import { standsatz } from './handelsstaende'

import type { Handelsvorgang } from '../apihandelsleitung'
import type { Ladezustand } from '../haken/useLaden'
import { kennungseingabe } from './kennungseingabe'

// wahlstil ist ein Optionsfeld — **mit sichtbarem Fokusring** (§6).
const wahlstil =
  'mt-1 accent-marke focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

const feldstil =
  'rounded-sm border border-kante bg-flaeche px-2 py-1 font-mono text-sm ' +
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

interface Handelswahleigenschaften {
  // vorgaenge sind die Handel mit dem Gegenüber — geladen im Handelsraum,
  // der sie nach einer Änderung nachlädt.
  vorgaenge: Ladezustand<Handelsvorgang[]>

  // gewaehlt ist die Kennung des gewählten Handels — leer heißt: ein neuer.
  gewaehlt: string
  waehlen: (vorgang: string) => void
}

export function Handelswahl({
  vorgaenge,
  gewaehlt,
  waehlen,
}: Handelswahleigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.handeln
  const gruppe = useId()
  const feld = useId()

  const [eingabe, setzeEingabe] = useState('')

  return (
    <div className="flex flex-col gap-2">
      <fieldset className="flex flex-col gap-1">
        <legend className="mb-1 text-sm font-medium">{eigene.wahlTitel}</legend>

        <label className="flex cursor-pointer items-start gap-2 rounded-sm border border-kante px-2 py-1.5 has-[:checked]:border-marke">
          <input
            type="radio"
            name={gruppe}
            value=""
            checked={gewaehlt === ''}
            onChange={() => {
              waehlen('')
            }}
            className={wahlstil}
          />
          <span className="flex flex-col font-medium">
            {eigene.wahlNeu}
            <span className="text-xs font-normal text-leise">
              {eigene.wahlNeuZusatz}
            </span>
          </span>
        </label>

        {vorgaenge.daten === undefined ? (
          vorgaenge.fehler === undefined ? (
            <Wirdgeladen />
          ) : (
            <Ladefehler
              fehler={vorgaenge.fehler}
              wiederholen={vorgaenge.nachladen}
            />
          )
        ) : vorgaenge.daten.length === 0 ? (
          <p className="text-sm text-leise">{eigene.wahlLeer}</p>
        ) : (
          vorgaenge.daten.map((vorgang) => (
            <label
              key={vorgang.vorgang}
              className="flex cursor-pointer items-start gap-2 rounded-sm border border-kante px-2 py-1.5 has-[:checked]:border-marke"
            >
              <input
                type="radio"
                name={gruppe}
                value={vorgang.vorgang}
                checked={gewaehlt === vorgang.vorgang}
                onChange={() => {
                  waehlen(vorgang.vorgang)
                }}
                className={wahlstil}
              />
              <span>
                {einsetzen(eigene.wahlEintrag, {
                  zustand: standsatz(eigene, vorgang.zustand),
                  revision: String(vorgang.revision),
                  positionen: zaehlwort(
                    vorgang.positionen.length,
                    texte.allgemein.zaehlwoerter.positionen,
                  ),
                })}
              </span>
            </label>
          ))
        )}
      </fieldset>

      <Aufklappkasten titel={eigene.wahlFachleute}>
        <div className="flex flex-col gap-2">
          <Erklaerungsabsatz
            text={eigene.wahlFachleuteErklaerung}
            klasse="max-w-prose text-sm text-leise"
          />

          {gewaehlt !== '' && (
            <p className="text-sm">
              {einsetzen(eigene.wahlGewaehlt, { kennung: gewaehlt })}
            </p>
          )}

          <div className="flex flex-wrap items-end gap-2">
            <label className="flex flex-col gap-1 text-sm" htmlFor={feld}>
              {eigene.wahlKennung}
              <input
                {...kennungseingabe}
                id={feld}
                type="text"
                className={feldstil}
                value={eingabe}
                onChange={(ereignis) => {
                  setzeEingabe(ereignis.target.value)
                }}
              />
            </label>

            <Knopf
              angehalten={eingabe.trim() === ''}
              aufKlick={() => {
                waehlen(eingabe.trim())
              }}
            >
              {eigene.wahlOeffnen}
            </Knopf>
          </div>
        </div>
      </Aufklappkasten>
    </div>
  )
}
