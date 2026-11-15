// **Ein Vorschlag am Feld** — AI-05.AK-06
// (#2042),
// Artboard *Karten*: die Vorschläge stehen an den Feldern Name und
// Flavortext, nicht als eigener Block.
//
// # Wer gefragt wird
//
// **Der Vorrang** (`anbieter: 'vorrang'`): eine kleine Textaufgabe geht zuerst
// an Apples Gerätemodell, dann an einen lokalen Anbieter, nie von selbst in
// die Cloud. Der Vorschlag nennt, wer geantwortet hat und auf welchem Weg.
//
// # Was er nicht tut
//
// **Schreiben, ohne dass jemand drückt** — Vorschlagen ändert die Karte nicht,
// erst „Übernehmen" schreibt, in einem Zug und mit Herkunft (A11, #1450).
//
// # Leise wie die Marken
//
// **Der Knopf erscheint mit dem Zeiger oder dem Fokus auf der Feldzeile**
// (`group` an ihrem `<label>`) und bleibt in der Tabfolge — sonst hätte die
// Kartenseite R2 gerissen. Ohne Hover steht er immer da (N1 des Nachlaufs
// über 6.13).
import { useId, useState } from 'react'

import { Knopf } from './Knopf'
import { kiFeldauftrag } from './kifeldauftrag'
import { KIEinrichtungsverweis } from './KIEinrichtungsverweis'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'
import { useVorrangreihe } from '../haken/useVorrangreihe'

import type { KIVorschlagAntwort } from '../antworten'
import type { WertDTO } from '../formen'

interface KIFeldvorschlagEigenschaften {
  karteID: string
  feld: string
  aufschrift: string

  // werte und formen beschreiben die Karte, die das Modell sehen soll.
  werte: readonly WertDTO[]
  formen: ReadonlyMap<string, { art: string }>

  // geaendert lädt die Karte nach einer Übernahme neu.
  geaendert: () => void
}

// leise blendet den Knopf aus, bis die Feldzeile Zeiger oder Fokus hat — und
// nur, wo ein Zeiger Hover kennt.
const leise =
  '[@media(hover:hover)]:opacity-0 focus-visible:opacity-100 group-hover:opacity-100 group-focus-within:opacity-100'

