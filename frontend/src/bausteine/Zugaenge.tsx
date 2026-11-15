// **Die Zugangsdaten fremder Dienste** — Werkstück 8 von Phase 4.8
// (#1321),
// entschieden in [0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md)
// und [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md).
//
// # Was hier nie steht
//
// **Der Wert.** Gelesen wird nur, **ob** eine Angabe hinterlegt ist; ein Weg,
// der den Schlüssel ins Fenster zurückgäbe, machte aus dem Schlüsselbund eine
// Durchreiche — der Wert stünde im JSON der Bindung, im Speicher des Browsers
// und in jedem Mitschnitt.
import { useId, useState } from 'react'

import { api } from '../api'
import { auspacken, type Umschlag } from '../umschlag'
import { meldungAus } from '../meldung'
import { dienstnamen } from './dienstnamen'
import { Aktionsknopf } from './Aktionsknopf'
import { Entscheidung } from './Entscheidung'
import { Knopf } from './Knopf'
import { Satz } from './Satz'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import type { ZugangsstandDTO } from '../apizugaenge'

const feldklassen =
  'rounded-md border border-kante px-2 py-1 focus-visible:outline ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'

export function Zugaenge() {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  const [eingaben, setzeEingaben] = useState<Record<string, string>>({})

  // **Über `useLaden` und nicht über einen eigenen Effekt** (§7): der Haken
  // hält den Wächter über den Stand, das Räumen beim Wechsel und das
  // Nachladen — dreimal dieselbe Zusage, die sonst hier noch einmal stünde.
  //
  // **Der erste Anlauf setzte den Stand im Effekt selbst**, und der Linter
  // hat es angestrichen („Avoid calling setState() directly within an
  // effect"). Er hatte recht: es war ein zweiter Ladeweg neben dem, den es
  // gibt.
  const stand = useLaden(holen, [])

  // **Nur die Dienste, die Zugangsdaten brauchen.** Eine leere Zeile für
  // Scryfall wäre eine Frage ohne Antwort.
  const mitZugang = (stand.daten ?? []).filter(
    (eintrag) => eintrag.angaben.length > 0,
  )

  if (stand.fehler !== undefined) {
    return (
      <p role="alert" className="mt-4 text-fehler">
        {texte.bausteine.zustaende.fehler} {stand.fehler}
      </p>
    )
  }

  if (mitZugang.length === 0) {
    return null
  }

  return (
    // **Ohne eigene Überschrift seit #2465**: die Zugangsdaten stehen in
    // einem Abschnitt der Einstellungen, und dessen Knopf trägt den Namen.
    <section className="flex flex-col gap-3">
      <p className="max-w-prose text-sm text-leise">
        {/* **Über `Satz` und nicht roh** — die Kataloge benutzen `**…**`, und
            roh im JSX stehen die Sternchen wörtlich im Fenster. Dasselbe ist
            schon in #1100
            geschehen; behoben wurde damals die Stelle, nicht die Klasse.
            Jetzt hält `src/auszeichnung.test.tsx` jede Seite dagegen. */}
        <Satz text={eigene.zugaengeHinweis} />
      </p>

      {mitZugang.map((eintrag) => (
        <Dienstzugang
          key={eintrag.dienst}
          stand={eintrag}
          eingaben={eingaben}
          setzeEingabe={(schluessel, wert) => {
            setzeEingaben((vorher) => ({ ...vorher, [schluessel]: wert }))
          }}
          nachladen={stand.nachladen}
        />
      ))}
    </section>
  )
}

// holen fragt die Bindung.
async function holen(): Promise<ZugangsstandDTO[]> {
  const antwort = await api().Zugaenge()

  return auspacken(antwort, () => antwort.zugaenge)
}

// Dienstzugang zeigt die Angaben eines Dienstes.
function Dienstzugang({
  stand,
  eingaben,
  setzeEingabe,
  nachladen,
}: {
  stand: ZugangsstandDTO
  eingaben: Record<string, string>
  setzeEingabe: (schluessel: string, wert: string) => void
  nachladen: () => void
}) {
  const name = dienstnamen[stand.dienst] ?? stand.dienst

  return (
    <fieldset
      aria-label={name}
      // **`border-kante` und nicht 400**: die 400 liegt hell bei 2,63:1
      // und reißt damit 1.4.11 (gemessen, `docs/KONTRASTE.md`). Die 500 hält
      // 4,35 hell und 4,23 dunkel — sie steht dort schon für genau diesen
      // Fall, „Rand einer Fläche auf dem Anwendungsgrund".
      className="flex flex-col gap-2 rounded-md border border-kante p-3"
    >
      <legend className="px-1 font-medium">{name}</legend>

      {stand.angaben.map((angabe) => (
        <Angabezeile
          key={angabe}
          dienst={stand.dienst}
          name={name}
          angabe={angabe}
          liegt={stand.hinterlegt.includes(angabe)}
          wert={eingaben[`${stand.dienst}.${angabe}`] ?? ''}
          setzeWert={(wert) => {
            setzeEingabe(`${stand.dienst}.${angabe}`, wert)
          }}
          nachladen={nachladen}
        />
      ))}
    </fieldset>
  )
}

