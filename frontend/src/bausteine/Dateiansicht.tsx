// **Eine Datei ansehen, ohne dass etwas geschieht** — UX-13
// (#1685),
// Marke `new`: MSE2 zeigt eine geschriebene Datei nicht, es schreibt sie.
//
// # Was hier nie geschieht
//
// **Nichts wird ausgeführt.** Ein Bild und ein SVG erscheinen über `<img>` —
// dort läuft in WebKit kein Skript und wird keine fremde Adresse geladen. Ein
// Text, HTML eingeschlossen, erscheint als Text in `<pre>`; React setzt ihn
// als Textknoten. Ein Paket oder Archiv erscheint als Liste seiner Namen.
// **Welche Art was zeigen darf, entscheidet Go** (`domain.AnsichtsregelFuer`),
// nicht diese Datei.
//
// # Und was an macOS geht
//
// „Im Finder zeigen" für jede Datei, „Öffnen" nur, wo die Regel es erlaubt —
// ein `.command` oder ein Programmbündel wird nicht gestartet. Go prüft das
// ein zweites Mal; der fehlende Knopf ist die Bequemlichkeit, nicht der
// Schutz.
import { useCallback, useState } from 'react'

import { api } from '../api'
import type { DateiansichtDTO } from '../apidateiansicht'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { alsGroesse } from '../seiten/Diagnosepaket'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte, useZeitzone } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { alsZeitpunkt } from '../zeitpunkt'
import { Knopf } from './Knopf'

interface DateiansichtEigenschaften {
  pfad: string
}

const uebergabe = 'uebergabe'

