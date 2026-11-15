// **Die Kommandozentrale** — Werkstück U12 von Phase 5.1
// (#1423).
//
// Ein Dialog über einem Befehl: was er tut, was er auf den anderen Flächen
// heißt, seine Felder aus dem Schema, die Vorschau dessen, was geschickt
// wird, die Bestätigung, wo etwas nicht ohne Weiteres zurückzunehmen ist,
// das Ergebnis als Text oder JSON — und der Verlauf dieser Sitzung.
//
// **Kein Dialog über einem Dialog** (§6): die Palette geht zu, bevor die
// Zentrale aufgeht; wer einen anderen Befehl will, geht über ⌘K zurück.
import { useState } from 'react'
import { Dialog } from '../bausteine/Dialog'
import { Knopf } from '../bausteine/Knopf'
import { huebsch } from '../befehle'
import { useKommandozentrale } from '../haken/useKommandozentrale'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Befehlsformular } from './Befehlsformular'
import { Befehlsverlauf } from './Befehlsverlauf'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import type { BefehlDTO } from '../apibefehle'

interface ZentraleEigenschaften {
  befehl: BefehlDTO
  schliessen: () => void
}

export function Kommandozentrale({
  befehl,
  schliessen,
}: ZentraleEigenschaften) {
  const texte = useTexte().bausteine.kommandozentrale
  const zentrale = useKommandozentrale(befehl)
  const [alsJson, setzeAlsJson] = useState(true)

  const ergebnis = zentrale.ergebnis
  const gerichtet = ergebnis === undefined ? undefined : huebsch(ergebnis.text)

  return (
    <Dialog
      bezeichnung={texte.titel}
      schliessen={schliessen}
      ausrichtung="oben"
    >
      <div className="flex w-full max-w-2xl flex-col gap-4 overflow-y-auto rounded-lg border-2 border-kante bg-flaeche p-5">
        <header className="flex flex-col gap-1">
          <h2 className="text-lg font-semibold">
            <code>{befehl.name}</code>
          </h2>
          <p className="text-sm text-leise">
            {befehl.schreibend ? texte.schreibend : texte.lesend}
            {befehl.gefaehrlich && ` · ${texte.gefaehrlich}`}
            {befehl.faehigkeit !== '' &&
              ` · ${texte.faehigkeit}: ${befehl.faehigkeit}`}
          </p>
          <Erklaerungsabsatz
            text={befehl.erklaerung}
            klasse="max-w-prose text-sm"
          />
          <dl className="grid grid-cols-[max-content_1fr] gap-x-4 text-sm text-leise">
            {befehl.cli !== '' && (
              <>
                <dt>{texte.cli}</dt>
                <dd>
                  <code>{befehl.cli}</code>
                </dd>
              </>
            )}
            {befehl.rest !== '' && (
              <>
                <dt>{texte.rest}</dt>
                <dd>
                  <code>{befehl.rest}</code>
                </dd>
              </>
            )}
          </dl>
        </header>

        <section aria-label={texte.parameter}>
          <h3 className="mb-2 text-sm font-semibold">{texte.parameter}</h3>
          <Befehlsformular
            parameter={zentrale.parameter}
            werte={zentrale.werte}
            befunde={zentrale.befunde}
            angehalten={zentrale.laeuft}
            setzeWert={zentrale.setzeWert}
          />
        </section>

        <section>
          <h3 className="mb-1 text-sm font-semibold">{texte.vorschau}</h3>
          {/* **Eine Region mit Namen und `tabIndex`, weil sie rollt** — das
              Muster aus `Diagnose.tsx` (Zweitblick über Phase 6.3, G10). */}
          <pre
            role="region"
            aria-label={texte.vorschau}
            tabIndex={0}
            className="max-h-40 overflow-auto rounded-md border border-kante p-2 text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            {zentrale.vorschau}
          </pre>
          {befehl.gefaehrlich && (
            <label className="mt-2 flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={zentrale.bestaetigt}
                disabled={zentrale.laeuft}
                onChange={(ereignis) => {
                  zentrale.setzeBestaetigt(ereignis.target.checked)
                }}
              />
              <span>{texte.bestaetigen}</span>
            </label>
          )}
        </section>

        <div className="flex flex-wrap items-center gap-2">
          <Knopf
            gewicht="haupt"
            angehalten={zentrale.laeuft || !zentrale.bereit}
            aufKlick={zentrale.ausfuehren}
          >
            {texte.ausfuehren}
          </Knopf>
          {zentrale.laeuft && (
            <>
              <Knopf gewicht="neben" aufKlick={zentrale.abbrechen}>
                {texte.abbrechen}
              </Knopf>
              <span role="status" className="text-sm text-leise">
                {texte.laeuft}
              </span>
            </>
          )}
          <Knopf gewicht="neben" aufKlick={schliessen}>
            {texte.schliessen}
          </Knopf>
        </div>

        {ergebnis !== undefined && gerichtet !== undefined && (
          <section aria-label={texte.ergebnis}>
            <div className="mb-1 flex flex-wrap items-baseline gap-2 text-sm">
              <h3 className="font-semibold">{texte.ergebnis}</h3>
              <span className="text-leise">
                {texte[ergebnis.ausgang]}
                {' · '}
                {einsetzen(texte.dauer, { ms: String(ergebnis.dauerMs) })}
              </span>
              {gerichtet.json && (
                <Knopf
                  gewicht="neben"
                  gedrueckt={alsJson}
                  aufKlick={() => {
                    setzeAlsJson((bisher) => !bisher)
                  }}
                >
                  {alsJson ? texte.alsText : texte.alsJson}
                </Knopf>
              )}
            </div>
            {ergebnis.fehler !== null ? (
              <p role="alert" className="text-sm text-fehler">
                {ergebnis.fehler.meldung}
              </p>
            ) : (
              <pre
                role="region"
                aria-label={texte.ausgabe}
                tabIndex={0}
                className="max-h-64 overflow-auto rounded-md border border-kante p-2 text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              >
                {alsJson ? gerichtet.text : ergebnis.text}
              </pre>
            )}
          </section>
        )}

        <Befehlsverlauf
          uebernehmen={(name, argumente) => {
            if (name === befehl.name) {
              zentrale.uebernehmen(argumente)
            }
          }}
        />
      </div>
    </Dialog>
  )
}
