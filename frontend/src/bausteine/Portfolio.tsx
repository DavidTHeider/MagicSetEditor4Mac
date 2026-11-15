// **Die Übersicht über die Sammlung** — Werkstück 3 von Phase 3.4
// (#1165).
//
// # Drei Angaben stehen neben jeder Summe
//
// Wie viele Exemplare eingegangen sind, wie viele **ohne Preis** blieben, und
// wie alt der älteste Preis ist. Ohne sie wäre die Zahl eine, für die niemand
// einsteht — die Quelle sagt das über ihre Preise ausdrücklich.
import { useEffect, useRef } from 'react'

import type { Portfolioantwort } from '../api'
import { alterstext } from './preisalter'
import { einsetzen } from '../sprache/einsetzen'
import type { Ladezustand } from '../haken/useLaden'
import { usePortfolio } from '../haken/usePortfolio'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { geldAusText } from '../geld'

interface PortfolioEigenschaften {
  // runde steigt, wenn sich die Sammlung geändert hat — dann wird neu
  // gerechnet.
  runde: number
}

export function Portfolio({ runde }: PortfolioEigenschaften) {
  return <Portfolioansicht stand={usePortfolio(runde)} />
}

// Portfolioansicht zeigt eine schon gefragte Übersicht — die Sammlung fragt
// sie einmal für Leiste und Navigator (AP-9, `usePortfolio`).
//
// **`anlass` führt hierher**: steigt er, rückt die Überschrift ins Bild und
// bekommt den Fokus — das Ziel von „Wert“ und „ohne Preis“ in der Leiste.
export function Portfolioansicht({
  stand,
  anlass = 0,
}: {
  stand: Ladezustand<Portfolioantwort>
  anlass?: number
}) {
  const texte = useTexte()
  const darstellung = useDarstellung()
  const eigene = texte.seiten.sammlung
  const ueberschrift = useRef<HTMLHeadingElement>(null)

  useEffect(() => {
    if (anlass > 0) {
      ueberschrift.current?.scrollIntoView({ block: 'start' })
      ueberschrift.current?.focus()
    }
  }, [anlass])

  // **Ein Nullwert statt zweier Rückfälle.** Wer `daten?.zeilen ?? []` und
  // `daten?.ohnePreis ?? 0` nebeneinander schreibt, hat zwei Zweige für
  // denselben Fall — und einer davon ist nicht zu betreten, sobald Zeilen da
  // sind (§3).
  const uebersicht = stand.daten ?? { zeilen: [], ohnePreis: 0 }
  const zeilen = uebersicht.zeilen

  return (
    <section className="flex flex-col gap-2">
      <h2 ref={ueberschrift} tabIndex={-1} className="text-sm font-medium">
        {eigene.uebersichtTitel}
      </h2>

      {stand.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {stand.fehler}
        </p>
      )}

      {zeilen.length === 0 && stand.fehler === undefined && (
        <p className="text-sm text-leise">{eigene.uebersichtLeer}</p>
      )}

      {zeilen.length > 0 && (
        <>
          <ul className="flex flex-col gap-1">
            {zeilen.map((zeile) => (
              <li key={zeile.waehrung} className="text-sm">
                <span className="font-medium">
                  {geldAusText(zeile.wert, zeile.waehrung, darstellung)}
                </span>{' '}
                <span className="text-leise">
                  ·{' '}
                  {einsetzen(
                    zeile.exemplare === 1
                      ? eigene.einExemplar
                      : eigene.exemplare,
                    { zahl: String(zeile.exemplare) },
                  )}{' '}
                  · {alterstext(eigene, zeile.alterStunden, zeile.quelle)}
                </span>
              </li>
            ))}
          </ul>

          {/* **Die Zahl ohne Preis steht immer da, auch wenn sie null ist.**
              Eine Summe ohne sie sähe aus wie eine über alles. */}
          <p className="text-sm text-leise">
            {einsetzen(eigene.ohnePreis, {
              zahl: String(uebersicht.ohnePreis),
            })}
          </p>

          <p className="max-w-3xl text-xs text-leise">
            {eigene.keineUmrechnung}
          </p>
        </>
      )}
    </section>
  )
}
