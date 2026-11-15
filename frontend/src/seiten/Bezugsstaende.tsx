// **Die bezogenen Stände** — G04
// (#2251).
//
// Was sie tun, steht in `haken/useBezugsstaende.ts`; hier, wie sie aussehen
// (§2): je Baum der bezogene, der geprüfte und — wenn gefragt — der neueste
// Stand seiner Quelle, oder warum sie sich nicht fragen ließ. **Drei Stände
// und ein Fehler, und keiner sieht aus wie der andere.**

import { Bindungsfehler } from '../api'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Knopf } from '../bausteine/Knopf'
import { Leerzustand } from '../bausteine/Leerzustand'
import { useBezugsstaende } from '../haken/useBezugsstaende'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useSprache, useTexte } from '../sprache/kontext'
import { alsKalendertag } from '../zeitpunkt'
import { Bezugsvorschau } from './Bezugsvorschau'

import type { Katalog } from '../sprache/kataloge'
import type { PaketbezugDTO } from '../paketbezuege'

interface Eigenschaften {
  // geaendert lädt die Paketliste nach einer Aktualisierung nach.
  geaendert: () => void

  // beziehen öffnet die Schublade „Paket beziehen …" — der Weg aus dem
  // Leerzustand.
  beziehen: () => void
}

const kurz = (stand: string) => stand.slice(0, 7)

// lagesatz sagt, wie ein Stand zu seiner Quelle steht.
function lagesatz(
  bezug: PaketbezugDTO,
  texte: Katalog,
  sprache: string,
): string {
  const t = texte.seiten.pakete.bezuege.lage

  switch (bezug.lage) {
    case 'neuer':
      return einsetzen(bezug.vorsprung === 1 ? t.neuerEins : t.neuer, {
        spitze: kurz(bezug.spitze),
        datum: alsKalendertag(bezug.spitzeZeit.slice(0, 10), sprache),
        vorsprung: String(bezug.vorsprung),
      })
    case 'abweichend':
      return einsetzen(t.abweichend, { spitze: kurz(bezug.spitze) })
    case 'unbekannt':
      return einsetzen(t.unbekannt, {
        grund:
          bezug.fehler === null
            ? ''
            : meldungAus(new Bindungsfehler(bezug.fehler, []), texte),
      })
    default:
      return t[bezug.lage]
  }
}

// paketzahl zählt die Pakete eines Baums — **keines ist ein eigener Satz**:
// der erste Commit einer Quelle kann nur ihre README tragen (g04bezuege).
function paketzahl(
  zahl: number,
  t: Katalog['seiten']['pakete']['bezuege'],
): string {
  if (zahl === 0) {
    return t.keinPaket
  }

  return zahl === 1 ? t.paketEins : einsetzen(t.pakete, { zahl: String(zahl) })
}

export function Bezugsstaende({ geaendert, beziehen }: Eigenschaften) {
  const texte = useTexte()
  const sprache = useSprache()
  const t = texte.seiten.pakete.bezuege
  const staende = useBezugsstaende(geaendert)
  // **Als Konstante**, damit die Prüfung unten auch im Rückruf gilt.
  const offen = staende.vorschau

  return (
    <div className="flex flex-col gap-3">
      <Erklaerungsabsatz text={t.erklaerung} klasse="max-w-prose text-leise" />

      <div>
        <Knopf angehalten={staende.laedt} aufKlick={staende.fragen}>
          {t.fragen}
        </Knopf>
      </div>

      {staende.fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {staende.fehler}
        </p>
      )}

      {staende.bezuege?.length === 0 && (
        <Leerzustand titel={t.ueberschrift} erklaerung={t.leer}>
          <Knopf aufKlick={beziehen}>{texte.seiten.pakete.plan.beziehen}</Knopf>
        </Leerzustand>
      )}

      <ul className="flex flex-col gap-3">
        {(staende.bezuege ?? []).map((bezug) => (
          <li
            key={bezug.name}
            className="rounded-md border border-kante p-3"
            aria-busy={staende.laeuft(bezug.name)}
          >
            <p className="font-medium">{bezug.name}</p>
            <p className="text-sm text-leise">{bezug.adresse}</p>
            <p className="mt-1 text-sm">
              {einsetzen(t.bezogen, { stand: kurz(bezug.stand) })} ·{' '}
              {bezug.geprueft === ''
                ? t.ohnePruefung
                : einsetzen(t.geprueft, { stand: kurz(bezug.geprueft) })}{' '}
              · {paketzahl(bezug.pakete, t)}
            </p>
            <p
              className="mt-1 text-sm"
              role={bezug.lage === 'unbekannt' ? 'alert' : undefined}
            >
              {lagesatz(bezug, texte, sprache)}
            </p>

            <div className="mt-2 flex flex-wrap gap-2">
              {(bezug.lage === 'neuer' || bezug.lage === 'abweichend') && (
                <Knopf
                  angehalten={staende.laeuft(bezug.name)}
                  bezeichnung={einsetzen(t.ansehenBezeichnung, {
                    name: bezug.name,
                  })}
                  aufKlick={() => {
                    void staende.ansehen(bezug)
                  }}
                >
                  {t.ansehen}
                </Knopf>
              )}
              {bezug.vorige && (
                <Knopf
                  angehalten={staende.laeuft(bezug.name)}
                  bezeichnung={einsetzen(t.zurueckBezeichnung, {
                    name: bezug.name,
                    stand: kurz(bezug.vorigerStand),
                  })}
                  aufKlick={() => {
                    void staende.zurueck(bezug)
                  }}
                >
                  {t.zurueck}
                </Knopf>
              )}
            </div>

            {staende.vorschauFuer === bezug.name &&
              staende.vorschauFehler !== undefined && (
                <p role="alert" className="mt-2 text-fehler">
                  {staende.vorschauFehler}
                </p>
              )}

            {staende.vorschauFuer === bezug.name && offen !== undefined && (
              <Bezugsvorschau
                vorschau={offen}
                vorigerStand={bezug.vorigerStand}
                laeuft={staende.laeuft(bezug.name)}
                aktualisieren={(bestaetigt) => {
                  void staende.aktualisieren(offen, bestaetigt)
                }}
                verwerfen={staende.verwerfen}
              />
            )}
          </li>
        ))}
      </ul>
    </div>
  )
}
