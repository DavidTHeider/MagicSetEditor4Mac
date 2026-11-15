// **Was aus einem Kartenfoto wurde** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Warum ein eigenes Bauteil
//
// **Weil `KIKartenerkennung.tsx` mit ihm über die Zeilengrenze aus §2 ginge**
// — und weil der Schnitt der Frage folgt: dort steht, **wie** gefragt wird,
// hier, **was die Antwort sagt**.
//
// # Und warum das Gelesene über den Kandidaten steht
//
// **Weil es die Begründung ist.** Die Kandidaten kommen aus einer Suche über
// die gelesenen Stellen; wer sie ohne sie läse, sähe eine Liste ohne Grund —
// und könnte nicht sehen, dass die Fußzeile abgeschnitten war.
import { useId } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { KIErkennungAntwort } from '../antwortenerkennung'

// artName gibt den Namen einer Evidenzart — roh, wo der Katalog sie nicht
// kennt.
//
// # Warum an einer Stelle und nicht an dreien
//
// **Weil nur eine der drei Stellen eine unbekannte Art sehen kann.** Die
// Beobachtungen kommen vom Modell und dürfen alles enthalten — dafür gibt es
// [domain.BefundErfundeneEvidenz]. Was nicht gelesen wurde und was einen
// Kandidaten stützt, bildet `internal/domain` aus seinen sechs; ein `?? art`
// dort wäre ein Zweig, den kein Aufruf betreten kann (§3).
//
// **Und roh statt weggelassen**: eine erfundene Art ist ein Befund, und der
// nützt nichts, wenn die Zeile dazu verschwindet.
function artName(namen: Record<string, string>, art: string): string {
  return namen[art] ?? art
}

interface BefundEigenschaften {
  antwort: KIErkennungAntwort
}

export function KIErkennungsbefund({ antwort }: BefundEigenschaften) {
  const eigene = useTexte().seiten.einstellungen

  // **Eine Kennung je Einbau und keine feste** (#1553): zwei Kästen auf
  // derselben Seite trügen sonst dieselbe `id`, und eine Beschriftung zeigte
  // auf den falschen.
  const gelesenID = useId()
  const offenID = useId()
  const kandidatenID = useId()

  // **Ein geschlossener Dialog ist keine Antwort** — und sieht ohne diesen
  // Zweig wie „kein Druck passt" aus.
  if (antwort.abgebrochen) {
    return <p className="text-sm text-leise">{eigene.kiErkennungLeer}</p>
  }

  return (
    <div className="flex flex-col gap-3">
      <section aria-labelledby={gelesenID}>
        <h4 id={gelesenID} className="text-sm font-medium">
          {eigene.kiErkennungGelesen}
        </h4>

        <ul className="mt-1 flex flex-col gap-1">
          {/* **Der Schlüssel trägt die Stelle mit** — dieselbe Art darf
              zweimal vorkommen: eine geteilte Karte hat zwei Namen, und ein
              Modell, das beide liest, hat richtig gelesen. Ein Schlüssel aus
              der Art allein wäre dann doppelt. */}
          {antwort.beobachtungen.map((beobachtung, stelle) => (
            <li
              key={`${String(stelle)}/${beobachtung.art}`}
              className="text-sm"
            >
              {artName(eigene.kiErkennungArten, beobachtung.art)}
              {': '}
              {/* **Drei Fälle und nicht zwei** (§8): unlesbar, gelesen und
                  leer, gelesen und gefüllt. Wer die ersten beiden
                  zusammenwürfe, sagte über eine abgeschnittene Fußzeile
                  dasselbe wie über ein leeres Feld. */}
              {!beobachtung.lesbar ? (
                <span className="text-leise">{eigene.kiErkennungUnlesbar}</span>
              ) : beobachtung.wert === '' ? (
                <span className="text-leise">{eigene.kiErkennungOhneWert}</span>
              ) : (
                /* **Auch der gelesene Wert steht in einem eigenen Knoten** —
                   wie die beiden Fälle darüber. Sonst hinge er als nackter
                   Text zwischen Beschriftung und Zeilenende, und wer ihn
                   sucht, sucht die ganze Zeile. */
                <span>{beobachtung.wert}</span>
              )}
            </li>
          ))}
        </ul>
      </section>

      {antwort.nichtGelesen.length > 0 && (
        <section aria-labelledby={offenID}>
          <h4 id={offenID} className="text-sm font-medium">
            {eigene.kiErkennungNichtGelesen}
          </h4>

          <p className="mt-1 text-sm text-leise">
            {antwort.nichtGelesen
              .map((art) => artName(eigene.kiErkennungArten, art))
              .join(', ')}
          </p>
        </section>
      )}

      {antwort.befunde.length > 0 && (
        <ul className="flex flex-col gap-1">
          {antwort.befunde.map((befund) => (
            <li
              key={`${befund.art}/${befund.stelle}`}
              role="alert"
              className="text-sm text-fehler"
            >
              {befund.meldung}
            </li>
          ))}
        </ul>
      )}

      <section aria-labelledby={kandidatenID}>
        <h4 id={kandidatenID} className="text-sm font-medium">
          {eigene.kiErkennungKandidaten}
        </h4>

        {antwort.kandidaten.length === 0 ? (
          <p className="mt-1 max-w-prose text-sm text-leise">
            {eigene.kiErkennungOhneKandidat}
          </p>
        ) : (
          <>
            {/* **Die Zahl der passenden Drucke ist die Unsicherheit** — sie
                ist gezählt und nicht geschätzt, und sie steht über der Liste,
                weil die Liste kürzer sein darf als sie. */}
            <p className="mt-1 text-sm text-leise">
              {antwort.eindeutig
                ? eigene.kiErkennungEinDruck
                : einsetzen(eigene.kiErkennungDrucke, {
                    drucke: String(antwort.drucke),
                  })}
              {antwort.weitere ? ` ${eigene.kiErkennungWeitere}` : ''}
            </p>

            <ul className="mt-1 flex flex-col gap-2">
              {antwort.kandidaten.map((kandidat) => (
                <li key={kandidat.kennung} className="text-sm">
                  <span className="font-medium">{kandidat.name}</span>
                  {` — ${kandidat.setname} (${kandidat.setcode}) ${kandidat.nummer}`}

                  <span className="block text-leise">
                    {kandidat.gestuetzt.length === 0
                      ? eigene.kiErkennungOhneStuetze
                      : einsetzen(eigene.kiErkennungGestuetzt, {
                          arten: kandidat.gestuetzt
                            .map((art) => artName(eigene.kiErkennungArten, art))
                            .join(', '),
                        })}
                  </span>
                </li>
              ))}
            </ul>
          </>
        )}
      </section>

      {/* **Der Satz steht immer da** — auch bei genau einem Treffer. Gerade
          dann: ein eindeutiger Kandidat ist die Stelle, an der jemand auf den
          Knopf drückte, den es nicht gibt. */}
      <p className="max-w-prose text-sm text-leise">
        {eigene.kiErkennungNichtsAngelegt}
      </p>

      <p className="text-sm text-leise">
        {einsetzen(eigene.kiErkennungModell, {
          modell: antwort.modell,
          anbieter: antwort.anbieter,
        })}
        {antwort.alter === ''
          ? ''
          : ` · ${einsetzen(eigene.kiErkennungStand, { alter: antwort.alter })}`}
      </p>
    </div>
  )
}
