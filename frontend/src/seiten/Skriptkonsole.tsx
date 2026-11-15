// Die **Skriptkonsole**
// (#1070).
//
// # Wofür
//
// MSE-Script ist vollständig portiert — und wer einen Ausdruck ausprobieren
// wollte, musste die Kommandozeile bemühen. MSE2 hat dafür einen eigenen
// Bereich (`src/gui/set/console_panel.cpp`, Fassung f9d9356 vom 2021-07-06),
// und §15 nennt seine Kommandozeile als **Orakel** dieses Projekts: eine
// eigene Konsole beantwortet dieselben Fragen über **dieses** Programm und
// macht die beiden vergleichbar.
//
// # Sie ändert nichts
//
// Die Sprache kann es nicht — `internal/schichten` weist jeden Zugriff auf das
// Dateisystem aus `msescript` ab. Die Erklärung sagt es, damit niemand es
// ausprobieren muss.
import { useEffect, useState } from 'react'

import { api } from '../fassade'
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Knopf } from '../bausteine/Knopf'
import { Quelltextfeld } from '../bausteine/Quelltextfeld'
import { Namensliste } from './Namensliste'
import { Skriptbeispiele } from './Skriptbeispiele'
import { Skriptergebnis } from './Skriptergebnis'
import { Skriptverlauf } from './Skriptverlauf'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { useLaden } from '../haken/useLaden'
import { useSkriptkonsole } from '../haken/useSkriptkonsole'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface SkriptkonsoleEigenschaften {
  // karteID ist die aufgeschlagene Karte — leer, wenn keine offen ist.
  karteID: string

  // zuDenKarten führt auf die Kartenlasche, wenn keine Karte offen ist.
  //
  // **Der Wechsler des Fensters und kein eigener Weg**: welche Seite es gibt
  // und wie man dorthin kommt, weiß der Rahmen — eine Lasche, die selbst
  // navigierte, hätte eine zweite Wahrheit darüber.
  zuDenKarten: (seite: 'karten') => void
}

