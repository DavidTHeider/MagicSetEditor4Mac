import { useCallback, useEffect, useState } from 'react'

import { Ablageraum } from './ablage'
import { Ablagehinweis } from './bausteine/Ablagehinweis'
import { Nutzereinstellungsraum } from './nutzereinstellungsraum'
import { Gestaltungsraum } from './gestaltungsraum'
import { Einstellungsraum } from './haken/einstellungsraum'
import { Schreibrecht } from './haken/schreibrecht'
import { Profilraum } from './profil'
import { Ruecknahmeraum } from './ruecknahme'
import { Anwendung } from './Anwendung'
import { Darstellungsstandraum } from './darstellungsstandraum'
import { Objekthandlungsraum } from './objekthandlungsraum'

import { Wiederherstellung } from './seiten/Wiederherstellung'
import { Sitzungstor } from './seiten/Sitzungstor'
import { Nurlesehinweis } from './bausteine/Nurlesehinweis'
import { useStartbefund, type Startstand } from './haken/useStartbefund'
import { useSprachmeldung } from './haken/useSprachmeldung'
import { useEigeneSprache } from './haken/useEigeneSprache'
import { spracheAus } from './sprache/einstellung'
import { darstellungAus, zeitzoneAus } from './sprache/darstellung'
import {
  setzeAnzeigeschrift,
  setzeErscheinungsbild,
  setzeSchriftgroesse,
} from './erscheinungsbild'
import { useEinstellungen, type Fragehaltung } from './haken/useEinstellungen'
import { useBestandsstand } from './haken/useBestandsstand'
import { Sprachraum } from './sprache/kontext'
import { vergissVorlagenentwuerfe } from './haken/useVorlagenpuffer'
import { leere as leereBefehlsverlauf } from './befehlsverlauf'
import { leere as leereSkriptverlauf } from './skriptverlauf'

