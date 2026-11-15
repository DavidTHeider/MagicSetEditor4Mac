// **Ein Deck verleihen oder leihen** — Werkstück G6 von Phase 5.5
// (#1463).
//
// # Drei Sätze, die dabeistehen und nicht bloß gelten
//
// **Weil eine Leihe nach mehr aussieht, als sie ist.** Das Issue verlangt sie
// ausdrücklich: „UI erklärt digitale Nutzungsfreigabe, keine physische
// Eigentums-/Bestandsänderung und kein technischer Kopierschutz."
//
//   - Es ist eine **Nutzungsfreigabe**: es wechselt nichts den Besitzer.
//   - Am **Bestand** ändert sich keine Karte, weder hier noch drüben.
//   - Es ist **kein Kopierschutz**: wer die Karten gesehen hat, hat sie
//     gesehen. Geregelt wird, womit dieses Programm eine Partie startet.
//
// Alle drei sind Zusagen, die dieses Programm hält — und alle drei wären
// missverständlich, wenn man sie verschwiege (§6, §10).
//
// # Drei Gruppen, und offen ist eine
//
// **Seit UXR-13** (#1921)
// steht die Leihe in einer Schublade der Spielenseite, und die gibt ihr den
// Namen. Darin führt das **Verleihen** — mit seinem Knopf als Hauptknopf —,
// das **Annehmen** eines Umschlags steht eingeklappt, und die **Leihen** samt
// Rücknahme darunter. Bis dahin stand alles offen untereinander unter dem
// Anfang einer Partie, ohne einen Knopf, der führte (R1, R3 in
// `docs/UIREVIEW.md`).
//
// # Ziel und Umschlag werden gewählt
//
// **Seit UXR-F15** (#2030)
// wählt ein Dateidialog, *wohin* der Umschlag gelegt wird und *welcher*
// angenommen wird. Ein Pfad von Hand ist auf macOS kein Bedienelement (R7);
// der gewählte steht lesbar daneben und lässt sich durch eine neue Wahl
// ändern. Abbrechen lässt den vorigen stehen (`usePfadwahl`). An Kommandozeile
// und MCP bleibt der Pfad ein Parameter — einen Dateiwahldialog ohne Fenster
// gibt es nicht (§16).
import { useId, useState } from 'react'

import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Auswahlfeld } from '../bausteine/Auswahlfeld'
import { Beschriftetesfeld } from '../bausteine/Beschriftetesfeld'
import { Knopf } from '../bausteine/Knopf'
import { Kontaktwahl } from '../bausteine/Kontaktwahl'
import { Ladefehler, Wirdgeladen } from '../bausteine/Ladehinweise'
import { Pfadwahl } from '../bausteine/Pfadwahl'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useDeckleihe } from '../haken/useDeckleihe'
import { useLaden } from '../haken/useLaden'
import { usePfadwahl } from '../haken/usePfadwahl'
import { Zahlenfeld } from '../bausteine/Zahlenfeld'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Leihenliste } from './Leihenliste'

import type { DeckDTO } from '../apidecks'

// wahlklassen ist die Deckwahl — **mit sichtbarem Fokusring** (§6) und dem
// Hintergrund der Fläche, damit sie auch im dunklen Erscheinungsbild lesbar
// bleibt.
const wahlklassen =
  'rounded-md border border-kante bg-flaeche px-2 py-1 text-text ' +
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

// Vorgabestunden ist, wie lange eine Leihe ohne Zutun gilt.
//
// **Vierundzwanzig Stunden.** Eine Partie dauert einen Abend, ein Turnier ein
// Wochenende — und wer länger will, trägt es ein. Eine Vorgabe von Wochen
// wäre eine Leihe, an die sich niemand erinnert, wenn sie abläuft.
export const Vorgabestunden = 24

// laengsteLeiheInStunden ist die längste Dauer, die Go fasst, in ganzen
// Stunden: 2⁶³ − 1 Nanosekunden (J05, #2269). Darüber lief die Umrechnung
// um, und 5.124.096 Stunden wurden zu 25 Minuten; Go weist sie jetzt selbst
// ab, das Feld sagt es vorher.
const laengsteLeiheInStunden = 2_562_047

