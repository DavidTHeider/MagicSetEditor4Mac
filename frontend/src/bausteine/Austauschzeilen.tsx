// **Der Unterschied als Tabelle** — Werkstück O3 von Phase 5.2
// (#1427).
//
// **Eine eigene Datei**, weil `Einstellungsaustausch.tsx` mit ihr über die
// Zeilengrenze ginge (§2). Der Schnitt folgt der Frage: dort steht, **was**
// geschieht, hier, **wie es aussieht**.
//
// # Warum eine Tabelle und keine Liste
//
// **Weil vier Angaben je Zeile in einer Spalte übereinander stünden.**
// Schlüssel, vorher, nachher, Herkunft — das ist eine Tabelle, und eine
// Tabelle sagt einem Vorleseprogramm auch, welche Spalte gerade dran ist.
import { useId } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Knopf } from './Knopf'

import type { Einfuhrbefund, Einfuhrzeile, Einstellungssicht } from '../api'

// zeilenname gibt den Namen, unter dem eine Zeile gewählt wird.
//
// **`<ebene>:<schluessel>`, weil ein Schlüssel in zwei Abschnitten stehen
// kann**: `sprache` gehört dem Menschen **und** steht als Übersteuerung im
// Profil. Eine Auswahl über den Schlüssel allein träfe beide oder die falsche.
export function zeilenname(zeile: Einfuhrzeile): string {
  return `${zeile.ebene}:${zeile.schluessel}`
}

// befundname gibt den Ort eines Befunds.
//
// **Ein Abschnitt ohne Schlüssel ist ein Befund über den Abschnitt selbst** —
// „diesen kennt dieses Programm nicht". Ein Doppelpunkt ohne etwas dahinter
// sähe aus, als fehlte etwas.
function befundname(befund: Einfuhrbefund): string {
  return befund.schluessel === ''
    ? befund.ebene
    : `${befund.ebene}:${befund.schluessel}`
}

interface Zeileneigenschaften {
  sicht: Einstellungssicht
  abgewaehlt: Record<string, boolean>
  umschalten: (name: string) => void
  angehalten: boolean
  uebernehmen: () => void
  verwerfen: () => void
}

export function Austauschzeilen({
  sicht,
  abgewaehlt,
  umschalten,
  angehalten,
  uebernehmen,
  verwerfen,
}: Zeileneigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen.austausch
  const tabellenID = useId()

  // herkunftsnamen übersetzt die drei Werte.
  //
  // **Der Typ sagt die Wahrheit**: ein Zugriff über `as keyof` behauptete,
  // jede Herkunft habe einen Text, und machte aus einer fehlenden ein stilles
  // `undefined` — die Zelle bliebe leer, ohne zu klagen.
  const herkunftsnamen: Record<string, string | undefined> = eigene.herkunft

  return (
    <section
      aria-labelledby={tabellenID}
      className="flex flex-col gap-3 rounded-md border border-kante p-3"
    >
      <h3 id={tabellenID} className="text-base font-medium">
        {eigene.ueberschrift}
      </h3>

      {sicht.zeilen.length === 0 ? (
        <p className="text-leise">{eigene.keineAenderung}</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-kante">
                <th scope="col" className="py-1 pr-2">
                  {eigene.spalteUebernehmen}
                </th>
                <th scope="col" className="py-1 pr-2">
                  {eigene.spalteSchluessel}
                </th>
                <th scope="col" className="py-1 pr-2">
                  {eigene.spalteVorher}
                </th>
                <th scope="col" className="py-1 pr-2">
                  {eigene.spalteNachher}
                </th>
                <th scope="col" className="py-1">
                  {eigene.spalteHerkunft}
                </th>
              </tr>
            </thead>

            <tbody>
              {sicht.zeilen.map((zeile) => {
                const name = zeilenname(zeile)

                return (
                  <tr key={name} className="border-b border-kante">
                    <td className="py-1 pr-2">
                      <input
                        type="checkbox"
                        checked={!abgewaehlt[name]}
                        disabled={angehalten}
                        aria-label={einsetzen(eigene.zeileWaehlen, {
                          schluessel: name,
                        })}
                        onChange={() => {
                          umschalten(name)
                        }}
                        className="size-4 accent-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                      />
                    </td>

                    <td className="py-1 pr-2 font-mono text-xs">{name}</td>
                    <td className="py-1 pr-2">{zeile.vorher}</td>
                    <td className="py-1 pr-2">
                      {zeile.nachher}

                      {/* **Der Neustartbedarf an der Zeile** — OPS-01.AK-02
                          (#1718). Derselbe Satz wie an der Einstellung
                          selbst, damit Vorschau und Seite nicht zweierlei
                          sagen; `sofort` und leer bleiben still. */}
                      {zeile.wirkung === 'neustart' && (
                        <span className="block text-xs text-leise">
                          {texte.seiten.einstellungen.wirktNachNeustart}
                        </span>
                      )}
                    </td>
                    <td className="py-1">
                      {herkunftsnamen[zeile.herkunft] ?? zeile.herkunft}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}

      {sicht.unveraendert > 0 && (
        <p className="text-leise">
          {einsetzen(eigene.unveraendert, {
            zahl: String(sicht.unveraendert),
          })}
        </p>
      )}

      {sicht.befunde.length > 0 && (
        <div className="flex flex-col gap-1">
          <h4 className="text-sm font-medium">{eigene.befunde}</h4>

          <ul className="flex list-disc flex-col gap-1 pl-5 text-leise">
            {sicht.befunde.map((befund) => (
              <li key={befundname(befund)}>
                <span className="font-mono text-xs">{befundname(befund)}</span>
                <span className="ml-2">{befund.grund}</span>
              </li>
            ))}
          </ul>
        </div>
      )}

      <div className="flex flex-wrap gap-2">
        <Knopf
          gewicht="haupt"
          angehalten={angehalten || sicht.zeilen.length === 0}
          aufKlick={uebernehmen}
        >
          {eigene.uebernehmen}
        </Knopf>

        <Knopf gewicht="leise" angehalten={angehalten} aufKlick={verwerfen}>
          {eigene.verwerfen}
        </Knopf>
      </div>
    </section>
  )
}
