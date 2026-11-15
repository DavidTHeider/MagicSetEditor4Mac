// The access-credential texts — split out with
// #1534.
//
// A catalogue may spread over several files as long as it arrives as one
// (§9) — the same shape as `seitenki.ts`.
export const zugaenge = {
  // The per-kind fallback — it applies where `dienstangaben` says nothing.
  //
  // “Patreon token” used to stand here and was wrong for CardTrader (#1534):
  // `domain.ZugangMarke` serves **two** services, and one string cannot be
  // right for both. A provider’s name therefore belongs in `dienstangaben`;
  // what stands here must hold for **every** service that uses the kind.
  angaben: {
    marke: 'Токен доступа',
    schluessel: 'Ключ API',
    team: 'Идентификатор команды',
  },

  // And per service, where the kind is not enough (#1534). Looked up, not
  // inferred: MTGJSON’s token exists only for Patreon supporters (0056),
  // CardTrader’s comes from its own account’s profile settings (0095).
  dienstangaben: {
    'mtgjson.marke': 'Токен Patreon',
    'cardtrader.marke': 'Токен из настроек профиля',
    'anthropic.schluessel': 'Ключ API',
  },

  zugangLiegt: 'сохранён',
  zugangFehlt: 'отсутствует',
  zugangSetzen: 'Сохранить',
  zugangLoeschen: 'Удалить',
  zugangLoeschenFrage: 'Удалить «{angabe}» для {name}?',
  zugangLoeschenErklaerung:
    'Значение будет удалено из связки ключей этого Mac. Получить его снова можно только у того, кто его выдал.',
  zugangLoeschenJa: 'Удалить',
}