export function Dateiansicht({ pfad }: DateiansichtEigenschaften) {
  const texte = useTexte()
  const eigene = texte.bausteine.dateiansicht
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()
  const vorgang = useVorgang()
  const [fehler, setzeFehler] = useState<string>()

  const holen = useCallback(async () => {
    const antwort = await api().DateiAnsehen(pfad)

    return auspacken(antwort, () => antwort)
  }, [pfad])

  const stand = useLaden<DateiansichtDTO>(holen, [pfad])

  if (stand.fehler !== undefined) {
    return (
      <p role="alert" className="text-fehler">
        {stand.fehler}
      </p>
    )
  }

  const datei = stand.daten

  if (datei === undefined) {
    return <p className="text-leise">{eigene.laedt}</p>
  }

  const uebergeben = (weg: 'finder' | 'oeffnen') =>
    vorgang.starte(uebergabe, async () => {
      try {
        const antwort =
          weg === 'finder'
            ? await api().DateiImFinderZeigen(datei.pfad)
            : await api().DateiMitProgrammOeffnen(datei.pfad)

        auspacken(antwort, () => undefined)
        setzeFehler(undefined)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  return (
    <section
      aria-label={eigene.ueberschrift}
      className="flex flex-col gap-3 rounded-md border border-kante p-3"
    >
      <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
        <dt className="text-leise">{eigene.name}</dt>
        <dd className="break-all">{datei.name}</dd>
        <dt className="text-leise">{eigene.art}</dt>
        <dd>{eigene.arten[datei.art]}</dd>
        {!datei.verzeichnis && (
          <>
            <dt className="text-leise">{eigene.groesse}</dt>
            <dd>{alsGroesse(datei.groesse, darstellung)}</dd>
          </>
        )}
        <dt className="text-leise">{eigene.geaendert}</dt>
        <dd>{alsZeitpunkt(datei.geaendert, darstellung, zeitzone)}</dd>
        <dt className="text-leise">{eigene.ort}</dt>
        <dd className="break-all">{datei.pfad}</dd>
      </dl>

      <Vorschau datei={datei} />

      <div className="flex flex-wrap gap-2">
        <Knopf
          gewicht="leise"
          bezeichnung={einsetzen(eigene.imFinderFuer, { name: datei.name })}
          angehalten={vorgang.laeuft(uebergabe)}
          aufKlick={() => {
            void uebergeben('finder')
          }}
        >
          {eigene.imFinder}
        </Knopf>

        {datei.oeffnen && (
          <Knopf
            gewicht="leise"
            bezeichnung={einsetzen(eigene.oeffnenFuer, { name: datei.name })}
            angehalten={vorgang.laeuft(uebergabe)}
            aufKlick={() => {
              void uebergeben('oeffnen')
            }}
          >
            {eigene.oeffnen}
          </Knopf>
        )}
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}
    </section>
  )
}

// Vorschau zeigt, was die Regel erlaubt — oder sagt, warum nichts.
function Vorschau({ datei }: { datei: DateiansichtDTO }) {
  const eigene = useTexte().bausteine.dateiansicht

  const grund = datei.verschluesselt
    ? eigene.verschluesselt
    : datei.zuGross
      ? eigene.zuGross
      : datei.unlesbar
        ? eigene.unlesbar
        : datei.art === 'unbekannt' && !datei.verzeichnis
          ? eigene.keineVorschau
          : undefined

  // **Ein gewöhnlicher Ordner hat keine Vorschau und keinen Grund dafür** —
  // er ist kein Programm, und was in ihm liegt, zeigt der Finder.
  if (datei.verzeichnis && datei.art === 'unbekannt') {
    return null
  }

  if (grund !== undefined) {
    return <p className="max-w-prose text-leise">{grund}</p>
  }

  if (datei.art === 'pdf') {
    return <p>{einsetzen(eigene.seiten, { anzahl: String(datei.seiten) })}</p>
  }

  if (datei.vorschau === 'bild') {
    return (
      <figure className="flex flex-col gap-1">
        <img
          src={datei.bild}
          alt={datei.name}
          className="max-h-80 max-w-full self-start rounded-sm border border-kante object-contain"
        />
        {datei.breite > 0 && (
          <figcaption className="text-sm text-leise">
            {einsetzen(eigene.masse, {
              breite: String(datei.breite),
              hoehe: String(datei.hoehe),
            })}
          </figcaption>
        )}
      </figure>
    )
  }

  if (datei.vorschau === 'text') {
    return (
      <>
        <pre
          role="region"
          tabIndex={0}
          aria-label={datei.name}
          className="max-h-80 overflow-auto rounded-sm bg-flaeche p-2 text-xs whitespace-pre-wrap"
        >
          {datei.text}
        </pre>
        {datei.gekuerzt && (
          <p className="text-sm text-leise">{eigene.gekuerzt}</p>
        )}
      </>
    )
  }

  return <Eintragsliste datei={datei} />
}

// Eintragsliste zeigt die Namen eines Pakets oder Archivs.
function Eintragsliste({ datei }: { datei: DateiansichtDTO }) {
  const eigene = useTexte().bausteine.dateiansicht

  return (
    <>
      <p className="text-sm">
        {einsetzen(eigene.eintraege, {
          anzahl: String(datei.eintraege.length),
        })}
      </p>
      {datei.unsicher > 0 && (
        <p className="max-w-prose text-sm text-leise">
          {einsetzen(eigene.unsicher, { anzahl: String(datei.unsicher) })}
        </p>
      )}
      {/* **Die Rolle an der Hülle und nicht an der Liste** — eine Liste mit
          `role="region"` wäre keine Liste mehr. */}
      <div
        role="region"
        tabIndex={0}
        aria-label={datei.name}
        className="max-h-60 overflow-auto text-xs"
      >
        <ul>
          {datei.eintraege.map((eintrag) => (
            <li key={eintrag.name} className="flex gap-2">
              <span className="break-all">
                {eintrag.verzeichnis ? `${eintrag.name}/` : eintrag.name}
              </span>
              {eintrag.unsicher && (
                <span className="text-leise">{eigene.unsicherEintrag}</span>
              )}
            </li>
          ))}
        </ul>
      </div>
    </>
  )
}
