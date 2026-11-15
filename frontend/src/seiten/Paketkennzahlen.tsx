// **Die Kennzahlen der Pakete** — AP-9, Teil 2
// (#2366),
// Befund B-15, Entwurf `Bereiche09`: installiert und mit Befund.
//
// # Die Ziele
//
// **Installiert** nimmt Suche, Art- und Herkunftsfilter samt einer Auswahl
// zurück — über denselben Weg, auf dem ein fehlendes Paket gesucht wird
// (`paketsuche`, mit leerer Suche); danach steht jedes Paket da. **Mit
// Befund** filtert die Liste auf genau die Pakete mit einem
// Abhängigkeitsbefund (`paketauswahl`) — dieselben, die an ihrer Zeile die
// Zahl der Vorbehalte tragen. **Null Befunde führen nirgends hin.**
//
// # Was die Leiste nicht trägt
//
// **„Aktualisierung“ aus dem Entwurf** — wie weit die Quelle eines
// bezogenen Pakets ist, fragt die Seite erst auf Knopfdruck im Netz
// (`Bezugsstaende`, G04); eine Zahl beim Öffnen hieße, ohne Anweisung zu
// fragen (§16, Regel 5).
import { useAnschluss } from '../anschlussraum'
import { Kennzahlleiste } from '../bausteine/Kennzahlleiste'
import type { Ladezustand } from '../haken/useLaden'
import type { AbhaengigkeitDTO, PaketeAntwort } from '../pakete'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'

export function Paketkennzahlen({
  pakete,
  befunde,
}: {
  pakete: Ladezustand<PaketeAntwort>
  // befunde sind die Abhängigkeitsbefunde der Seite — ihr Schlüssel ist
  // `name.art` (`service.Paketverwalter.Abhaengigkeiten`). **Samt ihrem
  // Fehler**: ohne ihn stünde die Zahl nach einem Lesefehler für immer als
  // „lädt“ da.
  befunde: Pick<Ladezustand<AbhaengigkeitDTO[]>, 'daten' | 'fehler'>
}) {
  const texte = useTexte()
  const eigene = texte.seiten.pakete.leiste
  const anschluss = useAnschluss()

  const liste = pakete.daten?.pakete
  const schluessel = new Set(befunde.daten?.map((befund) => befund.paket))
  const mitBefund = (liste ?? []).filter((paket) =>
    schluessel.has(`${paket.name}.${paket.art}`),
  )

  return (
    <Kennzahlleiste
      bezeichnung={eigene.bezeichnung}
      kennzahlen={[
        {
          schluessel: 'installiert',
          was: eigene.installiert,
          wert:
            pakete.fehler === undefined
              ? liste?.length.toString()
              : eigene.nichtGelesen,
          hin: () => {
            anschluss.uebergib({
              art: 'paketsuche',
              name: eigene.installiert,
              paket: '',
            })
          },
        },
        {
          schluessel: 'befund',
          was: eigene.befund,
          wert:
            befunde.fehler !== undefined
              ? eigene.nichtGelesen
              : befunde.daten === undefined
                ? undefined
                : String(schluessel.size),
          hin:
            mitBefund.length === 0
              ? undefined
              : () => {
                  anschluss.uebergib({
                    art: 'paketauswahl',
                    name: einsetzen(eigene.auswahl, {
                      pakete: zaehlwort(
                        mitBefund.length,
                        texte.allgemein.zaehlwoerter.pakete,
                      ),
                    }),
                    namen: mitBefund.map((paket) => paket.name),
                  })
                },
        },
      ]}
    />
  )
}
