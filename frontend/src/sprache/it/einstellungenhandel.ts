// **The trading limits** — work item `T1` of phase 5.6
// (#1466).
//
// A separate file, for the same reason as `einstellungennetz.ts` next to it:
// `seiteneinstellungen.ts` would cross the 400-line limit (§2). It stays
// **one** catalogue — the parts are spread in there.

// handelsnamen are the labels of the five limits.
export const handelsnamen = {
  'richtlinie.handel': 'Consenti gli scambi',
  'richtlinie.handel.markt': 'Consenti gli scambi tramite un fornitore esterno',
  'richtlinie.handel.menge': 'Quantità massima per riga',
  'richtlinie.handel.wert':
    'Valore massimo di un’offerta nell’unità più piccola (per EUR: centesimi)',
  'richtlinie.handel.frist': 'Durata massima del blocco in minuti',
}

// handelshilfe is the explaining sentence per limit.
export const handelshilfe = {
  'richtlinie.handel':
    'Disattiva insieme condivisione, prestito di mazzi, baratto e mercato. **Un limite, non una preferenza**: ciò che qui è disattivato non può essere riattivato in nessun profilo.',
  'richtlinie.handel.markt':
    'Per impostazione predefinita è **disattivato**. Il mercato si rivolge a un fornitore esterno e a denaro — tutto il resto resta tra due persone.',
  'richtlinie.handel.menge':
    'Quante copie può indicare al massimo una singola riga di un’offerta. **Vuoto significa nessun limite.**',
  'richtlinie.handel.wert':
    'Quanto può valere al massimo un’offerta nel complesso. Vale solo se la valutazione è **parte del contratto** — una semplice vista di mercato non la limita. **Vuoto significa nessun limite.**',
  'richtlinie.handel.frist':
    'Per quanto tempo un impegno può trattenere delle copie. **Vuoto significa nessun limite.**',
}

// handelswerte are the two positions of the switches.
export const handelswerte = {
  'richtlinie.handel.aus': 'Disattivato',
  'richtlinie.handel.an': 'Attivato',
  'richtlinie.handel.markt.aus': 'Disattivato',
  'richtlinie.handel.markt.an': 'Attivato',
}
