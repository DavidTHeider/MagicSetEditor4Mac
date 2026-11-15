// **Wer bei diesem Tausch ungefähr mehr bekommt** — TRADE-02b
// (#1951).
//
// **Ein Anhalt und keine Vorgabe**: der Satz darunter sagt, woher die Preise
// sind und wie alt, und dass sie nicht in den Vertrag gehen (AK-03).
//
// # Die Richtung, nicht das Vorzeichen
//
// **„1,95 € zu deinen Gunsten"** und nicht „+1,95 €": ein Vorzeichen sagt
// nicht, aus wessen Sicht gerechnet ist. Im Artboard stand die Richtung bei
// 19,90 gegen 17,95 umgekehrt — gerechnet wird hier erhalten weniger gegeben.
import { alsGeld } from '../geld'
import { alsZeitpunkt } from '../zeitpunkt'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte, useZeitzone } from '../sprache/kontext'

import type { Preisanhalt } from '../apihandelsleitung'

interface Eigenschaften {
  anhalt: Preisanhalt | undefined
  fehler: string | undefined
}

export function Handelsdifferenz({ anhalt, fehler }: Eigenschaften) {
  const eigene = useTexte().seiten.handeln.assistent
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()

  if (fehler !== undefined) {
    return (
      <p role="alert" className="text-sm text-fehler">
        {fehler}
      </p>
    )
  }

  if (anhalt === undefined) {
    return null
  }

  const waehrung = anhalt.hauptwaehrung
  const differenz = anhalt.differenzen.find(
    (eine) => eine.waehrung === waehrung,
  )

  if (differenz === undefined) {
    return (
      <p className="max-w-prose text-sm text-leise">{eigene.keinePreise}</p>
    )
  }

  const betrag = alsGeld(Math.abs(differenz.wert), waehrung, darstellung)
  const satz =
    differenz.wert > 0
      ? einsetzen(eigene.differenzDir, { betrag })
      : differenz.wert < 0
        ? einsetzen(eigene.differenzGegenueber, { betrag })
        : eigene.ausgeglichen

  // **Der älteste Preis dieser Währung** — eine Summe ist so frisch wie ihr
  // ältester Bestandteil.
  const summen = anhalt.seiten.flatMap((seite) =>
    seite.summen.filter((summe) => summe.waehrung === waehrung),
  )
  const aeltester = summen.reduce((bisher, summe) =>
    summe.erhoben < bisher.erhoben ? summe : bisher,
  )
  const ohnePreis = anhalt.zeilen.filter(
    (zeile) => !zeile.preise.some((preis) => preis.waehrung === waehrung),
  ).length

  return (
    <div className="flex flex-col gap-1 text-sm">
      <p role="status" className="font-medium">
        {satz}
      </p>

      {!differenz.vollstaendig && (
        <p className="text-leise">
          {einsetzen(eigene.differenzTeil, { zahl: String(ohnePreis) })}
        </p>
      )}

      <p className="max-w-prose text-leise">
        {einsetzen(eigene.preisHinweis, {
          datum: alsZeitpunkt(aeltester.erhoben, darstellung, zeitzone),
          quelle: aeltester.quelle,
        })}
      </p>
    </div>
  )
}