// Angabezeile ist eine Angabe samt ihren beiden Knöpfen.
function Angabezeile({
  dienst,
  name,
  angabe,
  liegt,
  wert,
  setzeWert,
  nachladen,
}: {
  dienst: string
  name: string
  angabe: string
  liegt: boolean
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

  // **Gelöscht wird erst nach einer Frage** — UX-10.AK-04 (#1682): der Wert
  // ist danach aus dem Schlüsselbund weg, und zurück gibt ihn nur der Dienst,
  // der ihn ausgegeben hat. Bis zum 2026-09-25 genügte ein Klick.
  const [fragt, setzeFragt] = useState(false)

  // **Der Name der Angabe kommt aus dem Katalog, nicht aus `domain`.**
  // `marke`, `schluessel` und `team` sind Bezeichner (§9) und gehören nicht
  // roh ins Fenster. Dass es zu jeder einen Namen gibt, hält
  // `internal/werkzeuge/einstellungstexte_test.go` gegen die Go-Seite;
  // `keintextimjsx.test.ts` kann es nicht sehen, weil hier eine Variable steht
  // und keine Zeichenkette.
  //
  // # Zwei Stufen, und die erste ist der Dienst
  //
  // **Weil eine Angabeart mehrere Dienste bedient** (#1534): `marke` steht
  // für MTGJSON **und** CardTrader. Bis zum 2026-09-12 gab es nur die
  // Angabeart, und an CardTrader stand „Patreon-Marke" — der Name eines
  // Unterstützerprogramms, mit dem CardTrader nichts zu tun hat.
  //
  // **Der Rückfall bleibt**, und er ist kein Rest: ein Dienst, den Go kennt
  // und der Katalog noch nicht, bekommt so einen Namen statt seines rohen
  // Bezeichners.
  const jeDienst: Record<string, string | undefined> = eigene.dienstangaben
  const namen: Record<string, string | undefined> = eigene.angaben
  const angabename = jeDienst[`${dienst}.${angabe}`] ?? namen[angabe] ?? angabe

  // **Ein Vorgang je Angabe** (§7, Regel 4) — und **einer** für Eintragen und
  // Entfernen: sie schließen einander aus, und getrennte Schlüssel ließen
  // beides an derselben Angabe gleichzeitig laufen.
  //
  // **Befund 5 des Zweitblicks.** `docs/SCHREIBWEGE.md` behauptete, zwei
  // Aufrufe für dieselbe Angabe seien nicht möglich — „das Feld ist nach dem
  // ersten leer". Geleert wird es erst **nach** der Antwort, und der Knopf
  // steht die ganze Zeit da.
  const schluessel = `${dienst}.${angabe}`

  // **Befund 3 des Zweitblicks.** Beide Wege gaben die Antwort ungelesen weg:
  // ein Schlüsselbundfehler verschwand (§8), und nachgeladen wurde trotzdem
  // (§7, Regel 3). Die Go-Seite prüft diesen Fehlerweg ausdrücklich — nur
  // ankommen konnte er nicht.
  const zug = async (arbeit: () => Promise<Umschlag>, danach: () => void) => {
    await vorgang.starte(schluessel, async () => {
      try {
        const antwort = await arbeit()

        auspacken(antwort, () => undefined)
        setzeFehler(undefined)
        danach()

        // **Geschrieben ist geschrieben** (§7, Regel 3): das Nachladen danach
        // ist Bestenfalls-Arbeit und steht deshalb hinter dem Erfolg.
        nachladen()
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  const setzen = () =>
    zug(
      () => api().ZugangSetzen(dienst, angabe, wert),
      () => {
        setzeWert('')
      },
    )

  const loeschen = () =>
    zug(
      () => api().ZugangLoeschen(dienst, angabe),
      () => undefined,
    )

  return (
    <div className="flex flex-col gap-1">
      <div className="flex flex-wrap items-end gap-2">
        <label htmlFor={feldID} className="flex flex-col gap-1">
          <span className="text-sm">
            {angabename}: {liegt ? eigene.zugangLiegt : eigene.zugangFehlt}
          </span>

          {/* **Ein Passwortfeld.** Wer den Schlüssel eintippt, soll ihn nicht
              auf dem Bildschirm stehen haben. */}
          <input
            id={feldID}
            type="password"
            value={wert}
            aria-label={`${name}: ${angabename}`}
            // **Der Fehler hängt am Feld** (§6): wer ihn hört, hört ihn beim
            // Feld, das ihn verursacht hat — nicht als anonymen Hinweis.
            aria-describedby={fehler === undefined ? undefined : fehlerID}
            onChange={(ereignis) => {
              setzeWert(ereignis.target.value)
            }}
            className={feldklassen}
          />
        </label>

        <Knopf
          bezeichnung={`${eigene.zugangSetzen}: ${name} ${angabename}`}
          angehalten={vorgang.laeuft(schluessel)}
          aufKlick={() => {
            void setzen()
          }}
        >
          {eigene.zugangSetzen}
        </Knopf>

        {liegt && (
          <Aktionsknopf
            aktion="loeschen"
            bezeichnung={`${eigene.zugangLoeschen}: ${name} ${angabename}`}
            angehalten={vorgang.laeuft(schluessel) || fragt}
            aufKlick={() => {
              setzeFragt(true)
            }}
          >
            {eigene.zugangLoeschen}
          </Aktionsknopf>
        )}
      </div>

      {liegt && fragt && (
        <Entscheidung
          frage={einsetzen(eigene.zugangLoeschenFrage, {
            angabe: angabename,
            name,
          })}
          erklaerung={eigene.zugangLoeschenErklaerung}
          bestaetigen={eigene.zugangLoeschenJa}
          aktion="loeschen"
          abbrechen={texte.bausteine.aktionen.abbrechen}
          laeuft={vorgang.laeuft(schluessel)}
          aufBestaetigen={() => {
            setzeFragt(false)
            void loeschen()
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
