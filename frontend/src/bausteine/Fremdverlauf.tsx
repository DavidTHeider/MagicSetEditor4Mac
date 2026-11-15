// **Den Preisverlauf bei einer fremden Quelle holen** — Werkstück A5 von
// Phase 5.3 (#1444).
//
// # Warum es neben »Preise holen« steht und nicht darin
//
// **Weil es etwas anderes holt.** `PreiseHolen` fragt nach dem **heutigen**
// Stand bei Scryfall; hier kommt eine **Reihe über die Zeit** von einer
// anderen Quelle, mit ihrem eigenen Zustandsvokabular. Beides in einen Knopf
// zu legen hieße, zwei Fragen mit einer Antwort zu beantworten.
import { useState } from 'react'

import { Knopf } from './Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { Fremdpreisantwort } from '../antworten'

interface FremdverlaufEigenschaften {
  printing: string

  // aufGeholt sagt dem Aufrufer, dass neue Stände dastehen.
  aufGeholt: () => void
}

// dieQuelle ist die einzige, die heute einen Verlauf hergibt.
const dieQuelle = 'scrydex'

export function Fremdverlauf({
  printing,
  aufGeholt,
}: FremdverlaufEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.sammlung

  const [bilanz, setzeBilanz] = useState<Fremdpreisantwort>()
  const [meldung, setzeMeldung] = useState<string>()

  const vorgang = useVorgang()
  const laeuft = vorgang.laeuft(printing)

  const holen = () => {
    void vorgang.starte(printing, async () => {
      setzeMeldung(undefined)

      try {
        const geholt = await api().Fremdpreisverlauf(
          dieQuelle,
          printing,
          '',
          [],
        )

        // **Die Bilanz wird gesetzt, bevor der Fehler geworfen wird**
        // (§7, Regel 3): was abgelegt ist, bleibt abgelegt.
        setzeBilanz(geholt)
        auspacken(geholt, () => undefined)

        aufGeholt()
      } catch (ursache: unknown) {
        setzeMeldung(meldungAus(ursache, texte))
      }
    })
  }

  return (
    <section className="basis-full flex flex-col gap-2">
      <h3 className="text-sm font-medium">{eigene.verlaufTitel}</h3>

      <Erklaerungsabsatz
        text={eigene.verlaufErklaerung}
        klasse="text-sm text-leise"
      />

      <Knopf
        bezeichnung={einsetzen(eigene.verlaufHolenBei, { quelle: dieQuelle })}
        aufKlick={holen}
        angehalten={laeuft}
      >
        {laeuft ? eigene.verlaufLaeuft : eigene.verlaufHolen}
      </Knopf>

      {meldung !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {meldung}
        </p>
      )}

      {bilanz?.fehler === null && (
        <p className="text-sm text-leise">
          {einsetzen(eigene.verlaufBilanz, {
            staende: String(bilanz.staende),
            fassung: bilanz.fassung,
          })}
        </p>
      )}

      {bilanz !== undefined && bilanz.befunde.length > 0 && (
        <ul className="flex flex-col gap-1 text-xs text-leise">
          {bilanz.befunde.map((befund) => (
            <li key={befund}>{befund}</li>
          ))}
        </ul>
      )}
    </section>
  )
}
