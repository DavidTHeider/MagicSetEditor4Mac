// **Einen Handel verbuchen** — Werkstück T6 von Phase 5.6
// (#1471).
//
// # Warum die Stimme des Gegenübers ein eigener Knopf bleibt
//
// **Weil ein Knopf, der alles täte, die zweite Unterschrift erfinden
// müsste** — und genau das verbietet dieses Werkstück. Verbuchen bewegt den
// eigenen Bestand, Abgleichen meldet der Gegenseite, dass hier fertig ist,
// die eigene Stimme bestätigt den Abschluss: das sind drei Schritte **dieser**
// Seite, und seit P13 geht sie ein Klick auf „Verbuchen"
// (#2381,
// Vermerk 0200) — jeder mit seiner eigenen Freigabe und seinem eigenen
// Fehlerweg (`handelszug.ts`). „Gegenseite bestätigt" bleibt allein.
//
// # Welche Knöpfe stehen — UXR-15
//
// **Was der Stand des Handels nennt** (#1923).
// Bis dahin standen alle drei in jeder Phase und meldeten vor der Übergabe
// einen Fehler; und wo der nächste Schritt einer von ihnen war, führte
// keiner (R1 in `docs/UIREVIEW.md`). Jetzt gilt:
//
//   - **Vor der Phase „Verbuchen" steht ein Satz** und kein Knopf — keiner
//     hätte etwas getan.
//   - **„Verbuchen" steht, wo `moegliche` es nennt** — also bis gebucht ist.
//     Bis zur Hüllenprobe `nat04handel` (#1672) stand es die ganze Phase
//     hindurch, denn ein zweiter Druck war der einzige Weg, den Beleg wieder
//     zu sehen. **Seither liest der Kasten ihn** (`Handelsbeleg`): beim
//     Aufschlagen, nach jedem eigenen Schritt und wenn ohne dieses Fenster
//     etwas ankam — Bert sah sonst nie, dass Anna gegengezeichnet hatte.
//   - **„Abgleich melden" steht, wo `moegliche` es nennt** — dieselbe Regel
//     wie bei den Schritten, aus `domain` und nicht nachgerechnet. Seit P13
//     ist es der Ausweg, wenn ein Zug nach dem Verbuchen hängen blieb.
//   - **„Abschluss bestätigen", solange der offene Beschluss das Bestätigen
//     ist und die eigene Stimme fehlt** (#2026) — er führt, denn `naechster`
//     nennt ihn.
//   - **„Gegenseite bestätigt", solange der offene Beschluss das Bestätigen
//     ist und die Stimme des Gegenübers fehlt** — `beschluss` und
//     `zustimmungen` kommen aus `domain`. `moegliche` gälte hier nicht: es
//     sagt, was **ich** gehen darf, und diese Stimme ist die des Gegenübers.
//   - **Der nächste Schritt führt** (`naechster`), wie bei den Schritten.
import { useId } from 'react'
import { einsetzen } from '../sprache/einsetzen'

import { Knopf } from './Knopf'
import { api } from '../fassade'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Gegenbelegzeile } from './Gegenbelegzeile'
import { auspacken } from '../umschlag'
import { useTexte } from '../sprache/kontext'
import { standsatz } from './handelsstaende'
import { Ladefehler } from './Ladehinweise'
import { useLaden } from '../haken/useLaden'
import type { Vorgang } from '../haken/useVorgang'
import {
  derAbschlussschritt,
  derVerbuchungsschritt,
} from '../apihandelsverbuchung'

import type { Handelsvorgang as Vorgangsstand } from '../apihandelsleitung'
import type { Zugauftrag } from '../handelszug'
import type { Handelszug } from '../haken/useHandelszug'

// derAbgleichsschritt ist der Schritt, der der Gegenseite meldet, dass hier
// fertig ist — er braucht keine Freigabe und steht deshalb nicht neben den
// zwei Schritten in `apihandelsverbuchung.ts`.
const derAbgleichsschritt = 'abgleichen'

// dieAbschlussphase ist die Phase, in der verbucht wird.
const dieAbschlussphase = 'verbuchen'

interface Handelsbuchungseigenschaften {
  kennung: string

  // name ist, wie das Gegenüber heißt — leer, wo keiner bekannt ist.
  name: string

