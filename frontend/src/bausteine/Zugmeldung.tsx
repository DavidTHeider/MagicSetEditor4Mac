// **Was ein Zug unterwegs sagt** — P13
// (#2381),
// Vermerk 0200: jeder Teil hat seinen eigenen Satz und, wo es einen gibt,
// seinen eigenen Ausweg. **Was gegangen ist, bleibt** — die Meldung sagt
// deshalb nie „nichts geschehen", sondern wo es hängt.
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Knopf } from './Knopf'

import type { Zugausgang, Zugschritt } from '../handelszug'

interface Zugmeldungseigenschaften {
  ausgang: Zugausgang | undefined

  // name ist, wie das Gegenüber heißt.
  name: string

  angehalten: boolean
  nochmal: (bisher: Zugausgang, legen: boolean) => void
}

export function Zugmeldung({
  ausgang,
  name,
  angehalten,
  nochmal,
}: Zugmeldungseigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.handeln
  const zug = eigene.zug

  if (ausgang === undefined) {
    return null
  }

  const { gescheitert, zustellung } = ausgang

  const schritte: Record<Zugschritt | '', string> = {
    '': '',
    anbieten: eigene.schrittAnbieten,
    zustimmen: eigene.schrittZustimmen,
    vorbereiten: eigene.schrittVorbereiten,
    uebergeben: eigene.schrittUebergeben,
    verbuchen: eigene.buchungVerbuchen,
    abgleichen: eigene.buchungAbgleichen,
    bestaetigen: eigene.buchungEigeneStimme,
  }

  const satz = (
    gescheitert: NonNullable<Zugausgang['gescheitert']>,
  ): string => {
    const meldung =
      gescheitert.ursache === null ? '' : meldungAus(gescheitert.ursache, texte)

    switch (gescheitert.teil) {
      case 'veraltet':
        return zug.veraltet
      case 'festhalten':
        return gescheitert.ursache === null
          ? zug.festhaltenOffen
          : einsetzen(zug.gescheitert, {
              schritt: eigene.zusagenFesthalten,
              meldung,
            })
      case 'lesen':
        return einsetzen(zug.lesen, { meldung })
      case 'ausgang':
        return einsetzen(zug.ausgang, { meldung })
      case 'schritt':
        return einsetzen(zug.gescheitert, {
          schritt: schritte[gescheitert.schritt],
          meldung,
        })
    }
  }

  return (
    <div className="flex flex-col gap-1">
      {gescheitert !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {satz(gescheitert)}
        </p>
      )}

      {gescheitert?.teil === 'ausgang' && (
        <div>
          <Knopf
            gewicht="leise"
            angehalten={angehalten}
            aufKlick={() => {
              nochmal(ausgang, true)
            }}
          >
            {eigene.ausgangSchicken}
          </Knopf>
        </div>
      )}

      {zustellung !== undefined && (
        <>
          <p role="alert" className="text-sm text-fehler">
            {einsetzen(zug.unterwegs, {
              name,
              meldung: meldungAus(zustellung, texte),
            })}
          </p>
          <div>
            <Knopf
              gewicht="leise"
              angehalten={angehalten}
              aufKlick={() => {
                nochmal(ausgang, false)
              }}
            >
              {eigene.ausgangZustellen}
            </Knopf>
          </div>
        </>
      )}

      {gescheitert === undefined &&
        zustellung === undefined &&
        ausgang.getragen > 0 && (
          <p role="status" className="text-sm text-leise">
            {einsetzen(zug.hinaus, { name })}
          </p>
        )}
    </div>
  )
}
