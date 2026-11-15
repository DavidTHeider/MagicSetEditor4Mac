// **Das Ergebnis der Skriptkonsole** — Quelltext, Art, Dauer, Vorbehalte, und
// eine Abbildung als Baum (Werkstück U13 von Phase 5.1, #1424).
import type { KonsolenEintrag } from '../antworten'
import type { Konsolenergebnis } from '../haken/useSkriptkonsole'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

interface ErgebnisEigenschaften {
  ergebnis: Konsolenergebnis
}

// Zweig zeigt einen Eintrag — und darunter seine eigenen, wenn er welche hat.
//
// **Immer aufgeklappt.** `card` und `set` sind Abbildungen über Abbildungen,
// und wer die Konsole fragt, will sehen, was drin ist — nicht erst klicken.
function Zweig({ eintrag }: { eintrag: KonsolenEintrag }) {
  const verzweigt = eintrag.eintraege.length > 0

  return (
    <li
      role="treeitem"
      aria-selected={false}
      aria-expanded={verzweigt ? true : undefined}
    >
      <span className="flex flex-wrap items-baseline gap-x-2">
        <span className="font-medium">{eintrag.name}</span>
        {!verzweigt && <code className="text-sm">{eintrag.quelltext}</code>}
        <span className="text-xs text-leise">{eintrag.art}</span>
      </span>
      {verzweigt && (
        <ul role="group" className="ml-4 border-l border-kante pl-3">
          {eintrag.eintraege.map((kind) => (
            <Zweig key={kind.name} eintrag={kind} />
          ))}
        </ul>
      )}
    </li>
  )
}

export function Skriptergebnis({ ergebnis }: ErgebnisEigenschaften) {
  const eigene = useTexte().seiten.karten.konsole

  return (
    <section className="flex flex-col gap-2">
      <h3 className="text-sm font-medium">{eigene.ergebnis}</h3>

      {/* **Der Quelltext und nicht der Text** — so zeigt MSE es
          (`src/cli/cli_main.cpp`, Zeile 234, Fassung f9d9356 vom
          2021-07-06): wer eine Konsole benutzt, will eine Zeichenkette von
          einem Namen unterscheiden. */}
      {/* **`text-sm/6` und nicht `/60`**: die alte Klasse setzte die Zeilenhöhe
          auf 15 rem, und ein einzeiliges Ergebnis stand in einem leeren Kasten
          — gesehen in der Sichtprüfung zu U13 (#1424). Ein langer Quelltext
          bricht um, statt die Seite zu verbreitern. */}
      <output className="rounded-sm bg-betont/60 px-2 py-1 font-mono text-sm/6 break-all whitespace-pre-wrap">
        {ergebnis.quelltext}
      </output>

      <p className="flex gap-3 text-sm text-leise">
        <span>{einsetzen(eigene.art, { art: ergebnis.art })}</span>
        <span>{einsetzen(eigene.dauer, { ms: String(ergebnis.dauerMs) })}</span>
      </p>

      {ergebnis.eintraege.length > 0 && (
        <ul role="tree" aria-label={eigene.baum} className="text-sm">
          {ergebnis.eintraege.map((eintrag) => (
            <Zweig key={eintrag.name} eintrag={eintrag} />
          ))}
        </ul>
      )}

      {ergebnis.befunde.length > 0 && (
        <>
          <h3 className="text-sm font-medium">{eigene.vorbehalte}</h3>
          <ul className="flex flex-col gap-1 text-sm text-leise">
            {ergebnis.befunde.map((befund) => (
              <li key={befund}>{befund}</li>
            ))}
          </ul>
        </>
      )}
    </section>
  )
}
