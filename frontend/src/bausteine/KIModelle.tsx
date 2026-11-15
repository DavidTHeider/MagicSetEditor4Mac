// **Was über KI-Modelle feststeht** — Werkstück A6 von Phase 5.3
// (#1445).
//
// # Warum eine leere Liste die richtige Auskunft ist
//
// **Weil eine Modellliste keine Fähigkeit ist.** Das Issue sagt es scharf:
// „Unbekannt bleibt unbekannt; Modelllisten garantieren keine Capability."
// Eine Seite, die zwölf Namen zeigt, von denen keiner geprüft ist, wäre
// schlechter als eine, die sagt: es steht keines fest.
//
// **Seit A7 (#1446)
// steht daneben, warum sie leer ist.** Fünf Einrichtungsstände unterscheiden
// „niemand hat eine Laufzeit eingeschaltet" von „es antwortet keine", „es ist
// keines installiert" und „keines kann etwas Belegtes". Eine leere Liste ohne
// diesen Satz wäre eine Seite, die nicht sagt, was als Nächstes zu tun ist
// (§6).
import { KIGeraetSitzung } from './KIGeraetSitzung'
import { KIMessung } from './KIMessung'
import { derOrtLokal } from './kiwahl'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { useLaden } from '../haken/useLaden'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { alsZahl } from '../lokal'

export function KIModelle() {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen
  const darstellung = useDarstellung()

  // **Gelesen wird beim Aufschlagen.** Ist eine Laufzeit eingeschaltet,
  // fragt der Ruf sie — auf der Schleifenadresse, und der Riegel am
  // Transport hält ihn auf, wenn der Schalter aus ist.
  const stand = useLaden(async () => {
    const antwort = await api().KIModelle()

    return auspacken(antwort, () => antwort)
  }, [])

  const katalog = stand.daten

  return (
    // **Ohne eigene Überschrift seit #2465**: die Modelle stehen in einem
    // Abschnitt der Einstellungen, und dessen Knopf trägt den Namen.
    <section className="flex flex-col gap-2">
      {stand.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {stand.fehler}
        </p>
      )}

      {katalog !== undefined && (
        <>
          {/* **Die Regel steht vor der Liste** — sie sagt, ob überhaupt
              etwas diesen Rechner verlassen darf. */}
          <p className="text-sm text-leise">
            {katalog.nurLokal ? eigene.kiNurLokal : eigene.kiAuchCloud}
          </p>

          {/* **Je Anbieter eine Zeile** (A8, #1447) — und **keine
              Zusammenfassung darüber**. Der Katalog trägt `einrichtung` als
              besten Stand über alle, und am Fenster stand er als eigener
              Absatz über der Liste: bei zwei abgeschalteten Anbietern dreimal
              derselbe Satz. Die Sichtprüfung hat ihn angestrichen; für die
              Flächen ohne Fenster bleibt das Feld.

              **Nicht flach darüber**: Einrichtung, Fassung und Adresse
              sprächen mit dem zweiten Anbieter von einem von zweien und
              sähen aus, als sprächen sie von beiden. */}
          <ul className="flex flex-col gap-1 text-sm text-leise">
            {katalog.anbieter.map((anbieter) => (
              <li key={anbieter.name} className="flex flex-col">
                <span>
                  {einsetzen(eigene.kiAnbieterZeile, {
                    anbieter: anbieter.name,
                    stand: einrichtungssatz(eigene, anbieter.einrichtung),
                  })}
                </span>

                {/* **Adresse und Fassung in je einer Zeile.** Zusammen in
                    einer wären sie ein Textknoten, der keine von beiden mehr
                    als Aussage lesbar macht — weder für einen Test noch für
                    eine Vorlesehilfe. */}
                <span>
                  {anbieter.name === geraeteanbieter
                    ? eigene.kiAdresseSocket
                    : einsetzen(eigene.kiAdresse, {
                        adresse: anbieter.adresse || eigene.kiAdresseVorgabe,
                      })}
                </span>

                {anbieter.laufzeit !== '' && (
                  <span>
                    {einsetzen(eigene.kiLaufzeitFassung, {
                      fassung: anbieter.laufzeit,
                    })}
                  </span>
                )}

                {anbieter.befunde.map((befund) => (
                  <span key={befund}>{befund}</span>
                ))}

                {/* **Apples Gerätemodell lässt sich für die Sitzung
                    abschalten** (AI-05.AK-04, #2042) — nicht, wenn die
                    Einstellung es ohnehin ausschaltet: dann wäre der Schalter
                    einer, der nichts tut. */}
                {anbieter.name === geraeteanbieter &&
                  (anbieter.sitzungAus || anbieter.einrichtung !== 'aus') && (
                    <KIGeraetSitzung
                      an={!anbieter.sitzungAus}
                      gemeldet={stand.nachladen}
                    />
                  )}

                {/* **Und was sich messen lässt, steht mit einem Knopf da**
                    (A9, #1448). Ohne ihn wären diese Modelle unerreichbar:
                    sie stehen nicht im Katalog, weil über sie nichts
                    feststeht — und feststellen kann es nur eine Messung. */}
                <KIMessung
                  key={anbieter.name}
                  anbieter={anbieter.name}
                  kostet={anbieter.ort !== derOrtLokal}
                  messbar={anbieter.messbar}
                  gemessen={anbieter.gemessen}
                  gemeldet={stand.nachladen}
                />
              </li>
            ))}
          </ul>

          {/* **Der Cloudstand hat drei Werte**, und der dritte ist der
              ehrliche: „keine Serverdatei" heißt ungewiss und nicht „an". */}
          <p className="text-sm text-leise">
            {cloudsatz(eigene, katalog.cloud.stand)}{' '}
            {einsetzen(eigene.kiCloudBeleg, { beleg: katalog.cloud.beleg })}
          </p>

          {katalog.befunde.length > 0 && (
            <ul className="flex flex-col gap-1 text-sm text-leise">
              {katalog.befunde.map((befund) => (
                <li key={befund}>{befund}</li>
              ))}
            </ul>
          )}

          {katalog.modelle.length === 0 ? (
            <p className="text-sm text-leise">{eigene.kimodelleKeine}</p>
          ) : (
            <ul className="flex flex-col gap-2 text-sm">
              {katalog.modelle.map((modell) => (
                <li
                  key={`${modell.anbieter}/${modell.alias}`}
                  className="flex flex-col gap-1"
                >
                  <span className="font-medium">
                    {modell.anbieter}/{modell.alias}
                  </span>

                  <span className="text-leise">
                    {modell.ort} · {modell.laufzeit}
                  </span>

                  {/* **Das Fenster mit seiner Herkunft** (AI-01.AK-02,
                      #1742) — und wo niemand es sagt, steht das da statt
                      „0 Tokens". */}
                  <span className="text-leise">
                    {modell.kontextfenster > 0
                      ? einsetzen(eigene.kiKontextfenster, {
                          tokens: alsZahl(
                            String(modell.kontextfenster),
                            darstellung,
                            0,
                          ),
                          beleg: modell.kontextbeleg,
                        })
                      : eigene.kiKontextfensterUnbekannt}
                  </span>

                  {modell.faehigkeiten.map((beleg) => (
                    <span key={beleg.faehigkeit} className="text-leise">
                      {einsetzen(eigene.kiBeleg, {
                        faehigkeit: beleg.faehigkeit,
                        datum: beleg.datum,
                        beleg: beleg.beleg,
                      })}
                    </span>
                  ))}

                  {/* **Unbekannte Herkunft steht da und fehlt nicht** — eine
                      leere Kette ist eine Auskunft. */}
                  {modell.herkunft.length === 0 && (
                    <span className="text-leise">
                      {eigene.kiHerkunftUnbekannt}
                    </span>
                  )}
                </li>
              ))}
            </ul>
          )}
        </>
      )}
    </section>
  )
}

