// **Ein Turnier in vier Schritten** — DOM-04.AK-01
// (#1699):
// *„Ein Assistent führt durch Format, Teilnehmer, Regeln und erste Runde;
// Pflichtangaben sind begründet und wiederverwendbare Vorlagen möglich."*
//
// Muster *Assistent* aus `docs/UX-ARCHITEKTUR.md`: eine Aufgabe, mehrere
// Schritte, ein Ergebnis. **Angelegt wird nach dem zweiten Schritt** und
// nicht am Ende: Teilnehmende und Runde sind Vorgänge am Turnier, die es
// schon gibt, und ein Assistent, der alles bis zum Schluss sammelte, hinterließe
// bei einem Fehler im vierten Schritt einen halben Stand.
//
// # In einer Schublade
//
// **Seit UXR-12** (#1920).
// Bis dahin stand der Assistent immer offen über der Liste — fünf Felder,
// auch über einem laufenden Turnier, dessen Werkbank er unter die Falz schob
// (`docs/UIREVIEW.md`). Jetzt öffnet ihn *Turnier anlegen …* im Kopf der
// Seite, wie *+ Aufnehmen …* das Formular der Sammlung (UXR-10).
//
// **Gezeichnet beim ersten Öffnen, danach nur verborgen** — wie die
// Schubladen der Sammlung: ein halb getippter Name übersteht das Schließen.
// **Ab dem dritten Schritt nicht**: dann gibt es das Turnier schon, und
// Schließen heißt *Später weitermachen*. Sonst bliebe die Werkbank für ein
// Turnier fort, dessen Assistent niemand mehr sieht.
import { Activity, useState, type ReactNode } from 'react'

import { api } from '../api'
import { auspacken } from '../umschlag'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Knopf } from './Knopf'
import { Ladefehler } from './Ladehinweise'
import { Schrittfolge } from './Schrittfolge'
import { Schublade } from './Schublade'
import { Turnieranlage } from './Turnieranlage'
import { Tooltip } from './Tooltip'
import { Turnierteilnehmer } from './Turnierteilnehmer'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import type { PaarungDTO, TurnierDTO } from '../apiturniere'
import type { Turnierstand } from '../haken/useTurniere'

interface AssistentEigenschaften {
  stand: Turnierstand

  // aufLauf sagt der Seite, für welches Turnier der Assistent gerade steht —
  // `null`, wenn für keines. **Solange er es tut, zeigt die Seite keine
  // Werkbank dafür**: zwei Stellen, an denen man dieselben Teilnehmenden
  // anmeldet, wären eine zu viel.
  aufLauf: (turnierID: string | null) => void

  // offen sagt, ob die Schublade steht; schliessen schließt sie.
  offen: boolean
  schliessen: () => void

  // meldung ist der Fehler der Seite. **Er steht in der Schublade**, solange
  // sie steht: darunter sähe ihn niemand (§6).
  meldung: ReactNode
}

export function Turnierassistent({
  stand,
  aufLauf,
  offen,
  schliessen,
  meldung,
}: AssistentEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere

  // **Beim Zeichnen gesetzt und nicht in einem Effekt**, wie im
  // `Aufklappkasten`: sonst stünde die Schublade erst einen Zug später da.
  const [gezeichnet, setzeGezeichnet] = useState(offen)
  if (offen && !gezeichnet) {
    setzeGezeichnet(true)
  }

  const [schritt, setzeSchritt] = useState(0)
  const [turnierID, setzeTurnierID] = useState<string | null>(null)

  // **Ein neuer Durchgang ist ein neues Formular** — der Zähler ist der
  // `key` der Anlage (§7, Regel 1).
  const [durchgang, setzeDurchgang] = useState(0)

  // **Der letzte Stand des eigenen Turniers bleibt stehen**, solange es neu
  // lädt. Nach jedem Anmelden liest die Seite nach, und `waehle` räumt dabei
  // erst (§7, Regel 1) — ohne diese Zeile baute sich der Schritt jedes Mal
  // ab und wieder auf, und das Feld für den nächsten Namen war kurz fort.
  // **Beim Zeichnen und nicht in einem Effekt**, wie in `useSymbolziel`.
  const [angelegt, setzeAngelegt] = useState<TurnierDTO | null>(null)
  const frisch =
    turnierID !== null && stand.offen?.id === turnierID ? stand.offen : null
  if (frisch !== null && frisch !== angelegt) {
    setzeAngelegt(frisch)
  }

  const beenden = () => {
    setzeSchritt(0)
    setzeTurnierID(null)
    setzeAngelegt(null)
    setzeDurchgang((bisher) => bisher + 1)
    aufLauf(null)
    schliessen()
  }

  if (!gezeichnet) {
    return null
  }

  return (
    <Activity mode={offen ? 'visible' : 'hidden'}>
      <Schublade
        titel={eigene.neu}
        schliessenBezeichnung={
          texte.allgemein.arbeitsplatz.schubladen.schliessen
        }
        schliessen={schritt >= 2 ? beenden : schliessen}
      >
        <div className="flex flex-col gap-3 px-4 pb-5 pt-3">
          {meldung}

          <Assistentenschritte
            stand={stand}
            schritt={schritt}
            setzeSchritt={setzeSchritt}
            durchgang={durchgang}
            angelegt={angelegt}
            anlegen={async (kennung) => {
              setzeTurnierID(kennung)
              aufLauf(kennung)
              setzeSchritt(2)
              await stand.waehle(kennung)
            }}
            beenden={beenden}
          />
        </div>
      </Schublade>
    </Activity>
  )
}

