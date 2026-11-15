// **Invite, compare, allow** — NET-01
// (#1705).
// Maschinell übersetzt.
export const einladung = {
  titel: '邀请某人',
  erklaerung:
    '如果你们在同一网络里找不到彼此，就把你的邀请码发给对方。对方联系得上的话，一个就够；否则就互相交换——之后你们核对三十位数字。',
  zeigen: '显示我的邀请码',
  eingeben: '输入邀请码',

  name: '别人应该如何看到你？',
  nameHinweis: '名称不能证明任何事——所以稍后你们要核对数字。',
  zweck: '这次邀请的用途是什么？',
  erzeugen: '生成邀请码',
  code: '你的邀请码',
  codeHinweis:
    '把它发给对方，例如通过消息。有效期至 {zeitpunkt}，且只能使用一次。',
  kopiert: '已拷贝。',

  eingabe: '对方的邀请码',
  eingabeBeispiel: 'mse4mac-kontakt:1:…',
  pruefen: '检查',

  bestaetigenTitel: '确认 {name}',
  vergleichen:
    '请互相核对这些数字——当面大声念出，或通过电话。如果一致，说明你们确实在与彼此交谈。',
  vergleichenHinweis:
    '对方认识你之后，会看到相同的数字——无论是通过你的邀请码，还是因为你主动告知了对方。',
  vorgaenger:
    '有人使用新密钥自称是 {name}。你之前对 {name} 的授权不适用于它——请重新核对。',
  bekannt: '你已经确认过这台电脑。',
  gesperrt: '你已屏蔽这台电脑。如果想重新添加，请先取消屏蔽。',
  angenommen: '你已经接受过这份邀请。要建立新连接，请索取新的邀请码。',
  eigenerName: '你想怎么称呼对方？',
  eigenerNameHinweis: '这个名称只在你这边有效，不会发送出去。',
  freigeben: '之后允许：{zweck}',
  nichtsGeteilt:
    '在你同意之前，不会共享任何内容。即使确认之后，对方也只能看到你明确共享的内容。',

  technik: '技术信息',
  technikFingerabdruck: '指纹',
  technikAdresse: '地址',
  technikOhneAdresse: '无——你们在同一网络中互相发现',
  technikAblauf: '有效期至',

  abbrechen: '取消',
  spaeter: '稍后核对',
  stimmt: '一致',
  aufgenommen: '已添加 {name}——尚未确认。',
  bestaetigt: '已确认 {name}。',

  lageSchluesselGeaendert: '密钥已更改——请重新核对',
  zeileVergleichen: '请核对这些数字',
  zeileStimmt: '一致',
  zeileVonHand: '手动核对指纹',
  vonHand: '通过标识符手动添加',

  // **One invitation is enough** (NET-01b, #1954).
  meinName: '对方应该如何看到你？',
  meinNameHinweis:
    '你会用这个名称向 {name} 表明身份。名称不能证明任何事——所以你们要核对数字。',
  gemeldet: '{name} 现在也认识你了——在你们核对数字之前，仍未确认。',
  nichtGemeldet: '要让 {name} 也认识你，请把你自己的邀请码发回去。',
  meldungGescheitert: '{name} 没有添加你：{grund}请把你自己的邀请码发回去。',
  meldegruende: {
    unbeantwortet: '没有收到回应。',
    verbraucht: '这份邀请在那边已被使用过。',
    abgelaufen: '这份邀请已过期。',
    fremd: '这份邀请并非来自那边。',
    gesperrt: '你在那边已被屏蔽。',
    gedrosselt: '尝试次数过多——请稍后再试。',
    fassung: '那边运行的是另一个版本。',
    gescheitert: '那边出了点问题。',
  },
}
