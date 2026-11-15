// **Ein Exemplar aufnehmen** — Werkstück 2 von Phase 3.4
// (#1165).
//
// # Warum das Formular offen dasteht und nicht in einem Dialog
//
// §6: keine verschachtelten Dialoge, und Bearbeitung im Seitenbereich. Wer
// zehn Karten einträgt, soll zehnmal tippen und nicht zehnmal einen Dialog
// öffnen.
import { useState } from 'react'

import { Aufklappkasten } from './Aufklappkasten'
import { Drucksuche } from './Drucksuche'
import { Beschriftetesfeld, eingabeklasse } from './Beschriftetesfeld'
import { Vorschlagsfeld } from './Vorschlagsfeld'
import { Waehrungsfeld } from './Waehrungsfeld'
import { Zahlenfeld } from './Zahlenfeld'
import { Feld } from './Feld'
import { Knopf } from './Knopf'
import { Zustandswahl } from './Zustandswahl'
import { Ausfuehrungswahl } from './Ausfuehrungswahl'
import { leeresExemplar } from './exemplarentwurf'
import { useDruckwahl } from '../haken/useDruckwahl'
import { useEinstellungen } from '../haken/useEinstellungen'
import { useVorgabe } from '../haken/useVorgabe'
import { preisAuslegen } from '../preiseingabe'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { drucksprachen } from '../drucksprachen'
import { waehrungsbefund } from '../waehrungsgestalt'

import type { ExemplarDTO } from '../api'

interface SammlungsformularEigenschaften {
  // aufAufnehmen sagt mit seinem Rückgabewert, ob es geglückt ist.
  //
  // **Ohne diese Auskunft räumte das Formular sich auch nach einem
  // Fehlschlag** — und die Eingabe wäre weg, genau dann, wenn man sie braucht.
  aufAufnehmen: (exemplar: ExemplarDTO) => Promise<boolean>

  laeuft: boolean

  // geleert räumt das Formular von außen — nach einer zusammengeführten
  // Dublette (DOM-03, #1698): die Aufnahme ist dann geschehen, nur nicht
  // über diesen Knopf. **Über den `key`**, wie die Vorgabe darunter.
  geleert?: number
}

// Sammlungsformular holt die Vorgabe und baut das Formular darüber.
//
// # Warum zwei Bauteile und nicht eines
//
// **Weil die Vorgabe später ankommt als der erste Zeichenvorgang.** Sie kommt
// aus dem Bestand; ein `useState`, das sie beim ersten Zeichnen liest, bekäme
// eine leere Zeichenkette und behielte sie. Ein `useEffect`, der nachzieht,
// wäre ein `setState` im Effekt — der React-Linter weist es ab, und er hat
// recht.
//
// **Der Weg dieses Projekts ist ein `key`** (§7, „Regel 1 und 2 haben drei
// Gestalten"): das innere Bauteil entsteht neu, sobald die Vorgabe dasteht,
// und sein Entwurf beginnt bei ihr. Ein Bauteil kann sich nicht selbst
// schlüsseln — deshalb zwei.
//
// **`ohne` heißt leer**: der Wert steht so im Bestand, und ihn als Zustand
// einzutragen hieße, „ohne" als Kartenzustand auszugeben.
export function Sammlungsformular(
  eigenschaften: SammlungsformularEigenschaften,
) {
  const einstellungen = useEinstellungen()
  const gewaehlt = einstellungen.wert('vorgabe.zustand')
  const vorgabe = gewaehlt === 'ohne' ? '' : gewaehlt

  // **Und der erste zugeordnete Lagerort** — Werkstück 15 (#1351), Punkt 7.
  //
  // **Der erste und nicht alle**: eine Zeile liegt an **einem** Ort. Was die
  // Zuordnung mehrwertig macht, ist, dass ein Profil mehrere führen darf; wo
  // eine neue Zeile anfängt, ist trotzdem einer.
  const lagerort = useVorgabe('lagerort')[0] ?? ''

  // **Und die Währung** — eine eigene Einstellung und keine Folge der
  // Sprache (LOC-03, #1689).
  const waehrung = einstellungen.wert('vorgabe.waehrung')

  return (
    <Formular
      key={`${vorgabe}:${lagerort}:${waehrung}:${String(eigenschaften.geleert)}`}
      vorgabe={vorgabe}
      lagerort={lagerort}
      waehrung={waehrung}
      {...eigenschaften}
    />
  )
}

