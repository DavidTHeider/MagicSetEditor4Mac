// **Die lokale Sitzung** — ID-02
// (#1701).
//
// **Ein Anhang und keine neue Katalogzahl** (§9): er wird in
// `seitennutzer.ts` eingesetzt. **Touch ID heißt hier lokale Entsperrung**
// und nirgends Passkey (AK-02).
export const sitzung = {
  ueberschrift: 'Sitzung',
  erklaerung:
    'Sperren zeigt nur noch die Sperranzeige, bis jemand mit dem Kennwort oder lokal über Touch ID entsperrt — auch nach einem Neustart. Abmelden beendet die Sitzung; danach wird gewählt, wer arbeitet. Alles bleibt auf diesem Rechner: keine zentrale Anmeldung, keine Bestätigung per E-Mail.',
  sperren: 'Sperren',
  abmelden: 'Abmelden',

  kennwortAendern: 'Kennwort ändern',
  kennwortAlt: 'Bisheriges Kennwort',
  kennwortAltHinweis: 'Leer lassen, wenn es noch keines gibt',
  kennwortNeu: 'Neues Kennwort',
  kennwortNeuHinweis: 'Mindestens zwölf Zeichen',
  kennwortSetzen: 'Kennwort setzen',
  kennwortGeaendert: 'Das Kennwort ist gesetzt.',

  gesperrtTitel: 'Gesperrt',
  gesperrtErklaerung:
    'Diese Anwendung ist gesperrt. Entsperrt wird mit dem Kennwort dessen, der gesperrt hat.',
  kennwort: 'Kennwort',
  entsperren: 'Entsperren',
  lokalEntsperren: 'Lokal entsperren (Touch ID)',
  lokalHinweis:
    'Touch ID entsperrt nur auf diesem Rechner. Es ist kein Passkey und meldet nirgends an.',

  vergessen: 'Kennwort vergessen?',
  vergessenErklaerung:
    'Ein Code vom Zettel setzt ein neues Kennwort — und sonst nichts. Den Zettel gab es beim Einrichten des zweiten Faktors; keine Rolle, keine Freigabe, kein Gegenüber ändert sich. Ohne Zettel bleibt nur eine Sicherung von vorher — über eine E-Mail-Adresse geht es nicht.',
  code: 'Code vom Zettel',
  neuSetzen: 'Kennwort neu setzen',

  abgemeldetTitel: 'Abgemeldet',
  abgemeldetErklaerung:
    'Wähle, wer arbeitet, und melde dich an. Wer kein Kennwort hat, kommt ohne herein.',
  wer: 'Wer arbeitet',
  anmelden: 'Anmelden',
}
