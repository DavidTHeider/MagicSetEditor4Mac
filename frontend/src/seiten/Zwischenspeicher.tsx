// **Die Zwischenspeicher sehen und leeren** — OPS-08
// (#1725).
//
// **Die Trefferquote ist die Auskunft**: ein Speicher, der selten trifft,
// kostet Platz, ohne etwas zu sparen. Deshalb steht sie vorn, und ohne eine
// einzige Abfrage steht keine — eine Quote von null wäre eine Behauptung.
import { useState } from 'react'

import { api } from '../fassade'
import { Knopf } from '../bausteine/Knopf'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { alsProzent } from '../lokal'
import { meldungAus } from '../meldung'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { alsGroesse } from './Diagnosepaket'

import type { ZwischenspeicherDTO } from '../apizwischenspeicher'

export function Zwischenspeicher() {
  const texte = useTexte()
  const eigene = texte.seiten.technik.zwischenspeicher
  const darstellung = useDarstellung()
  const vorgang = useVorgang()

  const [lauf, setzeLauf] = useState(0)
  const [geleert, setzeGeleert] = useState(false)
  const [fehler, setzeFehler] = useState<string>()

  const stand = useLaden(async () => {
    const antwort = await api().Zwischenspeicher()

    return auspacken(antwort, () => antwort.speicher)
  }, [lauf])

  const leeren = () =>
    vorgang.starte('zwischenspeicher', async () => {
      try {
        const antwort = await api().ZwischenspeicherLeeren()
        auspacken(antwort, () => null)
        setzeGeleert(true)
        setzeFehler(undefined)
        setzeLauf((bisher) => bisher + 1)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  const namen = eigene.namen as Record<string, string>
  const orte = eigene.orte as Record<string, string>

  const quote = (speicher: ZwischenspeicherDTO) => {
    const abfragen = speicher.treffer + speicher.fehlgriffe

    return abfragen === 0
      ? eigene.keineQuote
      : alsProzent(speicher.treffer / abfragen, darstellung)
  }

  const grenze = (speicher: ZwischenspeicherDTO) => {
    if (speicher.grenzeBytes > 0) {
      return alsGroesse(speicher.grenzeBytes, darstellung)
    }

    return speicher.grenzeEintraege > 0
      ? speicher.grenzeEintraege.toLocaleString(darstellung)
      : eigene.keineGrenze
  }

  return (
    <section className="flex flex-col gap-3 border-t border-kante pt-4">
      <h3 className="text-lg font-semibold">{eigene.ueberschrift}</h3>

      <Erklaerungsabsatz text={eigene.erklaerung} klasse="text-sm text-leise" />

      {stand.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {stand.fehler}
        </p>
      )}

      {stand.daten !== undefined && (
        <table className="w-full max-w-3xl text-sm">
          <caption className="sr-only">{eigene.ueberschrift}</caption>

          <thead>
            <tr className="text-left">
              <th scope="col">{eigene.spalteSpeicher}</th>
              <th scope="col">{eigene.spalteOrt}</th>
              <th scope="col" className="text-right">
                {eigene.spalteQuote}
              </th>
              <th scope="col" className="text-right">
                {eigene.spalteEintraege}
              </th>
              <th scope="col" className="text-right">
                {eigene.spalteGroesse}
              </th>
              <th scope="col" className="text-right">
                {eigene.spalteGrenze}
              </th>
            </tr>
          </thead>

          <tbody>
            {stand.daten.map((speicher) => (
              <tr key={speicher.name}>
                <td>{namen[speicher.name] ?? speicher.name}</td>
                <td>{orte[speicher.ort] ?? speicher.ort}</td>
                <td className="text-right tabular-nums">{quote(speicher)}</td>
                <td className="text-right tabular-nums">
                  {speicher.eintraege.toLocaleString(darstellung)}
                </td>
                <td className="text-right tabular-nums">
                  {alsGroesse(speicher.bytes, darstellung)}
                </td>
                <td className="text-right tabular-nums">{grenze(speicher)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <Erklaerungsabsatz
        text={eigene.leerenHinweis}
        klasse="text-sm text-leise"
      />

      <div>
        <Knopf
          bezeichnung={eigene.leeren}
          angehalten={vorgang.laeuft('zwischenspeicher')}
          aufKlick={() => {
            void leeren()
          }}
        >
          {eigene.leeren}
        </Knopf>
      </div>

      {geleert && (
        <p role="status" className="text-sm">
          {eigene.geleert}
        </p>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}
    </section>
  )
}
