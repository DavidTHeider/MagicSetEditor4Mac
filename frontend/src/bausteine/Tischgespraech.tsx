// **Das Gespräch am Tisch** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum Gespräch und Präsenz nebeneinander stehen
//
// **Weil sie dieselbe Frage beantworten**: wer ist außer mir noch da, und was
// tut er gerade. Getrennt gezeichnet müsste ein Mensch zwei Stellen im Auge
// behalten, um eine Antwort zu bekommen.
//
// # Und warum die Präsenz als Satz dasteht
//
// **Weil ein gezeichneter Mauszeiger für VoiceOver nichts ist.** „anna: zieht
// k7" ist dieselbe Auskunft, und sie steht jedem zur Verfügung (§6). Ein
// Zeiger **auf** dem Tisch ist die schönere Hälfte davon; er kommt dazu und
// nicht an ihre Stelle.
//
// # Warum das Eingabefeld hier steht und nicht `Beschriftetesfeld` ist
//
// **Weil es drei Dinge braucht, die jenes nicht hat**: eine Grenze für die
// Länge, die Eingabetaste als Absender und eine Beschriftung, die nur
// Vorleseprogramme hören. Jenen Baustein dafür zu erweitern hieße, eine Zeile
// in dreißig Formularen zu ändern, damit eine Chatleiste funktioniert (§2).
import { useId, useState } from 'react'

import { Knopf } from './Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { zeigerwort } from '../seiten/partiewoerter'

import type { Nebenlauf } from '../haken/usePartienebenbei'
import { Abschnittstitel } from './Abschnittstitel'

// Satzlaenge ist die Grenze, ab der ein Satz kein Satz mehr ist.
//
// **Dieselbe wie in `domain.Chatlaenge`.** Das Feld hält sie ein, damit
// niemand tippt, was hinterher abgewiesen wird; verlassen wird sich darauf
// nicht — die Go-Seite prüft noch einmal.
export const Satzlaenge = 500

export function Tischgespraech({ lauf }: { lauf: Nebenlauf }) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const ueberschrift = useId()

  return (
    <section
      aria-labelledby={ueberschrift}
      className="flex flex-col gap-2 rounded-sm border border-kante p-3"
    >
      <Abschnittstitel id={ueberschrift} className="text-sm font-semibold">
        {eigene.gespraech}
      </Abschnittstitel>

      <Praesenzleiste lauf={lauf} />
      <Gespraechsliste lauf={lauf} />
      <Sagefeld lauf={lauf} />

      {lauf.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {lauf.fehler}
        </p>
      )}
    </section>
  )
}

// Sagefeld ist die Zeile, in die jemand tippt.
function Sagefeld({ lauf }: { lauf: Nebenlauf }) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const [entwurf, setzeEntwurf] = useState('')
  const feld = useId()

  // **Erst leeren, dann senden.** Wer umgekehrt vorginge, ließe bei einer
  // langsamen Leitung Zeit für einen zweiten Druck auf die Eingabetaste — und
  // derselbe Satz ginge zweimal hinaus (§7, ein Vorgang je Datensatz).
  const absenden = () => {
    const text = entwurf.trim()
    if (text === '') return

    setzeEntwurf('')
    lauf.sagen(text)
  }

  return (
    <div className="flex items-center gap-2">
      <label className="sr-only" htmlFor={feld}>
        {eigene.gespraechEingabe}
      </label>

      <input
        id={feld}
        type="text"
        value={entwurf}
        maxLength={Satzlaenge}
        className="flex-1 rounded-sm border border-kante bg-flaeche px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        onChange={(ereignis) => {
          setzeEntwurf(ereignis.target.value)
        }}
        onKeyDown={(ereignis) => {
          if (ereignis.key !== 'Enter') return

          ereignis.preventDefault()
          absenden()
        }}
      />

      <Knopf
        gewicht="leise"
        angehalten={entwurf.trim() === ''}
        aufKlick={absenden}
      >
        {eigene.gespraechSenden}
      </Knopf>
    </div>
  )
}

// Praesenzleiste sagt, wer gerade was tut.
function Praesenzleiste({ lauf }: { lauf: Nebenlauf }) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  if (lauf.zeiger.length === 0) {
    return <p className="text-xs text-leise">{eigene.gespraechNiemandZeigt}</p>
  }

  return (
    <ul aria-live="polite" className="flex flex-wrap gap-2 text-xs text-leise">
      {lauf.zeiger.map((einer) => (
        <li key={einer.von} className="rounded-sm bg-kante/30 px-2 py-0.5">
          {einsetzen(
            einer.objekt === ''
              ? eigene.gespraechZeigt
              : eigene.gespraechZeigtAuf,
            {
              sitz: einer.von,
              tut: zeigerwort(eigene, einer.art),
              objekt: einer.objekt,
            },
          )}
        </li>
      ))}
    </ul>
  )
}

// Gespraechsliste zeigt, was gesagt wurde.
//
// **Die neuesten unten**, wie in jedem Gespräch — und die Liste rollt, statt
// die Seite zu strecken.
function Gespraechsliste({ lauf }: { lauf: Nebenlauf }) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  if (lauf.gespraech.length === 0) {
    return <p className="text-sm text-leise">{eigene.gespraechLeer}</p>
  }

  return (
    // **Eine Region mit Namen und `tabIndex`, weil sie rollt** — das
    // Muster aus `Diagnose.tsx` (Zweitblick über Phase 6.3, G10). **Um die
    // Liste und nicht auf ihr**: eine Region auf der `<ul>` nahm ihr die
    // Rolle, und jedes `<li>` stand ohne gültige Liste da (Nachlauf, N7).
    <div
      role="region"
      aria-label={eigene.gespraechVerlauf}
      tabIndex={0}
      className="max-h-40 overflow-y-auto focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
    >
      <ul aria-live="polite" className="flex flex-col gap-1 text-sm">
        {lauf.gespraech.map((zeile) => (
          <li key={zeile.nummer}>
            <span className="font-mono text-xs text-leise">{zeile.von}</span>{' '}
            {zeile.text}
          </li>
        ))}
      </ul>
    </div>
  )
}
