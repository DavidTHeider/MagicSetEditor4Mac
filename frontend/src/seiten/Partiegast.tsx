// **An einem fremden Tisch mitspielen** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Zwei Teile an zwei Orten
//
// **Der Beitritt steht in einer Schublade, der Platz auf der Seite** — UXR-13
// (#1921).
// Bis dahin stand beides als ein Abschnitt unter dem Anfang und der
// Partienliste: die Wahl des Wirts offen neben jedem Anfang, und wer saß,
// fand den fremden Tisch unter allem anderen. Jetzt öffnet *Bei jemandem
// mitspielen …* im Kopf der Seite die Schublade (`Gastbeitritt`), und wer
// sitzt, hat den Tisch oben auf der Seite (`Gastplatz`). **Den Lauf hält die
// Seite** (`usePartiegast`), damit beide dieselbe Wahrheit zeigen.
//
// # Warum diese Seite den Tisch nicht selbst zeichnet
//
// **Weil sie ihn nicht kennt.** Was hier ankommt, ist die **Sicht**, die der
// Wirt für diesen Sitz gerechnet hat — fertig projiziert. Gezeichnet wird sie
// von demselben Bauteil wie ein eigener Tisch (`Partietisch`), und genau das
// ist der Punkt: es gibt **eine** Darstellung und nicht zwei, die
// auseinanderlaufen können.
//
// # Und warum hier keine Kennung getippt wird
//
// **Weil beide ableitbar sind** (UX-08.AK-04, #1680): das Gegenüber steht in
// den Kontakten, und die Partie nennt der Wirt mit seinem ersten Wort
// (`service.Partieerfragen`); ein Tippfehler in der Partie hieße, dass der
// Tisch nie kommt — ohne ein Wort darüber. **Das Feld für die Partie
// bleibt**, unter „Die Partie selbst nennen": ein Wirt vor 0.6.0 begrüßt
// nicht.
//
// # Und warum der Abriss dasteht, bevor er eintritt
//
// **Weil er die Eigenschaft ist, die am meisten überrascht.** Reißt die
// Verbindung ab, verschwindet der Tisch — auch der letzte Stand. §6 verlangt
// die Erklärung dort, wo es geschieht, und nicht hinterher: beim Beitritt und
// am Platz.
import { useId, useState } from 'react'

import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Auswahlfeld } from '../bausteine/Auswahlfeld'
import {
  Beschriftetesfeld,
  eingabeklasse,
} from '../bausteine/Beschriftetesfeld'
import { mitbringAus } from '../spiel/mitbringen'
import { useMitbringwahl } from '../haken/useMitbringwahl'
import { Knopf } from '../bausteine/Knopf'
import { Partietisch } from './Partietisch'
import { Tischgespraech } from '../bausteine/Tischgespraech'
import { usePartienebenbei } from '../haken/usePartienebenbei'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Kontaktwahl } from '../bausteine/Kontaktwahl'

import type { PartieAntwort } from '../apipartie'
import type { Gastlauf } from '../haken/usePartiegast'
import type { Partielauf } from '../haken/usePartie'

// dasRechtAmTisch ist der Zweck, ohne den kein Beitritt gewählt wird — der,
// den `service.Peersitzungen.Oeffne` vor dem Wählen prüft.
const dasRechtAmTisch = 'partie'

// sitztAmFremdenTisch sagt, ob dieser Rechner gerade an einem fremden Tisch
// sitzt — **nach dem letzten Stand des Beitritts**, nicht nach dem Bild: ein
// Bild kann ankommen, bevor der Stand gefragt ist.
export function sitztAmFremdenTisch(lauf: Gastlauf): boolean {
  return lauf.stand?.laeuft === true
}

