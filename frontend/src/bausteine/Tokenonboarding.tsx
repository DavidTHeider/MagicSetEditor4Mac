// **So richtet man einen MCP-Client ein** — Werkstück O2 von Phase 5.2
// (#1426).
//
// **Ein Aufklappkasten und keine eigene Seite**: die Anleitung gehört zu den
// Token, denn mit einem beginnt sie. Vier Schritte und ein Beispiel für die
// Konfiguration eines Clients — das Beispiel ist Code und steht deshalb in
// beiden Sprachen gleich.
import { Aufklappkasten } from './Aufklappkasten'
import { useTexte } from '../sprache/kontext'

export function Tokenonboarding() {
  const eigene = useTexte().seiten.nutzer

  const schritte = [
    eigene.tokenOnboardingSchritt1,
    eigene.tokenOnboardingSchritt2,
    eigene.tokenOnboardingSchritt3,
    eigene.tokenOnboardingSchritt4,
  ]

  return (
    <Aufklappkasten titel={eigene.tokenOnboarding}>
      <ol className="flex max-w-prose list-decimal flex-col gap-1 pl-5 text-leise">
        {schritte.map((schritt) => (
          <li key={schritt}>{schritt}</li>
        ))}
      </ol>

      {/* **Eine Region mit Namen und Tabhalt, weil sie rollt** — das Muster
          aus `Diagnose.tsx` (Nachlauf über Phase 6.3, N10). */}
      <pre
        role="region"
        aria-label={eigene.tokenOnboardingBeispielTitel}
        tabIndex={0}
        className="mt-2 overflow-x-auto rounded-sm bg-flaeche px-3 py-2 font-mono text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      >
        <code>{eigene.tokenOnboardingBeispiel}</code>
      </pre>
    </Aufklappkasten>
  )
}
