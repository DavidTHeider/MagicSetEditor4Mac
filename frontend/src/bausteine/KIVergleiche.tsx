// **Die Vergleichskarten einer Balanceanalyse** — Werkstück A13.2b von Phase
// 5.3 (#1452).
//
// # Warum die Strategie ganz oben steht
//
// **Weil sie die Frage beantwortet, die jemand vor den Karten stellt**: wie
// kommen gerade diese hierher? Das Issue nennt sie „nachvollziehbar", und
// nachvollziehbar ist sie nur, wenn man sie sieht, bevor man den Vergleich
// liest.
//
// # Und warum keine Zahl dasteht
//
// **Weil es keine gibt.** „Kein objektiver Balancewert aus Ähnlichkeitsscore"
// ist hier keine Prüfung, sondern ein fehlendes Feld: ein Rang ordnet, er
// bewertet nicht ([0103](../../../docs/entscheidungen/0103-wie-zwei-trefferlisten-zusammenkommen.md)).
import { useId } from 'react'

import { useTexte } from '../sprache/kontext'

import type { KIBalanceDTO } from '../antworteneinschaetzung'

interface VergleicheEigenschaften {
  balance: KIBalanceDTO
}

export function KIVergleiche({ balance }: VergleicheEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  // **Von React und nicht fest** (A13.1b, #1452): eine feste Kennung
  // kollidiert, sobald zwei Kästen im selben Dokument stehen — ein
  // axe-Befund, bevor er ein sichtbarer Fehler ist.
  const ueberschrift = useId()

  return (
    <section
      className="flex flex-col gap-2 border-t border-kante pt-3"
      aria-labelledby={ueberschrift}
    >
      <h4 id={ueberschrift} className="text-sm font-medium">
        {eigene.kiEinschaetzungVergleiche}
      </h4>

      <p className="max-w-prose text-sm text-leise">
        {eigene.kiEinschaetzungStrategie}: {balance.strategie}
      </p>

      <p className="text-sm text-leise">
        {eigene.kiEinschaetzungFormat}:{' '}
        {balance.format === ''
          ? eigene.kiEinschaetzungOhneFormat
          : balance.format}
      </p>

      {balance.vergleiche.length === 0 ? (
        <p className="text-sm text-leise">
          {eigene.kiEinschaetzungOhneVergleich}
        </p>
      ) : (
        <ul className="flex flex-col gap-2">
          {balance.vergleiche.map((vergleich) => (
            <li key={vergleich.beleg} className="text-sm">
              <span className="font-medium">{vergleich.beleg}</span>{' '}
              {vergleich.name}
              <ul className="mt-0.5 list-disc pl-5 text-leise">
                {vergleich.unterschiede.map((unterschied) => (
                  <li key={unterschied}>{unterschied}</li>
                ))}
              </ul>
            </li>
          ))}
        </ul>
      )}

      {balance.nichtBetrachtet.length > 0 && (
        <div>
          <h5 className="text-sm font-medium">
            {eigene.kiEinschaetzungSynergien}
          </h5>

          <ul className="mt-1 flex flex-col gap-1">
            {balance.nichtBetrachtet.map((satz) => (
              <li key={satz} className="text-sm text-leise">
                {satz}
              </li>
            ))}
          </ul>
        </div>
      )}
    </section>
  )
}
