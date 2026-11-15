// **Ein Gegenüber bestätigen** — NET-01
// (#1705),
// nach dem Artboard *Gegenüber*.
//
// # Warum „stimmt überein" ein Knopf ist und kein Feld
//
// **Weil beide Seiten dieselben Ziffern sehen.** Beim Fingerabdruck liest
// jeder den eigenen vor, und der andere tippt ab; der Vergleichscode gehört
// dem Paar, und verglichen wird laut. Der Knopf schickt, was angezeigt war —
// der Dienst hält es gegen seine eigene Rechnung.
//
// # Warum das Häkchen nicht gesetzt ist
//
// **Weil ein Zweck eine eigene Entscheidung ist** — „Bestätigung des
// Gegenübers ersetzt nicht die Zustimmung" (AK-04), und ein vorgesetztes
// Häkchen wäre eine Zustimmung, die niemand gegeben hat.
import { useId, useState } from 'react'

import { Aufklappkasten } from './Aufklappkasten'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Knopf } from './Knopf'
import { eingabeklasse } from './Beschriftetesfeld'
import { alsZeitpunkt } from '../zeitpunkt'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte, useZeitzone } from '../sprache/kontext'
import { zwecksatz } from '../zwecksatz'

import type { Einladungsvorschau } from '../apikontakteinladung'
import type { Einladungstausch } from '../haken/useEinladungstausch'

interface Einladungsvergleichseigenschaften {
  vorschau: Einladungsvorschau
  tausch: Einladungstausch
}

export function Einladungsvergleich({
  vorschau,
  tausch,
}: Einladungsvergleichseigenschaften) {
  const texte = useTexte()
  const technik = texte.seiten.technik
  const eigene = technik.einladung
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()
  const titel = useId()
  const namensfeld = useId()
  const meldefeld = useId()

  const [eigenerName, setzeEigenerName] = useState(vorschau.name)

  // **Leer und nicht vorausgefüllt** (NET-01b): dieser Name geht hinaus,
  // und ein vorausgefüllter wäre einer, den niemand gewählt hat.
  const [meinName, setzeMeinName] = useState('')
  const [freigeben, setzeFreigeben] = useState(false)

  const gesperrt = vorschau.lage === 'widerrufen'
  const zu = gesperrt || vorschau.angenommen
  const laeuft = tausch.laeuft('annehmen')
  const fehler =
    tausch.fehler?.ort === 'annehmen' ? tausch.fehler.text : undefined

  return (
    <section
      aria-labelledby={titel}
      data-einladungsvergleich=""
      className="flex flex-col gap-3 rounded-sm border border-kante p-3"
    >
      <h4 id={titel} className="font-semibold">
        {/* **Nur der Name, wenn es nichts zu bestätigen gibt** — „Bert
            bestätigen" über „schon angenommen" widerspräche sich (gesehen
            an der Hüllenprobe `net01`). */}
        {zu
          ? vorschau.name
          : einsetzen(eigene.bestaetigenTitel, { name: vorschau.name })}
      </h4>

      {vorschau.angenommen && (
        <p role="alert" className="text-sm text-fehler">
          {eigene.angenommen}
        </p>
      )}

      {gesperrt && (
        <p role="alert" className="text-sm text-fehler">
          {eigene.gesperrt}
        </p>
      )}

      {!zu && (
        <>
          {vorschau.vorgaenger !== '' && (
            <p className="text-sm text-fehler">
              {einsetzen(eigene.vorgaenger, { name: vorschau.vorgaenger })}
            </p>
          )}

          {vorschau.lage === 'bestaetigt' && (
            <p className="text-sm text-leise">{eigene.bekannt}</p>
          )}

          <p className="max-w-prose text-sm">{eigene.vergleichen}</p>
          <p
            data-vergleichscode=""
            className="font-mono text-lg tracking-wider"
          >
            {vorschau.vergleichscode}
          </p>
          <Erklaerungsabsatz
            text={eigene.vergleichenHinweis}
            klasse="text-sm text-leise"
          />

          <label className="flex flex-col gap-1 text-sm" htmlFor={namensfeld}>
            {eigene.eigenerName}
            <input
              id={namensfeld}
              className={eingabeklasse}
              value={eigenerName}
              onChange={(ereignis) => {
                setzeEigenerName(ereignis.target.value)
              }}
            />
          </label>
          <Erklaerungsabsatz
            text={eigene.eigenerNameHinweis}
            klasse="text-sm text-leise"
          />

          {vorschau.einloesbar && (
            <>
              <label
                className="flex flex-col gap-1 text-sm"
                htmlFor={meldefeld}
              >
                {eigene.meinName}
                <input
                  id={meldefeld}
                  className={eingabeklasse}
                  value={meinName}
                  onChange={(ereignis) => {
                    setzeMeinName(ereignis.target.value)
                  }}
                />
              </label>
              <Erklaerungsabsatz
                text={einsetzen(eigene.meinNameHinweis, {
                  name: vorschau.name,
                })}
                klasse="text-sm text-leise"
              />
            </>
          )}

          {vorschau.zweck !== 'paarung' && (
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                className="focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                checked={freigeben}
                onChange={(ereignis) => {
                  setzeFreigeben(ereignis.target.checked)
                }}
              />
              {einsetzen(eigene.freigeben, {
                zweck: zwecksatz(technik, vorschau.zweck),
              })}
            </label>
          )}

          <Erklaerungsabsatz
            text={eigene.nichtsGeteilt}
            klasse="max-w-prose text-sm text-leise"
          />

          <Aufklappkasten titel={eigene.technik}>
            <dl className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-sm">
              <dt className="font-medium">{eigene.technikFingerabdruck}</dt>
              <dd className="font-mono">{vorschau.fingerabdruck}</dd>
              <dt className="font-medium">{eigene.technikAdresse}</dt>
              <dd className="break-all font-mono">
                {vorschau.adresse === ''
                  ? eigene.technikOhneAdresse
                  : vorschau.adresse}
              </dd>
              <dt className="font-medium">{eigene.technikAblauf}</dt>
              <dd>{alsZeitpunkt(vorschau.ablauf, darstellung, zeitzone)}</dd>
            </dl>
          </Aufklappkasten>
        </>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      <div className="flex flex-wrap gap-2">
        <Knopf gewicht="leise" angehalten={laeuft} aufKlick={tausch.verwerfen}>
          {eigene.abbrechen}
        </Knopf>

        {!zu && (
          <>
            <Knopf
              angehalten={laeuft}
              aufKlick={() => {
                tausch.annehmen(vorschau, eigenerName, false, false, meinName)
              }}
            >
              {eigene.spaeter}
            </Knopf>
            <Knopf
              gewicht="haupt"
              angehalten={laeuft}
              aufKlick={() => {
                tausch.annehmen(
                  vorschau,
                  eigenerName,
                  true,
                  freigeben,
                  meinName,
                )
              }}
            >
              {eigene.stimmt}
            </Knopf>
          </>
        )}
      </div>
    </section>
  )
}
