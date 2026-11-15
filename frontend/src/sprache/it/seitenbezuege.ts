// I testi degli **stati scaricati** nella pagina dei pacchetti — G04 (#2251).
// La loro fonte tedesca è `de/seitenbezuege.ts`.

export const bezuege = {
  ueberschrift: 'Stati scaricati',
  erklaerung:
    'Un pacchetto scaricato resta a un commit e non invecchia da solo. Chiedi alle fonti se ne esiste uno più recente: solo allora qualcosa passa per la rete.',
  fragen: 'Interroga le fonti',
  leer: 'Nessun pacchetto ancora scaricato. Scaricane uno con «Scarica pacchetto …»: dopo qui vedrai quanto è avanzata la sua fonte.',
  bezogen: 'Scaricato: {stand}',
  geprueft: 'Verificato: {stand}',
  ohnePruefung: 'Non è nell’elenco delle fonti verificate',
  pakete: '{zahl} pacchetti',
  paketEins: 'un pacchetto',
  keinPaket: 'nessun pacchetto',
  lage: {
    ungefragt: 'Non ancora interrogata.',
    aktuell: 'Allo stato più recente della fonte.',
    neuerEins: 'Stato più recente {spitze} del {datum}: un commit avanti.',
    neuer: 'Stato più recente {spitze} del {datum}: {vorsprung} commit avanti.',
    abweichend:
      'La fonte non è avanti: reimpostata o divergente (punta {spitze}).',
    unbekannt: 'Non è stato possibile interrogare la fonte: {grund}',
  },
  ansehenBezeichnung: 'Esamina l’aggiornamento … ({name})',
  zurueckBezeichnung: 'Torna allo stato conservato ({name}, {stand})',
  ansehen: 'Esamina l’aggiornamento …',
  aufgehoben: 'Conservato: {stand}',
  zurueck: 'Torna allo stato conservato',
  vorschauTitel: 'Aggiornamento di {name}',
  vonZu: '{von} → {zu}: {vorsprung} avanti, {rueckstand} indietro',
  paketzeile:
    '{paket}: {neu} nuovi, {geaendert} modificati, {entfernt} rimossi',
  keinePakete: 'Nessun pacchetto di modelli cambia.',
  uebrige: 'Altri {zahl} file non appartengono ad alcun pacchetto.',
  sets: 'Questi set nominano un pacchetto interessato: {sets}',
  keineSets: 'Nessun set nomina un pacchetto interessato.',
  aufheben: 'Lo stato attuale viene conservato e si può ripristinare.',
  vorigeWeicht: 'Lo stato già conservato {stand} viene sostituito.',
  gruende: {
    setVerliert: 'Un pacchetto nominato da un set perde file.',
    nichtVoraus:
      'La fonte non è avanti: l’aggiornamento sarebbe un passo indietro o di lato.',
    gekuerzt:
      'La fonte indica più di 300 file modificati; l’elenco sopra è troncato.',
  },
  bestaetigen: 'Ho visto i motivi e voglio aggiornare comunque.',
  ausfuehren: 'Aggiorna a {zu}',
  verwerfen: 'Non aggiornare',
}
