// **Wie weit jede Sprache ist** — H04
// (#2258).
//
// Unter der Sprachwahl, aufklappbar: je Sprache ihr Stand, ihre Herkunft und
// ob ein Mensch sie redigiert hat. Bis dahin sagte ein Satz alles für alle,
// und welche Sprache wann übersetzt war, stand nur in
// `docs/SPRACHINVENTAR.md`.
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import {
  bcp47,
  sprachen,
  sprachherkunft,
  sprachstaende,
} from '../sprache/kataloge'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { herkunftstext, redaktionstext } from './sprachstand'

export function Sprachstaende() {
  const texte = useTexte()
  const eigene = texte.seiten.sprachstaende
  const darstellung = useDarstellung()

  // **Die Namen, die auch die Wahl zeigt** — jede Sprache in ihrer eigenen.
  const namen: Record<string, string | undefined> =
    texte.seiten.einstellungen.werte

  return (
    <Aufklappkasten titel={eigene.titel}>
      <table className="text-left text-sm">
        <thead className="text-leise">
          <tr>
            <th scope="col" className="pr-3 font-medium">
              {eigene.kopf.sprache}
            </th>
            <th scope="col" className="pr-3 font-medium">
              {eigene.kopf.stand}
            </th>
            <th scope="col" className="pr-3 font-medium">
              {eigene.kopf.herkunft}
            </th>
            <th scope="col" className="font-medium">
              {eigene.kopf.redaktion}
            </th>
          </tr>
        </thead>
        <tbody>
          {sprachen.map((sprache) => (
            <tr key={sprache} className="align-top">
              <th
                scope="row"
                lang={bcp47[sprache]}
                className="pr-3 font-normal"
              >
                {namen[`sprache.${sprache}`]}
              </th>
              <td className="pr-3">{eigene.stand[sprachstaende[sprache]]}</td>
              <td className="pr-3">
                {herkunftstext(eigene, sprachherkunft[sprache], darstellung)}
              </td>
              <td>
                {redaktionstext(eigene, sprachherkunft[sprache], darstellung)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </Aufklappkasten>
  )
}
