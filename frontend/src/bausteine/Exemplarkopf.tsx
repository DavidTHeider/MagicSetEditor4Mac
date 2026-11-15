// **Der Kopf des Inspectors der Sammlung** — was das gewählte Exemplar ist:
// Bezeichnung, Druck, Menge und Zustand, Sprache und Ausführung, Lagerort.
// Aus `Sammlungsinspector.tsx` herausgelöst, als er mit den Handlungen für
// Leiste und Rechtsklick über die Zeilengrenze wuchs (AP-6, Teil 2, §2).
import type { ExemplarDTO } from '../api'
import { useTexte } from '../sprache/kontext'
import { Vorlaeufig } from './Vorlaeufig'
import { zustaende } from './Zustandswahl'

export function Exemplarkopf({ exemplar }: { exemplar: ExemplarDTO }) {
  const eigene = useTexte().seiten.sammlung

  return (
    <header className="flex flex-col gap-1">
      <h3 className="flex items-center gap-2 font-medium">
        {exemplar.bezeichnung}
        <Vorlaeufig vorlaeufig={exemplar.vorlaeufig} />
      </h3>
      {exemplar.name !== '' && (
        <span className="font-mono text-xs text-leise">
          {exemplar.printing}
        </span>
      )}
      <span className="text-sm">
        {exemplar.menge} × {eigene.zustaende[zustandsname(exemplar.zustand)]}
      </span>
      <span className="text-sm text-leise">
        {exemplar.sprache} · {eigene[exemplar.finish]}
      </span>
      {exemplar.lagerort !== '' && (
        <span className="text-sm text-leise">{exemplar.lagerort}</span>
      )}
    </header>
  )
}

// zustandsname nimmt einen bekannten Zustand und sonst den Rückfall — über
// dieselbe Liste wie die Wahl im Formular.
function zustandsname(zustand: string): (typeof zustaende)[number] {
  return zustaende.find((bekannt) => bekannt === zustand) ?? 'nearMint'
}