export function Deckleihe({
  kontaktstand = 0,
  geaendert,
}: {
  kontaktstand?: number

  // geaendert meldet ein geglücktes Schreiben an einer Leihe — P1b (#2370).
  geaendert?: (() => void) | undefined
}) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const lauf = useDeckleihe(geaendert)
  const verleihen = useId()
  const liste = useId()
  const deckwahl = useId()

  // **Das Deck wird gewählt, nicht getippt** (UX-08.AK-04, #1680) — dieselbe
  // Bindung wie auf der Deckseite, über alle Spiele: verliehen wird, was hier
  // liegt.
  const decks = useLaden<DeckDTO[]>(async () => {
    const antwort = await api().Decks('')

    return auspacken(antwort, () => antwort.decks)
  }, [])

  const [entleiher, setzeEntleiher] = useState('')
  const [deckID, setzeDeckID] = useState('')
  const [woran, setzeWoran] = useState('')
  const [stunden, setzeStunden] = useState(Vorgabestunden)

  // stundenGelten sagt, ob das Feld gerade eine Stundenzahl trägt, die
  // abgeschickt werden darf (F03, #2245). Bis dahin wurden „0", „-2" und
  // „0.5" still zur Vorgabe und „1.5" zu einer Stunde.
  const [stundenGelten, setzeStundenGelten] = useState(true)

  // **Gewählt und nicht getippt** (UXR-F15, #2030): zwei Dialoge, zwei Orte.
  // Das Ziel fragt der Sichern-Dialog, den Umschlag der Öffnen-Dialog — und
  // der liest ihn dabei, damit ein unlesbarer gar nicht erst dasteht.
  const zielwahl = usePfadwahl(() => api().DeckleiheZielWaehlen(), 'leihe-ziel')
  const umschlagwahl = usePfadwahl(
    () => api().UmschlagWaehlen(),
    'leihe-umschlag',
  )

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-2">
        <Erklaerungsabsatz
          text={eigene.leiheErklaerung}
          klasse="max-w-prose text-sm text-leise"
        />

        <Erklaerungsabsatz
          text={eigene.leiheKeinEigentum}
          klasse="max-w-prose text-xs text-leise"
        />

        <Erklaerungsabsatz
          text={eigene.leiheKeinKopierschutz}
          klasse="max-w-prose text-xs text-leise"
        />
      </div>

      {lauf.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {lauf.fehler}
        </p>
      )}

      <section aria-labelledby={verleihen} className="flex flex-col gap-2">
        <h3 id={verleihen} className="text-base font-semibold">
          {eigene.leiheVerleihenTitel}
        </h3>

        {/* **Gewählt und nicht getippt** (UX-08.AK-04, #1680): Entleiher und
            Deck kommen aus einer Wahl. **Die Partie bleibt nennbar** — sie
            ist die Ausnahme, und der Kasten sagt, wann sie gebraucht wird. */}
        <Kontaktwahl
          titel={eigene.leiheEntleiher}
          leer={eigene.leiheEntleiherLeer}
          gewaehlt={entleiher}
          waehlen={setzeEntleiher}
          stand={kontaktstand}
        />

        {/* **Warten, Fehlschlag, leer — und erst dann die Wahl**
            (UX-08.NAT-01): eine leere Auswahlliste, während die Decks noch
            kommen, sähe aus wie „es gibt keine". */}
        {decks.daten === undefined ? (
          decks.fehler === undefined ? (
            <Wirdgeladen />
          ) : (
            <Ladefehler fehler={decks.fehler} wiederholen={decks.nachladen} />
          )
        ) : decks.daten.length === 0 ? (
          <p className="text-sm text-leise">{eigene.leiheKeinDeck}</p>
        ) : (
          <div className="flex flex-col gap-1 text-sm">
            <label htmlFor={deckwahl}>{eigene.leiheDeck}</label>

            {/* **Ab fünfzehn Decks mit Suche** — UXR-F1
                (#2007),
                R7; die Beschriftung steht neben dem Feld und nicht um es. */}
            <Auswahlfeld
              id={deckwahl}
              bezeichnung={eigene.leiheDeck}
              eintraege={[
                { wert: '', name: eigene.leiheDeckWaehlen },
                ...decks.daten.map((deck) => ({
                  wert: deck.id,
                  name: deck.name,
                })),
              ]}
              wert={deckID}
              aendern={setzeDeckID}
              klasse={wahlklassen}
            />
          </div>
        )}

        {/* **Die Stunden sind eine ganze Zahl ab eins** (R7, F03) — ein
            Textfeld nahm „viele" an und verlieh dann stillschweigend für die
            Vorgabe. Die Vorgabe steht im Feld; wer sie löscht oder etwas
            anderes als eine ganze Stundenzahl tippt, hält den Knopf an. */}
        <Zahlenfeld
          beschriftung={eigene.leiheStunden}
          wert={stunden}
          kleinstens={1}
          groesstens={laengsteLeiheInStunden}
          ganzzahlig
          geaendert={setzeStunden}
          gilt={setzeStundenGelten}
        />

        <Pfadwahl
          bezeichnung={eigene.leiheZiel}
          pfad={zielwahl.pfad}
          kein={eigene.leiheZielKein}
          waehlen={eigene.leiheZielWaehlen}
          anderes={eigene.leiheZielAnderes}
          fehler={zielwahl.fehler}
          angehalten={zielwahl.laeuft}
          aufWahl={zielwahl.waehle}
        />

        <Aufklappkasten titel={eigene.leiheFachleute}>
          <div className="flex flex-col gap-2">
            <Erklaerungsabsatz
              text={eigene.leiheFachleuteErklaerung}
              klasse="max-w-prose text-sm text-leise"
            />

            <Beschriftetesfeld
              bezeichnung={eigene.leiheWoran}
              wert={woran}
              setzeWert={setzeWoran}
            />
          </div>
        </Aufklappkasten>

        <div>
          <Knopf
            gewicht="haupt"
            angehalten={
              lauf.laeuft ||
              !stundenGelten ||
              entleiher === '' ||
              deckID === '' ||
              zielwahl.pfad === ''
            }
            aufKlick={() => {
              lauf.anbieten(
                {
                  entleiher,
                  deckID,
                  bindungsart: 'partie',
                  woran: woran.trim(),
                  rechte: ['spielen'],
                  stunden,
                  modus: 'zuende',
                },
                zielwahl.pfad,
              )
            }}
          >
            {eigene.leiheAnbieten}
          </Knopf>
        </div>

        {lauf.abgelegt !== undefined && (
          <p role="status" className="text-sm">
            {eigene.leiheAbgelegt}{' '}
            <span className="font-mono text-xs">{lauf.abgelegt}</span>
          </p>
        )}
      </section>

      {/* **Das Annehmen eingeklappt** — offen ist eine Gruppe, nicht zwei
          (R3). Wer einen Umschlag bekommen hat, klappt es auf. */}
      <Aufklappkasten titel={eigene.leiheAnnehmenTitel}>
        <div className="flex flex-col gap-2">
          <Pfadwahl
            bezeichnung={eigene.leiheUmschlag}
            pfad={umschlagwahl.pfad}
            kein={eigene.leiheUmschlagKein}
            waehlen={eigene.leiheUmschlagWaehlen}
            anderes={eigene.leiheUmschlagAnderer}
            fehler={umschlagwahl.fehler}
            angehalten={umschlagwahl.laeuft}
            aufWahl={umschlagwahl.waehle}
          />

          <div>
            <Knopf
              gewicht="leise"
              angehalten={lauf.laeuft || umschlagwahl.pfad === ''}
              aufKlick={() => {
                lauf.annehmen(umschlagwahl.pfad)
              }}
            >
              {eigene.leiheAnnehmen}
            </Knopf>
          </div>
        </div>
      </Aufklappkasten>

      <section aria-labelledby={liste} className="flex flex-col gap-2">
        <h3 id={liste} className="text-base font-semibold">
          {eigene.leiheListeTitel}
        </h3>

        <div>
          <Knopf
            gewicht="leise"
            angehalten={lauf.laeuft}
            aufKlick={lauf.ansehen}
          >
            {eigene.leiheAnsehen}
          </Knopf>
        </div>

        <Leihenliste lauf={lauf} />

        <Erklaerungsabsatz
          text={eigene.leiheRuecknahmeGrenze}
          klasse="max-w-prose text-xs text-leise"
        />
      </section>
    </div>
  )
}
