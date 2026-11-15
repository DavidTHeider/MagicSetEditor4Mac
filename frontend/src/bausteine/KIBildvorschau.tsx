// **Was ein Modell vorgeschlagen hat, bevor es jemand nimmt** — Werkstück
// A12 von Phase 5.3
// (#1451).
//
// # Warum das Bild hier steht und nicht im Zuschnittfenster
//
// **Weil zwischen „da ist eines" und „so soll es liegen" eine Entscheidung
// liegt.** Wer den Vorschlag nicht will, verwirft ihn hier und sieht das
// Zuschnittfenster gar nicht erst.
//
// # Und die Herkunft steht dabei
//
// **Was der Dienst gesagt hat, steht da; was er nicht gesagt hat, steht
// nicht da.** Die Ergänzung des Issues verlangt es: „Fehlende Angaben als
// unbekannt führen." Eine Zeile `seed: —` behauptete, es gäbe einen.
import { Fragment } from 'react'

import { beschriftung } from './kibildangaben'
import { einsetzen } from '../sprache/einsetzen'

import type { KIBildvorschlagAntwort } from '../antworten'
import type { kataloge } from '../sprache/kataloge'

interface KIBildvorschauEigenschaften {
  vorschlag: KIBildvorschlagAntwort

  eigene: (typeof kataloge)['de']['seiten']['einstellungen']

  titel: string

  laeuft: boolean

  // zuschneiden öffnet das vorhandene Zuschnittfenster.
  zuschneiden: () => void

  verwerfen: () => void
}

export function KIBildvorschau({
  vorschlag,
  eigene,
  titel,
  laeuft,
  zuschneiden,
  verwerfen,
}: KIBildvorschauEigenschaften) {
  return (
    <figure className="flex flex-col gap-2 rounded-sm border border-kante p-2">
      <figcaption className="text-sm font-medium">{titel}</figcaption>

      {/* **Das Bild und nicht eine Verkleinerung davon**: was hier steht, ist
          genau das, was übernommen wird. */}
      <img
        className="max-h-64 self-start rounded-sm"
        src={vorschlag.vorschau}
        alt={titel}
      />

      <p className="text-xs text-leise">
        {einsetzen(eigene.kiBildMasse, {
          breite: String(vorschlag.breite),
          hoehe: String(vorschlag.hoehe),
          kilobytes: String(Math.round(vorschlag.bytes / 1024)),
        })}
      </p>

      <p className="text-xs text-leise">
        {einsetzen(eigene.kiBildAblauf, { ablauf: vorschlag.ablauf })}
      </p>

      {/* **Die Umschrift nur, wenn es eine gibt.** Ein leerer Satz „Gezeichnet
          wurde:" wäre eine Auskunft über nichts. */}
      {vorschlag.umschrift !== '' && (
        <p className="max-w-prose text-xs text-leise">
          {einsetzen(eigene.kiBildUmschrift, {
            umschrift: vorschlag.umschrift,
          })}
        </p>
      )}

      {/* **Eine Beschreibungsliste und keine Schablone.** Der Name einer
          Angabe ist ein Bezeichner dieses Programms; seine **Beschriftung**
          gehört in den Katalog, und was der Katalog nicht kennt, steht roh
          da — das ist ehrlicher als eine erfundene Übersetzung (§9). */}
      {vorschlag.angaben.length > 0 && (
        <dl className="grid grid-cols-[auto_1fr] gap-x-2 text-xs text-leise">
          {vorschlag.angaben.map((angabe) => (
            <Fragment key={angabe.name}>
              <dt className="font-medium">
                {beschriftung(angabe.name, eigene)}
              </dt>
              <dd>{angabe.wert}</dd>
            </Fragment>
          ))}
        </dl>
      )}

      <div className="flex gap-2">
        <button
          type="button"
          className="rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
          disabled={laeuft}
          onClick={zuschneiden}
        >
          {eigene.kiBildUebernehmen}
        </button>

        <button
          type="button"
          className="rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
          disabled={laeuft}
          onClick={verwerfen}
        >
          {eigene.kiBildVerwerfen}
        </button>
      </div>
    </figure>
  )
}
