// **Die Diagnose** — Werkstück 4 von Phase 3.5
// (#1166).
//
// # Warum es sie gibt
//
// `docs/ADMINISTRATION.md` nannte für einen Fehlerbericht
// `tail -n 200 …/mse.log`. **Die Zielgruppe dieser Anwendung kann das nicht**
// (Z9-01 im Audit) — sie mag Magic und sammelt Karten, und ein Terminal ist
// für sie keine Abkürzung, sondern eine Wand.
//
// # Vorschau und Zustimmung
//
// **Sie steht da, bevor jemand sie weitergibt.** Der Text wird auf Wunsch
// geholt, angezeigt und erst dann kopiert oder gesichert — drei Schritte, und
// jeder davon ist eine Handlung. Eine Anwendung, die von selbst etwas
// hinausschickte, wäre etwas anderes.
//
// **Was drinsteht und was nicht**, sagt die Seite selbst — nicht nur die
// Dokumentation. Wer etwas weitergibt, soll wissen, was er weitergibt.
import { useId, useState } from 'react'

import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { inDieZwischenablage } from '../zwischenablage'
import { api } from '../fassade'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { useAnschluss, useAufgenommen } from '../anschlussraum'
import { einsetzen } from '../sprache/einsetzen'

import type { Anschluss } from '../anschluss'

export function Diagnose() {
  const texte = useTexte()
  const eigene = texte.seiten.technik.diagnose
  const vorgang = useVorgang()

  const [text, setzeText] = useState<string>()
  const [fehler, setzeFehler] = useState<string>()
  const [kopiert, setzeKopiert] = useState(false)

  // **Ein Filter über die Zeilen** (NET-02, S13.NAT-04, #1706) — er ändert
  // nur, was gezeigt wird; kopiert und gesichert wird der ganze Text.
  const [filter, setzeFilter] = useState('')
  const filterfeld = useId()

  const holen = () =>
    vorgang.starte('diagnose', async () => {
      try {
        const antwort = await api().Diagnose()

        setzeText(auspacken(antwort, () => antwort.text))
        setzeFehler(undefined)
        setzeKopiert(false)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  // **Aus „Diagnose öffnen" kommt eine Vorgangskennung mit** — sie wird der
  // Filter, und der Text wird geholt: der Klick dort **war** die Bitte.
  const anschluss = useAnschluss()
  const mitgereist = mitVorgang(anschluss.offen) ? anschluss.offen : undefined

  useAufgenommen(mitgereist, (angekommen) => {
    setzeFilter(angekommen.vorgang)
    void holen()
  })

  const zeilen = text?.split('\n') ?? []
  const gezeigt =
    filter === '' ? zeilen : zeilen.filter((zeile) => zeile.includes(filter))

  const sichern = () =>
    vorgang.starte('diagnose', async () => {
      try {
        const antwort = await api().DiagnoseSichern()

        auspacken(antwort, () => null)
        setzeFehler(undefined)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  return (
    <section className="flex flex-col gap-3 border-t border-kante pt-4">
      <h3 className="text-lg font-semibold">{eigene.ueberschrift}</h3>

      <Erklaerungsabsatz text={eigene.erklaerung} klasse="text-sm text-leise" />

      {/* **Was drinsteht, steht vor dem Knopf.** Danach hätte es niemand
          gelesen, der schon geklickt hat — dieselbe Ordnung wie beim Satz
          über die Shell darüber. */}
      {/* **Was nicht darin steht, steht im ersten Satz**; die ganze Liste
          hinter dem Hilfesymbol (UX-15, #1771). */}
      <Erklaerungsabsatz text={eigene.inhalt} klasse="text-sm text-leise" />

      <div className="flex flex-wrap gap-2">
        <Knopf
          bezeichnung={eigene.anzeigen}
          aufKlick={() => {
            void holen()
          }}
        >
          {eigene.anzeigen}
        </Knopf>

        {text !== undefined && (
          <>
            <Aktionsknopf
              aktion="kopieren"
              bezeichnung={eigene.kopieren}
              aufKlick={() => {
                // **Auch das Nein wird gezeigt** — Befund 5 des Nachlaufs zum
                // Zweitblick über Phase 5.1: ohne Erlaubnis blieb hier eine
                // abgewiesene Zusage ohne Empfänger, und die Seite zeigte
                // nichts.
                void inDieZwischenablage(text).then((geglueckt) => {
                  setzeKopiert(geglueckt)
                  setzeFehler(geglueckt ? undefined : eigene.kopierfehler)
                })
              }}
            >
              {kopiert ? eigene.kopiert : eigene.kopieren}
            </Aktionsknopf>

            <Aktionsknopf
              aktion="sichern"
              bezeichnung={eigene.sichern}
              aufKlick={() => {
                void sichern()
              }}
            >
              {eigene.sichern}
            </Aktionsknopf>
          </>
        )}
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {/* **Ein `pre` und kein `textarea`.** Der Text wird gelesen und nicht
          geändert; ein Eingabefeld lüde dazu ein, ihn vor dem Weitergeben
          zurechtzulegen — und dann stimmte der Bericht nicht mehr.
          
          `tabIndex`, weil er rollt: ein rollender Bereich, den die Tastatur
          nicht erreicht, ist für manche gar nicht da (§6). */}
      <label className="flex flex-col gap-1 text-sm" htmlFor={filterfeld}>
        {eigene.filter}
        <input
          id={filterfeld}
          type="search"
          className="w-64 rounded-sm border border-kante bg-transparent px-2 py-1 font-mono text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          value={filter}
          onChange={(ereignis) => {
            setzeFilter(ereignis.target.value)
          }}
        />
      </label>

      {text !== undefined && filter !== '' && (
        <p className="text-sm text-leise">
          {einsetzen(eigene.zeilenGezeigt, {
            gezeigt: String(gezeigt.length),
            alle: String(zeilen.length),
          })}
        </p>
      )}

      {text !== undefined && (
        <pre
          role="region"
          tabIndex={0}
          aria-label={eigene.ueberschrift}
          className="max-h-64 overflow-auto rounded-sm bg-betont/60 p-3 font-mono text-xs whitespace-pre-wrap focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        >
          {gezeigt.join('\n')}
        </pre>
      )}
    </section>
  )
}

// Mitgereist ist ein Sprung zur Diagnose, der eine Vorgangskennung trägt.
type Mitgereist = Extract<Anschluss, { art: 'einstellungsgruppe' }> & {
  vorgang: string
}

// mitVorgang sagt, ob ein Anschluss eine Kennung mitbringt — **als
// Typwächter**, damit die Kennung danach nicht noch einmal auf `undefined`
// geprüft wird: ein Rückfall, den kein Lauf betreten kann, wäre eine
// Zusage, die niemand hält (§10).
function mitVorgang(offen: Anschluss | undefined): offen is Mitgereist {
  return offen?.art === 'einstellungsgruppe' && offen.vorgang !== undefined
}
