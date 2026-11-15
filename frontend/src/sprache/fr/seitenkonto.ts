// **The account kinds on the user page** — piece 6 of phase 4.12 (#1351).
export const konto = {
  eigenes: 'Ton propre compte',
  eigenesErklaerung:
    'Une adresse ou un numéro de téléphone est ici un identifiant : il dit qui se connecte. Le programme ne vérifie pas si la boîte aux lettres ou le numéro existe réellement — il n’envoie rien et n’a besoin d’aucun service externe pour cela.',

  kennung: 'Adresse ou numéro de téléphone',
  kennungHinweis: 'Par exemple nom@example.org ou +4915112345678',

  kennwort: 'Mot de passe',
  kennwortHinweis: 'Au moins douze caractères',
  kennwortAnmelden: 'Mot de passe',

  anzeigename: 'Nom affiché (facultatif)',
  anzeigenameHinweis: 'S’il reste vide, il est formé à partir de l’identifiant',

  eintragen: 'Créer le compte',
  eingetragen: 'Créé : {name}',

  anmelden: 'Se connecter',
  angemeldet: 'Connecté en tant que {name}',

  ohneAngaben: 'Sans informations',
  ohneAngabenErklaerung:
    'Un compte anonyme reçoit un identifiant aléatoire comme nom, pas de mot de passe et pas d’adresse. Et qui n’en veut pas du tout n’en a pas besoin : l’utilisateur du premier lancement est là, et l’application fonctionne.',
  anonym: 'Créer un compte anonyme',
  anonymAngelegt: 'Créé : {name}',

  kennwortNieZurueck:
    'Le mot de passe est stocké sous forme protégée et n’est plus jamais affiché. Qui l’oublie en définit un nouveau au lieu de relire l’ancien.',
} as const

// **The second factor** — piece 7 of phase 4.12 (#1351).
export const faktor = {
  titel: 'Second facteur',
  erklaerung:
    'Un second facteur est exigé à la connexion et au changement de profil. Le code d’une app d’authentification fonctionne partout ; Touch ID ne fonctionne que sur ce Mac.',

  totp: 'Code d’une app',
  biometrie: 'Biométrie',

  einrichten: 'Configurer',
  entfernen: 'Retirer',

  totpEinrichten: 'Configurer un code d’une app',
  totpEntfernen: 'Retirer le code d’une app',
  biometrieEinrichten: 'Configurer Touch ID ou Face ID',
  biometrieEntfernen: 'Retirer Touch ID ou Face ID',

  totpEntfernenFrage: 'Retirer le code d’une app ?',
  biometrieEntfernenFrage: 'Retirer Touch ID ou Face ID ?',
  entfernenErklaerung:
    'La connexion et le changement de profil ne le demandent plus. Qui le veut de nouveau le configure à nouveau — une app avec un nouveau secret.',
  entfernenJa: 'Retirer',

  einmal: 'Ceci ne s’affiche qu’une seule fois',
  einmalErklaerung:
    'Saisis ce secret dans ton app d’authentification et conserve les codes de récupération. Ni l’un ni l’autre ne peut être récupéré ensuite — qui les perd doit tout reconfigurer.',
  codesUeberschrift: '{zahl} codes de récupération, chacun valable une fois :',
  bestaetigen:
    'Saisis un code de ton app. Ce n’est qu’ensuite que le second facteur prend effet.',
} as const
