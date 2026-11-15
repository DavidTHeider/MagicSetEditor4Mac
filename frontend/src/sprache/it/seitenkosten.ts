// The **cost ledger** — work item O12 of phase 5.2
// (#1439).

export const kosten = {
  titel: 'Quanto può costare un incarico esterno',
  hinweis:
    'Gli incarichi inviati a servizi esterni possono costare denaro. Qui ' +
    'trovi ciò che è stato autorizzato — e quanto di questo è già impegnato.',

  kontoWaehlen: 'Account',
  zeitraumWaehlen: 'Periodo',

  zeitraeume: {
    tag: 'Oggi',
    monat: 'Questo mese',
  } as Record<string, string>,

  grenze: 'Autorizzato',
  reserviert: 'Impegnato',
  verbraucht: 'Speso',
  unklar: 'Da chiarire',
  verfuegbar: 'Disponibile',

  unklarHinweis:
    'Da chiarire significa che l’incarico è andato in timeout dopo che il ' +
    'fornitore poteva già averlo accettato. Conta come speso finché qualcuno ' +
    'non ha verificato.',

  ohneGrenze:
    'Per questo account non è autorizzato nulla. Senza un’autorizzazione non ' +
    'parte alcun incarico a pagamento.',

  lokalHinweis:
    'Questo è un limite su questo computer. Chi usa le stesse credenziali ' +
    'altrove consuma una quota di cui qui non c’è traccia.',

  grenzeSetzen: 'Modifica l’autorizzazione',
  betragTippen: 'Importo nell’unità più piccola (per EUR: centesimi)',
  waehrungTippen: 'Valuta',
  setzen: 'Autorizza',
  setztGerade: 'Impostazione in corso …',
  nullNimmtWeg: 'Zero revoca l’autorizzazione.',

  postenTitel: 'Gli ultimi incarichi',
  keinePosten: 'Ancora nulla registrato.',
  postenZeile: '{aufgabe} · {stand} · {betrag}',

  staende: {
    reserviert: 'impegnato',
    abgerechnet: 'saldato',
    freigegeben: 'nessun addebito',
    unklar: 'da chiarire',
  } as Record<string, string>,
}
