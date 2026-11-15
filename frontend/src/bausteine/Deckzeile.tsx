// **Eine Zeile der Deckübersicht** — herausgelöst aus `Deckliste.tsx`, als
// das offene Deck sein Menü „…“ und den Rechtsklick bekam (AP-6,
// #2363,
// Befund B-03). Dort steht, **welche** Decks es gibt und welches offen ist;
// hier, was an einem Deck zu tun ist.
//
// # Hinter einem Knopf und im Rechtsklick
//
// Umbenennen, Duplizieren, Teilen und Löschen standen seit UXR-11 (#1919)
// als vier Knöpfe am offenen Deck — im Bild vier Bedienelemente mehr für
// Handlungen, die man an einem Deck selten tut (`Bereiche09`: höchstens
// zwanzig bei 1.440 × 838). **Eine Liste für beide Wege** (U2, #1413):
// das Menü „…“ und der Rechtsklick an der Zeile zeigen dieselben Einträge.
// Ein Rechtsklick auf ein anderes Deck öffnet es zuerst (`useAnstoss`).
//
// **Liste und Zustand stehen eine Ebene höher** (#2469, `useDeckhandlung`):
// die Menüleiste meldet sie auch dann an, wenn diese Zeile fehlt. Hier steht,
// wie sie aussehen.
import type { KeyboardEvent, MouseEvent } from 'react'

import type { DeckDTO } from '../apidecks'
import { api } from '../fassade'
import type { Deckhandlung } from '../haken/useDeckhandlung'
import type { Deckstand } from '../haken/useDecks'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Aktionsknopf } from './Aktionsknopf'
import { Deckkopieformular } from './Deckkopie'
import { Entscheidung } from './Entscheidung'
import { Knopf } from './Knopf'
import { Kontextmenue, type Anstoss } from './Kontextmenue'
import type { Gestalt } from './Listenansicht'
import { Menue } from './Menue'
import { Pflichtfeld } from './Pflichtfeld'
import { Popover } from './Popover'
import { Teilen } from './Teilen'
import { auspacken } from '../umschlag'

// markenklassen zeichnen die Marke „offen“ wie die Marke an einem Druck —
// klein, gerahmt, in der Textfarbe (`docs/KONTRASTE.md`).
const markenklassen =
  'rounded-full border border-marke px-2 py-0.5 text-xs text-text'

interface Zeileneigenschaften {
  deck: DeckDTO
  offen: boolean
  gestalt: Gestalt
  stand: Deckstand
  handlung: Deckhandlung
  umbenennen: (
    deckID: string,
    fassung: string,
    name: string,
    notiz: string,
  ) => Promise<void>

  // anstoss, anstossVerbraucht und anstossen sind der Rechtsklick und ⇧F10
  // an der Zeile — `useAnstoss` in `Deckliste`.
  anstoss: Anstoss | undefined
  anstossVerbraucht: () => void
  anstossen: (
    ereignis: MouseEvent<HTMLElement> | KeyboardEvent<HTMLElement>,
  ) => void
}

