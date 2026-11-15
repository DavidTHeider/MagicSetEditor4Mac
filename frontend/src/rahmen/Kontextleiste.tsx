// **Die Kontextleiste: wer wo als was arbeitet** — Werkstück U4 von Phase
// 5.1 (#1415).
//
// **Sichtbar steht sie in der Leiste links**; liegt die Leiste oben
// (UX-14.AK-02), stehen Mensch und Profil nur für den Vorleser da, und
// sichtbar bleibt „nur lesen" — hier stand „jederzeit sichtbar", bis der
// Zweitblick über Phase 6.3 es anstrich (G19).
//
// # Warum in der Seitenleiste und nicht als Band über dem Inhalt
//
// **Weil sie für alles gilt und niemanden unterbrechen soll.** Ein Band über
// dem Inhalt ist für das, was sich ändert (eine Meldung, eine neue Fassung);
// wer arbeitet, ändert sich nur beim Wechsel — und dann baut die Anwendung
// die Leiste neu auf. Sie steht unter dem Namen der Anwendung, wo macOS den
// angemeldeten Menschen sucht.
//
// # Was sie sagt, und wie
//
//   - **Den Menschen mit Bild**, das Profil, und **wessen** Profil es ist,
//     wenn es nicht das eigene ist — Person A im Profil von B bleibt A.
//   - **Die Rolle als Wort** und „nur lesen" **als Text**, nicht als Farbe:
//     ein Vorleseprogramm liest Text, keinen Farbton (§6).
//   - **Einen Weg zu Nutzern und Profilen** — dort liegen Profilwechsel,
//     Kontozugang und Maschinenzugänge; die Leiste wiederholt sie nicht.
//     **„Profil wechseln …" und nicht „Nutzer und Profile"**: so heißt der
//     Eintrag der Seitenleiste, und zwei Knöpfe mit demselben Namen wären
//     für ein Vorleseprogramm einer.
//
// `role="region"` mit Namen, damit die Leiste als Orientierungspunkt
// anspringbar ist; der Zustand darin ist `role="status"`, damit ein Wechsel
// angesagt wird, ohne zu unterbrechen.
import { useId } from 'react'

import { Hinweiskasten } from '../bausteine/Hinweiskasten'
import { Nutzerbild } from '../bausteine/Nutzerbild'
import { useArbeitskontext } from '../haken/useArbeitskontext'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Seitenschluessel } from '../navigation'

interface KontextleisteEigenschaften {
  wechseln: (seite: Seitenschluessel) => void

  // waagerecht legt den Block als Zeile, für die Leiste oben (UX-14.AK-02).
  waagerecht?: boolean
}

const knopf =
  'w-full rounded-md px-3 py-1.5 text-left text-sm text-leise hover:bg-betont ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'

const zeilenknopf =
  'flex w-full min-w-0 items-center gap-2 rounded-md px-2 py-1 text-left ' +
  'hover:bg-betont focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

// **Oben steht der Block als Zeile** — UX-14.AK-02
// (#1686):
// dieselben Angaben, nebeneinander, und die Kante rechts statt unten.
const blockQuer = 'flex items-center gap-3 border-r border-kante px-3 py-1'
// **Oben stehen die Angaben für den Vorleser**: sichtbar bleiben Bild und
// Knopf — die Zeile braucht ihre Breite für die Bereiche.
const angabenQuer = 'sr-only'

