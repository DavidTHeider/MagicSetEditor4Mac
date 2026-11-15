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
    marke: 'アクセストークン',
    schluessel: 'APIキー',
    team: 'チームID',
  },

  // And per service, where the kind is not enough (#1534). Looked up, not
  // inferred: MTGJSON’s token exists only for Patreon supporters (0056),
  // CardTrader’s comes from its own account’s profile settings (0095).
  dienstangaben: {
    'mtgjson.marke': 'Patreonトークン',
    'cardtrader.marke': 'プロフィール設定のトークン',
    'anthropic.schluessel': 'APIキー',
  },

  zugangLiegt: '保存済み',
  zugangFehlt: '未設定',
  zugangSetzen: '保存',
  zugangLoeschen: '削除',
  zugangLoeschenFrage: '{name}の{angabe}を削除しますか？',
  zugangLoeschenErklaerung:
    'この値は、この Mac のキーチェーンから削除されます。再取得できるのは、発行元だけです。',
  zugangLoeschenJa: '削除',
}
