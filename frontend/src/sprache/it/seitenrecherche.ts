// The research texts — work package A16 of phase 5.3
// (#1515).
export const recherche = {
  rechercheTitel: 'Cerca',
  rechercheErklaerung:
    'Interroga testi delle carte, ruling e regolamento su questo computer — testo completo e vettori uniti. Il modello di embedding partecipa, in locale; nessun modello che scrive testi viene interpellato. Nulla lascia questo computer e nulla viene modificato.',

  rechercheFall: 'Per cosa',
  rechercheFallKarten: 'Carte',
  rechercheFallRegeln: 'Regole',
  rechercheFallRulings: 'Ruling',

  rechercheFrage: 'Domanda',
  rechercheSuchen: 'Cerca',
  rechercheSprache: 'Lingua',
  rechercheAlleSprachen: 'Qualsiasi',

  rechercheModus: 'Cercato',
  rechercheModusHybrid: 'testo completo e vettori uniti',
  rechercheModusLexikalisch:
    'solo testo completo — nessun modello di embedding',
  rechercheModusOhneIndex: 'per nulla — non c’è alcun indice di ricerca',

  rechercheLeereNichtsGefunden: 'Prova un’altra parola?',
  rechercheLeereOhneIndex:
    'Se ne può creare uno sotto “Impostazioni › Diagnostica”.',
  rechercheLeereOhneModell:
    'Senza un modello di embedding la ricerca trova parole, non somiglianze.',
  rechercheLeereOhneQuelle:
    'Questa domanda richiede una fonte che non è nell’indice.',
  rechercheLeereGesperrt: 'L’accesso a questa fonte non è stato autorizzato.',
  rechercheLeereBudget: 'Questa esecuzione ha raggiunto il suo limite.',

  rechercheHerkunft:
    'Revisione {revision}, valida {datum}, scaricata {abgerufen}',
  rechercheWeitere: '{anzahl} altre sezioni della stessa carta',
  rechercheOeffnen: 'Apri carta',
  rechercheEinschraenkung: 'Non eseguito',
  rechercheStaende: 'Revisioni',
  rechercheIndex: 'Indice di ricerca',
  rechercheMehr: 'Carica altro',

  aehnlichTitel: 'Carte simili',
  aehnlichErklaerung:
    'Cerca con il testo di questa carta, non con il suo nome. La carta stessa non compare nell’elenco.',
  aehnlichSuchen: 'Trova simili',
  aehnlichLeer: 'Ancora nulla cercato.',

  freigabeTitel: 'Cosa riceverebbe un modello',
  freigabeErklaerung:
    'Riunisce ciò che un modello vedrebbe — e non lo invia. Chi genera dopo ha visto prima con cosa.',
  freigabeAufgabe: 'Compito',
  freigabeAnbieter: 'Fornitore',
  freigabeModell: 'Modello',
  freigabeFenster: 'Finestra di contesto',
  freigabeReserve: 'Riserva per l’output',
  freigabeZeigen: 'Componi',
  freigabeOrtLokal: 'Calcolato su questo computer.',
  freigabeOrtCloud: 'Calcolato presso un fornitore esterno.',
  freigabeBudget: '{verbraucht} di {grenze} token, {reserve} messi da parte',
  freigabeZaehlweise: 'stimato secondo {zaehlweise}',
  freigabeZieladresse: 'Indirizzo di destinazione',
  freigabeKategorien: 'Categorie di dati',
  freigabePrivat: 'Sono inclusi contenuti propri.',
  freigabeOeffentlich: 'Solo fonti pubbliche.',
  freigabeBelege: 'Riscontri',
  freigabeKeineBelege:
    'Nessun riscontro — composto così, un modello non vedrebbe nulla.',
  freigabeVeraltet: 'Non è la revisione attiva',
  freigabeLaeuftAb: 'Valido fino a {zeit}',

  // **Releasing your own collection** — A13.4
  // (#1452).
  freigabeEigeneKarten: 'Cerca anche tra le mie carte',
  freigabeBestandErklaerung:
    'Senza questa spunta le tue carte restano escluse — la ricerca riguarda allora solo fonti pubbliche, e la risposta lo dice. Ciò che viene incluso lascia questo computer se il modello gira altrove.',
}
