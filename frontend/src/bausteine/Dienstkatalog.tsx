// **Was die fremden Dienste hier können — und ob sie es jetzt dürfen** —
// Werkstück A1 von Phase 5.3
// (#1440),
// entschieden in [0082](../../../docs/entscheidungen/0082-der-dienstkatalog.md).
//
// # Was hier steht
//
// Je Dienst die Fragen, für die es einen gebauten Weg gibt, mit ihrem
// Status aus `domain.Dienstlage` — und der Vertrag dazu, aufklappbar: wo
// der Dienst antwortet, wie sich die Anwendung ausweist, was es kostet, wie
// viel erlaubt ist, unter welchen Bedingungen und wann das nachgelesen wurde.
//
// # Der Nachweis ist kein Gütesiegel
//
// Bis zu diesem Werkstück führte `Dienstschalter.tsx` eine feste Liste
// „geprüfter" Dienste. Jetzt sagt die Go-Seite je Fähigkeit, **wann** und
// **mit welcher Fassung** ein Lauf gegen das Original ging — und ob das diese
// Fassung ist. Ein Lauf aus 0.3.0 steht als Lauf aus 0.3.0 da.
//
// # Lädt selbst, wie `Zugaenge`
//
// Über `useLaden` an einem Stand aus Schaltern und Wahlen: legt jemand einen
// Schalter um, lädt der Katalog nach (§7, Regel 2 über den Haken).
import { api } from '../api'
import { Aufklappkasten } from './Aufklappkasten'
import { auspacken } from '../umschlag'
import { dienstnamen } from './dienstnamen'
import { Satz } from './Satz'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import type { DienstfaehigkeitDTO, NachweisDTO } from '../dienstkatalog'

async function holen(): Promise<DienstfaehigkeitDTO[]> {
  const antwort = await api().Dienstkatalog()

  return auspacken(antwort, () => antwort.eintraege)
}

// nachDienst ordnet die Einträge ihrem Dienst zu — in Katalogreihenfolge.
function nachDienst(
  eintraege: DienstfaehigkeitDTO[],
): [string, DienstfaehigkeitDTO[]][] {
  const gruppen = new Map<string, DienstfaehigkeitDTO[]>()

  for (const eintrag of eintraege) {
    gruppen.set(eintrag.dienst, [
      ...(gruppen.get(eintrag.dienst) ?? []),
      eintrag,
    ])
  }

  return [...gruppen.entries()]
}

const statusklassen: Record<string, string> = {
  bereit: 'text-erfolg',
  aus: 'text-leise',
  nichtGewaehlt: 'text-leise',
}

export function Dienstkatalog({ stand }: { stand: string }) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen
  const katalog = useLaden(holen, [], false, stand)

  if (katalog.fehler !== undefined) {
    return (
      <p role="alert" className="mt-4 text-fehler">
        {texte.bausteine.zustaende.fehler} {katalog.fehler}
      </p>
    )
  }

  const eintraege = katalog.daten ?? []
  if (eintraege.length === 0) {
    return null
  }

  return (
    // **Ohne eigene Überschrift seit #2465**: der Katalog steht in einem
    // Abschnitt der Einstellungen, und dessen Knopf trägt den Namen.
    <section className="flex flex-col gap-3">
      <p className="max-w-prose text-sm text-leise">
        <Satz text={eigene.katalogHinweis} />
      </p>
      <div className="flex max-w-2xl flex-col gap-6">
        {nachDienst(eintraege).map(([dienst, faehigkeiten]) => (
          <Dienstblock
            key={dienst}
            dienst={dienst}
            faehigkeiten={faehigkeiten}
          />
        ))}
      </div>
    </section>
  )
}

function Dienstblock({
  dienst,
  faehigkeiten,
}: {
  dienst: string
  faehigkeiten: DienstfaehigkeitDTO[]
}) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen
  const name = dienstnamen[dienst] ?? dienst
  const lauf = faehigkeiten
    .map((eintrag) => eintrag.nachweis)
    .find((nachweis): nachweis is NachweisDTO => nachweis !== null)

  return (
    <article aria-label={name} className="flex flex-col gap-2">
      <h3 className="font-semibold">{name}</h3>
      <Nachweiszeile nachweis={lauf ?? null} />
      <ul className="flex flex-col gap-2">
        {faehigkeiten.map((eintrag) => (
          <Faehigkeit key={eintrag.frage} eintrag={eintrag} mitLauf={!!lauf} />
        ))}
      </ul>
      {/* eigene ist hier nur für den Hinweis der Ungeprüften nötig */}
      {lauf === undefined && (
        <Erklaerungsabsatz
          text={eigene.dienstUngeprueftHinweis}
          klasse="max-w-prose text-sm text-leise"
        />
      )}
      {/* **Der Weg zum Lauf steht unter „Diagnose"** — UX-08 S11.NAT-02
          (#1680):
          ein Pfad in den Quelltext ist keine Entscheidung, die hier jemand
          trifft, und er stand lang zwischen den Schaltern. **Nur, wenn es
          etwas zu wiederholen gibt** — ein Lauf mit dieser Fassung braucht
          keinen. */}
      {lauf !== undefined && !lauf.giltFuerDieseFassung && (
        <Aufklappkasten titel={eigene.katalogDiagnose}>
          <p className="text-sm">
            {einsetzen(eigene.katalogLauf, { lauf: lauf.lauf })}
          </p>
        </Aufklappkasten>
      )}
    </article>
  )
}