// Assistentenschritte zeigt die Schrittfolge und den Schritt darunter.
function Assistentenschritte({
  stand,
  schritt,
  setzeSchritt,
  durchgang,
  angelegt,
  anlegen,
  beenden,
}: {
  stand: Turnierstand
  schritt: number
  setzeSchritt: (schritt: number) => void
  durchgang: number
  angelegt: TurnierDTO | null
  anlegen: (kennung: string) => Promise<void>
  beenden: () => void
}) {
  const eigene = useTexte().seiten.turniere

  return (
    <>
      <Schrittfolge
        bezeichnung={eigene.assistent}
        schritte={[
          eigene.schrittTurnier,
          eigene.schrittRegeln,
          eigene.schrittTeilnehmende,
          eigene.schrittErsteRunde,
        ]}
        aktuell={schritt}
      />

      {schritt < 2 && (
        <Turnieranlage
          key={durchgang}
          stand={stand}
          schritt={schritt}
          setzeSchritt={setzeSchritt}
          aufAngelegt={anlegen}
        />
      )}

      {schritt === 2 && angelegt !== null && (
        <Teilnehmendenschritt
          stand={stand}
          turnier={angelegt}
          weiter={() => {
            setzeSchritt(3)
          }}
        />
      )}

      {schritt === 3 && angelegt !== null && (
        <Ersterundenschritt stand={stand} turnier={angelegt} fertig={beenden} />
      )}

      {schritt >= 2 && (
        <div>
          <Knopf gewicht="leise" aufKlick={beenden}>
            {eigene.verlassen}
          </Knopf>
        </div>
      )}
    </>
  )
}

// Teilnehmendenschritt meldet an — und geht erst weiter, wenn eine Runde
// möglich ist.
function Teilnehmendenschritt({
  stand,
  turnier,
  weiter,
}: {
  stand: Turnierstand
  turnier: TurnierDTO
  weiter: () => void
}) {
  const eigene = useTexte().seiten.turniere

  // **Zwei, die noch spielen** — wer ausgestiegen ist, wird nicht gepaart.
  const genug =
    turnier.teilnehmer.filter((teilnehmer) => !teilnehmer.ausgestiegen)
      .length >= 2

  return (
    <div className="flex flex-col gap-3">
      <Turnierteilnehmer stand={stand} turnier={turnier} />

      {!genug && <p className="text-sm text-leise">{eigene.zuWenige}</p>}

      <div>
        {/* **Der ruhende Knopf sagt, warum** — derselbe Satz wie darüber,
            als Beschreibung des Knopfs; ohne Grund braucht er keinen. */}
        {/* **Weiter führt** (R1) — das Anmelden darüber ist der Weg
            dorthin. */}
        {genug ? (
          <Knopf gewicht="haupt" aufKlick={weiter}>
            {eigene.weiter}
          </Knopf>
        ) : (
          <Tooltip text={eigene.zuWenige} breite="breit">
            {(griffe) => (
              <Knopf
                gewicht="haupt"
                angehalten
                beschreibung={griffe}
                aufKlick={weiter}
              >
                {eigene.weiter}
              </Knopf>
            )}
          </Tooltip>
        )}
      </div>
    </div>
  )
}

// Ersterundenschritt zeigt, wie die erste Runde ausfiele — und lost sie aus.
function Ersterundenschritt({
  stand,
  turnier,
  fertig,
}: {
  stand: Turnierstand
  turnier: TurnierDTO
  fertig: () => void
}) {
  const eigene = useTexte().seiten.turniere

  // **Ein Lesen und kein Schreiben** — der Vorschlag schafft keine Tatsachen.
  const vorschlag = useLaden<PaarungDTO[]>(async () => {
    const antwort = await api().TurnierPaarungsvorschlag(turnier.id)

    return auspacken(antwort, () => antwort.paarungen)
  }, [turnier.id])

  const namen = new Map(
    turnier.teilnehmer.map((teilnehmer) => [teilnehmer.id, teilnehmer.name]),
  )
  const wer = (kennung: string) => namen.get(kennung) ?? kennung

  const auslosen = async () => {
    // **Ein Kasten und keine Variable** — das Auslosen setzt ihn im Rückruf,
    // und erst danach steht fest, ob es geglückt ist.
    const ausgang = { gelegt: false }

    await stand.fuehre(`turnier:${turnier.id}`, async () => {
      auspacken(await api().TurnierRundeLegen(turnier.id), () => undefined)
      ausgang.gelegt = true
    })

    if (ausgang.gelegt) {
      fertig()
    }
  }

  return (
    <div className="flex flex-col gap-3">
      <Erklaerungsabsatz
        text={eigene.ersteRundeHinweis}
        klasse="text-sm text-leise"
      />

      {vorschlag.fehler !== undefined && (
        <Ladefehler
          fehler={vorschlag.fehler}
          wiederholen={vorschlag.nachladen}
        />
      )}

      <ol aria-label={eigene.vorschlagTitel} className="flex flex-col gap-1">
        {(vorschlag.daten ?? []).map((paarung) => (
          <li key={`${paarung.a}:${paarung.b}`}>
            {paarung.b === ''
              ? `${wer(paarung.a)} — ${eigene.bye}`
              : `${wer(paarung.a)} ${eigene.gegen} ${wer(paarung.b)}`}
          </li>
        ))}
      </ol>

      <div>
        <Knopf
          gewicht="haupt"
          angehalten={stand.vorgang.laeuft(`turnier:${turnier.id}`)}
          aufKlick={() => {
            void auslosen()
          }}
        >
          {eigene.ersteRundeAuslosen}
        </Knopf>
      </div>
    </div>
  )
}
