// The research texts — work package A16 of phase 5.3
// (#1515).
export const recherche = {
  rechercheTitel: 'Look it up',
  rechercheErklaerung:
    'Asks card texts, rulings and the rules on this machine — full text and vectors merged. The embedding model does take part, locally; no text-writing model is asked. Nothing leaves this machine, and nothing changes.',

  rechercheFall: 'What for',
  rechercheFallKarten: 'Cards',
  rechercheFallRegeln: 'Rules',
  rechercheFallRulings: 'Rulings',

  rechercheFrage: 'Question',
  rechercheSuchen: 'Look up',
  rechercheSprache: 'Language',
  rechercheAlleSprachen: 'Any',

  rechercheModus: 'Searched',
  rechercheModusHybrid: 'full text and vectors merged',
  rechercheModusLexikalisch: 'full text only — no embedding model',
  rechercheModusOhneIndex: 'not at all — there is no search index',

  rechercheLeereNichtsGefunden: 'Try another word?',
  rechercheLeereOhneIndex: 'One can be built under “Settings › Diagnostics”.',
  rechercheLeereOhneModell:
    'Without an embedding model the search finds words, not similarity.',
  rechercheLeereOhneQuelle:
    'This question needs a source that is not in the index.',
  rechercheLeereGesperrt: 'Access to this source has not been granted.',
  rechercheLeereBudget: 'This run reached its limit.',

  rechercheHerkunft: 'Revision {revision}, valid {datum}, fetched {abgerufen}',
  rechercheWeitere: '{anzahl} more sections of the same card',
  rechercheOeffnen: 'Open card',
  rechercheEinschraenkung: 'Not done',
  rechercheStaende: 'Revisions',
  rechercheIndex: 'Search index',
  rechercheMehr: 'Load more',

  aehnlichTitel: 'Similar cards',
  aehnlichErklaerung:
    'Searches with this card’s text, not its name. The card itself is not in the list.',
  aehnlichSuchen: 'Find similar',
  aehnlichLeer: 'Nothing searched yet.',

  freigabeTitel: 'What a model would get',
  freigabeErklaerung:
    'Assembles what a model would see — and does not send it. Whoever generates afterwards has seen what with.',
  freigabeAufgabe: 'Task',
  freigabeAnbieter: 'Provider',
  freigabeModell: 'Model',
  freigabeFenster: 'Context window',
  freigabeReserve: 'Output reserve',
  freigabeZeigen: 'Assemble',
  freigabeOrtLokal: 'Computed on this machine.',
  freigabeOrtCloud: 'Computed at an external provider.',
  freigabeBudget: '{verbraucht} of {grenze} tokens, {reserve} held back',
  freigabeZaehlweise: 'estimated by {zaehlweise}',
  freigabeZieladresse: 'Target address',
  freigabeKategorien: 'Data categories',
  freigabePrivat: 'Own content is included.',
  freigabeOeffentlich: 'Public sources only.',
  freigabeBelege: 'Citations',
  freigabeKeineBelege:
    'No citations — assembled this way a model would see nothing.',
  freigabeVeraltet: 'Not the active revision',
  freigabeLaeuftAb: 'Valid until {zeit}',

  // **Releasing your own collection** — A13.4
  // (#1452).
  freigabeEigeneKarten: 'Search my own cards as well',
  freigabeBestandErklaerung:
    'Without this tick your own cards stay out — the search then covers public sources only, and the answer says so. Whatever goes along leaves this machine if the model runs elsewhere.',
}
