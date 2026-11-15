// The **cost ledger** — work item O12 of phase 5.2
// (#1439).

export const kosten = {
  titel: 'What an outside job may cost',
  hinweis:
    'Jobs sent to outside services can cost money. What has been approved is ' +
    'shown here — and how much of it is already committed.',

  kontoWaehlen: 'Account',
  zeitraumWaehlen: 'Period',

  zeitraeume: {
    tag: 'Today',
    monat: 'This month',
  } as Record<string, string>,

  grenze: 'Approved',
  reserviert: 'Committed',
  verbraucht: 'Spent',
  unklar: 'Unsettled',
  verfuegbar: 'Available',

  unklarHinweis:
    'Unsettled means the job ran into a timeout after the provider may ' +
    'already have accepted it. It counts as spent until somebody has checked.',

  ohneGrenze:
    'Nothing is approved for this account. Without an approval no paid job ' +
    'goes out.',

  lokalHinweis:
    'This is a limit on this machine. Anyone using the same credentials ' +
    'elsewhere consumes quota this ledger knows nothing about.',

  grenzeSetzen: 'Change the approval',
  betragTippen: 'Amount in the smallest unit (cents for EUR)',
  waehrungTippen: 'Currency',
  setzen: 'Approve',
  setztGerade: 'Setting …',
  nullNimmtWeg: 'Zero withdraws the approval.',

  postenTitel: 'The most recent jobs',
  keinePosten: 'Nothing booked yet.',
  postenZeile: '{aufgabe} · {stand} · {betrag}',

  staende: {
    reserviert: 'committed',
    abgerechnet: 'settled',
    freigegeben: 'nothing charged',
    unklar: 'unsettled',
  } as Record<string, string>,
}
