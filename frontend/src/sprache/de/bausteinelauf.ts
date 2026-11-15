// **Der Stand eines Datei-Laufs** — FMT-03.AK-02 (#1692). Ein eigener Teil
// des Katalogs `bausteine`, weil der sonst über die Zeilengrenze liefe (§2, §9).
//
// **Die Einheit kommt als Schlüssel aus Go** (`domain.Einheit…`) und wird
// hier zum Wort — nach „von" im Dativ, ohne Gesamtzahl im Nominativ.
export const laufstand = {
  vonGesamt: '{erledigt} von {gesamt} {einheit}',
  einheitenVon: {
    karte: { eine: 'Karte', viele: 'Karten' },
    set: { eine: 'Set', viele: 'Sets' },
    paket: { eine: 'Paket', viele: 'Paketen' },
    bild: { eine: 'Bild', viele: 'Bildern' },
    schritt: { eine: 'Schritt', viele: 'Schritten' },
  },
  einheiten: {
    karte: { eine: 'Karte', viele: 'Karten' },
    set: { eine: 'Set', viele: 'Sets' },
    paket: { eine: 'Paket', viele: 'Pakete' },
    bild: { eine: 'Bild', viele: 'Bilder' },
    schritt: { eine: 'Schritt', viele: 'Schritte' },
  },
  laeuft: 'Läuft …',
  abbrechen: 'Abbrechen',
  wirdAngehalten: 'Wird angehalten …',
  abgebrochen:
    'Abgebrochen. Was schon ganz war, bleibt — halb Geschriebenes gibt es nicht.',
  bisDahin: 'Bis dahin: {stand}.',
} as const