export function Deckzeile({
  deck,
  offen,
  gestalt,
  stand,
  handlung,
  umbenennen,
  anstoss,
  anstossVerbraucht,
  anstossen,
}: Zeileneigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.decks
  const { umbenannt, setzeUmbenannt, fragen, kopie } = handlung
  const gefragt = handlung.gefragt === deck.id

  const laeuft = stand.vorgang.laeuft(`deck:${deck.id}`)

  const loeschen = async () => {
    await stand.fuehre(`deck:${deck.id}`, async () => {
      auspacken(await api().DeckLoeschen(deck.id), () => undefined)

      // **Das gelöschte Deck bleibt nicht aufgeschlagen** (§7, Regel 1) —
      // und eine unterwegs befindliche Leseantwort schlägt es nicht wieder
      // auf (Regel 2). Beides tut `vergiss`; hier steht nur, **wann**.
      stand.vergiss(deck.id)

      fragen('')
    })
  }
  const bezeichnung = einsetzen(eigene.handlungenAn, { name: deck.name })
  const eintraege = offen ? handlung.eintraege(deck) : []

  const umbenennend = umbenannt?.id === deck.id ? umbenannt : undefined

  return (
    <Kontextmenue
      // **Das offene Deck trägt den Rand der Marke** — gemessen gegen die
      // Fläche (`docs/KONTRASTE.md`); eine betonte Fläche trüge auch die
      // leisen Knöpfe darin, und deren Paar ist nicht gemessen.
      klasse={`${
        gestalt === 'kachel'
          ? 'flex h-full flex-col items-start gap-3 rounded-lg border bg-flaeche p-4'
          : 'flex flex-wrap items-center gap-3 rounded-md border p-3'
      } ${offen ? 'border-marke' : 'border-kante'}`}
      bezeichnung={bezeichnung}
      eintraege={eintraege}
      objekt={offen ? `deck/${deck.id}` : undefined}
      nurAngestossen
      anstoss={offen ? anstoss : undefined}
      anstossVerbraucht={anstossVerbraucht}
      anstossen={anstossen}
    >
      <span className="font-medium">{deck.name}</span>
      <span className="text-sm text-leise">{deck.spiel}</span>
      <span className="text-sm tabular-nums">
        {einsetzen(eigene.zaehlung, {
          haupt: String(deck.hauptdeck),
          side: String(deck.sideboard),
        })}
      </span>

      {/* **Am offenen Deck kein „Öffnen“** — AP-6 (#2363): es ist offen, und
          ein Knopf, der nichts mehr tut, zählte zu den Bedienelementen im
          Bild (R2). Die Marke sagt es in Worten, der Rand in der Farbe. */}
      {offen ? (
        <span className={markenklassen}>{eigene.offenMarke}</span>
      ) : (
        <Knopf
          bezeichnung={`${eigene.waehlen}: ${deck.name}`}
          aufKlick={() => {
            void stand.waehle(deck.id)
          }}
        >
          {eigene.waehlen}
        </Knopf>
      )}

      {/* **Was man an einem Deck tut, steht am offenen** — UXR-11 (#1919):
          sechs Bedienelemente je Zeile machten bei acht Decks 48 Knöpfe im
          Navigator. **Umbenennen und Kopieren sind eine kurze Eingabe an Ort
          und Stelle** — ein Popover am Menü (UX-08.AK-01, #1680): die Zeile
          bleibt, wo sie ist. */}
      {offen && (
        <Popover
          ausloeser={
            // **Mit Rahmen, wie „Öffnen“ daneben** — leise stand das Zeichen
            // ohne Kante neben einem Knopf mit einer und las sich wie
            // verirrte Punkte (Sichtprüfung am nativen Fenster, AP-6).
            <Menue bezeichnung={bezeichnung} eintraege={eintraege}>
              {eigene.handlungenKurz}
            </Menue>
          }
          offen={umbenennend !== undefined || kopie.name !== undefined}
          titel={einsetzen(
            umbenennend === undefined
              ? eigene.kopieTitel
              : eigene.umbenennenTitel,
            { name: deck.name },
          )}
          schliessen={() => {
            setzeUmbenannt(undefined)
            kopie.abbrechen()
          }}
        >
          {umbenennend !== undefined && (
            <>
              <Pflichtfeld
                beschriftung={eigene.neuerName}
                wert={umbenennend.name}
                geaendert={(name) => {
                  setzeUmbenannt({ id: deck.id, name })
                }}
              />

              <div className="flex flex-wrap gap-2">
                <Aktionsknopf
                  aktion="sichern"
                  gewicht="haupt"
                  angehalten={umbenennend.name.trim() === '' || laeuft}
                  aufKlick={() => {
                    void umbenennen(
                      deck.id,
                      deck.fassung,
                      umbenennend.name.trim(),
                      deck.notiz,
                    )
                  }}
                >
                  {eigene.nameSichern}
                </Aktionsknopf>

                <Knopf
                  aufKlick={() => {
                    setzeUmbenannt(undefined)
                  }}
                >
                  {eigene.umbenennenAbbrechen}
                </Knopf>
              </div>
            </>
          )}

          {kopie.name !== undefined && (
            <Deckkopieformular lauf={kopie} name={kopie.name} />
          )}
        </Popover>
      )}

      {offen && (
        <Teilen
          key={deck.id}
          art="deck"
          bezug={deck.id}
          anlass={handlung.teilanlass}
        />
      )}

      {offen && gefragt && (
        <div className="basis-full">
          <Entscheidung
            frage={eigene.loeschenFrage}
            bestaetigen={eigene.loeschenJa}
            aktion="loeschen"
            abbrechen={eigene.loeschenNein}
            laeuft={laeuft}
            aufBestaetigen={() => {
              void loeschen()
            }}
            aufAbbrechen={() => {
              fragen('')
            }}
          />
        </div>
      )}
    </Kontextmenue>
  )
}
