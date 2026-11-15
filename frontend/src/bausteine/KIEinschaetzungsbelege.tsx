// **Was eine Einschätzung hergibt** — Werkstück A13.2a von Phase 5.3
// (#1452).
//
// # Ein eigenes Bauteil
//
// **Weil `KIEinschaetzung.tsx` mit ihm über die Zeilengrenze aus §2 ginge** —
// und weil der Schnitt der Frage folgt: dort steht das **Formular**, hier,
// was zurückkam.
//
// # Die Reihenfolge ist eine Aussage
//
// **Vorbehalt, dann Satz, dann was daran nicht stimmt, dann die Quellen.**
// Wer die Befunde unter die Belege setzte, ließe jemanden erst die Quellen
// lesen und danach erfahren, dass die Antwort etwas anderes nennt.
import { KIVergleiche } from './KIVergleiche'

import { useTexte } from '../sprache/kontext'
import type { KIEinschaetzungAntwort } from '../antworteneinschaetzung'

interface BelegeEigenschaften {
  antwort: KIEinschaetzungAntwort
}

export function KIEinschaetzungsbelege({ antwort }: BelegeEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  return (
    <div className="flex flex-col gap-3 border-t border-kante pt-3">
      {/* **Der Vorbehalt zuerst** — er kommt aus `domain` und steht an jeder
          Antwort; ihn ans Ende zu setzen hieße, ihn dorthin zu setzen, wo
          eine lange Antwort abgeschnitten wird. */}
      <p className="max-w-prose text-sm text-leise">{antwort.vorbehalt}</p>

      <p className="max-w-prose whitespace-pre-wrap text-sm">{antwort.satz}</p>

      <p
        className={
          antwort.belegt ? 'text-sm text-leise' : 'text-sm text-fehler'
        }
        role={antwort.belegt ? undefined : 'alert'}
      >
        {antwort.belegt
          ? eigene.kiEinschaetzungBelegt
          : eigene.kiEinschaetzungUnbelegt}
      </p>

      {antwort.befunde.length > 0 && (
        <section aria-labelledby={`befunde-${antwort.art}`}>
          <h4 id={`befunde-${antwort.art}`} className="text-sm font-medium">
            {eigene.kiEinschaetzungBefunde}
          </h4>

          <ul className="mt-1 flex flex-col gap-1">
            {antwort.befunde.map((befund) => (
              <li key={`${befund.art}/${befund.stelle}`} className="text-sm">
                {befund.stelle === ''
                  ? befund.meldung
                  : `${befund.stelle} — ${befund.meldung}`}
              </li>
            ))}
          </ul>
        </section>
      )}

      {antwort.einschraenkungen.length > 0 && (
        <section aria-labelledby={`grenzen-${antwort.art}`}>
          <h4 id={`grenzen-${antwort.art}`} className="text-sm font-medium">
            {eigene.kiEinschaetzungEinschraenkungen}
          </h4>

          <ul className="mt-1 flex flex-col gap-1">
            {antwort.einschraenkungen.map((satz) => (
              <li key={satz} className="text-sm text-leise">
                {satz}
              </li>
            ))}
          </ul>
        </section>
      )}

      {antwort.balance !== null && <KIVergleiche balance={antwort.balance} />}

      {antwort.belege.length > 0 && (
        <section aria-labelledby={`belege-${antwort.art}`}>
          <h4 id={`belege-${antwort.art}`} className="text-sm font-medium">
            {eigene.kiEinschaetzungBelege}
          </h4>

          <ul className="mt-1 flex flex-col gap-2">
            {antwort.belege.map((beleg) => (
              <li key={beleg.kennung} className="text-sm">
                {/* **Die Kennung steht dabei** — sie ist es, die im Satz
                    oben in eckigen Klammern auftaucht. Ohne sie wäre die
                    Fußnote eine Zahl ohne Ziel. */}
                <span className="font-medium">{beleg.kennung}</span>{' '}
                {beleg.titel}
                <span className="text-leise">
                  {' '}
                  · {beleg.quelle}
                  {beleg.quelldatum === '' ? '' : ` · ${beleg.quelldatum}`}
                </span>
                <p className="mt-0.5 max-w-prose text-leise">
                  {beleg.ausschnitt}
                </p>
              </li>
            ))}
          </ul>
        </section>
      )}

      {antwort.korpusstaende.length > 0 && (
        <section aria-labelledby={`staende-${antwort.art}`}>
          <h4 id={`staende-${antwort.art}`} className="text-sm font-medium">
            {eigene.kiEinschaetzungStaende}
          </h4>

          <ul className="mt-1 flex flex-col gap-1">
            {antwort.korpusstaende.map((stand) => (
              <li key={stand.quelle} className="text-sm text-leise">
                {stand.quelle} · {stand.revision}
                {stand.datum === '' ? '' : ` · ${stand.datum}`}
              </li>
            ))}
          </ul>
        </section>
      )}
    </div>
  )
}
