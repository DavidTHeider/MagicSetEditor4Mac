// **Der Befund der offenen Datei, unter dem Editor** — FMT-02, S10.NAT-02 und
// Befund F12 (#1691).
//
// Nach dem Artboard `Werkstatt` steht er im **Zentrum**: unter dem Text, mit
// Zeile und einem Sprung dorthin. Bis FMT-02 stand er im Inspector — und der
// weicht bei engem Fenster zuerst; angestrichen war die Stelle dann, gesagt
// wurde nicht, warum (Befund F12).
import { Knopf } from '../bausteine/Knopf'
import { befundsatz } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Skriptstelle } from '../api'
import type { Skriptbefund } from '../haken/usePufferbefund'

interface Befundeigenschaften {
  // befund ist die letzte Prüfung — undefined, solange keine gelaufen ist.
  befund: Skriptbefund | undefined
  zurStelle: (stelle: Skriptstelle) => void
}

export function Vorlagenbefund({ befund, zurStelle }: Befundeigenschaften) {
  const alle = useTexte()
  const texte = alle.seiten.vorlagen

  if (befund === undefined) {
    return null
  }

  const { stelle } = befund

  if (stelle === null) {
    return <p className="text-sm text-leise">{texte.diagnoseFrei}</p>
  }

  return (
    <div
      role="status"
      className="flex flex-wrap items-start gap-2 rounded-sm border border-kante p-2 text-sm"
    >
      <span aria-hidden="true" className="font-bold text-hinweis">
        !
      </span>
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <span className="text-xs text-leise">
          {einsetzen(texte.diagnoseStelle, {
            zeile: String(stelle.zeile),
            spalte: String(stelle.spalte),
          })}
        </span>
        <span className="text-hinweis">
          {/* **Kennt der Katalog keinen, bleibt der deutsche Satz aus Go
              stehen** — besser unübersetzt als leer (§9). */}
          {befundsatz(befund.schluessel, befund.werte, alle) ?? befund.meldung}
        </span>
      </div>
      <Knopf
        gewicht="neben"
        aufKlick={() => {
          zurStelle(stelle)
        }}
      >
        {texte.zurStelle}
      </Knopf>
    </div>
  )
}
