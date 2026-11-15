// **Welche Exportvorlage** — aus `SetVorlagenausfuhr.tsx` gelöst, als sie
// eine Suche bekam (UXR-09,
// #1917);
// die Datei liefe sonst über die Zeilengrenze aus §2. Der Schnitt folgt der
// Frage: dort steht, was die Ausfuhr **tut**, hier, wie eine Vorlage
// **gewählt** wird.
//
// # Ab fünfzehn mit Suche
//
// **Gezählt am 2026-10-02**: `full-magic-pack` bringt für Magic achtzehn
// Vorlagen mit, und R7 aus `docs/UIREVIEW.md` verlangt ab fünfzehn Werten
// eine Auswahl mit Suche. **Die Gestalt kommt aus UXR-07** (Stilwahl): ein
// natives Auswahlfeld — in WKWebView belegt, mit dem Menü von macOS, der
// Tastatur und der Vorlesung —, davor ein Feld, das seine Einträge eingrenzt.
// Gesucht wird über den Namen **und** die Kennung des Pakets; was gewählt
// ist, bleibt angeboten, sonst fiele die Auswahl still auf ihren ersten
// Eintrag.
import { useId, useState } from 'react'

import { suchgrenze } from '../bausteine/Ausgabeauswahl'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import type { ExportvorlageDTO } from '../vorlagen'

interface VorlagenwahlEigenschaften {
  vorlagen: readonly ExportvorlageDTO[]
  gewaehlt: ExportvorlageDTO
  waehlen: (vorlage: ExportvorlageDTO) => void
}

const feld =
  'mt-1 block rounded-sm border border-kante bg-flaeche px-2 py-1 ' +
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Vorlagenwahl({
  vorlagen,
  gewaehlt,
  waehlen,
}: VorlagenwahlEigenschaften) {
  const texte = useTexte()
  const keine = useId()
  const [suche, setzeSuche] = useState('')

  const gesucht = suche.trim().toLowerCase()
  const passt = (vorlage: ExportvorlageDTO) =>
    vorlage.name.toLowerCase().includes(gesucht) ||
    vorlage.paket.toLowerCase().includes(gesucht)
  const treffer = vorlagen.filter(passt)
  const angeboten = vorlagen.filter(
    (vorlage) => vorlage === gewaehlt || passt(vorlage),
  )

  return (
    <div className="flex max-w-md flex-col gap-2">
      {vorlagen.length >= suchgrenze && (
        <label className="block">
          <span className="text-sm text-leise">
            {texte.seiten.ausgeben.vorlagenEingrenzen}
          </span>
          <input
            type="search"
            value={suche}
            onChange={(ereignis) => {
              setzeSuche(ereignis.target.value)
            }}
            className={feld}
          />
        </label>
      )}

      <label className="block">
        <span className="text-sm text-leise">
          {texte.seiten.sets.vorlageWaehlen}
        </span>
        <select
          value={gewaehlt.paket}
          aria-describedby={treffer.length === 0 ? keine : undefined}
          onChange={(ereignis) => {
            const naechste = vorlagen.find(
              (eine) => eine.paket === ereignis.target.value,
            )

            // **Die Auswahl kennt nur, was in ihr steht.** `find` gibt
            // deshalb immer etwas — der Zweig darunter ist die Typprüfung,
            // nicht ein Fall.
            /* v8 ignore next 3 -- unerreichbar: der Wert kommt aus dieser Liste */
            if (naechste) {
              waehlen(naechste)
            }
          }}
          className={feld}
        >
          {angeboten.map((eine) => (
            <option key={eine.paket} value={eine.paket}>
              {eine.name}
            </option>
          ))}
        </select>
      </label>

      {treffer.length === 0 && (
        <p id={keine} className="text-sm text-leise">
          {einsetzen(texte.bausteine.listenansicht.nichtsPasst, {
            suche: suche.trim(),
          })}
        </p>
      )}
    </div>
  )
}
