// Der **Editor einer Paketdatei** — das Zentrum der Vorlagenwerkstatt
// (#1164,
// Werkstück U7 von Phase 5.1: #1418).
//
// # Warum er nichts mehr lädt
//
// Bis U7 hielt dieses Bauteil Datei, Inhalt, Stand und Meldung selbst, und
// ein `key` auf dem Paket räumte alles beim Wechsel. Im Arbeitsplatz lesen
// **zwei** Regionen denselben Puffer — Vorschau und Diagnose stehen im
// Inspector —, und ein `key` räumt nur ein Bauteil. Der Puffer liegt deshalb
// in `useVorlagenpuffer`, und hier steht, wie er aussieht.
//
// # Was vor dem Speichern zu sehen ist
//
// Die Abnahme von #1418 verlangt einen **Diff vor dem Speichern**: wer eine
// Vorlage schreibt, schreibt sie für jedes Set, das das Paket benutzt — und
// soll vorher sehen, welche Zeilen anders sind als der gelesene Stand.
//
// # Warum der Fuß klebt
//
// **Das Feld wächst mit der Datei** — ein Stil hat schnell 130 Zeilen, und
// Sichern, Konflikt und Befund standen darunter außer Sicht. Gemessen an der
// Hüllenprobe `fmt02b` (FMT-02, #1691): bei 900 × 820 Punkten lag der Befund
// einer unbekannten Feldart unter dem unteren Rand, bei 1280 Punkten der
// Konflikt samt seinen zwei Knöpfen. Der Fuß bleibt deshalb am unteren Rand
// stehen, solange das Feld darüber hinausreicht, und rückt an seinen Platz
// unter dem Feld, sobald es endet — dort, wo das Artboard ihn zeichnet.
import { useState } from 'react'

import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { Quelltextfeld } from '../bausteine/Quelltextfeld'
import { Zeilendiff } from '../bausteine/Zeilendiff'
import { useTexte } from '../sprache/kontext'

import type { ReactNode } from 'react'
import type { Skriptstelle } from '../api'
import type { Sprungziel } from '../bausteine/quelltextsprung'
import type { Vorlagenpuffer } from '../haken/useVorlagenpuffer'

interface VorlagendateiEigenschaften {
  puffer: Vorlagenpuffer

  // stelle ist der erste Syntaxfehler in den Skripten — er wird angestrichen.
  stelle: Skriptstelle | null

  // namen sind die Vorschläge der Vervollständigung.
  namen: readonly string[]

  sprungziel: Sprungziel | null
  aufZeiger: (stelle: number) => void

  // befund steht im Fuß unter Sichern und einem Fehler — die Seite reicht ihn
  // herein, weil er aus der Prüfung kommt und nicht aus dem Puffer.
  befund?: ReactNode
}

export function Vorlagendatei({
  puffer,
  stelle,
  namen,
  sprungziel,
  aufZeiger,
  befund,
}: VorlagendateiEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.vorlagen
  const [zeigtAenderungen, setzeZeigtAenderungen] = useState(false)

  return (
    <section className="flex flex-col gap-2">
      {/* **Die Warnung zum Geltungsbereich steht in der Zeile darüber** —
          `Dateikopf.tsx` (UXR-18, #1926), weiter vor dem Feld. */}
      {/* **Zeilennummern und die Farben des Paketformats** (#1164, #1181) —
          `sprache="paket"`, weil eine Paketdatei kein MSE-Script ist: `game:
          magic` ist ein Schlüssel mit einem Wert; ein `script:`-Wert darin
          trägt die Arten der Sprache. **Der Syntaxfehler steht an seiner
          Stelle im Text** (#1324). */}
      <Quelltextfeld
        stelle={stelle}
        sprache="paket"
        bezeichnung={eigene.inhalt}
        wert={puffer.inhalt}
        setzeWert={puffer.setzeInhalt}
        namen={namen}
        sprungziel={sprungziel}
        aufZeiger={aufZeiger}
        zeilen={24}
      />

      <div className="sticky bottom-0 z-10 flex flex-col gap-2 border-t border-kante bg-grund py-2">
        <div className="flex flex-wrap items-center gap-3">
          <Aktionsknopf
            aktion="sichern"
            bezeichnung={eigene.speichern}
            aufKlick={() => {
              void puffer.speichern()
            }}
            gewicht="haupt"
            angehalten={!puffer.geaendert || puffer.schreibt}
          >
            {eigene.speichern}
          </Aktionsknopf>

          <Knopf
            gewicht="neben"
            ausgeklappt={zeigtAenderungen}
            angehalten={!puffer.geaendert}
            aufKlick={() => {
              setzeZeigtAenderungen((war) => !war)
            }}
          >
            {zeigtAenderungen
              ? eigene.aenderungenVerbergen
              : eigene.aenderungenZeigen}
          </Knopf>

          {/* **Nur was stimmt**: „Nicht gesichert." erst, wenn sich etwas
              geändert hat — unter einer frisch geöffneten Datei behauptete der
              Satz einen Zustand, den es nicht gab (Hüllenprobe `fmt02c`). */}
          {(puffer.gespeichert || puffer.geaendert) && (
            <p className="text-sm text-leise">
              {puffer.gespeichert ? eigene.gespeichert : eigene.ungespeichert}
            </p>
          )}
        </div>

        {puffer.fehler !== undefined && (
          <div className="flex flex-wrap items-center gap-3">
            <p role="alert" className="text-sm text-fehler">
              {puffer.fehler}
            </p>
            {/* **Neu laden gibt es nur beim Konflikt** — dort ist es der Weg,
              den die Meldung nennt; nach einem anderen Fehler stünde der
              Text zum Überschreiben bereit, und niemand hat ihn ersetzt. */}
            {puffer.fehler === eigene.konflikt && (
              <>
                <Knopf gewicht="neben" aufKlick={puffer.neuLaden}>
                  {eigene.neuLaden}
                </Knopf>
                <Knopf
                  gewicht="neben"
                  aufKlick={() => {
                    void puffer.behalten()
                  }}
                >
                  {eigene.behalten}
                </Knopf>
              </>
            )}
          </div>
        )}

        {befund}
      </div>

      {zeigtAenderungen && puffer.geaendert && (
        <section className="flex flex-col gap-1">
          <h3 className="text-sm font-medium">{eigene.aenderungen}</h3>
          <Zeilendiff alt={puffer.basis} neu={puffer.inhalt} />
        </section>
      )}
    </section>
  )
}
