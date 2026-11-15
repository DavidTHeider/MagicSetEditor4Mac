// **Betrieb und Vertrauen** — OPS-07
// (#1724).
//
// # Aus der Anwendung gelesen
//
// **Was hier steht, sagt `Betriebsauskunft`** — die Fassung aus dem Bau, der
// Datenstand aus dem Bestand, die Orte aus denselben Funktionen, die sie
// anlegen, und der Prüfstand aus einer Liste, die ein Wächter gegen ihre
// Belege hält (`internal/werkzeuge/pruefstand_test.go`). Eine Zahl, die hier
// von Hand stünde, veraltete still.
//
// # Eigene Datei
//
// Wie `Hilfedatenschutz`: `Hilfe.tsx` ginge sonst über 400 Zeilen (§2).
import { api } from '../fassade'
import { Flaeche } from '../bausteine/Flaeche'
import { Satz } from '../bausteine/Satz'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { alsZeitpunkt } from '../zeitpunkt'

import type { PruefpunktDTO } from '../apibetrieb'
import { Dateiverweis } from '../bausteine/Dateiverweis'

const issue = 'https://github.com/DavidTHeider/MagicSetEditor4Mac/issues/'

export function Hilfebetrieb() {
  const eigene = useTexte().hilfe.betrieb
  const darstellung = useDarstellung()

  const stand = useLaden(async () => {
    const antwort = await api().Betriebsauskunft()

    return auspacken(antwort, () => antwort)
  }, [])

  const worte = (tafel: Record<string, string>, schluessel: string) =>
    tafel[schluessel] ?? schluessel

  const beleg = (punkt: PruefpunktDTO) =>
    punkt.beleg.startsWith(issue)
      ? einsetzen(eigene.belegIssue, {
          nummer: punkt.beleg.slice(issue.length),
        })
      : punkt.beleg

  const auskunft = stand.daten

  return (
    <Flaeche ueberschrift={eigene.ueberschrift} sprungmarke="hilfe-betrieb">
      <p className="mb-4 max-w-prose text-leise">
        <Satz text={eigene.einleitung} />
      </p>

      {stand.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {stand.fehler}
        </p>
      )}

      {auskunft !== undefined && (
        <>
          <dl className="grid gap-3 sm:grid-cols-[12rem_minmax(0,1fr)]">
            <dt className="font-semibold">{eigene.fassungName}</dt>
            <dd className="max-w-prose text-leise">
              {einsetzen(eigene.fassungSatz, {
                fassung: auskunft.fassung,
                stand: auskunft.stand.slice(0, 7),
                gebaut: alsZeitpunkt(auskunft.gebaut, darstellung),
              })}
              {auskunft.veraendert && (
                <span className="ml-1">{eigene.fassungVeraendert}</span>
              )}
            </dd>

            <dt className="font-semibold">{eigene.datenName}</dt>
            <dd className="max-w-prose text-leise">
              {auskunft.migrationen === 0
                ? eigene.datenOhne
                : einsetzen(eigene.datenSatz, {
                    zahl: auskunft.migrationen.toLocaleString(darstellung),
                    letzte: auskunft.letzteMigration,
                  })}
            </dd>

            <dt className="font-semibold">{eigene.signaturName}</dt>
            <dd className="max-w-prose text-leise">
              <Satz text={eigene.signaturSatz} />
            </dd>

            <dt className="font-semibold">{eigene.aktualisierungName}</dt>
            <dd className="max-w-prose text-leise">
              {worte(eigene.aktualisierung, auskunft.aktualisierung)}
            </dd>

            {(
              [
                ['lizenzenName', 'lizenzenSatz'],
                ['empfaengerName', 'empfaengerSatz'],
                ['grenzenName', 'grenzenSatz'],
                ['kontaktName', 'kontaktSatz'],
              ] as const
            ).map(([name, satz]) => (
              <div key={name} className="contents">
                <dt className="font-semibold">{eigene[name]}</dt>
                <dd className="max-w-prose text-leise">
                  <Satz text={eigene[satz]} />
                </dd>
              </div>
            ))}
          </dl>

          <h3 className="mt-6 mb-2 font-semibold">{eigene.orteName}</h3>
          <dl className="grid gap-2 sm:grid-cols-[12rem_minmax(0,1fr)]">
            {auskunft.orte.map((ort) => (
              <div key={ort.name} className="contents">
                <dt>{worte(eigene.orte, ort.name)}</dt>
                <dd>
                  <span className="font-mono text-sm break-all text-leise">
                    {ort.pfad}
                  </span>

                  {/* **Ansehen und im Finder zeigen** — UX-13.AK-04 (#1685):
                      wer den Ort liest, will meist hin. */}
                  <Dateiverweis pfad={ort.pfad} />
                </dd>
              </div>
            ))}
          </dl>

          <h3 className="mt-6 mb-2 font-semibold">{eigene.pruefstandName}</h3>
          <p className="mb-3 max-w-prose text-sm text-leise">
            {eigene.pruefstandSatz}
          </p>

          <table className="w-full max-w-3xl text-sm">
            <caption className="sr-only">{eigene.pruefstandName}</caption>
            <thead>
              <tr className="text-left">
                <th scope="col">{eigene.spalteBereich}</th>
                <th scope="col">{eigene.spalteGegenstand}</th>
                <th scope="col">{eigene.spalteStand}</th>
                <th scope="col">{eigene.spalteBeleg}</th>
              </tr>
            </thead>
            <tbody>
              {auskunft.pruefstand.map((punkt) => (
                <tr
                  key={`${punkt.bereich}/${punkt.gegenstand}`}
                  className="align-top"
                >
                  <td className="pr-3 whitespace-nowrap">
                    {worte(eigene.bereiche, punkt.bereich)}
                  </td>
                  <td className="pr-3">
                    {worte(eigene.gegenstaende, punkt.gegenstand)}
                  </td>
                  <td className="pr-3">{worte(eigene.staende, punkt.stand)}</td>
                  <td className="font-mono text-xs break-all">
                    {beleg(punkt)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}
    </Flaeche>
  )
}
