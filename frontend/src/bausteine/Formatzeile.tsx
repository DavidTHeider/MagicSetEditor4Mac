// **Ein Format bearbeiten** — Befund 8 des Zweitblicks über Phase 4.6
// (#1293).
//
// # Warum es diese Datei gibt
//
// Die Liste zeigte die vier Zahlen nur **an**, und „Speichern" reichte
// dasselbe DTO unverändert zurück. Die Dokumentation versprach ein Format,
// „dessen Zahlen sich danach ändern lassen" — an allen vier Flächen. Die drei
// ohne Fenster konnten es, das Fenster nicht.
//
// # Was hier zu sehen sein muss
//
// **Null heißt dreierlei**, und der Hinweis daneben sagt es: keine Grenze,
// kein Sideboard, beliebig oft. Wer die Zahlen bearbeitbar macht, muss das
// zeigen.
//
// # Und was Werkstück U3 dazugelegt hat (#1414)
//
// **Die Regeln stehen in `internal/domain/format.go` und werden von dort
// gelesen, nicht erfunden:**
//
//   - `Validate`: `name` ist Pflicht (`SchluesselPflichtfeld`).
//   - `zahlenPruefen`: jede der vier Zahlen wird gegen `< 0` geprüft — „Die
//     Zahl ist %d — negativ gibt es nicht." **Ohne Obergrenze**; eine hier
//     erfundene wiese ein gültiges Format ab (§11).
//   - `Hoechstgroesse > 0 && < Mindestgroesse` ist ein Befund über **zwei**
//     Felder. Er lässt sich am einzelnen Feld nicht entscheiden und kommt
//     deshalb als Feldfehler aus Go zurück; hier steht er über
//     `verursachendesFeld` an dem Feld, das Go nennt.
//
// **Der Befund wird deshalb hier gehalten und nicht über [Deckstand.fuehre]
// gemeldet**: der Haken macht aus der Ursache einen Satz und verliert dabei
// den Feldbezug. Dasselbe Muster wie in `seiten/Setangaben.tsx` — §7 Regel 3
// (nachgeladen wird nur nach einem geglückten Schreiben) und Regel 4 (ein
// Vorgang je Datensatz) stehen deshalb hier ausgeschrieben.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { Entscheidung } from './Entscheidung'
import { Hinweiskasten } from './Hinweiskasten'
import { Aktionsknopf } from './Aktionsknopf'
import { Pflichtfeld } from './Pflichtfeld'
import { Zahlenfeld } from './Zahlenfeld'
import { istKonflikt, meldungAus, verursachendesFeld } from '../meldung'
import { useTexte } from '../sprache/kontext'

import type { Deckstand } from '../haken/useDecks'
import type { FormatDTO } from '../apidecks'

// zahlenfelder sind die vier, deren Null je etwas anderes bedeutet.
const zahlenfelder = [
  'mindestgroesse',
  'hoechstgroesse',
  'sideboardgroesse',
  'hoechstzahlGleich',
] as const

// feldnamen sind die Felder, die diese Zeile zeigt — mit **Gos** Schreibweise
// (`domain.Format.Validate` und `zahlenPruefen`). Was Go sonst nennt (`spiel`
// etwa) steht hier nicht und bekommt seinen Satz in der Zeile.
const feldnamen: readonly string[] = ['name', ...zahlenfelder, 'scryfallname']

// Befund ist eine Meldung samt dem Feld, an dem sie hängt — leer, wenn Go
// keines nennt.
//
// **konflikt** sagt, dass der Stand inzwischen ein anderer ist (UX-08.AK-02,
// #1680) — dann bietet die Zeile an, ihn zu holen.
interface Befund {
  feld: string
  meldung: string
  konflikt: boolean
}

