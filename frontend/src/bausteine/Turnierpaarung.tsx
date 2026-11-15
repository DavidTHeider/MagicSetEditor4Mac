// **Eine Paarung samt ihrem Ergebnis** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Als Spielstand und nicht als Sieger**: die Tiebreaker brauchen
// Spielpunkte (MTR Appendix C). Ein Bauteil, das nur „A gewinnt" anböte,
// machte zwei der vier Tiebreaker unberechenbar.
//
// # Was Werkstück U3 dazugelegt hat (#1414)
//
// **Die drei Kästchen sagen jetzt, was nicht stimmt.** Die Regeln stehen in
// `domain.Ergebnis.Pruefe` und werden von dort gelesen:
//
//	case e.SpieleA < 0 || e.SpieleB < 0 || e.Unentschieden < 0:
//	case e.SpieleA+e.SpieleB+e.Unentschieden == 0:
//	case hoechstens > 0 && e.SpieleA+e.SpieleB+e.Unentschieden > hoechstens:
//
// **Die Obergrenze gilt der Summe und nicht der einzelnen Zahl**, und wie hoch
// sie ist, sagt `service.hoechstensSpiele`: bei einem Einzelspielturnier eins,
// sonst null — und null heißt keine Grenze.
//
// **Der Text war bis heute schon aufgeräumt, bevor jemand ihn sah**: der
// `onChange` machte mit `Number.isFinite(gelesen) ? gelesen : 0` aus jedem
// Vertipper eine Null. Drei Nullen sind aber ein Stand, den Go abweist — der
// Fehler entstand also hier und wurde erst am anderen Ende sichtbar. Gehalten
// wird deshalb die **rohe** Zeichenkette (§3).
//
// # Eine Zeile, und das Eintragen dahinter
//
// **Seit UXR-12** (#1920),
// wie im Artboard *Turniere*: Namen, Stand und ein leiser Knopf *Ergebnis*.
// Vorher trug jede Paarung drei Kästchen, zwei Knöpfe und den Abschnitt
// *Hier gespielt* — bei sechzehn Teilnehmenden und vier Runden 96 Felder
// und 127 Bedienelemente in einer Lasche (R3 und R6 in `docs/UIREVIEW.md`).
// **Gezeichnet wird das Eintragen beim ersten Aufklappen** und danach nur
// verborgen, wie im `Aufklappkasten`: Zuklappen verändert keine Eingabe.
import { useId, useState } from 'react'

import { api } from '../api'
import { auspacken } from '../umschlag'
import { Hinweiskasten } from './Hinweiskasten'
import { Aktionsknopf } from './Aktionsknopf'
import { Knopf } from './Knopf'
import { Berichtigungsfolgen } from './Berichtigungsfolgen'
import { Turniermatch } from './Turniermatch'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import type {
  ErgebnisDTO,
  KorrekturfolgenDTO,
  PaarungDTO,
} from '../apiturniere'
import type { Turnierstand } from '../haken/useTurniere'

const zahlklassen =
  'w-16 rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-[invalid=true]:border-fehler'

// ganzAbNull trifft, was `Ergebnis.Pruefe` durchlässt: eine Ziffernfolge.
//
// **Leer zählt als null** — dasselbe, was Go aus einem fehlenden Wert macht,
// und wer eine Zahl ändert, löscht sie zuerst.
function ganzAbNull(text: string): boolean {
  const roh = text.trim()

  return roh === '' || /^\d+$/.test(roh)
}

// zahlAus liest eine der drei Zahlen — leer ist null.
function zahlAus(text: string): number {
  return text.trim() === '' ? 0 : Number(text)
}

