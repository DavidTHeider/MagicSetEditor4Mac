// **Ein installiertes Paket, wie die Bindung es gibt.**
//
// Beide Läufe über die Paketwerkstatt brauchen es — der über den Weg
// (`vorlagen.test.tsx`) und der über Wechsel und Fehler
// (`vorlagennebenlauf.test.tsx`). Zwei Kopien liefen an dem Tag auseinander,
// an dem das DTO ein Feld dazubekommt.
export function paketDTO(name: string) {
  return {
    art: 'stil',
    name,
    anzeigename: name,
    version: '2020-01-01',
    pfad: '/pakete/' + name,
    herkunft: '',
    pruefsumme: '',
    braucht: [],
    gruppe: '',
    rang: 0,
    artrang: 1,
    herkunftsart: 'mitgeliefert' as const,
    hauptdatei: 'style',
    geaendert: false,
    gebrauchtVon: [],
    benutzt: 0,
  }
}
