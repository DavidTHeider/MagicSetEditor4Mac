// **Invite, compare, allow** — NET-01
// (#1705).
// Maschinell übersetzt.
export const einladung = {
  titel: 'Inviter quelqu’un',
  erklaerung:
    'Envoie ton code d’invitation si vous ne vous trouvez pas sur le même réseau. Si ton correspondant est joignable, un seul suffit ; sinon vous échangez les deux — ensuite vous comparez trente chiffres.',
  zeigen: 'Montrer mon code',
  eingeben: 'Saisir un code',

  name: 'Comment les autres doivent-ils te voir ?',
  nameHinweis:
    'Un nom ne prouve rien — c’est pourquoi vous allez comparer des chiffres dans un instant.',
  zweck: 'Pour quel usage invites-tu ?',
  erzeugen: 'Créer le code',
  code: 'Ton code d’invitation',
  codeHinweis:
    'Envoie-le à ton correspondant, par exemple dans un message. Il est valable jusqu’à {zeitpunkt} et ne fonctionne qu’une seule fois.',
  kopiert: 'Copié.',

  eingabe: 'Code d’invitation de ton correspondant',
  eingabeBeispiel: 'mse4mac-kontakt:1:…',
  pruefen: 'Vérifier',

  bestaetigenTitel: 'Confirmer {name}',
  vergleichen:
    'Comparez ces chiffres entre vous — à voix haute, dans la même pièce ou au téléphone. S’ils sont identiques, vous parlez bien l’un avec l’autre.',
  vergleichenHinweis:
    'Ton correspondant voit les mêmes chiffres dès qu’il te connaît — par ton code ou parce que tu lui as signalé ta présence.',
  vorgaenger:
    'Cette personne se présente comme {name} avec une nouvelle clé. Ce que tu as autorisé jusqu’ici pour {name} ne s’y applique pas — comparez de nouveau.',
  bekannt: 'Tu as déjà confirmé cet ordinateur.',
  gesperrt:
    'Tu as bloqué cet ordinateur. Débloque-le d’abord si tu veux l’ajouter de nouveau.',
  angenommen:
    'Tu as déjà accepté cette invitation. Pour une nouvelle connexion, demande un nouveau code.',
  eigenerName: 'Comment veux-tu appeler ton correspondant ?',
  eigenerNameHinweis: 'Ce nom ne vaut que chez toi. Il ne sort pas.',
  freigeben: 'Ensuite autoriser : {zweck}',
  nichtsGeteilt:
    'Rien n’est partagé avant que tu donnes ton accord. Même après la confirmation, ton correspondant ne voit que ce que tu partages explicitement.',

  technik: 'Détails techniques',
  technikFingerabdruck: 'Empreinte',
  technikAdresse: 'Adresse',
  technikOhneAdresse: 'aucune — vous vous trouvez sur le même réseau',
  technikAblauf: 'Valable jusqu’à',

  abbrechen: 'Annuler',
  spaeter: 'Comparer plus tard',
  stimmt: 'Ça correspond',
  aufgenommen: '{name} a été ajouté — pas encore confirmé.',
  bestaetigt: '{name} est confirmé.',

  lageSchluesselGeaendert: 'Clé modifiée — compare à nouveau',
  zeileVergleichen: 'Comparez ces chiffres',
  zeileStimmt: 'Ça correspond',
  zeileVonHand: 'Comparer l’empreinte à la main',
  vonHand: 'Ajouter à la main avec un identifiant',

  // **One invitation is enough** (NET-01b, #1954).
  meinName: 'Comment ton correspondant doit-il te voir ?',
  meinNameHinweis:
    'C’est sous ce nom que tu te présentes à {name}. Un nom ne prouve rien — c’est pourquoi vous comparez les chiffres.',
  gemeldet:
    '{name} te connaît maintenant aussi — pas encore confirmé, tant que vous n’avez pas comparé les chiffres.',
  nichtGemeldet: 'Pour que {name} te connaisse aussi, renvoie ton propre code.',
  meldungGescheitert:
    'Ta demande n’a pas abouti chez {name} : {grund} Renvoie ton propre code.',
  meldegruende: {
    unbeantwortet: 'aucune réponse n’est arrivée.',
    verbraucht: 'l’invitation avait déjà été utilisée là-bas.',
    abgelaufen: 'l’invitation avait expiré.',
    fremd: 'l’invitation ne vient pas de là-bas.',
    gesperrt: 'ton accès y est bloqué.',
    gedrosselt: 'trop de tentatives — réessaie plus tard.',
    fassung: 'une autre version y tourne.',
    gescheitert: 'quelque chose a mal tourné là-bas.',
  },
}
