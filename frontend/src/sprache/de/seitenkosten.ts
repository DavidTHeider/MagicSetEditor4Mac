// Die Texte des **Kostenjournals** — Werkstück O12 von Phase 5.2
// (#1439).

export const kosten = {
  titel: 'Was ein Auftrag draußen kosten darf',
  hinweis:
    'Aufträge an fremde Dienste können Geld kosten. Was freigegeben ist, ' +
    'steht hier — und was davon schon gebunden ist.',

  kontoWaehlen: 'Konto',
  zeitraumWaehlen: 'Zeitraum',

  zeitraeume: {
    tag: 'Heute',
    monat: 'Diesen Monat',
  } as Record<string, string>,

  grenze: 'Freigegeben',
  reserviert: 'Angemeldet',
  verbraucht: 'Ausgegeben',
  unklar: 'Ungeklärt',
  verfuegbar: 'Verfügbar',

  unklarHinweis:
    'Ungeklärt heißt: der Auftrag lief in eine Frist, nachdem der Anbieter ' +
    'ihn schon angenommen haben konnte. Er zählt wie ausgegeben, bis jemand ' +
    'nachgesehen hat.',

  ohneGrenze:
    'Für dieses Konto ist nichts freigegeben. Ohne Freigabe geht kein ' +
    'kostenpflichtiger Auftrag hinaus.',

  lokalHinweis:
    'Das ist eine Grenze auf diesem Rechner. Wer denselben Zugang woanders ' +
    'benutzt, verbraucht Kontingent, von dem hier nichts steht.',

  grenzeSetzen: 'Freigabe ändern',
  betragTippen: 'Betrag in der kleinsten Einheit (bei EUR: Cent)',
  waehrungTippen: 'Währung',
  setzen: 'Freigeben',
  setztGerade: 'Wird gesetzt …',
  nullNimmtWeg: 'Null nimmt die Freigabe weg.',

  postenTitel: 'Die letzten Aufträge',
  keinePosten: 'Noch nichts gebucht.',
  postenZeile: '{aufgabe} · {stand} · {betrag}',

  staende: {
    reserviert: 'angemeldet',
    abgerechnet: 'abgerechnet',
    freigegeben: 'nichts angefallen',
    unklar: 'ungeklärt',
  } as Record<string, string>,
}