// geraeteanbieter ist der Name, unter dem Apples Gerätemodell im Katalog steht.
const geraeteanbieter = 'apple'

// Einstellungstexte ist der Katalog dieser Seite.
type Einstellungstexte = ReturnType<typeof useTexte>['seiten']['einstellungen']

// einrichtungssatz gibt den Satz zum Einrichtungsstand.
//
// **Eine Abbildung und kein `if`-Baum**: kommt ein sechster Stand dazu, fällt
// er hier als fehlender Schlüssel auf und nicht als leerer Absatz.
function einrichtungssatz(eigene: Einstellungstexte, stand: string): string {
  const saetze: Record<string, string> = {
    aus: eigene.kiEinrichtungAus,
    unerreichbar: eigene.kiEinrichtungUnerreichbar,
    ohnemodell: eigene.kiEinrichtungOhneModell,
    ohnefaehigkeit: eigene.kiEinrichtungOhneFaehigkeit,
    bereit: eigene.kiEinrichtungBereit,
  }

  return saetze[stand] ?? eigene.kiEinrichtungAus
}

// cloudsatz gibt den Satz zum Cloudstand.
function cloudsatz(eigene: Einstellungstexte, stand: string): string {
  const saetze: Record<string, string> = {
    aus: eigene.kiCloudAus,
    an: eigene.kiCloudAn,
    ungewiss: eigene.kiCloudUngewiss,
  }

  return saetze[stand] ?? eigene.kiCloudUngewiss
}
