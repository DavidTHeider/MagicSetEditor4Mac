// **Das Geheimnis eines Anmeldeanbieters** —
// #1401,
// Befund 3 des Zweitblicks über Phase 4.12.
//
// # Wofür es das gibt
//
// **GitHub verlangt neben der Kennung ein `client_secret`**, und bis hierher
// schrieb es keine Bindung: wer sich über GitHub anmelden wollte, musste die
// Schlüsselbundverwaltung öffnen. §6 sagt, ein leerer Zustand erklärt den
// nächsten Schritt — nicht: er verweist auf ein anderes Programm.
//
// # Was hier nie steht
//
// **Der Wert.** Gelesen wird nur, **ob** eines hinterlegt ist — dieselbe
// Zusage wie bei `Zugaenge` nebenan, und aus demselben Grund.
//
// # Wer keines darf, bekommt kein Feld
//
// **Microsoft verbietet ein Geheimnis für öffentliche Klienten.** Ein Feld,
// das man ausfüllen kann und das die Anmeldung danach scheitern lässt, ist
// schlimmer als keines; hier steht stattdessen der Satz, warum.
import { useId, useState } from 'react'

import { api } from '../api'
import { auspacken, type Umschlag } from '../umschlag'
import { meldungAus } from '../meldung'
import { Aktionsknopf } from './Aktionsknopf'
import { Entscheidung } from './Entscheidung'
import { Knopf } from './Knopf'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import type { AnbietergeheimnisDTO } from '../apizugaenge'

const feldklassen =
  'rounded-md border border-kante px-2 py-1 focus-visible:outline ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'

export function Anbietergeheimnisse() {
  const texte = useTexte()

  const [eingaben, setzeEingaben] = useState<Record<string, string>>({})

  const stand = useLaden(holen, [])

  if (stand.fehler !== undefined) {
    return (
      <p role="alert" className="mt-4 text-fehler">
        {texte.bausteine.zustaende.fehler} {stand.fehler}
      </p>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {(stand.daten ?? []).map((eintrag) => (
        <Geheimniszeile
          key={eintrag.anbieter}
          eintrag={eintrag}
          wert={eingaben[eintrag.anbieter] ?? ''}
          setzeWert={(wert) => {
            setzeEingaben((bisher) => ({ ...bisher, [eintrag.anbieter]: wert }))
          }}
          nachladen={stand.nachladen}
        />
      ))}
    </div>
  )
}

// holen liest den Stand.
async function holen(): Promise<AnbietergeheimnisDTO[]> {
  const antwort = await api().Anbietergeheimnisse()

  return auspacken(antwort, () => antwort.anbieter)
}

// Geheimniszeile ist ein Anbieter samt seinen Knöpfen.
function Geheimniszeile({
  eintrag,
  wert,
  setzeWert,
  nachladen,
}: {
  eintrag: AnbietergeheimnisDTO
  wert: string
  setzeWert: (wert: string) => void
  nachladen: () => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen
  const feldID = useId()
  const fehlerID = useId()
  const vorgang = useVorgang()

  const [fehler, setzeFehler] = useState<string>()

  // **Gelöscht wird erst nach einer Frage** — UX-10.AK-04 (#1682), derselbe
  // Grund wie bei den Zugängen: zurück gibt das Geheimnis nur der Anbieter.
  const [fragt, setzeFragt] = useState(false)

  const namen: Record<string, string | undefined> = eigene.anbietername
  const name = namen[eintrag.anbieter] ?? eintrag.anbieter

  if (eintrag.verboten) {
    return (
      <p className="max-w-prose text-sm text-leise">
        {name}: {eigene.geheimnisVerboten}
      </p>
    )
  }

  // **Ein Vorgang je Anbieter** (§7, Regel 4) — und **einer** für Eintragen
  // und Entfernen: sie schließen einander aus.
  const zug = async (arbeit: () => Promise<Umschlag>, danach: () => void) => {
    await vorgang.starte(eintrag.anbieter, async () => {
      try {
        const antwort = await arbeit()

        auspacken(antwort, () => undefined)
        setzeFehler(undefined)
        danach()

        // **Geschrieben ist geschrieben** (§7, Regel 3).
        nachladen()
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  const stand = eintrag.hinterlegt
    ? eigene.zugangLiegt
    : eintrag.pflicht
      ? eigene.geheimnisPflicht
      : eigene.zugangFehlt

  return (
    <div className="flex flex-col gap-1">
      <div className="flex flex-wrap items-end gap-2">
        <label htmlFor={feldID} className="flex flex-col gap-1">
          <span className="text-sm">
            {name}: {stand}
          </span>

          {/* **Ein Passwortfeld.** Wer das Geheimnis eintippt, soll es nicht
              auf dem Bildschirm stehen haben. */}
          <input
            id={feldID}
            type="password"
            value={wert}
            aria-label={`${eigene.geheimnis}: ${name}`}
            aria-describedby={fehler === undefined ? undefined : fehlerID}
            onChange={(ereignis) => {
              setzeWert(ereignis.target.value)
            }}
            className={feldklassen}
          />
        </label>

        <Knopf
          bezeichnung={`${eigene.zugangSetzen}: ${name}`}
          angehalten={vorgang.laeuft(eintrag.anbieter)}
          aufKlick={() => {
            void zug(
              () => api().AnbietergeheimnisSetzen(eintrag.anbieter, wert),
              () => {
                setzeWert('')
              },
            )
          }}
        >
          {eigene.zugangSetzen}
        </Knopf>

        {eintrag.hinterlegt && (
          <Aktionsknopf
            aktion="loeschen"
            bezeichnung={`${eigene.zugangLoeschen}: ${name}`}
            angehalten={vorgang.laeuft(eintrag.anbieter) || fragt}
            aufKlick={() => {
              setzeFragt(true)
            }}
          >
            {eigene.zugangLoeschen}
          </Aktionsknopf>
        )}
      </div>

      {eintrag.hinterlegt && fragt && (
        <Entscheidung
          frage={einsetzen(eigene.zugangLoeschenFrage, {
            angabe: eigene.geheimnis,
            name,
          })}
          erklaerung={eigene.zugangLoeschenErklaerung}
          bestaetigen={eigene.zugangLoeschenJa}
          aktion="loeschen"
          abbrechen={texte.bausteine.aktionen.abbrechen}
          laeuft={vorgang.laeuft(eintrag.anbieter)}
          aufBestaetigen={() => {
            setzeFragt(false)
            void zug(
              () => api().AnbietergeheimnisLoeschen(eintrag.anbieter),
              () => undefined,
            )
          }}
          aufAbbrechen={() => {
            setzeFragt(false)
          }}
        />
      )}

      {fehler !== undefined && (
        <div role="alert">
          <p id={fehlerID} className="text-sm text-fehler">
            {fehler}
          </p>
        </div>
      )}
    </div>
  )
}
