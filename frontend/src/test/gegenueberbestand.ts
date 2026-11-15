// **Ein Bestand an Gegenübern, wie ihn jemand nach ein paar Wochen hat** —
// UXR-16 (#1924).
//
// # Wozu
//
// Die Hüllenprobe `uirevue` geht die Seiten der Palette an, und keine davon
// ist *Gegenüber finden*: die Fläche steht als Kasten in *Handeln*, in
// *Spielen* und unter *Einstellungen › Verbindungen*. Der Probebestand kennt
// außerdem kein Gegenüber. Gemessen wird deshalb unter jsdom mit
// `test/bedienzaehlung.ts`, an Mengen, die ein Mensch nach ein paar Wochen
// Tauschen hat:
//
//   - **acht Gegenüber** in allen Lagen — drei warten auf den Vergleich (zwei
//     mit Code, eines nur mit Fingerabdruck), eines hat einen neuen
//     Schlüssel, drei sind bestätigt und dürfen Verschiedenes, eines ist
//     gesperrt;
//   - **je zwei erteilte Sichten** an den Bestätigten;
//   - **sechs Funde im Netz**, darunter zwei bekannte, einer davon an einer
//     neuen Adresse;
//   - **150 Kontakte** im Adressbuch und **eine vCard mit 200 Karten** —
//     Apple Kontakte gibt „alle Kontakte" als **eine** Datei aus.
import type { KartenvorschauDTO, KontaktDTO } from '../apikontakte'
import type { Nachbar, Peerkontakt } from '../apipeerkontakt'
import type { Sichtfreigabe } from '../apisichtfreigabe'

// zwecke sind die vier aus `domain`, umfaenge die fünf.
export const zwecke = ['paarung', 'teilen', 'handel', 'partie']
export const umfaenge = [
  'auswahl',
  'dubletten',
  'wuensche',
  'sammlung',
  'digitales',
]

// gegenueber baut einen Kontakt mit lesbarem Namen und eigener Kennung.
export function gegenueber(
  name: string,
  anders: Partial<Peerkontakt> = {},
): Peerkontakt {
  const stand = anders.stand ?? 'unbestaetigt'

  return {
    kennung: `b${name.toLowerCase()}`,
    fingerabdruck: '1111-2222-3333-4444-5555-6666-7777-8888',
    selbstgewaehlterName: `${name}s Mac`,
    eigenerName: name,
    stand,
    gesehen: '2026-09-30T18:00:00Z',
    rechte: [],
    adresse: '',
    erstmals: '2026-09-14T03:00:00Z',
    notiz: '',
    lage: stand === 'gesperrt' ? 'widerrufen' : stand,
    vergleichscode: '',
    ...anders,
  }
}

// achtGegenueber ist der Bestand dieser Messung, in der Reihenfolge, in der
// Go ihn liefert.
export function achtGegenueber(): Peerkontakt[] {
  return [
    gegenueber('Anna', {
      vergleichscode: '49311 67386 01529 95503 63193 56021',
      adresse: '192.168.1.42:41234',
    }),
    gegenueber('Bert', {
      vergleichscode: '11111 22222 33333 44444 55555 66666',
    }),
    gegenueber('Cleo'),
    gegenueber('Dora', {
      lage: 'schluesselGeaendert',
      vergleichscode: '77777 88888 99999 00000 11111 22222',
    }),
    gegenueber('Emil', {
      stand: 'bestaetigt',
      rechte: ['handel', 'partie'],
      notiz: 'tauscht fair',
      adresse: 'emils-mac.local:41234',
    }),
    gegenueber('Finn', { stand: 'bestaetigt', rechte: ['handel'] }),
    gegenueber('Greta', {
      stand: 'bestaetigt',
      rechte: ['handel', 'teilen', 'partie'],
    }),
    gegenueber('Hugo', { stand: 'gesperrt' }),
  ]
}

// zweiSichten sind die erteilten Sichten eines Bestätigten.
export function zweiSichten(kennung: string): Sichtfreigabe[] {
  return [
    {
      empfaenger: kennung,
      umfang: 'dubletten',
      zweck: 'handel',
      auswahl: [],
      behalten: 4,
      ablauf: '',
      erteilt: '2026-09-20T12:00:00Z',
    },
    {
      empfaenger: kennung,
      umfang: 'wuensche',
      zweck: 'handel',
      auswahl: [],
      behalten: 0,
      ablauf: '',
      erteilt: '2026-09-21T12:00:00Z',
    },
  ]
}

// fund baut einen Fund im selben Netz.
export function fund(name: string, anders: Partial<Nachbar> = {}): Nachbar {
  const klein = name.toLowerCase()

  return {
    kennung: `b${klein}`,
    fingerabdruck: '9f4b-8c8e-9025-1111-2222-3333-4444-5555',
    name: `${name}s Mac`,
    adresse: '192.168.1.50:41234',
    merkadresse: `${klein}s-mac.local:41234`,
    bekannt: false,
    stand: '',
    gemerkteAdresse: '',
    adresseGeaendert: false,
    ...anders,
  }
}

// sechsFunde sind die Funde dieser Messung.
export function sechsFunde(): Nachbar[] {
  return [
    fund('Ida'),
    fund('Jan'),
    fund('Kim'),
    fund('Lea'),
    fund('Emil', {
      bekannt: true,
      stand: 'bestaetigt',
      gemerkteAdresse: 'emils-mac.local:41234',
    }),
    fund('Finn', {
      bekannt: true,
      stand: 'bestaetigt',
      gemerkteAdresse: '192.168.1.9:41234',
      adresseGeaendert: true,
    }),
  ]
}

// adressbuchkontakt baut einen Kontakt aus Apple Kontakte.
function adressbuchkontakt(nummer: number): KontaktDTO {
  return {
    id: nummer + 1,
    fassung: 1,
    name: `Kontakt ${String(nummer + 1).padStart(3, '0')}`,
    wege: [
      { art: 'email', wert: `k${String(nummer)}@example.org` },
      { art: 'telefon', wert: `+49 170 ${String(nummer)}` },
    ],
    bild: null,
    bildart: '',
    ursprung: `k${String(nummer)}`,
    uebernommen: '2026-09-27T10:00:00Z',
  }
}

// adressbuch sind die 150 schon übernommenen Kontakte.
export function adressbuch(anzahl = 150): KontaktDTO[] {
  return Array.from({ length: anzahl }, (_, nummer) =>
    adressbuchkontakt(nummer),
  )
}

// vcard sind die 200 Karten einer ganzen Ausgabe — die ersten 150 kennt
// das Adressbuch schon.
export function vcard(anzahl = 200): KartenvorschauDTO[] {
  return Array.from({ length: anzahl }, (_, nummer) => {
    const kontakt = { ...adressbuchkontakt(nummer), id: 0, fassung: 0 }

    return nummer < 150
      ? { kontakt, abgleich: 'gleich', vorhanden: adressbuchkontakt(nummer) }
      : { kontakt, abgleich: 'neu', vorhanden: null }
  })
}