export function KIFeldvorschlag({
  karteID,
  feld,
  aufschrift,
  werte,
  formen,
  geaendert,
}: KIFeldvorschlagEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  const vorgang = useVorgang()
  const [vorschlag, setzeVorschlag] = useState<KIVorschlagAntwort>()
  const [fehler, setzeFehler] = useState<string>()

  const schluessel = `feldvorschlag/${karteID}/${feld}`

  // **Wen der Vorrang fragt** (#2054) — geladen, sobald jemand die Zeile
  // berührt, und als Beschreibung des Knopfs.
  // **Und was an der Reihe nicht stimmt** (Z4), etwa ein gewähltes Modell,
  // das fehlt — ein Satz für Auge und Vorleser.
  const vorrang = useVorrangreihe()
  const reiheID = useId()
  const reihesatz =
    [
      vorrang.reihe !== undefined && vorrang.reihe.length > 0
        ? einsetzen(eigene.kiFeldVorrangFragt, {
            reihe: vorrang.reihe.join(', '),
          })
        : '',
      ...vorrang.befunde,
    ]
      .filter((satz) => satz !== '')
      .join(' ') || undefined

  async function vorschlagen() {
    setzeFehler(undefined)
    setzeVorschlag(undefined)

    await vorgang.starte(schluessel, async () => {
      try {
        const antwort = await api().KIVorschlagen(karteID, {
          anbieter: 'vorrang',
          alias: '',
          auftrag: kiFeldauftrag(
            feld === 'name'
              ? eigene.kiFeldAuftragName
              : eigene.kiFeldAuftragFlavor,
            eigene.kiFeldKarte,
            werte,
            formen,
          ),
          nachweis: '',
          belegfrage: '',
          freigegeben: [],
          nur: [feld],
        })

        setzeVorschlag(auspacken(antwort, () => antwort))
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  async function uebernehmen(kennung: string) {
    setzeFehler(undefined)

    await vorgang.starte(schluessel, async () => {
      try {
        auspacken(
          await api().KIVorschlagUebernehmen(kennung, [feld]),
          () => null,
        )

        setzeVorschlag(undefined)
        geaendert()
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  // **Ein leerer Wert ist kein Vorschlag** — das echte Modell gab einmal
  // von vier Läufen einen Namen ohne Text (Vermerk 0166).
  const gefunden = vorschlag?.felder.find(
    (wert) =>
      wert.feld === feld && wert.werte.some((text) => text.trim() !== ''),
  )
  const abgelehnt = vorschlag?.ausgang === 'abgelehnt'

  return (
    <div className="flex flex-col gap-1 text-sm">
      <span
        className={vorschlag === undefined ? leise : undefined}
        onPointerEnter={vorrang.beruehrt}
        onFocus={vorrang.beruehrt}
      >
        <Knopf
          gewicht="neben"
          klein
          bezeichnung={einsetzen(eigene.kiFeldVorschlagenFuer, {
            feld: aufschrift,
          })}
          beschriebenVon={reihesatz === undefined ? undefined : reiheID}
          angehalten={vorgang.laeuft(schluessel)}
          aufKlick={() => {
            void vorschlagen()
          }}
        >
          {eigene.kiFeldVorschlagen}
        </Knopf>
        {reihesatz !== undefined && (
          <span id={reiheID} className="ml-2 text-leise">
            {reihesatz}
          </span>
        )}
      </span>

      {abgelehnt && (
        <p role="status" className="text-leise">
          {einsetzen(eigene.kiFeldAbgelehnt, {
            grund: vorschlag.befunde.join(' '),
          })}
        </p>
      )}

      {vorschlag !== undefined && !abgelehnt && gefunden === undefined && (
        <p role="status" className="text-leise">
          {eigene.kiFeldNichts}{' '}
          {/* **Warum, steht dabei** — die Befunde der Schemaprüfung, wie sie
              Go formuliert (§8); ohne sie bliebe nur raten. */}
          {vorschlag.befunde.join(' ')}
        </p>
      )}

      {vorschlag !== undefined && gefunden !== undefined && (
        <div className="flex flex-wrap items-center gap-2">
          <span>{gefunden.werte.join(', ')}</span>

          <Knopf
            gewicht="neben"
            klein
            angehalten={vorgang.laeuft(schluessel)}
            aufKlick={() => {
              void uebernehmen(vorschlag.kennung)
            }}
          >
            {eigene.kiFeldUebernehmen}
          </Knopf>

          <span className="text-leise">
            {einsetzen(eigene.kiFeldVon, {
              anbieter: `${vorschlag.anbieter}/${vorschlag.modell}`,
              weg: wegsatz(eigene, vorschlag.weg),
            })}
          </span>
        </div>
      )}

      {/* **Mit dem Weg zur Einrichtung** (AI-01.AK-04, #1742): eine

          KI-Meldung heißt meist, dass ein Anbieter oder Modell fehlt. */}

      {fehler !== undefined && (
        <div className="flex flex-col items-start gap-1">
          <p role="alert" className="text-fehler">
            {fehler}
          </p>
          <KIEinrichtungsverweis />
        </div>
      )}
    </div>
  )
}

// Einstellungstexte ist der Katalog, in dem die KI-Texte stehen.
type Einstellungstexte = ReturnType<typeof useTexte>['seiten']['einstellungen']

// wegsatz sagt, wie der Anbieter drankam — eine Abbildung, damit ein neuer
// Weg als fehlender Schlüssel auffällt.
function wegsatz(eigene: Einstellungstexte, weg: string): string {
  const saetze: Record<string, string> = {
    gewaehlt: eigene.kiFeldWegGewaehlt,
    vorrang: eigene.kiFeldWegVorrang,
    rueckfall: eigene.kiFeldWegRueckfall,
  }

  return saetze[weg] ?? weg
}
