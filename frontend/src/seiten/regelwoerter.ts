// **Vom Schlüssel zum Satz** — PLAY-02.AK-01
// (#1710).
//
// # Warum das hier steht und nicht in der Komponente
//
// **Weil es zwei Verbraucher hat**: die Regelschau zeigt die Namen, der
// Startassistent beschriftet damit seine Kästchen. Zwei Übersetzungen
// nebeneinander liefen beim ersten neuen Bereich auseinander.
//
// # Und warum ein unbekannter Schlüssel stehen bleibt
//
// **Weil eine Fläche, die ihn verschwiege, eine Regel verschwiege.** Kommt
// aus dem Kern ein Bereich, den dieses Fenster nicht kennt, steht sein
// Schlüssel da — das ist hässlich und ehrlich, und genau so fällt es auf
// (§6, §10).

// Wortbuch ist der Katalog, aus dem gelesen wird.
type Wortbuch = Record<string, unknown>

// grossAnfang macht aus `zonen` das Wortstück `Zonen`.
function grossAnfang(schluessel: string): string {
  return schluessel.charAt(0).toUpperCase() + schluessel.slice(1)
}

// wortOderSchluessel liest einen Eintrag — oder gibt den Schlüssel zurück.
function wortOderSchluessel(
  eigene: Wortbuch,
  name: string,
  schluessel: string,
): string {
  const wort = eigene[name]

  return typeof wort === 'string' ? wort : schluessel
}

// regelbereichwort gibt den Namen eines Bereichs.
export function regelbereichwort(eigene: Wortbuch, bereich: string): string {
  return wortOderSchluessel(
    eigene,
    `regelbereich${grossAnfang(bereich)}`,
    bereich,
  )
}

// regelgradwort gibt den Satz zu einem Grad.
export function regelgradwort(eigene: Wortbuch, grad: string): string {
  return wortOderSchluessel(eigene, `regelgrad${grossAnfang(grad)}`, grad)
}

// regelgruppenwort gibt den Namen eines Astes.
export function regelgruppenwort(eigene: Wortbuch, gruppe: string): string {
  return wortOderSchluessel(eigene, `regelgruppe${grossAnfang(gruppe)}`, gruppe)
}