  // lauf ist der Riegel des Raums — **derselbe wie für Schritte, Vorschläge
  // und Aktionen** (§7, Regel 4; Befund N3 des Nachlaufs über Phase 6.8).
  // Mit einem eigenen konnten ein Schritt und dieser Kasten zugleich
  // schreiben.
  lauf: Vorgang

  // vorgang ist der gewählte Handel — leer heißt: noch keiner (UX-08.AK-04).
  vorgang: string

  // stand ist, was der Raum von diesem Handel weiß — er sagt, welche Knöpfe
  // stehen (UXR-15, #1923).
  stand: Vorgangsstand | undefined

  // gebucht meldet einen geglückten Schritt: der Raum lädt den Handel nach,
  // damit Lage und Schritte darüber den neuen Stand zeigen.
  gebucht: () => void

  // zug ist der des Raums (P13) — derselbe wie an den Schritten.
  zug: Handelszug

  // nachgeladen zählt, was den Beleg bewegt haben kann — eigene Schritte und
  // was ohne dieses Fenster ankam. Der Kasten liest ihn dann neu.
  nachgeladen: number
}

export function Handelsbuchung({
  kennung,
  name,
  vorgang,
  lauf,
  stand: vertrag,
  gebucht,
  zug,
  nachgeladen,
}: Handelsbuchungseigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.handeln
  const ueberschrift = useId()
  const genannt = name === '' ? eigene.assistent.deinGegenueber : name

  // **Der Beleg wird gelesen und nicht aus dem letzten Klick genommen**
  // (Hüllenprobe `nat04handel`, #1672) — und nur in der Phase, in der es
  // einen geben kann. **Behalten**, weil sich nur der Stand bewegt: der
  // `key` am Kasten räumt beim Wechsel des Handels (§7, Regel 1).
  const abschluss = vertrag?.phase === dieAbschlussphase
  const beleg = useLaden(
    async () => {
      if (!abschluss) {
        return undefined
      }

      const antwort = await api().Handelsbeleg(vorgang)

      return auspacken(antwort, () => antwort)
    },
    [vorgang, abschluss, nachgeladen],
    true,
  )
  const stand = abschluss ? beleg.daten : undefined

  const kennzeichen = `handelsvorgang/${kennung}`

  // gewichtFuer gibt dem Schritt, den der Vertrag als nächsten nennt, das
  // Hauptgewicht.
  const naechster = vertrag?.naechster
  const gewichtFuer = (schritt: string) =>
    naechster === schritt ? 'haupt' : 'leise'

  // ziehe geht einen Zug ab diesem Schritt (Vermerk 0200).
  //
  // **Der Klick ist die Zustimmung** (T10,
  // #1475)
  // — zu jedem Schritt, den sein Knopf nennt; der Zug holt die Freigabe
  // unmittelbar vor ihm, gebunden an den Stand von dann. **Der Nachweis
  // entsteht am Fenster und nirgends sonst**: `Ausstellen` weist jeden
  // anderen Modus ab.
  //
  // **Nachladen ist Bestenfalls-Arbeit** (§7, Regel 3): der Raum lädt den
  // Handel nach, das Journal im eigenen Kasten liest nach — ein Lesefehler
  // dort nimmt die Buchung nicht zurück. Ging gar nichts, wird nicht
  // nachgeladen.
  //
  // **Jeder Knopf hier geht bis zur eigenen Abschlussstimme** — und nicht
  // weiter; die des Gegenübers ist der Knopf „Gegenseite bestätigt" für sich.
  // **Ein veralteter Stand wird neu gelesen** (Befund 1 des Zweitblicks über
  // Phase 8): der Mensch soll sehen, worüber er entscheidet.
  const ziehe = (
    stand: Vorgangsstand,
    auftrag: Omit<Zugauftrag, 'gegenueber' | 'stand' | 'bis'>,
  ) => {
    void zug.ziehen({ ...auftrag, bis: derAbschlussschritt, stand }, (neu) => {
      const teil = neu.gescheitert?.teil

      if (neu.gegangen.length > 0 || teil === 'lesen' || teil === 'veraltet') {
        gebucht()
      }
    })
  }

  return (
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-2">
      <h3 id={ueberschrift} className="text-sm font-medium">
        {eigene.buchungTitel}
      </h3>

      <Erklaerungsabsatz
        text={eigene.buchungErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {vertrag?.phase !== dieAbschlussphase ? (
        <p className="text-sm text-leise">{eigene.buchungNochNicht}</p>
      ) : (
        <>
          <div className="flex flex-wrap gap-2">
            {vertrag.moegliche.includes(derVerbuchungsschritt) && (
              <Knopf
                gewicht={gewichtFuer(derVerbuchungsschritt)}
                angehalten={lauf.laeuft(kennzeichen)}
                aufKlick={() => {
                  ziehe(vertrag, { schritt: derVerbuchungsschritt })
                }}
              >
                {eigene.buchungVerbuchen}
              </Knopf>
            )}

            {vertrag.moegliche.includes(derAbgleichsschritt) && (
              <Knopf
                gewicht={gewichtFuer(derAbgleichsschritt)}
                angehalten={lauf.laeuft(kennzeichen)}
                aufKlick={() => {
                  ziehe(vertrag, { schritt: derAbgleichsschritt })
                }}
              >
                {eigene.buchungAbgleichen}
              </Knopf>
            )}

            {/* **Die eigene Stimme** (#2026) — der Abschluss braucht die
              jeder Partei (`Alle` in der Matrix), und wessen sie ist, sagt
              der Dienst (`ich`). Bis dahin nannte die Lage sie als nächsten
              Schritt, und das Fenster bot nur die des Gegenübers an. */}
            {vertrag.beschluss === derAbschlussschritt &&
              vertrag.ich !== '' &&
              !vertrag.zustimmungen.includes(vertrag.ich) && (
                <Knopf
                  gewicht={gewichtFuer(derAbschlussschritt)}
                  angehalten={lauf.laeuft(kennzeichen)}
                  aufKlick={() => {
                    ziehe(vertrag, { schritt: derAbschlussschritt })
                  }}
                >
                  {eigene.buchungEigeneStimme}
                </Knopf>
              )}

            {/* **Die Stimme des Gegenübers** — wer sie abgibt, ist der
              Kontakt, den man gerade ansieht. Sie führt nie: der nächste
              Schritt, den `naechster` nennt, ist die **eigene** Stimme. */}
            {vertrag.beschluss === derAbschlussschritt &&
              !vertrag.zustimmungen.includes(kennung) && (
                <Knopf
                  gewicht="leise"
                  angehalten={lauf.laeuft(kennzeichen)}
                  aufKlick={() => {
                    ziehe(vertrag, {
                      schritt: derAbschlussschritt,
                      wer: kennung,
                    })
                  }}
                >
                  {eigene.buchungBestaetigen}
                </Knopf>
              )}
          </div>

          {/* **Was „Verbuchen" noch tut**, steht dabei (Vermerk 0200) — und
            dass die Stimme des Gegenübers ein eigener Schritt bleibt. */}
          {vertrag.moegliche.includes(derVerbuchungsschritt) && (
            <p className="max-w-prose text-xs text-leise">
              {einsetzen(eigene.zug.buchungHinweis, { name: genannt })}
            </p>
          )}
        </>
      )}

      {abschluss && beleg.fehler !== undefined && (
        <Ladefehler fehler={beleg.fehler} wiederholen={beleg.nachladen} />
      )}

      {/* **Der Stand wird angesagt** (§6): er ändert sich auf einen Klick
          hin oder wenn drüben etwas geschah, und wer die Seite nicht sieht,
          erführe es sonst nicht. */}
      {stand !== undefined && (
        <p role="status" className="text-sm text-leise">
          {eigene.buchungZustand}: {standsatz(eigene, stand.zustand)}
          {stand.hatBeleg && (
            <>
              {' · '}
              {eigene.buchungBeleg}
              {' · '}
              {stand.beleg.beidseitig
                ? eigene.buchungBeidseitig
                : eigene.buchungEinseitig}
            </>
          )}
        </p>
      )}

      {/* **Woran das liegt, steht dabei** —
          #1650:
          wer gegengezeichnet hat, oder warum nicht (#2500). */}
      {stand !== undefined && stand.hatBeleg && (
        <Gegenbelegzeile beleg={stand.beleg} kennung={kennung} name={name} />
      )}
    </section>
  )
}