function Formular({
  aufAufnehmen,
  laeuft,
  vorgabe,
  lagerort,
  waehrung,
}: SammlungsformularEigenschaften & {
  vorgabe: string
  lagerort: string
  waehrung: string
}) {
  const texte = useTexte()
  const eigene = texte.seiten.sammlung

  const sprachvorschlaege = drucksprachen.map((code) => ({
    wert: code,
    bezeichnung: texte.bausteine.drucksprachen.namen[code],
  }))

  const [entwurf, setzeEntwurf] = useState(() =>
    leeresExemplar(vorgabe, lagerort, waehrung),
  )

  const darstellung = useDarstellung()

  // **Der Befund am Preis** — LOC-03.AK-03
  // (#1689):
  // er gehört zu genau dem Text und der Währung, aus denen er stammt, und
  // geht, sobald sich eines davon ändert.
  const [preisbefund, setzePreisbefund] = useState<string>()

  // **Ob die Menge im Feld abgeschickt werden darf** — das Zahlenfeld reicht
  // eine 0 nicht weiter, und ohne diese Angabe schickte der Knopf still die
  // letzte gültige (UX-06.AK-03).
  const [mengeGilt, setzeMengeGilt] = useState(true)

  // **Ein Befund hält die weiteren Angaben offen** — UXR-10 (#1918). In
  // einer zugeklappten Gruppe sähe ihn niemand, und Aufnehmen schriebe
  // nicht, ohne dass einer sagt, warum. **Der am Preis bis zum nächsten
  // Aufnehmen**: er geht beim ersten Tastendruck, und die Gruppe klappte
  // sonst unter den Fingern zu. Die Währung meldet sich, solange sie tippt.
  const [festgehalten, setzeFestgehalten] = useState(false)
  const waehrungFalsch =
    waehrungsbefund(entwurf.erwerbswaehrung, true) !== undefined

  const setze = (teil: Partial<ExemplarDTO>) => {
    setzeEntwurf((vorher) => ({ ...vorher, ...teil }))
  }

  const druck = useDruckwahl(setzeEntwurf)

  const aufnehmen = () => {
    // **Der Preis geht kanonisch hinüber** — so, wie Go ihn liest
    // (`docs/LOKALISIERUNG.md`); gelesen wie beim Bearbeiten
    // (`preiseingabe.ts`). **Zwei Werte werden nicht geraten** (AK-03).
    const auslegung = preisAuslegen(
      entwurf.erwerbspreis,
      entwurf.erwerbswaehrung,
      darstellung,
      eigene.preisMehrdeutig,
    )

    if (auslegung.art === 'frage') {
      setzePreisbefund(auslegung.befund)
      setzeFestgehalten(true)

      return
    }

    const erwerbspreis = auslegung.kanonisch

    void aufAufnehmen({ ...entwurf, erwerbspreis }).then((geglueckt) => {
      // **Geräumt wird nur nach einem geglückten Zug.** Wer bei einem Fehler
      // ein leeres Formular vorfände, hätte seine Eingabe verloren (§7,
      // Regel 3).
      if (geglueckt) {
        setzeEntwurf(leeresExemplar(vorgabe, lagerort, waehrung))
        setzeFestgehalten(false)
        druck.neueRunde()
      }
    })
  }

  return (
    // **Ohne eigene Überschrift**: es steht in der Schublade *Ein Exemplar
    // aufnehmen* (UXR-10), und zwei Namen für dieselbe Sache sind einer zu
    // viel.
    <section className="flex flex-col gap-3">
      {/* **Offen steht, was ein Exemplar ausmacht** — UXR-10 (#1918), R3
          in `docs/UIREVIEW.md`: vierzehn Felder standen offen da. Was eine
          Zeile von der nächsten unterscheidet, steht vorn; Bezug, Erwerb
          und Notiz sind wahlfrei und stehen eingeklappt darunter. */}
      {/* **Der Name führt** — AP-10 (#2367), `Wege09`: die Druckkennung
          steht unter „Weitere Angaben“, der gewählte Druck bringt Name, Set
          und Nummer mit. */}
      <Drucksuche key={druck.runde} laeuft={laeuft} waehle={druck.gewaehlt} />

      <div className="flex flex-wrap items-end gap-2">
        {/* **Eine Menge ist eine ganze Zahl ab eins** (UX-06.AK-02) —
            dieselbe Regel wie `domain.Exemplar.Validate`. Das Zahlenfeld
            sagt es am Feld, statt eine Null an den Dienst zu schicken. */}
        <Zahlenfeld
          beschriftung={eigene.menge}
          wert={entwurf.menge}
          kleinstens={1}
          ganzzahlig
          geaendert={(menge) => {
            setze({ menge })
          }}
          gilt={setzeMengeGilt}
        />

        <Zustandswahl
          wert={entwurf.zustand}
          setzeWert={(wert) => {
            setze({ zustand: wert })
          }}
        />

        {/* **Vorschläge und keine Grenze** (UX-06.AK-01): die Codes, unter
            denen Scryfall Sprachen führt — und jeder andere, den jemand
            braucht. */}
        <Vorschlagsfeld
          beschriftung={eigene.sprache}
          wert={entwurf.sprache}
          setzeWert={(wert) => {
            setze({ sprache: wert })
          }}
          vorschlaege={sprachvorschlaege}
          beispiel={texte.bausteine.drucksprachen.beispiel}
          breite={5}
        />

        <Ausfuehrungswahl
          wert={entwurf.finish}
          setzeWert={(wert) => {
            setze({ finish: wert })
          }}
        />

        {/* **Der Lagerort ist Freitext** (Vermerk 0057): „Kiste 4", „Ordner
            Rot, Seite 12", „bei Anna". Jede Aufzählung, die wir vorgeben, ist
            bei jemandem falsch. */}
        <Beschriftetesfeld
          bezeichnung={eigene.lagerort}
          wert={entwurf.lagerort}
          setzeWert={(wert) => {
            setze({ lagerort: wert })
          }}
          breite={16}
        />

        {/* **Die Absicht als Ankreuzfeld.** Ein Wunsch ist derselbe Verbund
            wie ein Besitz — nur mit dem anderen Vorzeichen (0057). */}
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={entwurf.wunsch}
            onChange={(ereignis) => {
              setze({ wunsch: ereignis.target.checked })
            }}
            className="focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
          {eigene.wunsch}
        </label>
      </div>

      <Aufklappkasten
        titel={eigene.weitereAngaben}
        offenGehalten={festgehalten || waehrungFalsch}
      >
        <div className="flex flex-wrap items-end gap-2">
          {/* **Die Kennung für den, der sie zur Hand hat** — seit AP-10 hier
              und nicht mehr im Hauptweg. */}
          <Beschriftetesfeld
            bezeichnung={eigene.druck}
            wert={entwurf.printing}
            setzeWert={(wert) => {
              setze({ printing: wert })
            }}
            breite={36}
          />

          {/* **Der lesbare Bezug** (#1200) — wahlfrei: wer nur eine Kennung zur
              Hand hat, trägt sie ein und sieht sie in der Liste. Wer den Namen
              dazuschreibt, liest seine Sammlung später ohne Nachschlagen. */}
          <Beschriftetesfeld
            bezeichnung={eigene.name}
            wert={entwurf.name}
            setzeWert={(wert) => {
              setze({ name: wert })
            }}
            breite={20}
          />

          <Beschriftetesfeld
            bezeichnung={eigene.setcode}
            wert={entwurf.setcode}
            setzeWert={(wert) => {
              setze({ setcode: wert })
            }}
            breite={5}
          />

          <Beschriftetesfeld
            bezeichnung={eigene.nummer}
            wert={entwurf.nummer}
            setzeWert={(wert) => {
              setze({ nummer: wert })
            }}
            breite={5}
          />

          {/* **Über `Feld` und nicht über `Beschriftetesfeld`**, damit ein
              Befund am Feld stehen kann (§6) — und nicht als Hinweis
              irgendwo darunter. */}
          <Feld
            beschriftung={eigene.erwerbspreis}
            {...(preisbefund === undefined ? {} : { befund: preisbefund })}
          >
            {(verbindung) => (
              <input
                {...verbindung}
                value={entwurf.erwerbspreis}
                size={12}
                onChange={(ereignis) => {
                  setzePreisbefund(undefined)
                  setze({ erwerbspreis: ereignis.target.value })
                }}
                className={eingabeklasse}
              />
            )}
          </Feld>

          {/* **Klein geschrieben darf es sein**: `Exemplar.Normalisieren` hebt
              es an (Vermerk 0057). */}
          <Waehrungsfeld
            beschriftung={eigene.erwerbswaehrung}
            wert={entwurf.erwerbswaehrung}
            setzeWert={(wert) => {
              setzePreisbefund(undefined)
              setze({ erwerbswaehrung: wert })
            }}
            kleinErlaubt
          />

          {/* **Erwerbsdatum und Notiz gehören dazu** — sie standen an den
              anderen Flächen und fehlten hier. Eine Fläche, die weniger sagen
              kann als die anderen, führt nicht dieselbe Sammlung (§16).
              Angestrichen vom Zweitblick über Phase 3.4. */}
          <Beschriftetesfeld
            bezeichnung={eigene.erwerbsdatum}
            wert={entwurf.erwerbsdatum}
            setzeWert={(wert) => {
              setze({ erwerbsdatum: wert })
            }}
            art="date"
          />

          <Beschriftetesfeld
            bezeichnung={eigene.notiz}
            wert={entwurf.notiz}
            setzeWert={(wert) => {
              setze({ notiz: wert })
            }}
            breite={20}
          />
        </div>
      </Aufklappkasten>

      <div>
        <Knopf
          bezeichnung={eigene.aufnehmen}
          aufKlick={aufnehmen}
          gewicht="haupt"
          angehalten={
            laeuft || entwurf.printing === '' || !mengeGilt || waehrungFalsch
          }
        >
          {eigene.aufnehmen}
        </Knopf>
      </div>
    </section>
  )
}
