// **Nutzer und Profile** — Phase 4.12
// (#1351),
// nach Vermerk [0066](../../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md).
//
// # Drei Laschen, und der Schnitt ist der der Domäne
//
// *Allgemeines* ist, was der Person gehört; *Konto* ist, womit sie sich
// anmeldet; *Profile* ist die Arbeitsumgebung. Die Seite zeigt damit den Bau,
// statt ihn zu verstecken — auf Anweisung des Auftraggebers vom 2026-09-04.
//
// # Und daneben, wer hier arbeitet
//
// **Seit UXR-20** (#1928),
// nach dem Artboard `Nutzer`: links die Profilwahl und wer gerade arbeitet —
// der häufigste Handgriff der Seite, ohne eine Lasche zu öffnen —, rechts die
// Laschen. **Die Laschen bleiben**, wie sie verlangt waren; was selten ist,
// steht in ihnen eingeklappt. Wo das vom Artboard abweicht, steht in
// `docs/PROTOTYP-0.6.0.md`.
import { Anmeldung } from './Anmeldung'
import { Kontoanlage } from './Kontoanlage'
import { Kontoanonym } from './Kontoanonym'
import { Zweiterfaktor } from './Zweiterfaktor'
import { Sitzungsbereich } from './Sitzungsbereich'
import { Persoenliches } from './Persoenliches'
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Laschen } from '../bausteine/Laschen'
import { Nutzerangaben } from './Nutzerangaben'
import { Nutzeranlage } from './Nutzeranlage'
import { Nutzerpflege } from './Nutzerpflege'
import { Profilangaben } from './Profilangaben'
import { Profilfreigaben } from '../bausteine/Profilfreigaben'
import { Profilvorgaben } from '../bausteine/Profilvorgaben'
import { Zugriffstoken } from '../bausteine/Zugriffstoken'
import { Nutzerwahl } from './Nutzerwahl'
import { Profilwahl } from './Profilwahl'
import { useTexte } from '../sprache/kontext'

import type { AnbieterkontoDTO, NutzerDTO } from '../api'
import type { AnmeldeanbieterDTO } from '../apinutzer'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface Eigenschaften {
  nutzer: readonly NutzerDTO[]

  // gezeigt ist die Kennung des Profils, in dem gearbeitet wird.
  gezeigt: string

  // anbieter sind die Anbieter, über die man sich anmelden kann.
  anbieter: readonly string[]

  // anbieterstaende sagen, ob sie eingerichtet und wogegen sie geprüft sind.
  anbieterstaende?: readonly AnmeldeanbieterDTO[]

  // konten sind die Anbieterkonten des Nutzers, dem das gezeigte Profil gehört.
  konten: readonly AnbieterkontoDTO[]

  // eigner ist dieser Nutzer.
  eigner: string

  // arbeitet ist der Mensch, der gerade arbeitet — Werkstück 17 von Phase 4.12.
  arbeitet: string

  // bilder sind die Profilbilder als Daten-URI, je Nutzerkennung.
  bilder: Record<string, string>

  // badges sind die Abzeichen, je **Profil**kennung — Werkstück 9 (#1351).
  badges: Record<string, string>

  // neuLaden meldet nach oben, dass die Liste neu zu lesen ist.
  neuLaden: () => void
}

