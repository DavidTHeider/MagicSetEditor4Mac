// **Meinen Code zeigen** — NET-01
// (#1705).
//
// **Der Code steht in einem Feld zum Lesen und nicht in einem Absatz**:
// markieren, kopieren und in eine Nachricht legen soll ohne Rest gehen, und
// ein Feld nimmt beim Markieren keine Nachbarn mit.
import { useId, useState } from 'react'

import { Aktionsknopf } from './Aktionsknopf'
import { Identitaetanlegen } from './Identitaetanlegen'
import { Knopf } from './Knopf'
import { Teilen } from './Teilen'
import { eingabeklasse } from './Beschriftetesfeld'
import { alsZeitpunkt } from '../zeitpunkt'
import { einsetzen } from '../sprache/einsetzen'
import { inDieZwischenablage } from '../zwischenablage'
import { useDarstellung, useTexte, useZeitzone } from '../sprache/kontext'
import { zwecksatz } from '../zwecksatz'

import type { Einladungstausch } from '../haken/useEinladungstausch'

interface Einladungsausgabeeigenschaften {
  zwecke: string[]
  tausch: Einladungstausch
}

export function Einladungsausgabe({
  zwecke,
  tausch,
}: Einladungsausgabeeigenschaften) {
  const texte = useTexte()
  const technik = texte.seiten.technik
  const eigene = technik.einladung
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()
  const namensfeld = useId()
  const zweckfeld = useId()
  const hinweis = useId()
  const codefeld = useId()

  const [name, setzeName] = useState('')
  const [zweck, setzeZweck] = useState('handel')
  const [kopiert, setzeKopiert] = useState(false)

  // ausloeser ist der Fehler dieses Schritts — ihn bekommt das Anlegen der
  // Identität mit, damit sein Ende nur zu ihm gehört (§7, Regel 2).
  const ausloeser =
    tausch.fehler?.ort === 'erzeugen' ? tausch.fehler : undefined
  const fehler = ausloeser?.text
  const ausgestellt = tausch.ausgestellt

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap items-end gap-3">
        <label className="flex flex-col gap-1 text-sm" htmlFor={namensfeld}>
          {eigene.name}
          <input
            id={namensfeld}
            className={eingabeklasse}
            value={name}
            aria-describedby={hinweis}
            aria-invalid={fehler !== undefined}
            onChange={(ereignis) => {
              setzeName(ereignis.target.value)
            }}
          />
        </label>

        <label className="flex flex-col gap-1 text-sm" htmlFor={zweckfeld}>
          {eigene.zweck}
          <select
            id={zweckfeld}
            className={eingabeklasse}
            value={zweck}
            onChange={(ereignis) => {
              setzeZweck(ereignis.target.value)
            }}
          >
            {zwecke.map((moeglich) => (
              <option key={moeglich} value={moeglich}>
                {zwecksatz(technik, moeglich)}
              </option>
            ))}
          </select>
        </label>

        {/* **Führend, bis es einen Code gibt** (UXR-16, #1924): danach ist
            das Weitergeben dran — Kopieren und „Teilen …" —, und ein zweiter
            Code ersetzte den ersten. */}
        <Knopf
          gewicht={ausgestellt === undefined ? 'haupt' : 'neben'}
          angehalten={tausch.laeuft('erzeugen') || name.trim() === ''}
          aufKlick={() => {
            setzeKopiert(false)
            tausch.erzeugen(name, zweck)
          }}
        >
          {eigene.erzeugen}
        </Knopf>
      </div>

      <p id={hinweis} className="text-sm text-leise">
        {eigene.nameHinweis}
      </p>

      {/* **Fehlt die eigene Identität, steht ihr Anlegen hier** (G01): der
          Kasten, den die Meldung nannte, steht nicht in dieser Schublade. */}
      {ausloeser?.ohneIdentitaet === true ? (
        <Identitaetanlegen
          angelegt={() => {
            tausch.identitaetAngelegt(ausloeser)
          }}
        />
      ) : (
        fehler !== undefined && (
          <p role="alert" className="text-sm text-fehler">
            {fehler}
          </p>
        )
      )}

      {ausgestellt !== undefined && (
        <div className="flex flex-col gap-1">
          <label
            className="flex flex-col gap-1 text-sm font-medium"
            htmlFor={codefeld}
          >
            {eigene.code}
            <textarea
              id={codefeld}
              readOnly
              rows={3}
              value={ausgestellt.code}
              className="rounded-sm border border-kante bg-transparent p-2 font-mono text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            />
          </label>

          <p className="text-sm text-leise">
            {einsetzen(eigene.codeHinweis, {
              zeitpunkt: alsZeitpunkt(
                ausgestellt.ablauf,
                darstellung,
                zeitzone,
              ),
            })}
          </p>

          <div className="flex items-center gap-2">
            <Aktionsknopf
              aktion="kopieren"
              aufKlick={() => {
                void inDieZwischenablage(ausgestellt.code).then(setzeKopiert)
              }}
            />

            {kopiert && (
              <span role="status" className="text-sm text-leise">
                {eigene.kopiert}
              </span>
            )}
          </div>

          {/* **Die Einladung über Mail oder Nachrichten** — NET-04.AK-01. */}
          <Teilen
            key={ausgestellt.code}
            art="einladung"
            bezug={ausgestellt.code}
          />
        </div>
      )}
    </div>
  )
}