export function Kontextleiste({
  wechseln,
  waagerecht = false,
}: KontextleisteEigenschaften) {
  const texte = useTexte()
  const eigene = texte.bausteine.kontext
  const stand = useArbeitskontext()
  const profilVon = einsetzen(eigene.profilVon, { eigner: stand.eignername })
  const beschreibung = useId()

  // rolle ist das Wort zur Rolle — eine unbekannte bleibt, wie sie kam: sie
  // stammt aus dem Bestand, und ein leeres Wort verschwiege sie.
  const rolle = (name: string): string => {
    const woerter: Record<string, string | undefined> = eigene.rollen

    return woerter[name] ?? name
  }

  const eignerAnders = stand.kontext?.eignerID !== stand.kontext?.nutzerID

  if (waagerecht) {
    return (
      <section aria-label={eigene.bezeichnung} className={blockQuer}>
        {stand.fehler !== undefined && (
          <Hinweiskasten art="fehler" meldung={stand.fehler} />
        )}

        {stand.laedt && stand.kontext === undefined && (
          <p role="status" className="text-sm text-leise">
            {texte.bausteine.zustaende.laedt}
          </p>
        )}

        {stand.kontext !== undefined && (
          <div role="status" className="flex items-center gap-3">
            <Nutzerbild
              name={stand.nutzername}
              datenURI={stand.bild}
              badgeURI={stand.badge}
            />

            <div className={angabenQuer}>
              <span>{stand.nutzername}</span>
              <span>{stand.profilname}</span>
              {eignerAnders && <span>{profilVon}</span>}
              <span>
                {rolle(stand.kontext.rolle)}
                {stand.kontext.nurLesend && ` · ${eigene.nurLesen}`}
              </span>
            </div>

            {/* **Oben steht „nur lesen" für sich** (UX-14.AK-02): die
                Angaben davor sind für den Vorleser da, und dieses Zeichen
                ändert, was man tun kann. Angesagt wird es dort schon einmal. */}
            {stand.kontext.nurLesend && (
              <span aria-hidden="true" className="text-sm text-fehler">
                {eigene.nurLesen}
              </span>
            )}
          </div>
        )}

        <button
          type="button"
          className={knopf}
          onClick={() => {
            wechseln('nutzer')
          }}
        >
          {eigene.zuDenNutzern}
        </button>
      </section>
    )
  }

  // **Links eine Zeile, und sie ist der Knopf** — AP-11
  // (#2368),
  // B-31: „Bild + Name, Rolle im Tooltip“, wie `Ruhe09` die Leiste zeichnet.
  // Drei Zeilen und ein Knopf darunter nahmen die halbe Höhe der Leiste; ein
  // Knopf **neben** der Zeile ließ in 232 Punkten von „Bo · Mein Profil“ nur
  // „Bo · M…“ übrig. Jetzt hat der Name die Breite, und der Weg zu den
  // Nutzern ist die Zeile selbst — beschrieben als „Profil wechseln …“.
  //
  // **Der Eigner eines fremden Profils und „nur lesen“ bleiben zu sehen**,
  // klein darunter: das eine sagt, wessen Arbeit man ändert, das andere, ob
  // man es kann. Die Rolle steht im Titel und für den Vorleser.
  const zeile = einsetzen(eigene.zeile, {
    name: stand.nutzername,
    profil: stand.profilname,
  })

  return (
    <section
      aria-label={eigene.bezeichnung}
      className="flex flex-col gap-2 border-b border-kante px-2 py-1"
    >
      {stand.fehler !== undefined && (
        <Hinweiskasten art="fehler" meldung={stand.fehler} />
      )}

      {stand.laedt && stand.kontext === undefined && (
        <p role="status" className="px-2 text-sm text-leise">
          {texte.bausteine.zustaende.laedt}
        </p>
      )}

      {stand.kontext === undefined ? (
        // **Ohne Kontext bleibt der Weg**, mit ausgeschriebenem Ziel.
        <button
          type="button"
          className={knopf}
          onClick={() => {
            wechseln('nutzer')
          }}
        >
          {eigene.zuDenNutzern}
        </button>
      ) : (
        <div role="status">
          <button
            type="button"
            aria-describedby={beschreibung}
            title={`${zeile} — ${rolle(stand.kontext.rolle)}`}
            className={zeilenknopf}
            onClick={() => {
              wechseln('nutzer')
            }}
          >
            {/* **Das Bild spricht nicht mit**: der Name steht im Text, und
                „Profilbild von …“ verlängerte nur den Namen des Knopfes. */}
            <span aria-hidden="true" className="inline-flex shrink-0">
              <Nutzerbild
                name={stand.nutzername}
                datenURI={stand.bild}
                badgeURI={stand.badge}
              />
            </span>

            <span className="flex min-w-0 flex-1 flex-col text-sm">
              {/* **Gekürzt wird die Zeichnung, nicht der Name** — H03
                  (#2257): er bleibt im Titel lesbar. */}
              <span className="truncate font-medium" title={zeile}>
                {zeile}
              </span>
              <span className="sr-only">{rolle(stand.kontext.rolle)}</span>
              {(eignerAnders || stand.kontext.nurLesend) && (
                <span className="flex min-w-0 gap-2 text-xs">
                  {eignerAnders && (
                    <span className="truncate text-leise" title={profilVon}>
                      {profilVon}
                    </span>
                  )}
                  {stand.kontext.nurLesend && (
                    <span className="shrink-0 text-fehler">
                      {eigene.nurLesen}
                    </span>
                  )}
                </span>
              )}
            </span>

            <span aria-hidden="true" className="shrink-0 text-leise">
              ›
            </span>
          </button>
          <span id={beschreibung} className="sr-only">
            {eigene.zuDenNutzern}
          </span>
        </div>
      )}
    </section>
  )
}
