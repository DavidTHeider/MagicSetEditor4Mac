// **Invite, compare, allow** — NET-01
// (#1705).
export const einladung = {
  titel: 'Invite someone',
  erklaerung:
    'Send your invitation code if you can’t find each other on the same network. If the other person is reachable, one is enough; otherwise you swap both — then you compare thirty digits.',
  zeigen: 'Show my code',
  eingeben: 'Enter a code',

  name: 'How should others see you?',
  nameHinweis:
    'A name proves nothing — that is why you will compare digits in a moment.',
  zweck: 'What are you inviting them to?',
  erzeugen: 'Create code',
  code: 'Your invitation code',
  codeHinweis:
    'Send it to the other person, for example as a message. It is valid until {zeitpunkt} and works only once.',
  kopiert: 'Copied.',

  eingabe: 'Their invitation code',
  eingabeBeispiel: 'mse4mac-kontakt:1:…',
  pruefen: 'Check',

  bestaetigenTitel: 'Confirm {name}',
  vergleichen:
    'Compare these digits with each other — out loud, in the same room or on the phone. If they match, you really are talking to each other.',
  vergleichenHinweis:
    'The other person sees the same digits once they know you — through your code or because you announced yourself.',
  vorgaenger:
    'This claims to be {name} with a new key. Whatever you allowed {name} before does not apply — compare again.',
  bekannt: 'You have already confirmed this computer.',
  gesperrt:
    'You have blocked this computer. Unblock it first if you want to add it again.',
  angenommen:
    'You have already accepted this invitation. Ask for a new code for a new connection.',
  eigenerName: 'What do you want to call them?',
  eigenerNameHinweis: 'The name only applies on your side. It never leaves.',
  freigeben: 'Then allow: {zweck}',
  nichtsGeteilt:
    'Nothing is shared before you agree. Even after confirming, they only see what you explicitly share.',

  technik: 'Technical details',
  technikFingerabdruck: 'Fingerprint',
  technikAdresse: 'Address',
  technikOhneAdresse: 'none — you will find each other on the same network',
  technikAblauf: 'Valid until',

  abbrechen: 'Cancel',
  spaeter: 'Compare later',
  stimmt: 'They match',
  aufgenommen: '{name} has been added — not yet confirmed.',
  bestaetigt: '{name} is confirmed.',

  lageSchluesselGeaendert: 'Key changed — please compare again',
  zeileVergleichen: 'Compare these digits',
  zeileStimmt: 'They match',
  zeileVonHand: 'Compare the fingerprint by hand',
  vonHand: 'Add by hand with an identifier',

  // **One invitation is enough** (NET-01b, #1954).
  meinName: 'How should the other person see you?',
  meinNameHinweis:
    'This is the name you announce yourself to {name} with. A name proves nothing — that is why you compare the digits.',
  gemeldet:
    '{name} knows you now as well — not yet confirmed until you have compared the digits.',
  nichtGemeldet: 'So that {name} knows you too, send your own code back.',
  meldungGescheitert:
    '{name} did not add you: {grund} Send your own code back.',
  meldegruende: {
    unbeantwortet: 'there was no answer.',
    verbraucht: 'the invitation had already been redeemed there.',
    abgelaufen: 'the invitation had expired.',
    fremd: 'the invitation did not come from there.',
    gesperrt: 'you are blocked there.',
    gedrosselt: 'too many attempts — try again later.',
    fassung: 'a different version is running there.',
    gescheitert: 'something went wrong there.',
  },
}
