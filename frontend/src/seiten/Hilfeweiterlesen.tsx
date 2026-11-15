// **Weiterlesen** — das Gebiet der Hilfe, das auf Handbuch und
// Administration zeigt. Aus `Hilfe.tsx` herausgelöst, als G02 es über 400
// Zeilen trug (§2) — wie Datenschutz, Betrieb und Schnittstellen.
import { Flaeche } from '../bausteine/Flaeche'
import { Satz } from '../bausteine/Satz'
import { useTexte } from '../sprache/kontext'

export function Hilfeweiterlesen() {
  const hilfe = useTexte().hilfe

  return (
    <Flaeche
      ueberschrift={hilfe.weiterlesen.ueberschrift}
      sprungmarke="hilfe-weiterlesen"
    >
      <p className="mb-4 max-w-prose text-leise">
        <Satz text={hilfe.weiterlesen.einleitung} />
      </p>

      <dl className="grid grid-cols-[max-content_1fr] gap-x-6 gap-y-3">
        <div className="contents">
          <dt className="font-semibold">{hilfe.weiterlesen.handbuchName}</dt>
          <dd className="max-w-prose text-leise">
            <Satz text={hilfe.weiterlesen.handbuchSatz} />
          </dd>
        </div>

        <div className="contents">
          <dt className="font-semibold">
            {hilfe.weiterlesen.administrationName}
          </dt>
          <dd className="max-w-prose text-leise">
            <Satz text={hilfe.weiterlesen.administrationSatz} />
          </dd>
        </div>
      </dl>
    </Flaeche>
  )
}
