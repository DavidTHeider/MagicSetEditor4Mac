// **Die Formate** — Werkstück 7 von Phase 4.6
// (#1293),
// nach Vermerk 0061: „Daten, nicht Code".
//
// # Warum hier eine Bauart zur Wahl steht
//
// Weil die Zahlen dreier Bauarten **nachgelesen** sind — CR 100.2a, 100.4a,
// 903.5a/b/e und 100.2b. Alles andere wäre geraten, und geraten wird hier
// nicht (§10). Wer ein Format ohne Bauart anlegt, setzt seine Zahlen selbst.
//
// # Ein Format ist eine Zeile, sein Formular steht auf Wunsch
//
// **UXR-11** (#1919).
// Jedes Format stand als offenes Formular da — Name, vier Grenzen, „Sideboard
// ohne Grenze", Scryfall-Name, Sichern, Löschen. Bei zehn Formaten siebzig
// Felder auf einer Lasche (R3 in `docs/UIREVIEW.md`). Jetzt trägt die Zeile
// den Namen und „Bearbeiten", und das Formular steht für **eines** zugleich.
import { useEffect, useId, useState } from 'react'

import { useSpielnamen } from '../haken/useSpielnamen'
import { Vorschlagsfeld } from './Vorschlagsfeld'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Formatzeile } from './Formatzeile'
import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'

import type { BauartDTO } from '../apidecks'
import type { Deckstand } from '../haken/useDecks'
import { Pflichtfeld } from './Pflichtfeld'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

export function Formatverwaltung({ stand }: { stand: Deckstand }) {
  const texte = useTexte()
  const eigene = texte.seiten.decks

  const [bauarten, setzeBauarten] = useState<BauartDTO[]>([])

  // bearbeitet ist das eine Format, dessen Formular gerade steht.
  const [bearbeitet, setzeBearbeitet] = useState<string>()
  const formularID = useId()
  const bearbeiten = texte.bausteine.aktionskatalog.aktionen.bearbeiten.kurz
  const [bauart, setzeBauart] = useState('konstruiert')
  const [name, setzeName] = useState('')
  const [spiel, setzeSpiel] = useState('magic')
  const spielnamen = useSpielnamen()
  const [scryfallname, setzeScryfallname] = useState('')

  // **Die Bauarten hängen an keinem Bestand** — sie stehen in `domain`, und
  // der Ruf kann nicht scheitern. Deshalb einmal beim Erscheinen und nicht
  // über `useLaden`.
  useEffect(() => {
    void (async () => {
      const antwort = await api().Bauarten()

      setzeBauarten(antwort.bauarten)
    })()
  }, [])

  const anlegen = async () => {
    await stand.fuehre('format-anlegen', async () => {
      const antwort = await api().FormatAusBauart(
        bauart,
        name,
        spiel,
        scryfallname,
      )

      auspacken(antwort, () => antwort.format)

      setzeName('')
      setzeScryfallname('')
    })
  }

  return (
    <section aria-label={eigene.formate} className="flex flex-col gap-3">
      <h2 className="text-lg font-medium">{eigene.formate}</h2>

      {stand.formate.length === 0 ? (
        <p>{eigene.formateLeer}</p>
      ) : (
        <ul className="flex flex-col gap-3">
          {stand.formate.map((format) => (
            // **Der `key` trägt die Fassung mit** (Befund 3 des Nachlaufs
            // über Phase 4.6). Hing er allein an der Kennung, blieb der
            // Entwurf in `Formatzeile` über ein Speichern hinweg stehen —
            // samt seiner **alten** Fassung, und der zweite Handgriff endete
            // im Konflikt. Das ist die Gestalt aus `docs/SCHREIBWEGE.md`: ein
            // `key`, der das Bauteil abräumt, löst §7 Regel 1 und 2 zugleich.
            <li
              key={`${format.id}:${format.fassung}`}
              className="flex flex-col gap-2"
            >
              <div className="flex flex-wrap items-center gap-3">
                <span className="font-medium">{format.name}</span>

                <Knopf
                  klein
                  gewicht="leise"
                  bezeichnung={`${bearbeiten}: ${format.name}`}
                  ausgeklappt={bearbeitet === format.id}
                  steuert={`${formularID}-${format.id}`}
                  aufKlick={() => {
                    setzeBearbeitet(
                      bearbeitet === format.id ? undefined : format.id,
                    )
                  }}
                >
                  {bearbeiten}
                </Knopf>
              </div>

              {/* **Jedes Format ist bearbeitbar** — Befund 8 des Zweitblicks
                  über Phase 4.6. Vorher zeigte die Liste die vier Zahlen nur
                  an, und „Speichern" reichte dasselbe DTO unverändert
                  zurück: eine Fähigkeit, die die Dokumentation behauptete und
                  das Fenster nicht hatte. */}
              {bearbeitet === format.id && (
                <div id={`${formularID}-${format.id}`}>
                  <Formatzeile stand={stand} format={format} />
                </div>
              )}
            </li>
          ))}
        </ul>
      )}

      <Erklaerungsabsatz
        text={eigene.formatLoeschenHinweis}
        klasse="text-sm text-leise"
      />

      <h3 className="font-medium">{eigene.formatNeu}</h3>

      <div className="flex flex-wrap items-end gap-3">
        <label className="flex flex-col gap-1">
          <span className="text-sm">{eigene.bauart}</span>

          <select
            value={bauart}
            onChange={(ereignis) => {
              setzeBauart(ereignis.target.value)
            }}
            className="rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            {bauarten.map((eine) => (
              <option key={eine.bauart} value={eine.bauart}>
                {eine.bauart}
              </option>
            ))}
          </select>
        </label>

        {/* **Pflichtfeld mit Satz am Feld** — Werkstück F von Phase 4.10 (§6). */}
        <Pflichtfeld
          beschriftung={eigene.name}
          wert={name}
          geaendert={setzeName}
        />

        {/* **Vorschläge aus den installierten Spielen, Freitext bleibt** —
            UX-06.AK-01 (#1678), der Grund steht in `useSpielnamen`. */}
        <Vorschlagsfeld
          beschriftung={eigene.spiel}
          wert={spiel}
          setzeWert={setzeSpiel}
          vorschlaege={spielnamen}
          breite={12}
        />

        <label className="flex flex-col gap-1">
          <span className="text-sm">{eigene.scryfallname}</span>

          <input
            value={scryfallname}
            onChange={(ereignis) => {
              setzeScryfallname(ereignis.target.value)
            }}
            className="rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
        </label>

        <Knopf
          gewicht="haupt"
          angehalten={stand.vorgang.laeuft('format-anlegen')}
          aufKlick={() => {
            void anlegen()
          }}
        >
          {eigene.formatNeu}
        </Knopf>
      </div>

      <Erklaerungsabsatz
        text={eigene.bauartHinweis}
        klasse="text-sm text-leise"
      />

      <Erklaerungsabsatz
        text={eigene.scryfallnameHinweis}
        klasse="text-sm text-leise"
      />

      <Erklaerungsabsatz
        text={eigene.grenzenHinweis}
        klasse="text-sm text-leise"
      />
    </section>
  )
}
