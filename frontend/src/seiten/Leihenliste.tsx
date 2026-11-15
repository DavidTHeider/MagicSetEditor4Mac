// **Die Leihen eines Profils, mit dem, was mit ihnen geht** — aus
// `Deckleihe.tsx` herausgelöst, als UXR-F15
// (#2030)
// die Seite über die Zeilengrenze aus §2 hob.
//
// **Der Schnitt folgt der Frage**: dort steht, wie eine Leihe ausgestellt und
// angenommen wird; hier, was der Stand zeigt und welcher Knopf an welcher
// Zeile steht — *widerrufen* an der verliehenen, *zurückgeben* an der
// geliehenen.
import { Knopf } from '../bausteine/Knopf'
import { useTexte } from '../sprache/kontext'

import type { useDeckleihe } from '../haken/useDeckleihe'
import type { Leihe } from '../apileihe'

// Leihenliste zeigt die drei Zustände: ungeholt, leer, voll.
export function Leihenliste({
  lauf,
}: {
  lauf: ReturnType<typeof useDeckleihe>
}) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  // **Der Grund, aus dem die Liste nach einem geglückten Schritt fehlt**
  // (#1960) — an ihr und nicht oben am Schritt.
  if (lauf.listenfehler !== undefined) {
    return (
      <p role="alert" className="text-sm text-fehler">
        {lauf.listenfehler}
      </p>
    )
  }

  if (lauf.leihen === undefined) {
    return null
  }

  if (lauf.leihen.length === 0) {
    return <p className="text-sm text-leise">{eigene.leiheKeine}</p>
  }

  return (
    <ul className="flex flex-col gap-2 text-sm">
      {lauf.leihen.map((leihe) => (
        <li key={leihe.kennung}>
          <Leihzeile leihe={leihe} lauf={lauf} />
        </li>
      ))}
    </ul>
  )
}

// Leihzeile zeigt eine Leihe samt dem, was mit ihr geht.
function Leihzeile({
  leihe,
  lauf,
}: {
  leihe: Leihe
  lauf: ReturnType<typeof useDeckleihe>
}) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  return (
    <div className="flex flex-col gap-1">
      <span className="flex flex-wrap items-baseline gap-3">
        <span>
          {leihe.verliehen ? eigene.leiheVerliehen : eigene.leiheGeliehen}{' '}
          <span className="font-mono text-xs">
            {leihe.verliehen ? leihe.entleiher : leihe.verleiher}
          </span>
        </span>

        <span className="text-leise">
          {leihe.karten} {eigene.leiheKarten}
        </span>

        <span className="text-leise">
          {eigene.leiheBis} {leihe.bis}
        </span>

        <span className="text-leise">
          {leihe.verbraucht === ''
            ? eigene.leiheOffen
            : `${eigene.leiheVerbraucht} ${leihe.verbraucht}`}
        </span>

        {leihe.widerrufen && (
          <span className="text-leise">{eigene.leiheZurueckgenommen}</span>
        )}
      </span>

      <span className="text-xs text-leise">
        {leihe.modus === 'zuende'
          ? eigene.leiheModusZuende
          : eigene.leiheModusAbbruch}
      </span>

      {!leihe.widerrufen && (
        <div>
          <Knopf
            gewicht="leise"
            angehalten={lauf.laeuft}
            aufKlick={() => {
              if (leihe.verliehen) {
                lauf.widerrufen(leihe.kennung)
              } else {
                lauf.zurueckgeben(leihe.kennung)
              }
            }}
          >
            {leihe.verliehen
              ? eigene.leiheWiderrufen
              : eigene.leiheZurueckgeben}
          </Knopf>
        </div>
      )}
    </div>
  )
}
