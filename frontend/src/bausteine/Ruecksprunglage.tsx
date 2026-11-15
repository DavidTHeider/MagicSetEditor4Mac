// **Was ein Zurückspielen kostet** — Werkstück O7 von Phase 5.2
// (#1431).
//
// # Zwei Richtungen, und die zweite ist die teurere
//
//  1. Was seit der Sicherung dazukam, **verschwindet** — Zahl für Zahl.
//  2. Was die Sicherung **mehr** hat, steht danach wieder da, ohne dass jemand
//     weiß, ob es noch jemandem gehört.
//
// Verlorene Karten kann man neu eintragen; eine zweimal verkaufte nicht. Eine
// Rückfrage, die nur die erste zeigt, verschweigt die teurere.
import type { RuecksprunglageAntwort } from '../apirueckkehr'
import { alsZeitpunkt } from '../zeitpunkt'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useZeitzone, useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

export function Ruecksprunglage({ lage }: { lage: RuecksprunglageAntwort }) {
  const worte = useTexte().seiten.ruecksprung
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()

  // **−1 heißt „diese Seite sagt nichts darüber"** und ist etwas anderes als
  // 0. Eine Zahl daneben, die nicht stimmt, wäre schlimmer als ein Strich.
  const zahl = (wert: number) => (wert < 0 ? worte.unbekannt : String(wert))

  return (
    <div className="mt-4">
      <h3 className="text-sm font-medium">{worte.vergleichTitel}</h3>

      {lage.datenstand !== '' && (
        <p className="mt-1 text-sm text-leise">
          {worte.datenstand}:{' '}
          {alsZeitpunkt(lage.datenstand, darstellung, zeitzone)}
        </p>
      )}

      <div className="mt-2 overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-leise">
              <th scope="col" className="pr-4 font-normal">
                {worte.spalteGegenstand}
              </th>
              <th scope="col" className="pr-4 font-normal">
                {worte.spalteSicherung}
              </th>
              <th scope="col" className="pr-4 font-normal">
                {worte.spalteJetzt}
              </th>
              <th scope="col" className="font-normal">
                {worte.spalteVerlust}
              </th>
            </tr>
          </thead>

          <tbody>
            {lage.zeilen.map((zeile) => (
              <tr key={zeile.gegenstand}>
                <td className="pr-4">
                  {worte.gegenstaende[zeile.gegenstand] ?? zeile.gegenstand}
                </td>
                <td className="pr-4 tabular-nums">{zahl(zeile.inSicherung)}</td>
                <td className="pr-4 tabular-nums">{zahl(zeile.jetzt)}</td>
                <td className="tabular-nums">
                  {zeile.unbekannt ? worte.unbekannt : String(zeile.verlust)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* **Und der Weg hinaus steht bei dem, was verloren geht** — die
          Abnahme verlangt, dass eigene Änderungen vor der Rückkehr ausfahrbar
          sind, und das nützt nur, wer davon weiß. */}
      <p className="mt-2 max-w-prose text-sm text-leise">
        {worte.vorherAusfahren}
      </p>

      {/* **Der Satz zur Handprüfung steht über den Mengen** — er sagt, dass
          die Liste darunter unvollständig sein kann, und das gehört davor. */}
      {lage.handpruefungNoetig && (
        <p className="mt-3 text-sm text-hinweis">{worte.handpruefung}</p>
      )}

      {lage.unklar.length > 0 && (
        <div className="mt-3">
          <h3 className="text-sm font-medium">{worte.unklarTitel}</h3>

          <Erklaerungsabsatz
            text={worte.unklarHinweis}
            klasse="mt-1 max-w-prose text-sm text-leise"
          />

          <ul className="mt-2 text-sm">
            {lage.unklar.map((menge) => (
              <li
                key={`${menge.schluessel}/${menge.zustand}/${menge.sprache}/${menge.finish}`}
              >
                {menge.name === ''
                  ? einsetzen(worte.ohneNamen, {
                      menge: String(menge.menge),
                      schluessel: menge.schluessel,
                    })
                  : einsetzen(worte.unklarZeile, {
                      menge: String(menge.menge),
                      name: menge.name,
                    })}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
