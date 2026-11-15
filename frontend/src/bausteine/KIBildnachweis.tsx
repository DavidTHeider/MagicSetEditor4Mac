// **Was über das Bild feststeht, das jetzt im Feld steht** — Werkstück A12
// von Phase 5.3
// (#1451).
//
// # Warum hier und nicht im Herkunftskasten
//
// **Weil die Angaben nur für Bilder gelten.** [KIHerkunft] beantwortet für
// **jedes** Feld dieselbe Frage — wer, womit, wann. Größe, Güte und
// Umschrift gelten nur für ein erzeugtes Bild, und sie dort anzuhängen hieße,
// je Feld eine zweite Abfrage zu stellen, die für die meisten leer bliebe.
//
// **Es ist der „Bildnachweis" aus der Ergänzung des Issues**: „Herkunft
// gehört zum temporären Vorschlag und bei bestätigter Übernahme zum
// Bildnachweis."
import { useId } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'

import { beschriftung } from './kibildangaben'

interface KIBildnachweisEigenschaften {
  karteID: string
  feld: string

  // gefragt sagt, ob gelesen werden darf — zugeklappt wird nicht gefragt.
  gefragt: boolean

  // stand zählt die Übernahmen; nach jeder wird neu gelesen.
  stand: number
}

export function KIBildnachweis({
  karteID,
  feld,
  gefragt,
  stand,
}: KIBildnachweisEigenschaften) {
  const eigene = useTexte().seiten.einstellungen
  const ueberschrift = useId()

  const nachweis = useLaden(async () => {
    if (!gefragt || feld === '') {
      return undefined
    }

    const antwort = await api().KIBildHerkunft(karteID, feld)

    return auspacken(antwort, () => antwort.angaben)
  }, [gefragt, karteID, feld, stand])

  if (nachweis.fehler !== undefined) {
    return (
      <p role="alert" className="text-sm text-fehler">
        {nachweis.fehler}
      </p>
    )
  }

  // **Ohne Angaben steht nichts da.** Ein Bild von der Platte hat keine, und
  // „keine Angaben" wäre eine Auskunft über nichts.
  if (nachweis.daten === undefined || nachweis.daten.length === 0) {
    return null
  }

  return (
    <section
      className="rounded-sm border border-kante p-2"
      aria-labelledby={ueberschrift}
    >
      <h3 id={ueberschrift} className="text-sm font-medium">
        {eigene.kiBildNachweisTitel}
      </h3>

      <dl className="mt-1 grid grid-cols-[auto_1fr] gap-x-2 text-xs text-leise">
        {nachweis.daten.map((angabe) => (
          <div key={angabe.name} className="contents">
            <dt className="font-medium">{beschriftung(angabe.name, eigene)}</dt>
            <dd>{angabe.wert}</dd>
          </div>
        ))}
      </dl>
    </section>
  )
}
