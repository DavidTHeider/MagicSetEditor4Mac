// **Der Abschnitt „Über"** — aus `Hilfe.tsx` gelöst, als die Seite mit
// UXR-22 (#1930)
// an die Zeilengrenze kam (§2). Fassung, Lizenz, Rechte, Quelle (#1168).
import { Flaeche } from '../bausteine/Flaeche'
import { Satz } from '../bausteine/Satz'
import { useTexte } from '../sprache/kontext'

export function Hilfeueber({ fassung }: { fassung: string }) {
  const { hilfe } = useTexte()

  return (
    <Flaeche ueberschrift={hilfe.ueber.ueberschrift} sprungmarke="hilfe-ueber">
      <dl className="grid grid-cols-[max-content_1fr] gap-x-6 gap-y-3">
        <div className="contents">
          <dt className="font-semibold">{hilfe.ueber.fassung}</dt>
          <dd className="text-leise tabular-nums">{fassung}</dd>
        </div>

        <div className="contents">
          <dt className="font-semibold">{hilfe.ueber.lizenzName}</dt>
          <dd className="max-w-prose text-leise">
            <Satz text={hilfe.ueber.lizenzSatz} />
          </dd>
        </div>

        <div className="contents">
          <dt className="font-semibold">{hilfe.ueber.rechteName}</dt>
          <dd className="max-w-prose text-leise">
            <Satz text={hilfe.ueber.rechteSatz} />
          </dd>
        </div>

        <div className="contents">
          <dt className="font-semibold">{hilfe.ueber.quelleName}</dt>
          <dd className="max-w-prose text-leise">
            <a
              href="https://github.com/DavidTHeider/MagicSetEditor4Mac"
              target="_blank"
              rel="noreferrer"
              className="underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            >
              {hilfe.ueber.quelleSatz}
            </a>
          </dd>
        </div>
      </dl>
    </Flaeche>
  )
}
