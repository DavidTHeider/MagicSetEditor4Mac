// **Ein Handel von der Eröffnung bis zur Übergabe** — Werkstück T1 von Phase
// 5.6 (#1466),
// angeschlossen mit
// #1637.
//
// # Wo die Schritte stehen
//
// **Seit TRADE-02 in `Handelsschritte`, über allen Tafeln des Assistenten**
// (#1716)
// — und nur die, die der Vertrag nennt. Hier bleibt, was hin und her geht:
// die Positionen, eröffnen, speichern, vorschlagen und der Stand.
//
// # Was vom Stand vorn steht — UXR-15
//
// **Was ein Mensch liest** (#1923):
// Zustand, Art, die Vereinbarung und wer zugestimmt hat — „du" und das
// Gegenüber beim Namen. **Revision, Abdruck, die Parteien mit ihrer Kennung
// und „Stand ansehen" stehen in einem Kasten für Fachleute**: neben Lage
// und Spalten zeigte „Stand ansehen" nichts mehr, was nicht schon dastand,
// und die Parteien standen als rohe Peerkennung da. Er bleibt der Weg, einen
// Handel zu laden, der über seine Kennung geöffnet wurde und in der Liste
// noch nicht steht.
import { useId, useState } from 'react'

import { Handelszusammenstellung } from './Handelszusammenstellung'
import { Aktionsknopf } from './Aktionsknopf'
import { Aufklappkasten } from './Aufklappkasten'
import { Knopf } from './Knopf'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useLaden } from '../haken/useLaden'
import { Handelsstandanzeige } from './Handelsstandanzeige'

import type {
  Handelsposition,
  Handelsvorgang as Vorgangsstand,
  Handelsvorgangantwort,
} from '../apihandelsleitung'
import type { Handelszug } from '../haken/useHandelszug'
import type { Vorgang } from '../haken/useVorgang'

// dieTauschart ist die Art, mit der dieses Fenster eröffnet.
//
// **Eine und nicht vier**: Teilen und Deckleihe haben ihre eigenen Wege
// (`T3`, `G6`), und ein Markthandel ist eine Handlung bei einem fremden
// Anbieter. Was hier entsteht, ist der Tausch zwischen zwei Menschen.
const dieTauschart = 'tausch'

interface Handelsvorgangeigenschaften {
  kennung: string

  // name ist, wie das Gegenüber heißt — leer, wo keiner bekannt ist.
  name: string

  // vorgang ist der gewählte Handel — leer heißt: ein neuer (UX-08.AK-04).
  vorgang: string

  // stand ist, was der Handelsraum von ihm weiß — die Liste oder die Antwort
  // von vorhin. Ein eigener Ruf hier überschreibt ihn.
  stand: Vorgangsstand | undefined

  // geaendert meldet einen geschriebenen Stand nach oben: der Raum wählt ihn
  // und lädt die Liste nach.
  geaendert: (neu: Vorgangsstand) => void

  // gesehen meldet einen gelesenen Stand — der Raum trägt ihn ein, ohne die
  // Liste nachzuladen: gelesen ist nichts geändert.
  gesehen: (neu: Vorgangsstand) => void

  // zug ist der des Raums (P13) — „Anbieten" geht ihn nach dem Eröffnen.
  zug: Handelszug

  // lauf ist der Riegel des Raums, geteilt mit den Schritten (§7, Regel 4).
  lauf: Vorgang
}