export function Skriptkonsole({
  karteID,
  zuDenKarten,
}: SkriptkonsoleEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.konsole

  // **Was die Konsole tut, steht im Haken** (§2) — Lauf, Abbruch, Dauer,
  // Verlauf (Werkstück U13 von Phase 5.1, #1424). Hier steht, wie sie
  // aussieht.
  const konsole = useSkriptkonsole(karteID)
  const { ausdruck, setzeAusdruck, stelle } = konsole

  // zeigeNamen sagt, ob jemand die volle Liste sehen will.
  const [zeigeNamen, setzeZeigeNamen] = useState(false)

  // **Die Namen kommen für die Vervollständigung von selbst** — aber erst,
  // wenn jemand einen Bezeichner anfängt
  // (#1164).
  //
  // Der Grund für „nicht beim Aufschlagen" bleibt: wer `1 + 1` rechnet, hat
  // nicht nach 1.644 Namen gefragt. Wer `ca` tippt, schon — und ein Vorschlag,
  // der erst nach dem Aufklappen der Namen käme, käme zu spät.
  //
  // **Ein Lader für beide Anlässe.** Das Aufklappen setzt nur einen Schalter
  // (`Namensbedarf`); zwei
  // Wege zu denselben Namen liefen an dem Tag auseinander, an dem einer von
  // beiden etwas anders macht (§10).
  //
  // **Der Haken steht vor dem frühen Ausstieg**, weil Haken unbedingt laufen
  // müssen.
  const brauchtNamen =
    karteID !== '' && (zeigeNamen || /[\p{L}_][\p{L}\p{N}_]/u.test(ausdruck))

  const namensstand = useLaden(async () => {
    if (!brauchtNamen) {
      return []
    }

    const antwort = await api().Skriptnamen(karteID)

    return auspacken(antwort, () => antwort.namen)
  }, [brauchtNamen, karteID])

  const namen = namensstand.daten ?? []

  if (karteID === '') {
    return (
      <section className="flex flex-col items-start gap-2">
        <h2 className="text-xl font-semibold">{eigene.ueberschrift}</h2>
        <p className="text-sm text-leise">{eigene.keineKarte}</p>

        {/* **Ein leerer Zustand nennt den nächsten Schritt** (§6, Werkstück C
            von Phase 4.10). Der Satz darüber sagte, **was fehlt**; wer die
            Konsole sucht, musste selbst wissen, dass der Weg über die
            Kartenlasche geht. */}
        <Knopf
          aufKlick={() => {
            zuDenKarten('karten')
          }}
        >
          {eigene.zuDenKarten}
        </Knopf>
      </section>
    )
  }

  const auswerten = () => {
    void konsole.auswerten()
  }

  return (
    <section className="flex flex-col gap-4">
      <header className="flex flex-col gap-1">
        <h2 className="text-xl font-semibold">{eigene.ueberschrift}</h2>
        <Erklaerungsabsatz
          text={eigene.erklaerung}
          klasse="text-sm text-leise"
        />
      </header>

      <div className="flex flex-col gap-2">
        {/* **Dasselbe Feld wie im Vorlagenbereich** (#1164): Zeilennummern,
            Farbe aus dem Lexer, und die gemeldete Stelle angestrichen. Der
            Zeiger darunter bleibt daneben stehen — er ist vorlesbar, die
            Markierung ist es nicht. **⌘↩ wertet aus** (#1424). */}
        <Quelltextfeld
          bezeichnung={eigene.ausdruck}
          wert={ausdruck}
          setzeWert={setzeAusdruck}
          stelle={stelle}
          namen={namen}
          zeilen={3}
          platzhalter={eigene.platzhalter}
          aufAusfuehren={auswerten}
        />

        <div className="flex flex-wrap items-center gap-3">
          {/* **Die führende Handlung** (R1 aus `docs/UIREVIEW.md`, UXR-08):
              bis dahin sah sie aus wie jedes der neun Beispiele. */}
          <Knopf
            gewicht="haupt"
            bezeichnung={eigene.auswerten}
            angehalten={konsole.laeuft}
            aufKlick={auswerten}
          >
            {eigene.auswerten}
          </Knopf>

          {konsole.laeuft ? (
            <>
              <Knopf
                gewicht="neben"
                bezeichnung={eigene.abbrechen}
                aufKlick={konsole.abbrechen}
              >
                {eigene.abbrechen}
              </Knopf>
              <span role="status" className="text-sm text-leise">
                {eigene.laeuft}
              </span>
            </>
          ) : (
            <span className="text-xs text-leise">{eigene.kuerzel}</span>
          )}
        </div>
      </div>

      {konsole.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {konsole.fehler}
          {stelle !== null && (
            <>
              {' '}
              <span className="font-mono">
                {einsetzen(eigene.stelle, {
                  zeile: String(stelle.zeile),
                  spalte: String(stelle.spalte),
                })}
              </span>
            </>
          )}
        </p>
      )}

      {/* **Abgebrochen ist kein Fehler** — es ist, was jemand wollte; deshalb
          ein Status und kein Alarm. */}
      {konsole.abgebrochen && (
        <p role="status" className="text-sm text-leise">
          {eigene.abgebrochen}
        </p>
      )}

      {/* **Kein eigener Block für die angestrichene Zeile**: das Feld selbst
          markiert die Stelle (#1164). Das Issue sagt es: „Fehlermarkierung an
          der gemeldeten Stelle, **nicht als Liste daneben**"; zweimal
          dasselbe ist die Liste daneben.

          **Vorlesbar bleibt es**: die Meldung darüber steht in einem
          `role="alert"` und nennt Zeile und Spalte. Ein Block wäre nur die
          sichtbare Hälfte, und die kann das Feld besser. */}

      {/* **Das Ergebnis steht unter dem Knopf, der es hervorbringt** —
          UXR-08 (#1916).
          Bis dahin stand es unter den Beispielen und der Namensliste: wer die
          Namen offen hatte, sah nach dem Auswerten sechzig Knöpfe und kein
          Ergebnis. */}
      {konsole.ergebnis !== undefined && (
        <Skriptergebnis ergebnis={konsole.ergebnis} />
      )}

      {/* **Die zwei Hilfen sind eingeklappt** — für den, der die Sprache
          noch nicht kennt, und für ihn einen Klick entfernt. **Beispiele vor
          der Namensliste** (#1164): wer die Sprache nicht kennt, braucht
          einen Ausdruck, der etwas tut — eine Liste von 1.644 Namen hilft
          erst dem, der weiß, wonach er sucht. */}
      <Aufklappkasten titel={eigene.beispiele}>
        <Skriptbeispiele einsetzen={setzeAusdruck} />
      </Aufklappkasten>

      <Aufklappkasten titel={eigene.namen}>
        <div className="flex flex-col gap-2">
          {/* **Erst das Aufklappen holt die Namen** — der Kasten zeichnet
              seinen Inhalt beim ersten Öffnen, und dieser meldet dann den
              Bedarf. Bis UXR-08 tat das ein eigener Knopf „Namen anzeigen",
              der sich nicht wieder schließen ließ. */}
          <Namensbedarf melden={setzeZeigeNamen} />

          <Erklaerungsabsatz
            text={eigene.namenErklaerung}
            klasse="text-sm text-leise"
          />

          {/* **Der Fehler steht bei den Namen und nicht bei der Auswertung**
              (§6): er sagt etwas über die Liste, nicht über den Ausdruck. */}
          {namensstand.fehler !== undefined && (
            <p role="alert" className="text-sm text-fehler">
              {namensstand.fehler}
            </p>
          )}

          {/* **Eingegrenzt und gedeckelt** — warum, steht am Bauteil: an
              einer Karte mit eingebundener Vorlage sind es 1.644 Namen
              (#1164). */}
          <Namensliste
            namen={namen}
            einsetzen={(name) => {
              setzeAusdruck(ausdruck + name)
            }}
          />
        </div>
      </Aufklappkasten>

      {/* **Der Verlauf steht unter dem Ergebnis**, so wie MSE2s Konsole
          jeden Befehl in ihre Nachrichtenliste legt (`ConsolePanel::exec`,
          `src/gui/set/console_panel.cpp`, Fassung f9d9356 vom 2021-07-06). */}
      <Skriptverlauf laden={setzeAusdruck} />
    </section>
  )
}

// Namensbedarf meldet, dass jemand die Namen sehen will — beim Zeichnen, und
// gezeichnet wird es erst, wenn der Kasten zum ersten Mal aufgeht.
//
// **Ein Bauteil ohne Gestalt und kein Schalter am Kasten**: `Aufklappkasten`
// wird von vielen Seiten benutzt, und was es zusagt — gezeichnet beim ersten
// Öffnen —, ist genau der Anlass, den diese Seite braucht.
function Namensbedarf({ melden }: { melden: (gebraucht: boolean) => void }) {
  useEffect(() => {
    melden(true)
  }, [melden])

  return null
}
