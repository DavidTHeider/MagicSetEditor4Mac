// **„Das Bauteil zeichnet nichts" — und was Mantine daran ändert** — UX-07
// (#1679).
//
// # Wofür es diesen Helfer gibt
//
// Drei Läufe sagen zu, dass ein Bauteil **gar nichts** zeichnet: die
// Herkunftszeile ohne Herkunft und die Verträglichkeit ohne Urteil. Sie
// prüften das über `toBeEmptyDOMElement()` am Behälter — und das stimmte,
// solange der Behälter nur das Bauteil enthielt.
//
// **Seit `Sprachraum` auch Mantine aufspannt, enthält er zwei Dinge.**
// `MantineProvider` legt seine Merkmale in ein `<style
// data-mantine-styles>`-Element **im Baum** ab, nicht im Kopf des Dokuments.
// Der Behälter ist damit nie leer, und die drei Läufe wurden rot — ohne dass
// sich an den Bauteilen etwas geändert hätte.
//
// # Warum nicht das Stilelement verhindern
//
// **Weil es die Merkmale trägt, aus denen jedes Mantine-Bauteil seine Maße
// nimmt.** `withCssVariables={false}` nähme sie weg, und dann stünde jedes
// Overlay ohne Maß da. Die Vorrichtung darf nicht großzügiger sein als der
// Betrieb (§3) — und der Betrieb hat dieses Element.
//
// # Was er hält und was nicht
//
// **Er sagt: außer Mantines Merkmalen steht nichts da.** Er sagt **nicht**,
// dass das Bauteil nichts getan hat — ein Effekt ohne sichtbare Wirkung käme
// hier durch, wie er auch durch `toBeEmptyDOMElement` käme.

// mantinestil ist das Merkmal, an dem Mantines Stilelement zu erkennen ist.
const mantinestil = 'style[data-mantine-styles]'

// zeichnetNichts sagt, ob außer Mantines Merkmalen nichts im Behälter steht.
export function zeichnetNichts(behaelter: HTMLElement): boolean {
  return [...behaelter.children].every((kind) => kind.matches(mantinestil))
}