function Nachweiszeile({ nachweis }: { nachweis: NachweisDTO | null }) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  if (nachweis === null) {
    return (
      <p className="text-sm text-hinweis">
        <strong>{eigene.dienstUngeprueft}.</strong>
      </p>
    )
  }

  return (
    <p className="text-sm">
      {einsetzen(eigene.katalogNachweis, {
        datum: nachweis.datum,
        fassung: nachweis.fassung,
      })}{' '}
      {nachweis.giltFuerDieseFassung ? (
        <span>{eigene.katalogNachweisAktuell}</span>
      ) : (
        <span className="text-hinweis">{eigene.katalogNachweisFremd}</span>
      )}
    </p>
  )
}

function Faehigkeit({
  eintrag,
  mitLauf,
}: {
  eintrag: DienstfaehigkeitDTO
  mitLauf: boolean
}) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen
  const fragen = eigene.katalogFragen as Record<string, string | undefined>
  const status = eigene.katalogStatus as Record<string, string | undefined>
  const angaben = eigene.angaben as Record<string, string | undefined>

  const grund =
    eintrag.status === 'zugangFehlt'
      ? einsetzen(eigene.katalogAngabeFehlt, {
          angabe: angaben[eintrag.angabe] ?? eintrag.angabe,
        })
      : eintrag.status === 'gestoert' || eintrag.status === 'gesperrt'
        ? eintrag.grund
        : ''

  return (
    <li className="flex flex-col gap-1 rounded-md border border-kante px-3 py-2">
      <div className="flex flex-wrap items-baseline gap-x-3 gap-y-1">
        <span className="font-medium">
          {fragen[eintrag.frage] ?? eintrag.frage}
        </span>
        <span
          className={['text-sm', statusklassen[eintrag.status] ?? ''].join(' ')}
        >
          {status[eintrag.status] ?? eintrag.status}
        </span>
        {grund !== '' && <span className="text-sm text-leise">{grund}</span>}
        {mitLauf && (
          <span className="text-sm text-leise">
            {eintrag.nachweis === null
              ? eigene.katalogNichtImLauf
              : eigene.katalogImLauf}
          </span>
        )}
      </div>
      <Vertrag eintrag={eintrag} />
    </li>
  )
}

function Vertrag({ eintrag }: { eintrag: DienstfaehigkeitDTO }) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen
  const felder = eigene.katalogFelder
  const wirkungen = eigene.katalogWirkung as Record<string, string | undefined>
  const vertrag = eintrag.vertrag

  return (
    <details className="text-sm">
      <summary className="cursor-pointer text-leise">
        {eigene.katalogVertrag}
      </summary>
      <dl className="mt-2 grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1">
        <dt className="text-leise">{felder.basisadresse}</dt>
        <dd>{vertrag.basisadresse}</dd>
        <dt className="text-leise">{felder.endpunkte}</dt>
        <dd>{vertrag.endpunkte.join(', ')}</dd>
        <dt className="text-leise">{felder.zugang}</dt>
        <dd>{vertrag.zugang}</dd>
        <dt className="text-leise">{felder.tarif}</dt>
        <dd>{vertrag.tarif}</dd>
        <dt className="text-leise">{felder.grenze}</dt>
        <dd>{vertrag.grenze}</dd>
        <dt className="text-leise">{felder.frist}</dt>
        <dd>{vertrag.frist}</dd>
        <dt className="text-leise">{felder.wirkung}</dt>
        <dd>{wirkungen[eintrag.wirkung] ?? eintrag.wirkung}</dd>
        <dt className="text-leise">{felder.bedingungen}</dt>
        <dd>
          <a
            href={vertrag.bedingungen}
            target="_blank"
            rel="noreferrer"
            className="underline"
          >
            {vertrag.bedingungen}
          </a>
        </dd>
        <dt className="text-leise">{felder.abgerufen}</dt>
        <dd>{vertrag.abgerufen}</dd>
        <dt className="text-leise">{felder.beleg}</dt>
        <dd>
          <a
            href={vertrag.beleg}
            target="_blank"
            rel="noreferrer"
            className="underline"
          >
            {vertrag.beleg}
          </a>
        </dd>
      </dl>
    </details>
  )
}