// Gastbeitritt ist die Schublade: bei wem, und los.
//
// **Der Beitritt führt** (R1): *Mitspielen* ist der Hauptknopf der Schublade.
// Bis UXR-13 stand er leise, gleichrangig mit „Nachsehen" und den Knöpfen
// der Leihe darunter.
export function Gastbeitritt({
  lauf,
  kontaktstand = 0,
}: {
  lauf: Gastlauf
  kontaktstand?: number
}) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const [gegenueber, setzeGegenueber] = useState('')
  const [partie, setzePartie] = useState('')

  // **Womit man sich setzt** — P9c (#2377). Vorgewählt, bis jemand wählt.
  const wahl = useMitbringwahl()
  const [mitbringen, setzeMitbringen] = useState<string>()
  const mitWert = mitbringen ?? wahl.vorgabe ?? ''
  const mitFeld = useId()

  return (
    <div className="flex flex-col gap-3">
      <Erklaerungsabsatz
        text={eigene.gastErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {/* **Gewählt und nicht getippt** (UX-08.AK-04, #1680): das Gegenüber
          aus den Kontakten, die Partie nennt der Wirt. **Für einen Wirt vor
          0.6.0** bleibt das Feld für die Partie erreichbar — mit dem Satz,
          wann es gebraucht wird. */}
      <Kontaktwahl
        titel={eigene.gastWahl}
        leer={eigene.gastWahlLeer}
        recht={dasRechtAmTisch}
        gewaehlt={gegenueber}
        waehlen={setzeGegenueber}
        stand={kontaktstand}
      />

      <div className="flex flex-col gap-1">
        <label htmlFor={mitFeld} className="text-sm font-medium">
          {eigene.gastMitbringen}
        </label>
        <Auswahlfeld
          id={mitFeld}
          bezeichnung={eigene.gastMitbringen}
          eintraege={wahl.eintraege}
          wert={mitWert}
          aendern={setzeMitbringen}
          klasse={eingabeklasse}
        />
      </div>

      <Aufklappkasten titel={eigene.gastFachleute}>
        <div className="flex flex-col gap-2">
          <Erklaerungsabsatz
            text={eigene.gastFachleuteErklaerung}
            klasse="max-w-prose text-sm text-leise"
          />

          <Beschriftetesfeld
            bezeichnung={eigene.gastPartie}
            wert={partie}
            setzeWert={setzePartie}
          />
        </div>
      </Aufklappkasten>

      <div className="flex flex-wrap items-center gap-3">
        <Knopf
          gewicht="haupt"
          angehalten={lauf.laeuft || gegenueber === ''}
          aufKlick={() => {
            lauf.beitreten(gegenueber, partie, mitbringAus(mitWert))
          }}
        >
          {eigene.gastBeitreten}
        </Knopf>

        <Knopf gewicht="leise" angehalten={lauf.laeuft} aufKlick={lauf.ansehen}>
          {eigene.gastNachsehen}
        </Knopf>
      </div>

      {lauf.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {lauf.fehler}
        </p>
      )}

      <p className="max-w-prose text-xs text-leise">{eigene.gastAbriss}</p>
    </div>
  )
}

// Gastplatz ist der fremde Tisch auf der Seite — **nur, wenn es einen gibt**:
// wer sitzt, wem ein Bild kam oder wessen Leitung abriss.
//
// **Oben auf der Seite**, über dem eigenen Anfang: wer an einem fremden Tisch
// sitzt, spielt gerade dort. Bis UXR-13 stand er unter der Partienliste, mit
// einem Commanderdeck mehrere Bildschirmhöhen tief.
export function Gastplatz({ lauf }: { lauf: Gastlauf }) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen
  const ueberschrift = useId()

  const sitzt = sitztAmFremdenTisch(lauf)

  // **Ein Abriss ist kein Aufstehen.** Ohne diesen Satz verschwände der Platz
  // still, und niemand wüsste, ob er hinausgeworfen wurde oder ob die Leitung
  // wegbrach (§6).
  //
  // **Und eine Absage ist kein Abriss** (P9c, #2377): der Wirt hat gesagt,
  // warum er niemanden hinsetzt — das steht dann statt des Satzes da.
  const absage = sitzt ? '' : (lauf.stand?.absage ?? '')
  const abgerissen = !sitzt && lauf.stand?.pausiert === true && absage === ''

  // **Der eigene Sitz kommt aus dem Tischbild** (`sicht.fuer`) und nicht aus
  // dem Beitritt: der Gast schreibt beim Beitreten keinen Sitz, den vergibt
  // der Wirt. Solange noch kein Bild da ist, steht hier nichts — und es steht
  // nur im eigenen Fenster, denn was hinausgeht, setzt der Wirt ohnehin neu.
  const gespraech = usePartienebenbei(
    lauf.tisch?.partie.fuer ?? '',
    'nebenbei-gast',
    lauf.tisch?.partie.id ?? lauf.stand?.partie ?? '',
  )

  if (!sitzt && !abgerissen && absage === '' && lauf.tisch === undefined) {
    return null
  }

  return (
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-3">
      <h2 id={ueberschrift} className="text-lg font-semibold">
        {eigene.gastPlatz}
      </h2>

      {sitzt && (
        <p className="text-sm">
          {eigene.gastSitzt}{' '}
          <span className="font-mono text-xs">{lauf.stand?.gegenueber}</span>
        </p>
      )}

      {abgerissen && (
        <p role="status" className="max-w-prose text-sm text-leise">
          {eigene.gastPausiert}
        </p>
      )}

      {absage !== '' && (
        <div role="status" className="flex max-w-prose flex-col gap-1 text-sm">
          <p>{eigene.gastAbgesagt}</p>
          <p className="text-leise">{absage}</p>
        </div>
      )}

      <div className="flex flex-wrap items-center gap-3">
        {sitzt && (
          <Knopf
            gewicht="leise"
            angehalten={lauf.laeuft}
            aufKlick={lauf.aufstehen}
          >
            {eigene.gastAufstehen}
          </Knopf>
        )}

        <Knopf gewicht="leise" angehalten={lauf.laeuft} aufKlick={lauf.ansehen}>
          {eigene.gastNachsehen}
        </Knopf>
      </div>

      {lauf.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {lauf.fehler}
        </p>
      )}

      {(sitzt || lauf.tisch !== undefined) && <Gasttisch lauf={lauf} />}

      {/* **Erst am Tisch.** Wer noch nirgends sitzt, hat niemanden, dem er
          etwas sagen könnte — ein Feld dafür wäre eine Zusage ohne Gegenüber
          (§6). */}
      {sitzt && <Tischgespraech lauf={gespraech} />}

      <p className="max-w-prose text-xs text-leise">{eigene.gastAbriss}</p>
    </section>
  )
}

