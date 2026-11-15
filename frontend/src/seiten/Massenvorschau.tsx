// **Die Vorschautafel der Massenänderung** — Werkstück U6 von Phase 5.1
// (#1417):
// je Karte Bisher und Danach, und ein Wort, wo nichts zu tun ist. **Eine
// Tabelle, keine Liste**: drei Spalten, die ein Vorleseprogramm als Spalten
// nennt, statt drei Wörter in einer Zeile.
//
// **Zehn Zeilen vorn, der Rest stückweise** (UXR-06,
// #1914):
// wer ein ganzes Grundset anhakt, bekam 300 Zeilen untereinander, rund
// 9.000 Punkte unter dem Arbeitsplatz (gerechnet; R4 aus
// `docs/UIREVIEW.md`). Die Zusammenfassung darüber zählt alle; angewendet
// wird auf alle.
import { useState } from 'react'

import type { MassenvorschauAntwort } from '../api'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Faltfuss } from '../bausteine/Faltliste'

// faltvorn und faltschritt: zehn Karten stehen da, ein Klick zeigt fünfzig
// mehr — für die Auswahl wie für die Vorschau (`Massenaenderung.tsx`).
export const faltvorn = 10
export const faltschritt = 50

interface MassenvorschauEigenschaften {
  vorschau: MassenvorschauAntwort
  // konflikt nennt die Karten, die seit der Vorschau anders dastehen.
  konflikt: readonly string[]
}

export function Massenvorschau({
  vorschau,
  konflikt,
}: MassenvorschauEigenschaften) {
  const texte = useTexte()
  const worte = texte.seiten.karten.massen
  const [gezeigt, setzeGezeigt] = useState(faltvorn)
  const geaendert = vorschau.zeilen.filter((zeile) => zeile.ausschluss === '')
  const gleich = vorschau.zeilen.length - geaendert.length

  // **Was inzwischen anders dasteht, steht vorn** — hinter der Falte sähe
  // niemand, welche Karte den Konflikt trägt.
  const zeilen =
    konflikt.length === 0
      ? vorschau.zeilen
      : [
          ...vorschau.zeilen.filter((zeile) =>
            konflikt.includes(zeile.karteID),
          ),
          ...vorschau.zeilen.filter(
            (zeile) => !konflikt.includes(zeile.karteID),
          ),
        ]

  return (
    <div className="mt-3">
      <Erklaerungsabsatz
        text={worte.vorschauErklaerung}
        klasse="max-w-prose text-leise"
      />
      <p className="mt-1">
        {geaendert.length === 0
          ? worte.nichtsZuTun
          : einsetzen(worte.zusammenfassung, {
              geaendert: String(geaendert.length),
              gleich: String(gleich),
            })}
      </p>
      <table className="mt-2 w-full border-collapse text-sm">
        <thead>
          <tr className="border-b border-kante text-left">
            <th scope="col" className="py-1 pr-3 font-medium">
              {worte.karte}
            </th>
            <th scope="col" className="py-1 pr-3 font-medium">
              {worte.alt}
            </th>
            <th scope="col" className="py-1 font-medium">
              {worte.neu}
            </th>
          </tr>
        </thead>
        <tbody>
          {zeilen.slice(0, gezeigt).map((zeile) => {
            const inKonflikt = konflikt.includes(zeile.karteID)

            return (
              <tr key={zeile.karteID} className="border-b border-kante">
                <td className="py-1 pr-3">
                  {zeile.name === '' ? zeile.karteID : zeile.name}
                  {/* **Ein Wort, keine Farbe**: rot allein liest sich nicht
                      vor. */}
                  {inKonflikt && (
                    <span className="ml-2 text-fehler">
                      {worte.inzwischenGeaendert}
                    </span>
                  )}
                </td>
                <td className="py-1 pr-3 text-leise">
                  {zeile.gabAlt ? zeile.alt : worte.keinWertBisher}
                </td>
                <td className="py-1">
                  {zeile.ausschluss === '' ? zeile.neu : worte.bleibtGleich}
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
      {vorschau.zeilen.length > faltvorn && (
        <Faltfuss
          gezeigt={Math.min(gezeigt, vorschau.zeilen.length)}
          gesamt={vorschau.zeilen.length}
          schritt={faltschritt}
          weiter={() => {
            setzeGezeigt(gezeigt + faltschritt)
          }}
        />
      )}
    </div>
  )
}