export function Turnierpaarung({
  stand,
  turnierID,
  fassung,
  paarung,
  namen,
  hoechstens,
  spaeterGelegt,
  nurLesen = false,
}: {
  stand: Turnierstand
  turnierID: string

  // fassung ist der Stand, aus dem diese Zeile gelesen wurde.
  fassung: string
  paarung: PaarungDTO
  namen: Map<string, string>

  // hoechstens ist, über wie viele Spiele ein Match hier geht — **0 heißt
  // keine Grenze**, wie in `service.hoechstensSpiele`.
  hoechstens: number

  // spaeterGelegt sagt, dass nach dieser Runde schon eine gepaart wurde —
  // dann ist jedes Ergebnis hier eine Berichtigung (DOM-04.AK-02, #1699).
  spaeterGelegt: boolean

  // nurLesen heißt: das Turnier ist abgeschlossen — der Spielstand steht da,
  // eingetragen wird nichts (DOM-04.AK-04, #1699).
  nurLesen?: boolean
}) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere

  const befundID = useId()
  const eintragenID = useId()

  const [offen, setzeOffen] = useState(false)

  // **Einmal offen, bleibt gezeichnet** — beim Zeichnen gesetzt und nicht in
  // einem Effekt, wie im `Aufklappkasten`.
  const [gezeichnet, setzeGezeichnet] = useState(false)
  if (offen && !gezeichnet) {
    setzeGezeichnet(true)
  }

  const [spieleA, setzeSpieleA] = useState(
    String(paarung.ergebnis?.spieleA ?? 0),
  )
  const [spieleB, setzeSpieleB] = useState(
    String(paarung.ergebnis?.spieleB ?? 0),
  )
  const [remis, setzeRemis] = useState(
    String(paarung.ergebnis?.unentschieden ?? 0),
  )

  const wer = (kennung: string) => namen.get(kennung) ?? kennung

  // frage ist eine Berichtigung, die auf ihr Ja wartet — samt dem Stand,
  // der dann eingetragen wird.
  const [frage, setzeFrage] = useState<{
    stand: ErgebnisDTO | null
    folgen: KorrekturfolgenDTO
  }>()

  const schreiben = async (stand2: ErgebnisDTO | null) => {
    await stand.fuehre(`turnier:${turnierID}`, async () => {
      auspacken(
        await api().TurnierErgebnisSetzen(
          turnierID,
          fassung,
          paarung.id,
          stand2,
        ),
        () => undefined,
      )

      // **Eingetragen, und das Eintragen klappt zu** — sonst stünden nach
      // einer Runde acht offene Kästen da. Das Nachladen baut die Werkbank
      // meistens ohnehin neu auf, aber nicht immer (`turnierpaarung.test`).
      // Ein Nein lässt es offen: dort steht, was zu tun ist.
      setzeOffen(false)
    })
  }

  // setzen trägt ein — **eine Berichtigung sagt vorher, was sie verschiebt**
  // (DOM-04.AK-02). Berichtigung heißt: hier steht schon ein Ergebnis, oder
  // nach dieser Runde wurde schon gepaart. Verschiebt sie nichts, geht sie
  // ohne Frage hinein; das erste Ergebnis der laufenden Runde fragt nie.
  const setzen = async (stand2: ErgebnisDTO | null) => {
    if (paarung.ergebnis === null && !spaeterGelegt) {
      await schreiben(stand2)

      return
    }

    // **Ein Kasten und keine Variable** — gesetzt wird er im Rückruf.
    const antwort: { folgen: KorrekturfolgenDTO | null } = { folgen: null }

    await stand.lies(`turnier:${turnierID}`, async () => {
      const gelesen = await api().TurnierKorrekturfolgen(
        turnierID,
        paarung.id,
        stand2,
      )

      antwort.folgen = auspacken(gelesen, () => gelesen.folgen)
    })

    const { folgen } = antwort

    if (folgen === null) {
      return
    }

    if (folgen.verschoben.length === 0 && folgen.spaetereRunden.length === 0) {
      await schreiben(stand2)

      return
    }

    setzeFrage({ stand: stand2, folgen })
  }

  const laeuft = stand.vorgang.laeuft(`turnier:${turnierID}`)

  const rohe = [spieleA, spieleB, remis]
  const alleZahlen = rohe.every(ganzAbNull)
  const summe = rohe.reduce((bisher, text) => bisher + zahlAus(text), 0)

  // **Der erste Satz, der zutrifft** — in der Reihenfolge, in der auch
  // `Ergebnis.Pruefe` prüft.
  const befund = !alleZahlen
    ? eigene.ergebnisZahlen
    : summe === 0
      ? eigene.ergebnisOhneSpiel
      : hoechstens > 0 && summe > hoechstens
        ? einsetzen(eigene.ergebnisZuViele, {
            hoechstens: String(hoechstens),
          })
        : undefined

  // **Ein Befund über die Summe hängt an allen drei Kästchen** — welches zu
  // viel hat, sagt die Regel nicht; ein Befund an einem allein zeigte auf ein
  // Feld, das für sich genommen in Ordnung ist.
  const ungueltig = (text: string) =>
    alleZahlen ? befund !== undefined : !ganzAbNull(text)

  // **Ein Bye trägt kein Ergebnis** — er *ist* eines (2–0, Appendix C).
  if (paarung.b === '') {
    return (
      <span>
        {wer(paarung.a)} — {eigene.bye}
      </span>
    )
  }

  const paarungstext = `${wer(paarung.a)} ${eigene.gegen} ${wer(paarung.b)}`

  // **Der Stand an der Zeile** — wie im Artboard: was gespielt ist, steht
  // als Spielstand da, was nicht, sagt es.
  const spielstand =
    paarung.ergebnis === null ? (
      <span className="text-sm text-leise">{eigene.offen}</span>
    ) : (
      <span className="tabular-nums">
        {`${String(paarung.ergebnis.spieleA)}:${String(paarung.ergebnis.spieleB)}:${String(paarung.ergebnis.unentschieden)}`}
      </span>
    )

  // **Abgeschlossen: nur der Stand** — Namen und Spielstand, kein Feld.
  if (nurLesen) {
    return (
      <span className="flex flex-wrap gap-3">
        <span className="min-w-52">{paarungstext}</span>
        {spielstand}
      </span>
    )
  }

  return (
    <div className="flex flex-col gap-2">
      {/* **Eine Mindestbreite, damit die Zeilen fluchten.** Ohne sie
          verschiebt jeder Name den Stand dahinter, und eine Runde mit acht
          Tischen liest sich als Treppe statt als Liste — angestrichen in der
          Sichtprüfung zu #1294. `min-w` und nicht `w`: ein langer Name wird
          nicht abgeschnitten, er schiebt dann eben. */}
      <div className="flex flex-wrap items-center gap-3">
        <span className="min-w-52">{paarungstext}</span>

        {spielstand}

        <Knopf
          gewicht="leise"
          bezeichnung={`${eigene.ergebnis}: ${paarungstext}`}
          ausgeklappt={offen}
          steuert={eintragenID}
          aufKlick={() => {
            setzeOffen(!offen)
          }}
        >
          {eigene.ergebnis}
        </Knopf>
      </div>

      {gezeichnet && (
        <div
          id={eintragenID}
          hidden={!offen}
          className="flex flex-wrap items-center gap-2 border-l-2 border-kante pl-3"
        >
          {/* **Die Aufschriften stehen im `aria-label` und nicht daneben.**
              Sichtbar wäre jede von ihnen die dritte Wiederholung derselben
              Paarung in einer Zeile — darüber steht sie schon, und drei
              Kästchen in der Form `2 : 0 : 0` liest jeder, der je eine
              Deckliste gesehen hat. Angesagt werden sie trotzdem einzeln,
              und **mit** der Paarung: »Spiele für den Ersten« stünde in
              einer Runde sonst achtmal gleich da. */}
          <span className="text-sm">{eigene.spiele}</span>

          {(
            [
              [eigene.spieleA, spieleA, setzeSpieleA],
              [eigene.spieleB, spieleB, setzeSpieleB],
              [eigene.unentschiedenSpiele, remis, setzeRemis],
            ] as const
          ).map(([aufschrift, wert, setze], stelle) => (
            <span key={aufschrift} className="flex items-center gap-2">
              {stelle > 0 && <span aria-hidden="true">:</span>}

              <input
                value={wert}
                inputMode="numeric"
                aria-label={`${aufschrift} — ${paarungstext}`}
                aria-invalid={ungueltig(wert)}
                aria-describedby={befund === undefined ? undefined : befundID}
                onChange={(ereignis) => {
                  setze(ereignis.target.value)
                }}
                className={zahlklassen}
              />
            </span>
          ))}

          <Knopf
            bezeichnung={`${eigene.eintragen}: ${paarungstext}`}
            angehalten={laeuft || befund !== undefined}
            aufKlick={() => {
              void setzen({
                spieleA: zahlAus(spieleA),
                spieleB: zahlAus(spieleB),
                unentschieden: zahlAus(remis),
              })
            }}
          >
            {eigene.eintragen}
          </Knopf>

          {paarung.ergebnis !== null && (
            // **Zurücknehmen hängt nicht am Befund.** Es schickt `null` und
            // liest die drei Kästchen gar nicht; eine Sperre wäre hier eine
            // ohne Grund.
            <Aktionsknopf
              aktion="zuruecknehmen"
              bezeichnung={`${eigene.zuruecknehmen}: ${paarungstext}`}
              angehalten={laeuft}
              aufKlick={() => {
                void setzen(null)
              }}
            >
              {eigene.zuruecknehmen}
            </Aktionsknopf>
          )}

          {befund !== undefined && (
            <div className="basis-full">
              <Hinweiskasten art="fehler" kennung={befundID} meldung={befund} />
            </div>
          )}

          {frage !== undefined && (
            <div className="basis-full">
              <Berichtigungsfolgen
                folgen={frage.folgen}
                laeuft={laeuft}
                aufEintragen={() => {
                  setzeFrage(undefined)
                  void schreiben(frage.stand)
                }}
                aufVerwerfen={() => {
                  setzeFrage(undefined)
                }}
              />
            </div>
          )}

          {/* **Der zweite Weg zu einem Ergebnis** (G7, #1464) — neben dem
              Eintragen von Hand und nicht an seiner Stelle. Der `key` trägt
              die Paarungskennung: wer die Runde neu legt, bekommt ein leeres
              Bauteil und nicht die halb ausgefüllten Felder der vorigen
              Begegnung (§7, Regel 1). **Seit UXR-12 hinter *Ergebnis***:
              gespielt wird hier an einem Tisch von acht, und sein Abschnitt
              mit zwei Absätzen stand an jedem. */}
          <div className="basis-full">
            <Turniermatch
              key={paarung.id}
              turnierID={turnierID}
              paarungID={paarung.id}
              teilnehmer={[paarung.a, paarung.b]}
            />
          </div>
        </div>
      )}
    </div>
  )
}
