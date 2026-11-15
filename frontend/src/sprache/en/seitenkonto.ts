// **The account kinds on the user page** — piece 6 of phase 4.12 (#1351).
export const konto = {
  eigenes: 'Your own account',
  eigenesErklaerung:
    'An address or a phone number is an identifier here: it says who is signing in. Whether the mailbox or the number actually exists is not checked — the program sends nothing and needs no outside service for it.',

  kennung: 'Address or phone number',
  kennungHinweis: 'For example who@example.org or +4915112345678',

  kennwort: 'Password',
  kennwortHinweis: 'At least twelve characters',
  kennwortAnmelden: 'Password',

  anzeigename: 'Display name (optional)',
  anzeigenameHinweis: 'Left empty, it is taken from the identifier',

  eintragen: 'Create account',
  eingetragen: 'Created: {name}',

  anmelden: 'Sign in',
  angemeldet: 'Signed in as {name}',

  ohneAngaben: 'Without any details',
  ohneAngabenErklaerung:
    'An anonymous account gets a random identifier as its name, no password and no address. And whoever wants none at all needs none: the user from the first start is there, and the application works.',
  anonym: 'Create anonymous account',
  anonymAngelegt: 'Created: {name}',

  kennwortNieZurueck:
    'The password is stored hashed and never shown again. Whoever forgets it sets a new one instead of reading the old one.',
} as const

// **The second factor** — piece 7 of phase 4.12 (#1351).
export const faktor = {
  titel: 'Second factor',
  erklaerung:
    'A second factor is required when signing in and when switching profiles. A code from an authenticator app works everywhere; Touch ID only works on this Mac.',

  totp: 'Code from an app',
  biometrie: 'Biometrics',

  einrichten: 'Set up',
  entfernen: 'Remove',

  totpEinrichten: 'Set up a code from an app',
  totpEntfernen: 'Remove the code from an app',
  biometrieEinrichten: 'Set up Touch ID or Face ID',
  biometrieEntfernen: 'Remove Touch ID or Face ID',

  totpEntfernenFrage: 'Remove the code from an app?',
  biometrieEntfernenFrage: 'Remove Touch ID or Face ID?',
  entfernenErklaerung:
    'Signing in and switching profiles no longer ask for it. Whoever wants it back sets it up again — an app with a new secret.',
  entfernenJa: 'Remove',

  einmal: 'This is shown exactly once',
  einmalErklaerung:
    'Enter this secret in your authenticator app and keep the recovery codes. Neither can be retrieved afterwards — whoever loses them sets up again.',
  codesUeberschrift: '{zahl} recovery codes, each valid once:',
  bestaetigen:
    'Enter a code from your app. Only then does the second factor take effect.',
} as const
