// **Die Vorschau vor dem Einlesen eines Sets** — FMT-03.AK-01
// (#1692).
//
// Was sie tut, steht in `useSetvorschau`; hier, wie sie aussieht: **eine**
// führende Handlung (*Einlesen*, bei einer schon offenen Datei
// *Aufschlagen*), daneben das Verwerfen — und darüber, warum sie überhaupt
// dasteht: Stellen, die sich nicht lesen lassen, und Pakete, die fehlen.
import type { Setvorschaulauf } from '../haken/useSetvorschau'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'
import { Flaeche } from '../bausteine/Flaeche'
import { Knopf } from '../bausteine/Knopf'
import { abhaengigkeitssatz } from './abhaengigkeitssatz'

interface Eigenschaften {
  lauf: Setvorschaulauf
  angehalten: boolean
}

export function Setvorschau({ lauf, angehalten }: Eigenschaften) {
  const texte = useTexte()
  const t = texte.seiten.sets.vorschau

  if (lauf.fehler !== undefined) {
    return (
      <p role="alert" className="mt-3 text-fehler">
        {texte.bausteine.zustaende.fehler} {lauf.fehler}
      </p>
    )
  }

  const vorschau = lauf.vorschau
  if (vorschau === undefined) {
    return null
  }

  const karten = zaehlwort(vorschau.karten, texte.allgemein.zaehlwoerter.karten)

  return (
    <div className="mt-3">
      <Flaeche
        ueberschrift={einsetzen(t.ueberschrift, { name: vorschau.name })}
      >
        <p className="text-leise">
          {vorschau.stil === ''
            ? einsetzen(t.umfangOhneStil, { karten, spiel: vorschau.spiel })
            : einsetzen(t.umfang, {
                karten,
                spiel: vorschau.spiel,
                stil: vorschau.stil,
              })}
        </p>

        {vorschau.schonOffen !== '' && (
          <p className="mt-2 max-w-prose">{t.schonOffen}</p>
        )}

        {vorschau.befunde.length > 0 && (
          <>
            <p className="mt-2 max-w-prose">{t.befunde}</p>
            <ul className="mt-1 space-y-1 font-mono text-sm">
              {vorschau.befunde.map((befund) => (
                <li key={befund}>{befund}</li>
              ))}
            </ul>
          </>
        )}

        {vorschau.bedarf.length > 0 && (
          <>
            <p className="mt-2 max-w-prose">{t.bedarf}</p>
            <ul className="mt-1 space-y-1">
              {vorschau.bedarf.map((befund) => (
                <li key={`${befund.paket}/${befund.braucht}/${befund.art}`}>
                  {abhaengigkeitssatz(befund, texte)}
                </li>
              ))}
            </ul>
          </>
        )}

        <div className="mt-3 flex flex-wrap gap-2">
          <Knopf
            gewicht="haupt"
            angehalten={angehalten}
            aufKlick={() => {
              lauf.einlesen(vorschau.pfad)
            }}
          >
            {vorschau.schonOffen === '' ? t.einlesen : t.aufschlagen}
          </Knopf>
          <Knopf aufKlick={lauf.verwerfen}>{t.verwerfen}</Knopf>
        </div>
      </Flaeche>
    </div>
  )
}
