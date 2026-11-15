// **Die Verlaufsliste der Kartenseite** — Werkstück U6 von Phase 5.1
// (#1417).
//
// Je Schritt: wann, welche Karte, welches Feld, Bisher und Danach, wer — und
// zwei Handgriffe: **genau diesen Schritt zurücknehmen** (#1393 am Fenster)
// und bei einem Feldschritt **beide Karten nebeneinander sehen**. Eine
// Massenänderung ist ein Schritt; die Zeile zeigt ihre jüngste Karte und
// sagt, wie viele es waren.
//
// **Eine Tabelle, keine Liste**: sechs Spalten, die ein Vorleseprogramm als
// Spalten nennt.
//
// **Die jüngsten zehn, der Rest stückweise** (UXR-06,
// #1914).
// Ein Set hält 200 Schritte (`GrenzeVerlauf`); untereinander unter der
// Karte waren das — gemessen bei 1.280 × 800 — 308 Bedienelemente auf der
// Seite und eine Rolltiefe von 10,6 (R6 und R4 aus `docs/UIREVIEW.md`).
// Der oberste Schritt, den ⌘Z bewegt, steht immer da.
import { useState } from 'react'

import type { KartenzeileDTO, VerlaufsschrittDTO } from '../api'
import type { Katalog } from '../sprache/kataloge'
import { useVerlaufsliste } from '../haken/useVerlaufsliste'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Ladefeld } from './Ladefeld'
import { Verlaufsvergleich } from './Verlaufsvergleich'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Faltfuss } from '../bausteine/Faltliste'

interface VerlaufslisteEigenschaften {
  setID: string
  karten: readonly KartenzeileDTO[]
  verlaufsstand: number
  zuruecknehmen?: ((vorgang: string) => void) | undefined
}

// vorn und jeKlick: zehn Schritte stehen da, ein Klick zeigt fünfzig mehr.
const vorn = 10
const jeKlick = 50

// wann kürzt den Zeitpunkt auf Minute und Sekunde — die Nanosekunden aus dem
// Bestand sagen niemandem etwas.
function wann(zeitpunkt: string): string {
  return zeitpunkt.replace('T', ' ').slice(0, 19)
}