// **`App` stellt den Sprachraum her und sonst nichts.**
//
// # Warum die Anwendung eine Ebene tiefer steht
//
// Ein Haken, der `useTexte()` ruft, liest den **Kontext** — und wer ihn über
// dem `Sprachraum` ruft, bekommt dessen Vorgabewert: Deutsch. Solange `App`
// beides tat, war `useSetSpeichern` und `useVerlauf` genau das passiert: die
// Oberfläche stand auf Englisch, ihre Fehlermeldungen kamen aus dem deutschen
// Katalog.
//
// Gefunden hat es die **Sichtprüfung am laufenden Fenster**, nicht ein Test —
// im Fenster stand „This could not be loaded. Der gesuchte Eintrag wurde nicht
// gefunden." (#133).
//
// **Für einen der Fälle stand die Falle schon im Quelltext angemerkt** und war
// dort mit `kataloge[sprache]` umgangen. Zwei Haken daneben blieben übrig —
// ein Wächter, der die halbe Zusage hält, macht die andere unsichtbar (§10).
// Deshalb wird hier die **Ursache** geräumt und nicht der dritte Fall: über
// dieser Zeile gibt es keinen Haken mehr, der einen Katalog braucht.
export function App() {
  // **Die Einstellungen kommen aus dem Bestand** (#245). Bis dahin lag die
  // Sprache im Browserspeicher; der Grund dort war gut und wird anders
  // eingelöst — siehe [useEinstellungen].
  // **`bestandsstand` zählt, wie oft der Bestand ersetzt wurde.** Ohne ihn
  // blieben die Einstellungen des ersetzten Stands stehen — samt Sprache
  // (Zweitblick über Phase S).
  const [bestandsstand, setzeBestandsstand] = useState(0)

  // neuLesen lässt alles neu lesen, was vom Bestand kommt — derselbe Weg wie
  // nach dem Ersetzen, und seit UX-09 (#1681) nach dem Zurücksetzen der
  // Darstellung aus der Menüleiste.
  const neuLesen = useCallback(() => {
    setzeBestandsstand((bisher) => bisher + 1)
  }, [])

  // **`profilstand` zählt die Profilwechsel**, und er ist der `key` der
  // Anwendung: ein Wechsel baut sie neu auf. Warum, steht in `profil.tsx` —
  // §7, Regel 1, und beim Profil ist das Bezugsobjekt **alles**.
  const [profilstand, setzeProfilstand] = useState(0)

  // **Was beim Start geschah** (#962). Ging der Bestand nicht auf, öffnet das
  // Fenster trotzdem — mit dem Grund im Klartext und dem Weg zur Sicherung.
  const start = useStartbefund(bestandsstand)

  // **Und erst dann werden die Einstellungen gefragt.**
  //
  // Ohne offenen Bestand sind die Nähte nicht gesetzt; ein Aufruf darüber
  // läuft in einen Nullzeiger, Wails fängt ihn ab und ruft den Rückweg **nie**
  // — das Fenster wartete für immer, und diese Seite käme nie zum Vorschein.
  // Nachgemessen an einer laufenden Anwendung mit einer unbrauchbaren
  // Bestandsdatei.
  //
  // **Und nach jedem Profilwechsel** — die Summe wächst, sobald einer der
  // beiden Zähler wächst. Gelesen wurde bis zum Zweitblick über Phase 6.3
  // nur beim Start und nach einem ersetzten Bestand, und Profil B trug
  // Farbwelt, eigene Farben und Sprache von A (S5).
  //
  // **Und nach einer fremden Änderung** (Befund 2 des Gesamtreviews über
  // 0.6.0): ein Schalter, den Kommandozeile, MCP oder REST umlegen, wirkte
  // erst beim nächsten anderen Anlass. Der Zähler steht deshalb hier und
  // geht nach unten an die Seiten (#156).
  const fremderstand = useBestandsstand()
  const einstellungen = useEinstellungen(
    bestandsstand + profilstand + fremderstand,
    fragehaltung(start),
    // **Wessen Einstellungen** (U08, #2124): ein ersetzter Bestand oder ein
    // anderes Profil räumt bei einem Lesefehler, ein fremdes Schreiben nicht.
    `${String(bestandsstand)}/${String(profilstand)}`,
  )

  // **Die Sprache: die Wahl, sonst das System, sonst Deutsch** (§9). Der
  // Bestand gibt für sie **keinen** Vorgabewert her — was ohne Wahl gilt, weiß
  // nur das Fenster.
  const gewaehlt = spracheAus(einstellungen.wert('sprache'), navigator.language)

  // **Und eine eigene Übersetzung darüber** — LOC-02
  // (#1688).
  // Mit ihr spricht das Fenster ihre Sätze, und was fehlt, ihre Basis.
  const eigene = useEigeneSprache(einstellungen.wert('sprache.eigene'))
  const sprache = eigene?.basis ?? gewaehlt

  // **Und die Darstellung ganz** — LOC-03.AK-03
  // (#1689):
  // die Sprache nimmt von `de-CH` nur `de`, die Darstellung behält die
  // Gegend. Die Zone ist eine eigene Einstellung; ohne sie gilt die des
  // Rechners.
  const darstellung = darstellungAus(
    einstellungen.wert('darstellung'),
    navigator.language,
    sprache,
  )
  const zeitzone = zeitzoneAus(einstellungen.wert('zeitzone'))

  // **Die Sprache dem Vorleser und der nativen Menüleiste melden — auf beiden
  // Wegen.** Sie saß in `Anwendung`, und das Wiederherstellungsfenster meldete
  // deshalb nichts: die Seite sprach die Sprache des Systems, die Leiste die,
  // mit der Go gestartet war — Deutsch (Befund 2 des Zweitblicks über Phase
  // 6.2). Hier steht sie über beiden Zweigen.
  //
  // **Erst, wenn die Sprache feststeht** — aus demselben Grund, aus dem
  // darunter erst dann gezeichnet wird: sonst spränge die Leiste von der
  // Systemsprache auf die gewählte.
  const meldefehler = useSprachmeldung(
    einstellungen.geladen && start.geladen ? sprache : undefined,
    eigene?.kennung,
  )

  // **Erscheinungsbild, Schriftgröße und Anzeigeschrift stehen am
  // Wurzelelement**, nicht in React: alle drei gelten für alles, auch für den
  // Rollbalken, den die Anwendung gar nicht zeichnet.
  useEffect(() => {
    setzeErscheinungsbild(
      document.documentElement,
      einstellungen.wert('erscheinungsbild'),
    )
    setzeSchriftgroesse(
      document.documentElement,
      einstellungen.wert('schriftgroesse'),
    )
    setzeAnzeigeschrift(
      document.documentElement,
      einstellungen.wert('schriftart'),
    )
  }, [einstellungen])

  // **Gezeichnet wird erst, wenn die Sprache feststeht.** Sonst stünde die
  // Oberfläche einen Wimpernschlag auf Deutsch und spränge dann auf Englisch —
  // und ein Vorleseprogramm läse den ersten Satz mit der falschen Aussprache.
  // Ein **Fehler** beim Laden zählt als „geladen": dann gelten die
  // Vorgabewerte, und warten hilft nicht mehr.
  if (!einstellungen.geladen || !start.geladen) {
    return null
  }

  // **Ohne Bestand gibt es nur einen Weg**, und der führt über dieselben
  // Bindungen wie sonst. Die Einstellungen sind dann Vorgabewerte — der
  // Bestand, aus dem sie kämen, ist ja nicht offen —, und die Sprache kommt
  // vom System (§9).
  if (!start.offen) {
    return (
      <Sprachraum
        sprache={sprache}
        katalog={eigene?.katalog}
        darstellung={darstellung}
        zeitzone={zeitzone}
      >
        <Wiederherstellung
          grund={start.grund}
          schluessel={start.schluessel}
          rueckweg={start.rueckweg}
          bestandErsetzt={() => {
            setzeBestandsstand((bisher) => bisher + 1)
          }}
        />
      </Sprachraum>
    )
  }

  return (
    <Sprachraum
      sprache={sprache}
      katalog={eigene?.katalog}
      darstellung={darstellung}
      zeitzone={zeitzone}
    >
      {/* **Der Ablageraum steht über der Anwendung** (#891): dort meldet sich
          der Horcher an, hier trägt die Kartenseite ein, was sie mit einem
          Bild vorhat. */}
      {/* **Der Hinweis steht über allem** — Werkstück 16 von Phase 4.12
          (#1351). Er gilt für die ganze Anwendung, solange der Zeiger in
          einem fremden Profil steht, in das jemand als Leser gekommen ist.

          **Mit dem Profilschlüssel**: ein Wechsel baut ihn neu, und die
          Antwort gehört zu dem Profil, aus dem sie stammt (§7, Regel 1). */}
      {/* **Das Tor der Sitzung steht über der Anwendung** — ID-02 (#1701):
          gesperrt oder abgemeldet steht sie gar nicht da. Wer öffnet, liest
          den Bestand neu — gesperrt kamen nur Vorgabewerte an. */}
      <Sitzungstor geoeffnet={neuLesen}>
        {/* **Das Schreibrecht steht über allem** (G10, #2276): einmal je
            Profilstand gefragt, gelesen vom Streifen und von jedem Knopf,
            der schreibt. */}
        <Schreibrecht stand={profilstand}>
          <Nurlesehinweis />

          <Profilraum
            gewechselt={() => {
              // **Ein Entwurf gehört dem Profil, in dem er entstand**
              // (UX-03.AK-04, #1675) — im nächsten wären es fremde Daten.
              vergissVorlagenentwuerfe()
              // **Und die Verläufe von Befehlen und Skripten** (Nachlauf über
              // Phase 6.3, N6) — sie tragen Argumente des vorigen Menschen.
              leereBefehlsverlauf()
              leereSkriptverlauf()
              setzeProfilstand((bisher) => bisher + 1)
            }}
          >
            <Ablageraum>
              {/* **Was ein Ablegen tut, steht vor dem Loslassen da** (AP-4,
                  #2361) — für das ganze Fenster, wie das Ablegen selbst. */}
              <Ablagehinweis />
              {/* **Der Rücknahmeraum steht über der Anwendung** (#1340): dort hängt
            der Menübefehl, hier trägt der Symboleditor seinen eigenen Verlauf
            ein. Die offene Zeichnung geht vor, solange es sie gibt. */}
              <Ruecknahmeraum>
                {/* **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12
                (#1351): `data-bewegung` und `data-kontrast` am
                Wurzelelement, und dieselben Werte für die
                Einstellungsseite. **Mit demselben Schlüssel wie die
                Anwendung**: wer das Profil wechselt, kann ein anderer
                Mensch sein. */}
                {/* **Ein Einstellungsstand für alles darunter** — Befund G14
                (#1665): wer `useEinstellungen()` ruft, bekommt diesen, statt
                einen eigenen zu laden. */}
                <Einstellungsraum stand={einstellungen}>
                  <Nutzereinstellungsraum
                    key={profilstand}
                    stand={bestandsstand}
                  >
                    {/* **Und was die Arbeitsumgebung aussehen lässt** — UX-09
                    (#1681): Farbwelt, Fensterbild, Schrift und eigene
                    Merkmale, unter dem Nutzerraum, weil hoher Kontrast dort
                    steht und alles hier sticht. **Mit dem Profilschlüssel**:
                    ein gewähltes Fensterbild hielt sonst die Adresse des
                    vorigen Menschen fest (Zweitblick über Phase 6.3, S5). */}
                    <Gestaltungsraum
                      key={profilstand}
                      einstellungen={einstellungen}
                      neuLesen={neuLesen}
                    >
                      {/* **Die Teile des Fensters melden sich hier** — UXR-F14
                      (#2029): Seitenleiste und Arbeitsplatz sagen der nativen
                      Menüleiste, wie sie dastehen. **Über der Anwendung**, denn
                      `useSeitenleiste` ruft sie selbst. */}
                      <Darstellungsstandraum>
                        {/* **Und das gewählte Objekt seine Handlungen** —
                        GESTE (#2358): dieselbe Liste wie im Rechtsklick steht
                        als Menü in der Leiste. Unter dem Schlüssel des Profils,
                        damit ein Wechsel mit „nichts gewählt" beginnt. */}
                        <Objekthandlungsraum>
                          <Anwendung
                            key={profilstand}
                            sprache={sprache}
                            fassung={start.fassung}
                            rueckspielrest={start.rueckspielrest}
                            einstellungen={einstellungen}
                            meldefehler={meldefehler}
                            bestandErsetzt={() => {
                              setzeBestandsstand((bisher) => bisher + 1)
                            }}
                            fremderstand={fremderstand}
                          />
                        </Objekthandlungsraum>
                      </Darstellungsstandraum>
                    </Gestaltungsraum>
                  </Nutzereinstellungsraum>
                </Einstellungsraum>
              </Ruecknahmeraum>
            </Ablageraum>
          </Profilraum>
        </Schreibrecht>
      </Sitzungstor>
    </Sprachraum>
  )
}

// fragehaltung übersetzt den Startbefund in das, was der Einstellungshaken
// tun soll.
//
// **Warten, solange nichts feststeht** — sonst gälten für einen Wimpernschlag
// die Vorgabewerte, und die Oberfläche spränge in der Sprache um.
// **Gefragt wird, solange nichts dagegen spricht.**
//
// Nicht erst, wenn der Startbefund da ist: der Ladelauf soll mit dem Zeichnen
// beginnen und nicht eine Antwort später. Steht der Bestand offen — der
// Normalfall —, war das Warten reiner Verlust; steht er es nicht, hängt die
// Anfrage ohnehin, und der Lauf verwirft sie (§7, Regel 2).
function fragehaltung(start: Startstand): Fragehaltung {
  return !start.geladen || start.offen ? 'fragen' : 'ohne'
}