export function Handelsvorgang({
  kennung,
  name,
  vorgang,
  stand: bekannt,
  geaendert,
  gesehen,
  zug,
  lauf,
}: Handelsvorgangeigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.handeln

  const ueberschrift = useId()

  // **Die Zeilen des Standes stehen da** — TRADE-02b
  // (#1951).
  // Bis dahin begann jeder gewählte Handel mit einem leeren Formular, und
  // was vereinbart war, stand nirgends. **Ein Gegenangebot beginnt beim
  // letzten Angebot**; wer schon geändert hat, behält seine Zeilen, bis sein
  // Schreiben glückt.
  const [positionen, setzePositionenRoh] = useState<Handelsposition[]>(
    bekannt?.positionen ?? [],
  )
  const [bearbeitet, setzeBearbeitet] = useState(false)
  const [uebernommen, setzeUebernommen] = useState(bekannt?.abdruck)

  if (!bearbeitet && bekannt !== undefined && bekannt.abdruck !== uebernommen) {
    setzeUebernommen(bekannt.abdruck)
    setzePositionenRoh(bekannt.positionen)
  }

  const setzePositionen = (neu: Handelsposition[]) => {
    setzeBearbeitet(true)
    setzePositionenRoh(neu)
  }

  // **Geräumt heißt: nach einem gescheiterten Ruf steht kein Stand da**,
  // auch nicht der bekannte — das wäre die Auskunft von vorhin. Einen
  // eigenen Stand führt dieses Bauteil nicht mehr (TRADE-02): was es liest
  // oder schreibt, geht an den Raum, und die Schritte daneben sehen es auch.
  const [geraeumt, setzeGeraeumt] = useState(false)
  const stand = geraeumt ? undefined : bekannt
  const [fehler, setzeFehler] = useState<string>()

  const kennzeichen = `handelsvorgang/${kennung}`

  // **Die eigene Kennung wird gelesen und nicht geraten** — sie ist die
  // Seite »von dir« einer Position, und der Vertrag prüft sie gegen die
  // Parteien.
  //
  // **Über [useLaden] und nicht über ein eigenes `useEffect`**: er führt den
  // Zähler, an dem eine späte Antwort erkennt, dass sie nicht mehr gilt (§7,
  // Regel 2) — ein Wächter daneben von Hand wäre die zweite Fassung
  // derselben Regel.
  const eigeneKennung = useLaden(async () => {
    const antwort = await api().Peeridentitaet()

    return auspacken(antwort, () => antwort.kennung)
  }, [])

  // **Ein Lesefehler ist hier kein Grund zu klagen**: dann steht die Auswahl
  // ohne die eigene Kennung da, und der nächste Ruf sagt es deutlicher — mit
  // dem Satz des Vertrags statt einem über den Schlüsselbund.
  const ich = eigeneKennung.daten ?? ''

  // nenne gibt einer Partei ihren Namen: „du", das Gegenüber, wie es heißt,
  // und wer sonst dasteht mit dem Anfang seiner Kennung — dieselbe Regel wie
  // in der Liste (`partnername`).
  const nenne = (partei: string) =>
    partei === ich
      ? eigene.vorgangDu
      : partei === kennung && name !== ''
        ? name
        : partei.slice(0, 12)

  // fahre führt einen Ruf und trägt seinen Stand ein.
  //
  // **Ein Vorgang je Datensatz** (§7, Regel 4) und **eine Antwort gehört zu
  // dem Stand, aus dem sie stammt** (Regel 2) — beides hält `useVorgang`.
  //
  // **Was schreibt, meldet sich nach oben** (`geaendert`): der Raum wählt den
  // Handel und lädt die Liste nach. Ein Blick auf den Stand schreibt nichts.
  const fahre = (
    ruf: () => Promise<Handelsvorgangantwort>,
    schreibt = true,
  ): Promise<void> =>
    lauf.starte(kennzeichen, async () => {
      setzeFehler(undefined)

      try {
        const antwort = await ruf()
        const neu = auspacken(antwort, () => antwort.vorgang)

        setzeGeraeumt(false)
        setzeBearbeitet(false)

        if (schreibt) {
          geaendert(neu)
        } else {
          gesehen(neu)
        }
      } catch (ursache) {
        setzeGeraeumt(true)
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  // anbieten eröffnet und geht den Zug des Anbietens (P13, Vermerk 0200):
  // anbieten, zustimmen, hinaus. **Scheitert schon das Eröffnen**, steht der
  // Fehler hier; was danach scheitert, sagt der Zug — und der Entwurf steht
  // (§7, Regel 3).
  const anbieten = (): Promise<void> =>
    zug.ziehen(
      {
        stand: async () => {
          setzeFehler(undefined)

          try {
            const antwort = await api().HandelEroeffnen({
              vorgang: '',
              art: dieTauschart,
              gegenueber: [kennung],
              positionen,
              bewertungGeteilt: false,
            })

            return auspacken(antwort, () => antwort.vorgang)
          } catch (ursache) {
            setzeFehler(meldungAus(ursache, texte))

            return undefined
          }
        },
        schritt: 'anbieten',
        bis: 'zustimmen',
      },
      (neu) => {
        geaendert(neu.stand)
      },
    )

  return (
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-2">
      <h3 id={ueberschrift} className="text-sm font-medium">
        {eigene.vorgangTitel}
      </h3>

      {/* **Kurz gefasst** (UX-15, QA01 #2382): seit *Handeln* einen Handel
          von selbst aufschlägt (P13), steht der Absatz im ersten Blick. */}
      <Erklaerungsabsatz text={eigene.vorgangErklaerung} />

      <Handelszusammenstellung
        ich={ich}
        kennung={kennung}
        name={name}
        positionen={positionen}
        setzePositionen={setzePositionen}
        bearbeitbar={
          vorgang === '' ||
          stand?.zustand === 'entwurf' ||
          stand?.moegliche.includes('gegenangebot') === true
        }
      />

      {/* **Ein neuer Handel wird eröffnet, ein gewählter fortgeführt** — die
          Kennung des neuen vergibt die Leitung (UX-08.AK-04). */}
      {vorgang === '' ? (
        <div className="flex flex-wrap gap-2">
          <Knopf
            gewicht="leise"
            angehalten={lauf.laeuft(kennzeichen)}
            aufKlick={() => {
              void fahre(() =>
                api().HandelEroeffnen({
                  vorgang: '',
                  art: dieTauschart,
                  gegenueber: [kennung],
                  positionen,
                  bewertungGeteilt: false,
                }),
              )
            }}
          >
            {eigene.vorgangEroeffnen}
          </Knopf>
          {/* **Eröffnen, Anbieten und Zustimmen in einem Zug** (P12, P13,
              Vermerk 0200): der Name sagt die Zustimmung — eine, die
              niemand sieht, wäre keine. Scheitert nur das Anbieten, steht
              der Entwurf trotzdem da (§7, Regel 3). */}
          <Knopf
            gewicht="haupt"
            angehalten={lauf.laeuft(kennzeichen)}
            aufKlick={() => {
              void anbieten()
            }}
          >
            {eigene.zug.anbietenZustimmen}
          </Knopf>
        </div>
      ) : (
        <>
          <div className="flex flex-wrap gap-2">
            {/* **Ein Entwurf wird gespeichert und nicht vorgeschlagen**
                (TRADE-01.AK-03, #1715) — er ist nie hinausgegangen und
                bleibt Revision 1. */}
            {stand?.zustand === 'entwurf' ? (
              <Aktionsknopf
                aktion="sichern"
                gewicht="leise"
                angehalten={lauf.laeuft(kennzeichen)}
                aufKlick={() => {
                  void fahre(() =>
                    api().HandelEntwurfAendern({
                      vorgang,
                      art: stand.art,
                      gegenueber: [kennung],
                      positionen,
                      bewertungGeteilt: stand.bewertungGeteilt,
                    }),
                  )
                }}
              >
                {eigene.liste.entwurfSpeichern}
              </Aktionsknopf>
            ) : (
              // **Vorschlagen nur, wo ein Gegenangebot geht** (TRADE-02,
              // #1716): nach der Einigung gibt es nichts mehr vorzuschlagen.
              stand?.moegliche.includes('gegenangebot') === true && (
                <Knopf
                  gewicht="haupt"
                  angehalten={lauf.laeuft(kennzeichen)}
                  aufKlick={() => {
                    void fahre(() =>
                      api().HandelVorschlagen(vorgang, positionen),
                    )
                  }}
                >
                  {eigene.vorgangVorschlagen}
                </Knopf>
              )
            )}
          </div>
        </>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {stand !== undefined && (
        <Handelsstandanzeige stand={stand} nenne={nenne} />
      )}

      {vorgang !== '' && (
        <Aufklappkasten titel={eigene.vorgangFachleute}>
          <div className="flex flex-col items-start gap-2 text-sm text-leise">
            {stand !== undefined && (
              <>
                <p>
                  {eigene.vorgangRevision} {stand.revision} ·{' '}
                  {eigene.vorgangAbdruck} {stand.abdruck.slice(0, 12)}
                </p>

                <p>
                  {eigene.vorgangParteien}: {stand.parteien.join(', ')}
                </p>
              </>
            )}

            <Knopf
              gewicht="leise"
              angehalten={lauf.laeuft(kennzeichen)}
              aufKlick={() => {
                void fahre(() => api().HandelStand(vorgang), false)
              }}
            >
              {eigene.vorgangAnsehen}
            </Knopf>
          </div>
        </Aufklappkasten>
      )}
    </section>
  )
}