export function Formatzeile({
  stand,
  format,
}: {
  stand: Deckstand
  format: FormatDTO
}) {
  const texte = useTexte()

  // **Die Rückfrage vor dem Löschen** — UX-10.AK-04.
  const [fragt, setzeFragt] = useState(false)
  const eigene = texte.seiten.decks

  // **Der Entwurf steht neben dem Stand, nicht darin** (§7, Regel 1): ein
  // Nachladen räumt ihn mit, weil der `key` in der Liste an Kennung **und
  // Fassung** hängt.
  //
  // **Die Fassung stand dort erst seit dem 2026-09-01.** Vorher hing der
  // `key` allein an der Kennung — die sich beim Speichern nicht ändert —,
  // und dieser Satz behauptete einen Mechanismus, den es nicht gab: der
  // Entwurf blieb stehen, und der zweite Speichervorgang schickte die alte
  // Fassung gegen einen Bestand, der schon weiter war (Befund 3 des
  // Nachlaufs über Phase 4.6).
  const [entwurf, setzeEntwurf] = useState(format)

  const [befund, setzeBefund] = useState<Befund>()

  const speichern = async () => {
    // §7, Regel 4: ein Vorgang je Datensatz.
    const geglueckt = await stand.vorgang.starte(
      `format:${format.id}`,
      async () => {
        try {
          const antwort = await api().FormatSetzen(entwurf)

          auspacken(antwort, () => antwort.format)

          return true
        } catch (ursache: unknown) {
          setzeBefund({
            feld: verursachendesFeld(ursache, ''),
            meldung: meldungAus(ursache, texte),
            konflikt: istKonflikt(ursache),
          })

          return false
        }
      },
    )

    // §7, Regel 3: nachgeladen wird **nur** nach einem geglückten Schreiben —
    // sonst überschriebe ein Lesefehler die Meldung, die sagt, was zu tun ist.
    if (geglueckt === true) {
      setzeBefund(undefined)
      stand.neuLaden()
    }
  }

  const loeschen = async () => {
    await stand.fuehre(`format:${format.id}`, async () => {
      auspacken(await api().FormatLoeschen(format.id), () => undefined)
    })
  }

  const laeuft = stand.vorgang.laeuft(`format:${format.id}`)

  // fuerFeld gibt den Befund, wenn er zu diesem Feld gehört.
  const fuerFeld = (feld: string) =>
    befund?.feld === feld ? befund.meldung : undefined

  return (
    <div className="flex flex-wrap items-end gap-3">
      <Pflichtfeld
        beschriftung={eigene.name}
        wert={entwurf.name}
        angehalten={laeuft}
        befund={fuerFeld('name')}
        geaendert={(name) => {
          setzeEntwurf({ ...entwurf, name })
        }}
      />

      {zahlenfelder.map((feld) => (
        <div key={feld} className="w-28">
          {/* **`kleinstens={0}` und keine Obergrenze** — genau das, was
              `zahlenPruefen` sagt. `ganzzahlig`, weil das Feld dahinter in Go
              ein `int` ist; eine Kommazahl käme dort als Umwandlungsfehler an
              statt als Befund am Feld. */}
          <Zahlenfeld
            beschriftung={eigene[feld]}
            wert={entwurf[feld]}
            kleinstens={0}
            ganzzahlig
            schritt={1}
            angehalten={laeuft}
            befund={fuerFeld(feld)}
            geaendert={(wert) => {
              setzeEntwurf({ ...entwurf, [feld]: wert })
            }}
          />
        </div>
      ))}

      <label className="flex items-center gap-2 pb-2">
        <input
          type="checkbox"
          checked={entwurf.sideboardOffen}
          onChange={(ereignis) => {
            setzeEntwurf({
              ...entwurf,
              sideboardOffen: ereignis.target.checked,
            })
          }}
          className="size-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />

        <span className="text-sm">{eigene.sideboardOffen}</span>
      </label>

      <label className="flex flex-col gap-1">
        <span className="text-sm">{eigene.scryfallname}</span>

        <input
          value={entwurf.scryfallname}
          aria-invalid={fuerFeld('scryfallname') !== undefined}
          onChange={(ereignis) => {
            setzeEntwurf({ ...entwurf, scryfallname: ereignis.target.value })
          }}
          className="rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-[invalid=true]:border-fehler"
        />
      </label>

      <Aktionsknopf
        aktion="sichern"
        bezeichnung={`${eigene.formatSichern}: ${format.name}`}
        angehalten={laeuft}
        aufKlick={() => {
          void speichern()
        }}
      >
        {eigene.formatSichern}
      </Aktionsknopf>

      {/* **Die Rückfrage vor dem Löschen** — UX-10.AK-04
          (#1682).

          **Weil es keine Rücknahme gibt**: ein Format trägt Name, Zahlen und
          seine Liste erlaubter Sets, und der Verlauf eines Sets kennt es
          nicht. Der Satz darunter gehört vor den Klick und nicht **neben**
          den Knopf. */}
      {fragt ? (
        <Entscheidung
          frage={einsetzen(eigene.formatLoeschenFrage, { name: format.name })}
          erklaerung={eigene.formatLoeschenHinweis}
          bestaetigen={eigene.formatLoeschenJa}
          aktion="loeschen"
          abbrechen={texte.bausteine.aktionen.abbrechen}
          laeuft={laeuft}
          aufBestaetigen={() => {
            setzeFragt(false)
            void loeschen()
          }}
          aufAbbrechen={() => {
            setzeFragt(false)
          }}
        />
      ) : (
        <Aktionsknopf
          aktion="loeschen"
          bezeichnung={`${eigene.formatLoeschen}: ${format.name}`}
          angehalten={laeuft}
          aufKlick={() => {
            setzeFragt(true)
          }}
        >
          {eigene.formatLoeschen}
        </Aktionsknopf>
      )}

      {/* **Was zu keinem Feld gehört, steht in der Zeile** — ein Konflikt über
          die Fassung etwa nennt kein Feld, und ein Satz ohne Adresse gehört
          wenigstens an den Datensatz, um den es geht (§6). */}
      {/* **Und ein Konflikt bietet den heutigen Stand an** (UX-08.AK-02):
          der Entwurf bleibt stehen, bis jemand ihn holt — dann räumt der
          `key` aus Kennung und Fassung die Zeile ab (§7, Regel 1). */}
      {befund !== undefined && !feldnamen.includes(befund.feld) && (
        <div className="basis-full">
          <Hinweiskasten
            art="fehler"
            meldung={befund.meldung}
            handlung={
              befund.konflikt
                ? {
                    beschriftung: texte.bausteine.zustaende.neuLaden,
                    aufKlick: () => {
                      setzeBefund(undefined)
                      stand.neuLaden()
                    },
                  }
                : undefined
            }
          />
        </div>
      )}
    </div>
  )
}
