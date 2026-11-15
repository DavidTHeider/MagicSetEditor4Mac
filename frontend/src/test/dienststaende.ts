// **Zwei Stände eines Dienstpostens** — für die Läufe um `Dienstposten`,
// aus `dienstposten.test.tsx` herausgelöst, als G08 (#2272) die Datei über
// die Zeilengrenze aus §2 brachte.
import type { DienstAntwort } from '../apidienste'

export const steht: DienstAntwort = {
  stand: {
    art: 'rest',
    laeuft: false,
    adresse: '',
    tls: false,
    marke: '',
    namen: [],
    adressen: [],
    fingerabdruck: '',
    zeugnisAussteller: '',
    zeugnisBis: '',
    lauscher: [],
    fehlende: [],
  },
  fehler: null,
}

export const laeuft: DienstAntwort = {
  stand: {
    art: 'rest',
    laeuft: true,
    adresse: 'https://[::]:54591',
    tls: true,
    marke: 'die-marke',
    namen: ['mein.mac'],
    adressen: ['192.168.178.51:54591'],
    fingerabdruck: '4e0b-621d',
    zeugnisAussteller: '',
    zeugnisBis: '',
    lauscher: ['https://[::]:54591'],
    fehlende: [],
  },
  fehler: null,
}
