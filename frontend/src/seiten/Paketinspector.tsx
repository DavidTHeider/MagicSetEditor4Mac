// **Der Inspector eines Pakets** — DOM-01.AK-01
// (#1696),
// gezeichnet nach dem Artboard `Pakete`: Status, Woher, Hängt ab von — und
// die Handlungen darunter.
//
// **Was er tut, steht in `Paketliste` und `usePaketplan`; wie er aussieht,
// hier** (§2). Er ändert nichts selbst: Löschen und Zurück öffnen den Plan.
//
// # Die Nutzung nur bei Spielen und Stilen
//
// Welches Set ein Spiel oder einen Stil benutzt, steht im Setkopf; welches
// eine Symbolschrift oder eine Exportvorlage braucht, steht nirgends, wo
// diese Anwendung es ohne Öffnen jedes Sets lesen könnte. Eine Null dort
// hieße „keines", wo es „unbekannt" ist — deshalb steht dort ein anderer Satz.
import { useEffect, useRef, type ReactNode } from 'react'

import type { PaketDTO } from '../api'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { Laufstand } from '../bausteine/Laufstand'
import type { Paketweitergabezustand } from '../haken/usePaketweitergabe'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

interface Eigenschaften {
  paket: PaketDTO | undefined
  laeuft: boolean
  aufLoeschen: (paket: PaketDTO) => void
  aufZurueck: (paket: PaketDTO) => void

  // weitergabe gibt das Paket als `.mse-installer` hinaus — FMT-02.AK-04
  // (#1691); was dabei herauskam, steht unter den Handlungen.
  weitergabe: Paketweitergabezustand
}

const mitSets = new Set(['mse-game', 'mse-style'])

export function Paketinspector({
  paket,
  laeuft,
  aufLoeschen,
  aufZurueck,
  weitergabe,
}: Eigenschaften) {
  const texte = useTexte()
  const t = texte.seiten.pakete.plan

  // **Die Meldung rollt sich in den Blick** — der Inspector ist eine eigene
  // Rollfläche, und am nativen Fenster stand sie unter ihrem sichtbaren
  // Rand (Hüllenprobe `fmt02e`): wer weitergab, sah keine Antwort. Nur
  // wenn sie sich ändert, nicht bei jedem Zeichnen.
  const meldungsknoten = useRef<HTMLParagraphElement>(null)
  const meldungstext =
    paket === undefined ? undefined : weitergabe.meldung(paket)?.text

  useEffect(() => {
    meldungsknoten.current?.scrollIntoView({ block: 'nearest' })
  }, [meldungstext])

  if (paket === undefined) {
    return <p className="max-w-prose text-leise">{t.inspectorLeer}</p>
  }

  const nutzung = !mitSets.has(paket.art)
    ? t.setsUnbekannt
    : paket.benutzt === 0
      ? t.keinSet
      : paket.benutzt === 1
        ? t.einSet
        : einsetzen(t.vieleSets, { anzahl: String(paket.benutzt) })

  const meldung = weitergabe.meldung(paket)

  const woher = {
    mitgeliefert: t.woherMitgeliefert,
    bezogen: t.woherBezogen,
    uebernommen: t.woherUebernommen,
    eigen: t.woherEigen,
  }[paket.herkunftsart]

  return (
    <div className="flex flex-col gap-3">
      <div>
        <h3 className="text-lg font-semibold">{paket.name}</h3>
        <p className="text-sm text-leise">
          {paket.version === ''
            ? einsetzen(t.kopfOhneFassung, { art: paket.art })
            : einsetzen(t.kopf, { art: paket.art, version: paket.version })}
        </p>
      </div>

      <div className="text-sm">
        <p className="font-medium">
          <span aria-hidden="true">✓ </span>
          {t.eingebunden}
        </p>
        <p>{nutzung}</p>
        {paket.geaendert && <p className="mt-1">{t.geaendert}</p>}
      </div>

      <Angabe titel={t.woher}>
        <p>{einsetzen(woher, { herkunft: paket.herkunft })}</p>
      </Angabe>

      {paket.pruefsumme !== '' && (
        <Angabe titel={t.pruefsumme}>
          <p className="break-all font-mono">{paket.pruefsumme}</p>
        </Angabe>
      )}

      <Angabe titel={t.lizenz}>
        <p>{t.lizenzUnbekannt}</p>
      </Angabe>

      <Angabe titel={t.ort}>
        <p className="break-all font-mono">{paket.pfad}</p>
      </Angabe>

      <Angabe titel={t.haengtAb}>
        <Namen namen={paket.braucht} leer={t.nichts} />
      </Angabe>

      <Angabe titel={t.gebrauchtVon}>
        <Namen namen={paket.gebrauchtVon} leer={t.nichts} />
      </Angabe>

      {/* **Nach Häufigkeit, das Löschen zuletzt** — UXR-17 (#1925), wie im
          Artboard `Pakete`: die gefährliche Handlung steht am Ende. Hier
          stand sie vorn, die seltenste und die einzige, die etwas wegnimmt.
          Weitergeben ändert nichts, Zurück tauscht nur die Fassung. */}
      <div className="flex flex-wrap gap-2">
        <Knopf
          angehalten={weitergabe.laeuft(paket)}
          aufKlick={() => {
            void weitergabe.weitergeben(paket)
          }}
        >
          {t.weitergeben}
        </Knopf>
        <Knopf
          angehalten={laeuft}
          aufKlick={() => {
            aufZurueck(paket)
          }}
        >
          {t.zurueck}
        </Knopf>
        <Aktionsknopf
          aktion="loeschen"
          angehalten={laeuft}
          aufKlick={() => {
            aufLoeschen(paket)
          }}
        >
          {t.loeschen}
        </Aktionsknopf>
      </div>

      <Laufstand art="weitergabe" />

      {meldung !== undefined && (
        <p
          ref={meldungsknoten}
          role={meldung.fehler ? 'alert' : 'status'}
          className={
            meldung.fehler
              ? 'break-all text-sm text-fehler'
              : 'break-all text-sm text-leise'
          }
        >
          {meldung.text}
        </p>
      )}
    </div>
  )
}

interface Angabeneigenschaften {
  titel: string
  children: ReactNode
}

function Angabe({ titel, children }: Angabeneigenschaften) {
  return (
    <div className="text-sm">
      <p className="font-medium text-leise">{titel}</p>
      {children}
    </div>
  )
}

function Namen({ namen, leer }: { namen: readonly string[]; leer: string }) {
  if (namen.length === 0) {
    return <p>{leer}</p>
  }

  return (
    <ul className="font-mono">
      {namen.map((name) => (
        <li key={name}>{name}</li>
      ))}
    </ul>
  )
}
