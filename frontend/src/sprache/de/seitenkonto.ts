// **Die Kontoarten auf der Nutzerseite** — Werkstück 6 von Phase 4.12
// (#1351).
//
// **Ein eigener Anhang und keine neue Katalogzahl** (§9): es bleiben fünf
// Kataloge; dieser Teil wird in `seitennutzer.ts` eingesetzt, wie
// `seitenkarten.ts` in `seiten.ts`.
export const konto = {
  eigenes: 'Ein eigenes Konto',
  eigenesErklaerung:
    'Eine Adresse oder eine Rufnummer ist hier eine Kennung: sie sagt, wer sich anmeldet. Ob es das Postfach oder die Nummer wirklich gibt, prüft das Programm nicht — es verschickt nichts und braucht dafür keinen fremden Dienst.',

  kennung: 'Adresse oder Rufnummer',
  kennungHinweis: 'Zum Beispiel wer@example.org oder +4915112345678',

  kennwort: 'Kennwort',
  kennwortHinweis: 'Mindestens zwölf Zeichen',
  kennwortAnmelden: 'Kennwort',

  anzeigename: 'Anzeigename (freiwillig)',
  anzeigenameHinweis: 'Bleibt es leer, wird er aus der Kennung gebildet',

  eintragen: 'Konto anlegen',
  eingetragen: 'Angelegt: {name}',

  anmelden: 'Anmelden',
  angemeldet: 'Angemeldet als {name}',

  // **Das anonyme Konto und der Weg ohne jedes.**
  ohneAngaben: 'Ohne Angaben',
  ohneAngabenErklaerung:
    'Ein anonymes Konto bekommt eine Zufallskennung als Namen, kein Kennwort und keine Adresse. Und wer gar keines will, braucht keines: der Nutzer vom ersten Start ist da, und die Anwendung arbeitet.',
  anonym: 'Anonymes Konto anlegen',
  anonymAngelegt: 'Angelegt: {name}',

  // **Was das Kennwort nicht tut.**
  kennwortNieZurueck:
    'Das Kennwort wird gehütet gespeichert und nie wieder angezeigt. Wer es vergisst, setzt ein neues, statt das alte zu lesen.',
} as const

// **Der zweite Faktor** — Werkstück 7 von Phase 4.12
// (#1351).
export const faktor = {
  titel: 'Zweiter Faktor',
  erklaerung:
    'Ein zweiter Faktor wird beim Anmelden und beim Profilwechsel verlangt. Der Code aus einer Authenticator-App geht überall; Touch ID geht nur auf diesem Rechner.',

  totp: 'Code aus einer App',
  biometrie: 'Biometrie',

  einrichten: 'Einrichten',
  entfernen: 'Entfernen',

  totpEinrichten: 'Einen Code aus einer App einrichten',
  totpEntfernen: 'Den Code aus einer App entfernen',
  biometrieEinrichten: 'Touch ID oder Face ID einrichten',
  biometrieEntfernen: 'Touch ID oder Face ID entfernen',

  // **Gefragt wird vor dem Entfernen** — UX-10.AK-04 (#1682): danach
  // verlangen Anmeldung und Profilwechsel einen Faktor weniger, und eine App
  // braucht beim Wiedereinrichten ein neues Geheimnis.
  totpEntfernenFrage: 'Den Code aus einer App entfernen?',
  biometrieEntfernenFrage: 'Touch ID oder Face ID entfernen?',
  entfernenErklaerung:
    'Anmeldung und Profilwechsel fragen danach nicht mehr danach. Wer ihn wieder will, richtet ihn neu ein — eine App mit einem neuen Geheimnis.',
  entfernenJa: 'Entfernen',

  einmal: 'Das steht genau einmal hier',
  einmalErklaerung:
    'Trage dieses Geheimnis in deine Authenticator-App ein und bewahre die Wiederherstellungscodes auf. Beides ist danach nicht mehr abrufbar — wer es verliert, richtet neu ein.',
  codesUeberschrift: '{zahl} Wiederherstellungscodes, jeder gilt einmal:',
  bestaetigen:
    'Gib einen Code aus deiner App ein. Erst danach gilt der zweite Faktor.',
} as const
