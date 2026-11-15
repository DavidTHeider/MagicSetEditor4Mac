// **Alles mit einem Knopf** — G08
// (#2272).
//
// Über den fünf Zielen: **eine** Vorschau über alle Dateien und Zeilen, die
// hinzukommen, ein Knopf, der alles einrichtet, was es gibt und was fehlt, und
// einer, der alles Eingerichtete zurücknimmt. Danach steht je Ziel, was ihm
// geschah — ein Teilfehler mit seinem Grund, ein fremder Eintrag benannt.
// **Was es tut und was nie**, steht in `internal/app/schnittstellenalle.go`.
import { Bindungsfehler } from '../api'
import type { SchnittstelleDTO, SchnittstellenausgangDTO } from '../antworten'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Knopf } from './Knopf'

export function SchnittstellenAlle({
  ziele,
  ausgaenge,
  laeuft,
  einrichten,
  widerrufen,
}: {
  ziele: SchnittstelleDTO[]
  ausgaenge: SchnittstellenausgangDTO[] | undefined
  laeuft: boolean
  einrichten: () => void
  widerrufen: () => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.technik.schnittstellen
  const namen: Record<string, string | undefined> = eigene.namen
  const ausgangsnamen: Record<string, string | undefined> =
    eigene.alle.ausgaenge

  const fehlend = ziele.filter(
    (ziel) => ziel.stand === 'offen' || ziel.stand === 'veraltet',
  )
  const zuruecknehmbar = ziele.some(
    (ziel) => ziel.stand === 'eingerichtet' || ziel.stand === 'veraltet',
  )

  return (
    <div className="flex flex-col gap-2 rounded-sm border border-kante p-2">
      <h4 className="font-medium">{eigene.alle.titel}</h4>
      <p className="max-w-prose text-sm text-leise">{eigene.alle.satz}</p>

      {fehlend.length === 0 ? (
        <p className="text-sm">{eigene.alle.vorschauLeer}</p>
      ) : (
        <details className="text-sm">
          <summary>
            {einsetzen(eigene.alle.vorschau, {
              anzahl: String(fehlend.length),
            })}
          </summary>
          <ul className="mt-1 flex flex-col gap-2">
            {fehlend.map((ziel) => (
              <li key={ziel.name} className="flex flex-col gap-1">
                <span className="font-medium">
                  {namen[ziel.name] ?? ziel.name}
                </span>
                <span className="font-mono text-xs break-all text-leise">
                  {ziel.datei}
                </span>
                <pre className="font-mono text-xs whitespace-pre-wrap break-all">
                  {ziel.zeilen.join('\n')}
                </pre>
              </li>
            ))}
          </ul>
        </details>
      )}

      <div className="flex flex-wrap gap-2">
        <Knopf
          angehalten={laeuft || fehlend.length === 0}
          aufKlick={einrichten}
        >
          {eigene.alle.einrichten}
        </Knopf>
        <Knopf angehalten={laeuft || !zuruecknehmbar} aufKlick={widerrufen}>
          {eigene.alle.widerrufen}
        </Knopf>
      </div>

      {ausgaenge !== undefined && (
        <div role="status" className="flex flex-col gap-1 text-sm">
          <span className="font-medium">{eigene.alle.ergebnis}</span>
          <ul className="flex flex-col gap-1">
            {ausgaenge.map((eines) => (
              <li
                key={eines.name}
                data-ausgang={eines.ausgang}
                className="flex flex-wrap gap-x-2"
              >
                <span>{namen[eines.name] ?? eines.name}</span>
                <span className="text-leise">
                  {ausgangsnamen[eines.ausgang] ?? eines.ausgang}
                </span>
                {eines.fehler !== null && (
                  <span className="text-fehler">
                    {meldungAus(new Bindungsfehler(eines.fehler), texte)}
                  </span>
                )}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