// Gasttisch zeigt, was der Wirt zuletzt gerechnet hat.
//
// **Über dasselbe Bauteil wie ein eigener Tisch.** Ein zweites, das dieselbe
// Sicht anders zeichnete, wäre eine zweite Wahrheit über dieselbe Partie —
// und beim ersten Unterschied wüsste niemand, welche stimmt (§16).
function Gasttisch({ lauf }: { lauf: Gastlauf }) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const tisch = lauf.tisch

  if (tisch === undefined) {
    return <p className="text-sm text-leise">{eigene.gastOhneTisch}</p>
  }

  return (
    <Partietisch lauf={gastalsLauf(lauf, tisch)} partie={tisch.partie} fremd />
  )
}

// gastalsLauf macht aus dem Gastlauf einen Partielauf für den Tisch.
//
// # Warum eine Umsetzung und nicht ein zweiter Tisch
//
// **Weil `Partietisch` eine Partie zeichnet und nicht weiß, woher sie kommt.**
// Was der Gast **anders** kann, sind genau zwei Dinge: befehlen heißt hier
// **wünschen**, und die drei Standwechsel gehören dem Wirt — ein Gast hält
// keine fremde Partie an. Die Umsetzung sagt das, statt es zu verstecken.
// **Das Bild kommt als Argument und nicht aus dem Lauf.** Der Aufrufer hat es
// schon in der Hand — es dort noch einmal aus `lauf.tisch` zu holen, hieße
// einen Fall zu behandeln, den es an dieser Stelle nicht gibt: ohne Bild wird
// gar nicht gezeichnet (§10).
export function gastalsLauf(lauf: Gastlauf, tisch: PartieAntwort): Partielauf {
  return {
    anfangen: nichts,
    ansehen: nichts,
    befehlen: lauf.wuenschen,

    // **Ein Gast hält die Partie nicht an.** Das ist die Entscheidung des
    // Wirts, und eine Schaltfläche, die nichts täte, wäre eine Zusage ohne
    // Einlösung (§10) — die Tischfläche hält sie über `erlaubt` verborgen.
    wechseln: nichts,
    raeumen: nichts,
    laeuft: () => lauf.laeuft,
    partie: tisch.partie,

    // **Ein Gast wechselt keine Sicht** — er sitzt an seinem Sitz, und
    // `ansehen` tut hier nichts.
    wechselt: false,

    // **Die erlaubten Befehle kommen aus dem Tischbild.** Was ein Sitz darf,
    // rechnet der Wirt; der Gast liest es ab — und wenn einmal eine
    // Zugstruktur dazukommt, steht hier, dass er gerade nichts darf, ohne
    // dass die Sicht sich ändert.
    erlaubt: tisch.erlaubt,

    // **Und die Auskunft je Objekt ebenso** — gerechnet beim Wirt (P3,
    // #2371).
    handlungen: tisch.handlungen,

    // **Ein Gast wartet auf den Wirt** — seine Wünsche stehen der Reihe nach
    // an (P5, #2373).
    wartend: lauf.wartend,

    // **Die Befunde des Wirts gehen nicht mit** (G8, #1465).
    //
    // **Weil sie über **seinen** Rechner sprechen**: ein Profil, das dort
    // fehlt, sagt hier niemandem etwas — und ein Gast könnte die Bindung
    // ohnehin nicht richten. Was er sähe, wäre eine Auffälligkeit, gegen die
    // er nichts tun kann (§6).
    befunde: [],

    ereignisse: lauf.zuege,
    ereignisluecke: lauf.luecke,
    fehler: lauf.fehler,

    // **Der Gast hat nur einen Weg, an dem etwas schiefgeht** — wünschen.
    // Ein Vorgangsname, der eine Handlung nennt, die es hier nicht gibt,
    // wäre eine Zusage ohne Einlösung (§10).
    fehlerVorgang: undefined,
  }
}

// nichts tut nichts — mit einem Rumpf, weil eine leere Funktion ein
// Lintbefund ist und weil der Rumpf sagt, dass sie mit Absicht leer ist.
function nichts(): void {
  return undefined
}
