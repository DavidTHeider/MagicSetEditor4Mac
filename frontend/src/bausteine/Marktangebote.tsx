// **Die Angebote eines Marktplatzes** — Werkstück A4 von Phase 5.3
// (#1443).
//
// # Warum es einen eigenen Knopf braucht
//
// **Weil der Abruf ins Netz geht.** Die abgelegten Preise stehen beim
// Aufschlagen da; ein Angebot gibt es nur frisch, und wer eine Kartenliste
// öffnete, löste sonst je Zeile einen Abruf aus (§16, und dasselbe Argument
// wie bei `Preise`).
//
// # Was hier **nicht** steht
//
// **Kein Kaufknopf.** Gelesen wird, was andere anbieten; was ein Konto tut,
// ist Sache von T11
// (#1476).
import { useState } from 'react'

import { Knopf } from './Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { geldAusText } from '../geld'
import { useVorgang } from '../haken/useVorgang'

import type { Marktantwort } from '../antworten'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

interface MarktangeboteEigenschaften {
  printing: string
}

// dieQuelle ist der einzige Marktplatz, den es heute gibt.
//
// **Als Konstante und nicht als Text im JSX**: sie ist ein **Name** der
// Anwendung und kein Oberflächentext — derselbe Grund wie bei `Preisbulk`.
const dieQuelle = 'cardtrader'

export function Marktangebote({ printing }: MarktangeboteEigenschaften) {
  const texte = useTexte()
  const darstellung = useDarstellung()
  const eigene = texte.seiten.sammlung

  const [antwort, setzeAntwort] = useState<Marktantwort>()
  const [meldung, setzeMeldung] = useState<string>()

  const vorgang = useVorgang()

  // **Ein Vorgang je Druck** (§7, Regel 4): ein zweiter Ruf, während der erste
  // läuft, ginge gegen einen Stand, den das Backend noch nicht kennt.
  const laeuft = vorgang.laeuft(printing)

  const holen = () => {
    void vorgang.starte(printing, async () => {
      setzeMeldung(undefined)

      try {
        const geholt = await api().Marktangebote(
          dieQuelle,
          printing,
          '',
          '',
          [],
        )

        setzeAntwort(geholt)
        auspacken(geholt, () => undefined)
      } catch (ursache: unknown) {
        setzeMeldung(meldungAus(ursache, texte))
      }
    })
  }

  return (
    <section className="basis-full flex flex-col gap-2">
      <h3 className="text-sm font-medium">{eigene.marktTitel}</h3>

      <Erklaerungsabsatz
        text={eigene.marktErklaerung}
        klasse="text-sm text-leise"
      />

      <Knopf
        bezeichnung={einsetzen(eigene.marktHolenBei, { quelle: dieQuelle })}
        aufKlick={holen}
        angehalten={laeuft}
      >
        {laeuft ? eigene.marktLaeuft : eigene.marktHolen}
      </Knopf>

      {meldung !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {meldung}
        </p>
      )}

      {antwort?.fehler === null && antwort.zugeordnet && (
        <p className="text-sm text-leise">
          {einsetzen(eigene.marktZugeordnet, { fassung: antwort.fassung })}
        </p>
      )}

      {antwort?.fehler === null && antwort.angebote.length === 0 && (
        <p className="text-sm text-leise">{eigene.marktLeer}</p>
      )}

      {antwort !== undefined && antwort.angebote.length > 0 && (
        <ul className="flex flex-col gap-1 text-sm">
          {antwort.angebote.map((angebot) => (
            <li key={angebot.kennung} className="flex flex-wrap gap-x-2">
              <span className="font-medium">
                {geldAusText(angebot.wert, angebot.waehrung, darstellung)}
              </span>
              <span className="text-leise">
                {einsetzen(eigene.marktStueck, {
                  menge: String(angebot.menge),
                })}
              </span>
              <span className="text-leise">
                {angebot.finish} · {angebot.zustand} · {angebot.sprache}
              </span>
              <span className="text-leise">
                {angebot.verkaeufer}
                {angebot.land === '' ? '' : ` (${angebot.land})`}
              </span>
              {angebot.buendel > 1 && (
                <span className="text-leise">
                  {einsetzen(eigene.marktBuendel, {
                    buendel: String(angebot.buendel),
                  })}
                </span>
              )}
              {angebot.signiert && (
                <span className="text-leise">{eigene.marktSigniert}</span>
              )}
              {angebot.veraendert && (
                <span className="text-leise">{eigene.marktVeraendert}</span>
              )}
              {angebot.gradiert && (
                <span className="text-leise">{eigene.marktGradiert}</span>
              )}
              {/* **Der Urlaub steht dabei und filtert nicht weg**: ein Posten,
                  der da ist und nicht kommt, ist eine andere Auskunft als
                  keiner. */}
              {angebot.imUrlaub && (
                <span className="text-leise">{eigene.marktImUrlaub}</span>
              )}
            </li>
          ))}
        </ul>
      )}

      {antwort !== undefined && antwort.befunde.length > 0 && (
        <div className="text-xs text-leise">
          <p>{eigene.marktBefunde}</p>
          <ul className="mt-1 space-y-1">
            {antwort.befunde.map((befund) => (
              <li key={befund}>{befund}</li>
            ))}
          </ul>
        </div>
      )}
    </section>
  )
}
