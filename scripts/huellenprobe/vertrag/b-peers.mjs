// **Peers, Einladung, Sichten und Adressbuch** — eine Gruppe des
// Vertragsrundgangs (B01, #2220).
//
// Eine Einzelanwendung hat kein Gegenüber. Was sie hat, ist ein Bestand, in
// den man **fremde Kennungen eintragen** kann: ein Peerkontakt ist ein
// Eintrag mit einer Kennung, einem Abdruck und einem Stand, und nichts davon
// verlangt, dass drüben jemand sitzt. Die Kennungen stellt `hilfen-b.mjs`
// her — gültig, aber nur in diesem Skript vorhanden; **es geht kein Byte ins
// Netz**.
//
// # Die Reihenfolge ist der Weg eines Kontakts
//
// aufnehmen (unbestätigt) → bestätigen (mit dem Abdruck, den `Peerkontakte`
// nennt) → Zweck erlauben → Sicht erteilen → alles zurücknehmen → sperren →
// entsperren. Die Adresse wird gesetzt und **sofort wieder gelöscht**: ein
// Kontakt mit Adresse ist einer, zu dem etwas wählen könnte.
//
// **Die Einladung** ist von einem zweiten fremden Peer unterschrieben und
// nicht einlösbar (sie nennt keine Adresse): angenommen wird sie, gemeldet
// wird nichts. Die eigene Einladung nähme das Programm nicht an.
//
// **Das Adressbuch** hat mit Peers nichts zu tun (ein Kontakt dort ist ein
// Eintrag aus einer vCard) und steht hier nur, weil die Liste es so ordnet.
//
// # Wie weit das geprüft ist
//
// Die Abfolge ist gegen den App-Bau gefahren worden (`app.NeueBindungen` über
// einem frisch gestarteten `app.NeueApp`, Wegwerfprobe vom 2026-10-09, nicht
// eingecheckt), die Argumentverbünde strikt gegen die Go-Typen und die
// `nimm`-Ausdrücke gegen die serialisierten Antworten. **Nicht geprüft** ist,
// was nur das native Fenster hat: die Wails-Naht selbst, die echten
// Dateidialoge (dort stand ein eingesetzter Dateiwähler) und jeder
// Unterprozess.

import { b, einladungVon, fremder } from './hilfen-b.mjs'

// kontakt ist der erste fremde Peer dieser Gruppe — der mit der Adresse, der
// Notiz und der Sicht.
const kontakt = (k) => fremder(k, 'kontakt').kennung
const einlader = (k) => fremder(k, 'einlader')

// liste ist die Antwort aller Kontaktwege: die ganze Liste mit Abdruck.
const liste = 'a.kontakte.map((c) => ({ kennung: c.kennung, fingerabdruck: c.fingerabdruck, stand: c.stand }))'

export const bPeers = {
  name: 'Peers, Einladung, Sichten, Adressbuch',
  rufe: [
    // Die eigene Identität gibt es nach dem ersten Ruf; ohne sie fragt kein
    // Weg, der unterschreiben muss (Einladung, Handel, Leihe).
    { name: 'PeeridentitaetSicherstellen', nimm: 'a.kennung', nach: (w, k) => { b(k).meineKennung = w } },

    {
      name: 'PeerkontaktAufnehmen', argumente: (k) => [kontakt(k), 'Fremder Kontakt'], nimm: liste,
      nach: (w, k) => { b(k).abdruck = w.find((c) => c.kennung === kontakt(k)).fingerabdruck },
    },
    { name: 'PeerkontaktBestaetigen', argumente: (k) => [kontakt(k), b(k).abdruck] },
    { name: 'PeerzweckErlauben', argumente: (k) => [kontakt(k), 'handel'] },
    { name: 'PeerkontaktAdresse', argumente: (k) => [kontakt(k), '127.0.0.1:9'] },
    { name: 'PeerkontaktAdresse', argumente: (k) => [kontakt(k), ''] },
    { name: 'PeerkontaktNotiz', argumente: (k) => [kontakt(k), 'aus dem Rundgang'] },

    // **Sichten** setzen einen bestätigten Kontakt voraus; der Umfang
    // `sammlung` braucht weder Auswahl noch Behaltemenge.
    {
      name: 'SammlungFreigeben',
      argumente: (k) => [{ empfaenger: kontakt(k), umfang: 'sammlung', zweck: 'handel', auswahl: [], behalten: 0, ablauf: '', erteilt: '' }],
    },
    { name: 'SammlungsfreigabeWiderrufen', argumente: (k) => [kontakt(k), 'handel', 'sammlung'] },
    { name: 'PeerzweckEntziehen', argumente: (k) => [kontakt(k), 'handel'] },
    { name: 'PeerkontaktSperren', argumente: (k) => [kontakt(k)] },
    { name: 'PeerkontaktEntsperren', argumente: (k) => [kontakt(k)] },

    // **Eine Einladung von einem Fremden**, ohne Adresse — nicht einlösbar,
    // also ohne Meldung. Ohne Vergleichscode bleibt der Kontakt unbestätigt.
    {
      name: 'KontakteinladungAnnehmen', argumente: (k) => [einladungVon(einlader(k), { name: 'Einladender', zweck: 'paarung' }), 'Eingeladen', '', false, ''],
    },

    // --- Adressbuch: ein Kontakt aus einer vCard, ohne Datei ---
    {
      name: 'KontakteUebernehmen',
      argumente: () => [[{
        id: 0, fassung: 0, name: 'Vertragskontakt', wege: [{ art: 'email', wert: 'vertrag@example.test' }],
        bild: null, bildart: '', ursprung: 'vertrag-kontakt-1', uebernommen: '',
      }]],
    },
    {
      name: 'Kontakte', nimm: "a.kontakte.filter((c) => c.name === 'Vertragskontakt').map((c) => ({ id: c.id, fassung: c.fassung }))",
      nach: (w, k) => { b(k).adresskontakt = w[0] },
    },
    { name: 'KontaktLoeschen', argumente: (k) => [b(k).adresskontakt.id, b(k).adresskontakt.fassung] },
  ],
}