export function Verlaufsliste({
  setID,
  karten,
  verlaufsstand,
  zuruecknehmen,
}: VerlaufslisteEigenschaften) {
  const texte = useTexte()
  const worte = texte.seiten.karten.verlauf
  const schritte = useVerlaufsliste(setID, verlaufsstand)
  const [verglichen, setzeVerglichen] = useState<VerlaufsschrittDTO>()
  const [zahl, setzeZahl] = useState(vorn)

  const name = (karteID: string) =>
    karten.find((karte) => karte.id === karteID)?.name ?? karteID

  const urheber: Record<string, string> = {
    fenster: worte.urheberFenster,
    server: worte.urheberServer,
    kommandozeile: worte.urheberKommandozeile,
    dienst: worte.urheberDienst,
  }

  const handlung: Record<string, string> = {
    reihenfolge: worte.handlungReihenfolge,
    schluesselwort: worte.handlungSchluesselwort,
  }

  return (
    <>
      <Erklaerungsabsatz
        text={worte.erklaerung}
        klasse="mb-2 max-w-prose text-leise"
      />
      <Ladefeld
        laedt={schritte.laedt}
        fehler={schritte.fehler}
        leer={schritte.daten?.length === 0}
      >
        <p className="max-w-prose text-leise">{worte.leer}</p>
      </Ladefeld>
      {/* **Die Tabelle rollt in sich, nicht die Seite** — UX-14.AK-03
          (#1686): bei 1.024 × 700 sprengten sechs Spalten die Mitte.
          **Erreichbar und benannt**, sonst rollte ihn nur, wer eine Maus
          hat. */}
      {schritte.daten !== undefined && schritte.daten.length > 0 && (
        <div
          role="region"
          aria-label={worte.ueberschrift}
          tabIndex={0}
          className="overflow-x-auto focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        >
          <table className="w-full border-collapse text-sm">
            <thead>
              <tr className="border-b border-kante text-left">
                <th scope="col" className="py-1 pr-3 font-medium">
                  {worte.wann}
                </th>
                <th scope="col" className="py-1 pr-3 font-medium">
                  {worte.karte}
                </th>
                <th scope="col" className="py-1 pr-3 font-medium">
                  {worte.feld}
                </th>
                <th scope="col" className="py-1 pr-3 font-medium">
                  {worte.bisher}
                </th>
                <th scope="col" className="py-1 pr-3 font-medium">
                  {worte.danach}
                </th>
                <th scope="col" className="py-1 font-medium">
                  {worte.wer}
                </th>
              </tr>
            </thead>
            <tbody>
              {schritte.daten.slice(0, zahl).map((schritt) => (
                <tr
                  key={schritt.id}
                  className={
                    'border-b border-kante align-top' +
                    (schritt.zurueckgenommen ? ' text-leise' : '')
                  }
                >
                  <td className="py-1 pr-3 whitespace-nowrap">
                    {wann(schritt.zeitpunkt)}
                  </td>
                  <td className="py-1 pr-3">
                    {name(schritt.karteID)}
                    {schritt.umfang > 1 && (
                      <span className="ml-1 text-leise">
                        {einsetzen(worte.karten, {
                          anzahl: String(schritt.umfang),
                        })}
                      </span>
                    )}
                  </td>
                  <td className="py-1 pr-3">
                    {schritt.handlung === 'feld'
                      ? schritt.feld
                      : schritt.handlung === 'karte'
                        ? kartenhandlung(schritt, worte)
                        : handlung[schritt.handlung]}
                  </td>
                  <td className="py-1 pr-3">
                    {schritt.handlung === 'karte' ? (
                      kartenstand(schritt.gabVorher, worte)
                    ) : schritt.gabVorher ? (
                      schritt.vorher
                    ) : (
                      <span className="text-leise">{worte.keinWert}</span>
                    )}
                  </td>
                  <td className="py-1 pr-3">
                    {schritt.handlung === 'karte' ? (
                      kartenstand(schritt.gabNachher, worte)
                    ) : schritt.gabNachher ? (
                      schritt.nachher
                    ) : (
                      <span className="text-leise">{worte.keinWert}</span>
                    )}
                    {schritt.zurueckgenommen && (
                      <span className="ml-2">{worte.zurueckgenommen}</span>
                    )}
                  </td>
                  <td className="py-1">
                    <span>{urheber[schritt.urheber] ?? schritt.urheber}</span>
                    <span className="ml-2 inline-flex gap-2">
                      {/* **Nur, was noch gilt, lässt sich zurücknehmen**, und
                        nur mit einem Vorgang gezielt — eine Einzeländerung
                        trägt keinen, sie geht über ⌘Z. */}
                      {zuruecknehmen !== undefined &&
                        !schritt.zurueckgenommen &&
                        schritt.vorgang !== '' && (
                          <Aktionsknopf
                            aktion="zuruecknehmen"
                            klein
                            aufKlick={() => {
                              zuruecknehmen(schritt.vorgang)
                            }}
                          >
                            {worte.zuruecknehmen}
                          </Aktionsknopf>
                        )}
                      {schritt.handlung === 'feld' && (
                        <button
                          type="button"
                          aria-pressed={verglichen?.id === schritt.id}
                          onClick={() => {
                            setzeVerglichen(
                              verglichen?.id === schritt.id
                                ? undefined
                                : schritt,
                            )
                          }}
                          className="rounded-sm border border-kante px-2 py-0.5 text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                        >
                          {worte.vergleichen}
                        </button>
                      )}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {schritte.daten !== undefined && schritte.daten.length > vorn && (
        <Faltfuss
          gezeigt={Math.min(zahl, schritte.daten.length)}
          gesamt={schritte.daten.length}
          schritt={jeKlick}
          weiter={() => {
            setzeZahl(zahl + jeKlick)
          }}
        />
      )}
      {verglichen !== undefined && (
        <Verlaufsvergleich
          schritt={verglichen}
          kartenname={name(verglichen.karteID)}
          schliessen={() => {
            setzeVerglichen(undefined)
          }}
        />
      )}
    </>
  )
}

// kartenhandlung sagt, wohin ein Kartenschritt ging — **angelegt oder
// gelöscht**, wie die Fassade es mitgibt (G01, `g01einsteiger`). Steht die
// Richtung nicht fest, bleibt der Satz, der beides nennt.
function kartenhandlung(
  schritt: VerlaufsschrittDTO,
  worte: Katalog['seiten']['karten']['verlauf'],
): string {
  if (!schritt.gabVorher && schritt.gabNachher) {
    return worte.handlungKarteAngelegt
  }

  if (schritt.gabVorher && !schritt.gabNachher) {
    return worte.handlungKarteGeloescht
  }

  return worte.handlungKarte
}

// kartenstand sagt, ob die Karte auf dieser Seite des Schritts dastand.
function kartenstand(
  gab: boolean,
  worte: Katalog['seiten']['karten']['verlauf'],
) {
  return gab ? (
    worte.karteVorhanden
  ) : (
    <span className="text-leise">{worte.keineKarte}</span>
  )
}
