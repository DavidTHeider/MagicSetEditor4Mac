// **Ein Symbol in den Fassungen dieses Sets** — DOM-02.AK-02
// (#1697):
// die Größen- und Farbvorschau.
//
// Je Fassung des Stils eine Zeile mit dem Symbol in drei Größen — gezeichnet
// vom Rasterer der Karten, damit es hier so aussieht wie an der Typzeile. Was
// sich nicht zeichnen ließ, fehlt, statt als leeres Bild dazustehen.
import { api } from '../api'
import { Ladefehler } from '../bausteine/Ladehinweise'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'

interface FassungenEigenschaften {
  id: string
  name: string
  setID: string
}

export function Symbolfassungen({ id, name, setID }: FassungenEigenschaften) {
  const eigene = useTexte().seiten.karten.symbolbibliothek

  const fassungen = useLaden(async () => {
    const antwort = await api().BibliothekssymbolFassungen(id, setID)

    return auspacken(antwort, () => antwort.fassungen)
  }, [id, setID])

  if (fassungen.fehler !== undefined) {
    return (
      <Ladefehler fehler={fassungen.fehler} wiederholen={fassungen.nachladen} />
    )
  }

  if (fassungen.daten === undefined) {
    return null
  }

  if (fassungen.daten.length === 0) {
    return (
      <p className="max-w-prose text-sm text-leise">{eigene.keineFassungen}</p>
    )
  }

  return (
    <ul
      aria-label={einsetzen(eigene.fassungsliste, { name })}
      className="flex flex-col gap-2"
    >
      {fassungen.daten.map((fassung) => (
        <li key={fassung.name} className="flex flex-wrap items-end gap-3">
          <span className="w-32 text-sm">{fassung.name}</span>
          {fassung.bilder
            .filter((bild) => bild.vorschau !== '')
            .map((bild) => (
              <img
                key={bild.groesse}
                src={bild.vorschau}
                alt={einsetzen(eigene.fassungsbild, {
                  name,
                  fassung: fassung.name,
                  groesse: String(bild.groesse),
                })}
                width={bild.groesse}
                height={bild.groesse}
              />
            ))}
        </li>
      ))}
    </ul>
  )
}
