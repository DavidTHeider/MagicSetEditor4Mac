// **Der Inspector der Vorlagenwerkstatt** — Testkarte, Vorschau und die
// Herkunft eines Namens (Werkstück U7 von Phase 5.1,
// #1418).
//
// **Die Diagnose steht seit FMT-02 unter dem Editor** (`Vorlagenbefund.tsx`,
// #1691): der Inspector weicht bei engem Fenster zuerst, und nach dem
// Artboard `Werkstatt` gehört der Befund ohnehin zum Text.
//
// **Was er tut, entscheidet die Seite; wie es aussieht, steht hier** (§2).
// Jeder Knopf ruft nach oben — „Zur Definition", eine Verwendung —, und die
// Seite setzt den Zeiger im Editor.
import { useId, useMemo } from 'react'

import { Knopf } from '../bausteine/Knopf'
import { Suchauswahl } from '../bausteine/Suchauswahl'
import type { Suchwahl } from '../bausteine/suchwahl'
import { Vorlagenvorschau } from './Vorlagenvorschau'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

import type { ProbekarteDTO, Skriptstelle, SymbolverwendungDTO } from '../api'
import type { Symbolstelle } from '../bausteine/symbolstelle'

interface VorlageninspectorEigenschaften {
  paket: string
  datei: string

  // gueltig ist der letzte Text ohne Befund — `undefined`, wenn es keinen gab.
  gueltig: string | undefined
  veraltet: Skriptstelle | null

  karteID: string
  setzeKarteID: (karteID: string) => void
  karten: ProbekarteDTO[]
  kartenFehler: string | undefined

  symbol: Symbolstelle | null
  verwendungen: SymbolverwendungDTO[]
  verwendungenLaden: boolean
  oeffnen: (datei: string, zeile: number, spalte: number) => void
}

export function Vorlageninspector({
  paket,
  datei,
  gueltig,
  veraltet,
  karteID,
  setzeKarteID,
  karten,
  kartenFehler,
  symbol,
  verwendungen,
  verwendungenLaden,
  oeffnen,
}: VorlageninspectorEigenschaften) {
  const texte = useTexte().seiten.vorlagen
  const definition = symbol?.symbol
  const testkarteID = useId()

  // **Die Testkarte als Auswahl mit Suche** — UXR-F1
  // (#2007),
  // R7: ein Set trägt hunderte Karten, und das native Auswahlfeld hatte keine
  // Suche. **Die erste Zeile ist die Vorgabe** — ohne gewählte Karte probt
  // die Werkstatt mit der Probekarte des Spiels.
  const eintraege = useMemo<Suchwahl[]>(
    () => [
      { wert: '', name: texte.testkarteVorgabe },
      ...karten.map((karte) => ({
        wert: karte.id,
        name:
          karte.name === ''
            ? einsetzen(texte.testkarteOhneNamen, {
                nummer: String(karte.nummer),
              })
            : karte.name,
      })),
    ],
    [karten, texte.testkarteVorgabe, texte.testkarteOhneNamen],
  )

  return (
    <div className="flex flex-col gap-4">
      {/* **Die Vorschau oben, die Testkarte darunter** — wie im Artboard
          `Werkstatt` (UXR-18, #1926): die Karte ist das Hauptobjekt (§6), und
          die Wahl, womit geprobt wird, steht als ihre Unterschrift darunter. */}
      {gueltig !== undefined && (
        <Vorlagenvorschau
          paket={paket}
          datei={datei}
          inhalt={gueltig}
          karteID={karteID}
          veraltet={veraltet}
        />
      )}

      <section className="flex flex-col gap-2">
        <label htmlFor={testkarteID} className="text-sm font-medium">
          {texte.testkarte}
        </label>
        <Suchauswahl
          id={testkarteID}
          bezeichnung={texte.testkarte}
          eintraege={eintraege}
          wert={karteID}
          aendern={setzeKarteID}
        />
        <Erklaerungsabsatz
          text={texte.testkarteErklaerung}
          klasse="text-xs text-leise"
        />
        {kartenFehler !== undefined && (
          <p role="alert" className="text-sm text-fehler">
            {kartenFehler}
          </p>
        )}
      </section>

      <section className="flex flex-col gap-1">
        <h3 className="text-sm font-medium">{texte.herkunft}</h3>
        {symbol === null && (
          <p className="text-sm text-leise">{texte.herkunftKeine}</p>
        )}
        {symbol !== null && symbol.symbol === undefined && (
          <p className="text-sm text-leise">
            {einsetzen(texte.herkunftUnbekannt, { name: symbol.name })}
          </p>
        )}
        {symbol !== null && definition !== undefined && (
          <>
            <p className="font-mono text-sm">{symbol.name}</p>
            <p className="text-sm">
              {einsetzen(texte.herkunftStelle, {
                art: texte.symbolarten[definition.art] ?? texte.symbolartSonst,
                datei: definition.datei,
                zeile: String(definition.zeile),
              })}
            </p>
            <div>
              <Knopf
                gewicht="neben"
                aufKlick={() => {
                  oeffnen(definition.datei, definition.zeile, definition.spalte)
                }}
              >
                {texte.zurDefinition}
              </Knopf>
            </div>
          </>
        )}
        {symbol !== null && (
          <div className="flex flex-col gap-1">
            <h4 className="text-xs font-medium text-leise">
              {texte.verwendungen}
            </h4>
            {verwendungenLaden && (
              <p className="text-xs text-leise">{texte.verwendungLaden}</p>
            )}
            {!verwendungenLaden && verwendungen.length === 0 && (
              <p className="text-xs text-leise">{texte.verwendungenKeine}</p>
            )}
            {verwendungen.length > 0 && (
              <ul className="flex flex-col gap-0.5">
                {verwendungen.map((verwendung) => (
                  <li key={`${verwendung.datei}:${String(verwendung.zeile)}`}>
                    <button
                      type="button"
                      onClick={() => {
                        oeffnen(verwendung.datei, verwendung.zeile, 1)
                      }}
                      className="w-full rounded-sm px-1 text-left font-mono text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                    >
                      <span className="text-leise">
                        {verwendung.datei}:{verwendung.zeile}
                      </span>{' '}
                      {verwendung.text}
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}
      </section>
    </div>
  )
}
