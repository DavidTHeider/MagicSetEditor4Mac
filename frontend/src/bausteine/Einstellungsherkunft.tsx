// **Was an einer Einstellung gilt, wessen Entscheidung es war und wann es
// wirkt** — herausgelöst aus `Wahl.tsx` für Befund M15 des Zweitblicks über
// Phase 6.3 (#1665).
//
// # Warum ein eigener Baustein
//
// **Weil nicht jede Einstellung eine Auswahlliste ist.** Die Belegung der
// Ziffern ist eine Tafel, die eigenen Farben ein Textfeld mit Befund — und
// beide zeichneten weder Ebene noch Herkunft noch Wirkzeitpunkt (UX-04.AK-02),
// weil diese Auskunft nur in `Wahl` stand. Hier steht sie einmal, und alle
// drei zeigen dieselben Sätze.
import { useEinstellungsraum } from '../haken/einstellungsraum'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

// Herkunftsangaben ist, was ein Eintrag über sich sagt — **wahlweise**: die
// Stiloptionen, die dieselbe Auswahlliste benutzen, haben keine Herkunft.
export interface Herkunftsangaben {
  // schluessel sagt, welche Einstellung es ist — für die Meldung eines
  // gescheiterten Schreibens an genau ihr.
  schluessel?: string | undefined
  wert: string
  wirksam?: string | undefined
  ebene?: string | undefined
  herkunft?: string | undefined
  wirkung?: string | undefined
}

export function Einstellungsherkunft({
  eintrag,
  beschriftung = (wert) => wert,
  ohneWert = false,
}: {
  eintrag: Herkunftsangaben

  // ohneWert lässt den Wert aus dem Satz — für eine Einstellung, deren Wert
  // im Feld daneben steht: die Belegung als Tafel, die eigenen Farben als
  // Zeilen CSS. In einem Satz läse sich beides wie ein Fehler.
  ohneWert?: boolean

  // beschriftung macht aus einem Wert, was dasteht — die Auswahlliste nennt
  // „dunkel" beim Namen, die Belegung zeigt ihn roh.
  beschriftung?: (wert: string) => string
}) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen
  const fehlschlag = useEinstellungsraum()?.fehlschlag

  // **Der Typ sagt die Wahrheit** — ein unbekanntes Wort trifft keinen
  // Eintrag und gibt `undefined`, statt still leer zu sein.
  const herkunftNamen: Record<string, string | undefined> = eigene.herkunftNamen
  const ebenenNamen: Record<string, string | undefined> = eigene.ebenenNamen

  // **Eine Zeile, und nur, was abweicht** — AP-11
  // (#2368),
  // B-18, Entwurf `Ruhe09`. Bis dahin standen unter jedem Wert „Wirksam ist …
  // — gesetzt von …, Ebene …“ und „Wirkt sofort.“, und im ersten Blick
  // passten vier Einstellungen.
  //
  // - **Was gilt**, wo die Wahl erbt — sonst steht es im Feld. **Und wo eine
  //   Schranke die Wahl hält**: eine Wahl, die nicht erbt, gilt sonst
  //   unverändert (`domain.Verschmolzen`); weicht sie ab, hat „nur lokal“
  //   einen Dienst abgeschaltet (`domain.mitNurLokal`) oder eine strengere
  //   Richtlinie gewonnen (`domain.richtliniengeltung`). Der Schalter steht
  //   dann auf „an“, und ohne den Satz hielte man den Dienst für eingeschaltet.
  // - **Wer es gesetzt hat**, als Marke und nur, wo es nicht die Vorgabe ist;
  //   der ganze Satz mit der Ebene ist ihr Titel und für das Vorleseprogramm.
  // - **Wann es wirkt**, nur beim Neustart: das Sofortige ist die Regel und
  //   sagte das Selbstverständliche (UX-04.AK-02 bekommt hier eine engere
  //   Lesart, eingetragen in `docs/PROTOTYP-0.9.0.md`).
  const wirksamer = eintrag.wirksam ?? eintrag.wert
  const wirksam = ohneWert
    ? undefined
    : eintrag.wert !== erben
      ? wirksamer === eintrag.wert
        ? undefined
        : einsetzen(eigene.wirksamGehalten, { wert: beschriftung(wirksamer) })
      : wirksamer === ''
        ? eigene.wirksamOffenKurz
        : einsetzen(eigene.wirksamKurz, { wert: beschriftung(wirksamer) })
  const satz =
    eintrag.herkunft === undefined ||
    eintrag.herkunft === 'vorgabe' ||
    eintrag.ebene === undefined
      ? undefined
      : {
          marke: einsetzen(eigene.herkunftMarke, {
            herkunft: herkunftNamen[eintrag.herkunft] ?? eintrag.herkunft,
          }),
          ganz: einsetzen(eigene.herkunftZeile, {
            herkunft: herkunftNamen[eintrag.herkunft] ?? eintrag.herkunft,
            ebene: ebenenNamen[eintrag.ebene] ?? eintrag.ebene,
          }),
        }
  const neustart = eintrag.wirkung === 'neustart'
  const gescheitert =
    fehlschlag !== undefined &&
    eintrag.schluessel !== undefined &&
    fehlschlag.schluessel === eintrag.schluessel

  return (
    <>
      {(wirksam !== undefined || satz !== undefined || neustart) && (
        <p className="mt-1 flex max-w-prose flex-wrap items-center gap-x-2 gap-y-1 text-sm text-leise">
          {wirksam !== undefined && <span>{wirksam}</span>}
          {satz !== undefined && (
            <>
              <span
                data-herkunftsmarke
                aria-hidden="true"
                title={satz.ganz}
                className="rounded-full border border-kante px-2 text-xs"
              >
                {satz.marke}
              </span>
              <span className="sr-only">{satz.ganz}</span>
            </>
          )}
          {neustart && <span>{eigene.wirktNachNeustart}</span>}
        </p>
      )}
      {/* **Die Meldung an der Einstellung, die sie verursacht hat** (§6) —
          ein abgewiesener Ordner sagt hier, dass das Laufwerk fehlt. */}
      {gescheitert && (
        <p role="alert" className="mt-1 max-w-prose text-sm text-fehler">
          {meldungAus(fehlschlag.ursache, texte)}
        </p>
      )}
    </>
  )
}

// erben ist `domain.Erben` — eine Wahl und kein Wert, der gilt.
const erben = 'erben'
