// **Was aus der Umgebung kommt** — Werkstück N5 von Phase 5.4
// (#1571),
// eingedampft für UX-15.AK-04 und AK-05
// (#1771).
//
// # Warum die sechs Werte hier nicht stehen
//
// **Sie stehen darunter als Einstellungen** — drei Proxyzeilen und die drei
// Fristen. Der Kasten steht in der Gruppe *Verbindungen*; zeigte er sie mit
// ihrer Herkunft noch einmal, stünden sie einmal als Auskunft und einmal zum
// Ändern da. Das ist der Fall, für den es
// #1771
// gibt — „das führt zu einer unnötig aufgeblähten Seite, die die nötigen
// Informationen nicht kompakt darstellt".
//
// # Was bleibt
//
// **Der eine Satz, den kein Feld sagen kann**: dass ein Wert aus der
// **Umgebung dieses Prozesses** kommt — `HTTPS_PROXY` und seine Geschwister
// — und das Feld darunter übersteuert. Wer `proxy.firma:3128` liest und das
// Feld leer findet, hielte es sonst für einen Fehler.
//
// **Und kommt keiner von dort, steht hier gar nichts.** Ein Kasten, der
// „nichts Besonderes" meldet, ist Platz für nichts (§6).
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { NetzstandAntwort, NetzfeldDTO } from '../apinetz'

export function Netzweg() {
  const texte = useTexte()
  const eigene = texte.seiten.technik

  const stand = useLaden<NetzstandAntwort>(async () => {
    const antwort = await api().Netzstand()

    return auspacken(antwort, () => antwort)
  }, [])

  const ausDerUmgebung =
    stand.daten === undefined ? [] : umgebungswerte(stand.daten, eigene)

  // **Ein Fehler steht da, auch wenn sonst nichts steht** — er ist eine
  // Auskunft, die kein Feld hat (§8).
  if (stand.fehler !== undefined) {
    return (
      <p role="alert" className="text-fehler">
        {texte.bausteine.zustaende.fehler} {stand.fehler}
      </p>
    )
  }

  if (ausDerUmgebung.length === 0) {
    return undefined
  }

  return (
    <section className="rounded-lg border border-kante bg-flaeche p-4">
      <h3 className="text-lg font-semibold">{eigene.netzwegTitel}</h3>

      <Erklaerungsabsatz
        text={eigene.netzwegErklaerung}
        klasse="mt-1 max-w-prose text-leise"
      />

      <ul className="mt-3 flex flex-col gap-1 text-sm">
        {ausDerUmgebung.map((eine) => (
          <li key={eine} className="break-all font-mono">
            {eine}
          </li>
        ))}
      </ul>
    </section>
  )
}

// umgebungswerte nennt die Werte, die aus der Umgebung kommen.
//
// **Nur die aus der Umgebung**: ein Wert aus den Einstellungen steht im Feld
// darunter, und zweimal wäre einmal zu viel (AK-04).
function umgebungswerte(
  daten: NetzstandAntwort,
  eigene: {
    netzwegProxyHTTP: string
    netzwegProxyHTTPS: string
    netzwegProxyAus: string
    netzwegAusUmgebung: string
  },
): string[] {
  const felder: { name: string; feld: NetzfeldDTO }[] = [
    { name: eigene.netzwegProxyHTTP, feld: daten.proxyHTTP },
    { name: eigene.netzwegProxyHTTPS, feld: daten.proxyHTTPS },
    { name: eigene.netzwegProxyAus, feld: daten.proxyAus },
  ]

  return felder
    .filter((eines) => eines.feld.herkunft === 'umgebung')
    .map((eines) =>
      einsetzen(eigene.netzwegAusUmgebung, {
        name: eines.name,
        wert: eines.feld.wert,
      }),
    )
}