export function Nutzer({
  nutzer,
  gezeigt,
  anbieter,
  anbieterstaende = [],
  konten,
  eigner,
  arbeitet,
  bilder,
  badges,
  neuLaden,
}: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  // **Der Mensch, dem das gezeigte Profil gehört.** Er ist der, der hier
  // arbeitet; einen anderen zu ändern, ist eine Verwaltungsaufgabe und steht
  // unter *Nutzer und Profile bearbeiten*.
  const eigenerNutzer = nutzer.find((einer) => einer.id === eigner)

  // **Das gezeigte Profil** — die Rolle, in der gerade gearbeitet wird. Sein
  // Abzeichen liegt über dem Nutzerfoto (Werkstück 9, Punkt 13).
  const gezeigtesProfil = eigenerNutzer?.profile.find(
    (eines) => eines.id === gezeigt,
  )

  // **Die Überschrift steht in `Nutzerseite`** — sie gilt auch, während
  // geladen wird (§6, und der Wächter über die Ebenen); der Vorbehalt über
  // der Seite ist seit UXR-20 ihre Kopferklärung.
  //
  // **Nebeneinander, solange der Platz reicht** — gefragt wird der Behälter
  // und nicht das Fenster, wie beim Ausgabeassistenten: darunter steht die
  // Spalte über den Laschen, und nichts rollt waagerecht (R5).
  return (
    <div className="@container">
      <div className="flex flex-col gap-6 @3xl:flex-row @3xl:items-start">
        {/* **Wer hier arbeitet** — die linke Spalte des Artboards `Nutzer`
            (UXR-20, #1928): in welchem Profil gearbeitet wird und wer
            arbeitet, beides ohne eine Lasche zu öffnen. Zwei Listen und
            nicht eine, weil es zwei Gesten sind (`Nutzerwahl.tsx`). Vorher
            stand die Profilwahl in der Lasche *Profile* und die Nutzerwahl
            in *Allgemeines*, unter einer zweiten Liste derselben Menschen. */}
        <div className="grid gap-6 rounded-md border border-kante p-3 @2xl:grid-cols-2 @3xl:w-72 @3xl:shrink-0 @3xl:grid-cols-1">
          <Profilwahl nutzer={nutzer} gezeigt={gezeigt} />

          {/* **Wer arbeitet, ist eine Frage über den Menschen** —
              Werkstück 17 (#1400); die Liste darüber stellt Fragen über die
              Arbeitsumgebung. */}
          <Nutzerwahl
            nutzer={nutzer}
            arbeitet={arbeitet}
            bilder={bilder}
            badges={badges}
            gezeigt={gezeigt}
          />
        </div>

        <div className="min-w-0 flex-1">
          <Laschen
            bezeichnung={eigene.titel}
            gebiete={[
              {
                schluessel: 'allgemein',
                name: eigene.laschen.allgemein,
                inhalt: (
                  <div className="flex flex-col gap-4">
                    <Erklaerungsabsatz
                      text={eigene.allgemeinErklaerung}
                      klasse="max-w-prose text-leise"
                    />

                    {/* **Die Angaben des Menschen, der hier arbeitet** —
                        Werkstück 8 (#1351).

                        **Der `key` trägt Kennung *und* Fassung**, und das löst
                        §7 Regel 1 und 2 in einer Gestalt: ein Wechsel des
                        Eigners räumt den halb eingetippten Entwurf ab, und nach
                        einem geglückten Schreiben beginnt er bei dem, was jetzt
                        im Bestand steht — sonst ginge der nächste Ruf gegen eine
                        Fassung, die es nicht mehr gibt. */}
                    {eigenerNutzer !== undefined && (
                      <Nutzerangaben
                        key={`${eigenerNutzer.id}:${eigenerNutzer.fassung}`}
                        nutzer={eigenerNutzer}
                        geaendert={neuLaden}
                      />
                    )}

                    {/* **Umbenennen, Löschen und Anlegen eingeklappt** — UXR-20
                        (#1928): beides ist eine Verwaltungsaufgabe, und mit drei
                        Menschen und sieben Profilen standen hier zehn Zeilen mit
                        „Bearbeiten" unter dem Formular. */}
                    <Aufklappkasten titel={eigene.pflege}>
                      <Nutzerpflege
                        nutzer={nutzer}
                        bilder={bilder}
                        geaendert={neuLaden}
                      />
                    </Aufklappkasten>

                    <Aufklappkasten titel={eigene.neuerNutzer}>
                      <Nutzeranlage angelegt={neuLaden} />
                    </Aufklappkasten>
                  </div>
                ),
              },
              {
                schluessel: 'konto',
                name: eigene.laschen.konto,
                inhalt: (
                  <div className="flex flex-col gap-4">
                    <Erklaerungsabsatz
                      text={eigene.kontoErklaerung}
                      klasse="max-w-prose text-leise"
                    />

                    {/* **Das Alltägliche vorn** — UXR-20 (#1928): sperren,
                        abmelden und der zweite Faktor stehen offen; ein Konto
                        eintragen, eines ohne Angaben und die Anbieter sind
                        seltene Wege und stehen eingeklappt dahinter. Vorher
                        begann die Lasche mit dem Eintragen eines Kontos und
                        endete mit der Sitzung. */}
                    <Sitzungsbereich />

                    <Zweiterfaktor nutzerID={eigner} />

                    <Aufklappkasten titel={eigene.konto.eigenes}>
                      <Kontoanlage geaendert={neuLaden} />
                    </Aufklappkasten>

                    <Aufklappkasten titel={eigene.konto.ohneAngaben}>
                      <Kontoanonym geaendert={neuLaden} />
                    </Aufklappkasten>

                    <Aufklappkasten titel={eigene.anmeldung}>
                      <Anmeldung
                        anbieter={anbieter}
                        staende={anbieterstaende}
                        konten={konten}
                        nutzerID={eigner}
                        geaendert={neuLaden}
                      />
                    </Aufklappkasten>
                  </div>
                ),
              },
              {
                schluessel: 'profil',
                name: eigene.laschen.profil,
                inhalt: (
                  <div className="flex flex-col gap-4">
                    <Erklaerungsabsatz
                      text={eigene.profilErklaerung}
                      klasse="max-w-prose text-leise"
                    />

                    {/* **Die Angaben des Profils, in dem gearbeitet wird** —
                        Werkstück 9 (#1351). **Mit `key` aus Kennung und
                        Fassung**, wie beim Nutzer: ein Wechsel räumt den
                        Entwurf ab, und nach einem geglückten Schreiben beginnt
                        er beim neuen Stand. */}
                    {gezeigtesProfil !== undefined && (
                      <Profilangaben
                        key={`${gezeigtesProfil.id}:${gezeigtesProfil.fassung}`}
                        profil={gezeigtesProfil}
                        badgeURI={badges[gezeigtesProfil.id] ?? ''}
                        geaendert={neuLaden}
                      />
                    )}

                    {/* **Vorgaben, Freigaben und Token eingeklappt** — UXR-20
                        (#1928), R3 aus `docs/UIREVIEW.md`: mit drei Menschen,
                        fünf Vorgaben und drei Token zeichnete diese Lasche 36
                        Felder. Offen bleiben die Angaben; das Seltene steht
                        dahinter, und ein Kasten fragt den Bestand erst, wenn
                        jemand ihn öffnet.

                        **Je Abschnitt ein eigener Schreibweg** (§7, Regel 4) —
                        deshalb drei Abschnitte und nicht ein Formular. **Ein
                        `key` aus der Kennung allein**: ein Profilwechsel räumt
                        sie ab (§7, Regel 1), eine neue Fassung der Angaben
                        nicht. **Mit einem Vorsatz je Abschnitt**: drei
                        Geschwister mit demselben Schlüssel sind für React einer
                        zu viel (gefunden am 2026-09-24 über
                        `test/einrichtung.ts`). */}
                    {gezeigtesProfil !== undefined && (
                      <>
                        <Aufklappkasten titel={eigene.vorgaben}>
                          <Profilvorgaben
                            key={`vorgaben:${gezeigtesProfil.id}`}
                            profilID={gezeigtesProfil.id}
                          />
                        </Aufklappkasten>

                        {/* **Wer außer dem Eigner an diesem Profil etwas darf**
                            — Werkstück 11 (#1351). */}
                        <Aufklappkasten titel={eigene.freigaben}>
                          <Profilfreigaben
                            key={`freigaben:${gezeigtesProfil.id}`}
                            profilID={gezeigtesProfil.id}
                            nutzer={nutzer}
                            eigner={gezeigtesProfil.nutzerID}
                          />
                        </Aufklappkasten>

                        {/* **Die Zugriffstoken daneben und nicht im Formular** —
                            Werkstück 10 (#1351): dort wären sie ein zweiter
                            Schreibweg auf dieselbe Fassung. */}
                        <Aufklappkasten titel={eigene.token}>
                          <Zugriffstoken
                            key={`token:${gezeigtesProfil.id}`}
                            profilID={gezeigtesProfil.id}
                          />
                        </Aufklappkasten>
                      </>
                    )}
                  </div>
                ),
              },
              {
                // **Was dem Menschen gehört und nicht dem Profil** — am
                // 2026-09-06 von der Einstellungsseite hierher gezogen. Der
                // Grund steht in `Persoenliches.tsx`.
                schluessel: 'persoenlich',
                name: eigene.laschen.persoenlich,
                inhalt: <Persoenliches />,
              },
            ]}
          />
        </div>
      </div>
    </div>
  )
}
