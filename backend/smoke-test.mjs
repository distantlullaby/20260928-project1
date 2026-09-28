// 端到端冒烟测试：注册赠送 → 发帖冻结 → 追加 → 解读 → 采纳结算 → 流水校验
const BASE = 'http://127.0.0.1:8088/api';

async function req(path, { method = 'GET', token, body } = {}) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });
  let data = null;
  const text = await res.text();
  try { data = text ? JSON.parse(text) : null; } catch { data = text; }
  return { status: res.status, data };
}

const assert = (cond, msg) => {
  if (!cond) { console.error('❌ ASSERT FAIL:', msg); process.exitCode = 1; }
  else console.log('✅', msg);
};

(async () => {
  // 注册或登录两个用户（测试可重复执行）
  async function loginOrRegister(username, password, nickname) {
    let r = await req('/auth/register', { method: 'POST', body: { username, password, nickname } });
    if (r.status === 409) r = await req('/auth/login', { method: 'POST', body: { username, password } });
    return r;
  }
  const alice = await loginOrRegister('alice', 'alice123', '爱丽丝');
  const bob = await loginOrRegister('bob', 'bob12345', '鲍勃');
  assert(alice.status === 200 || alice.status === 201, `alice 认证成功 (${alice.status})`);
  assert(bob.status === 200 || bob.status === 201, `bob 认证成功 (${bob.status})`);
  const TA = alice.data.token, TB = bob.data.token;

  // 错误密码
  const bad = await req('/auth/login', { method: 'POST', body: { username: 'alice', password: 'wrong' } });
  assert(bad.status === 401, '错误密码返回 401');

  // 题库
  const q = await req('/quiz/questions');
  assert(q.data.questions.length === 28, `题库共 28 题 (实际 ${q.data.questions.length})`);

  // 提交测评：偶数题选 A、奇数题选 B
  const answers = {};
  for (let i = 1; i <= 28; i++) answers[i] = i % 2 === 0 ? 'A' : 'B';
  let r = await req('/quiz/submit', { method: 'POST', token: TA, body: { answers } });
  assert(r.status === 201, `测评提交成功 type=${r.data.typeCode}`);
  const rid = r.data.resultId;
  console.log('   维度占比:', r.data.dimensions.map(d =>
    `${d.negative}${d.negPercent}%/${d.positive}${d.posPercent}%`).join('  '));
  assert(!!r.data.profile.summary && r.data.profile.traits.length === 4, '返回类型画像');

  // 未答完
  r = await req('/quiz/submit', { method: 'POST', token: TA, body: { answers: { 1: 'A' } } });
  assert(r.status === 400, '未答完返回 400');
  r = await req('/quiz/submit', { method: 'POST', body: { answers } });
  assert(r.status === 401, '未登录提交返回 401');

  // 查 alice 当前余额（老数据可能已被前序测试花掉）
  let me = (await req('/me', { token: TA })).data;
  const balanceBefore = me.user.coinBalance;
  console.log('   alice 发帖前余额:', balanceBefore);

  // 发悬赏帖 30
  r = await req('/posts', { method: 'POST', token: TA, body: { assessmentResultId: rid, bounty: 30, note: '求解读' } });
  assert(r.status === 201, `发帖成功 id=${r.data.post.id}`);
  const pid = r.data.post.id;
  me = (await req('/me', { token: TA })).data;
  assert(me.user.coinBalance === balanceBefore - 30, `发帖冻结 30，余额 ${balanceBefore}→${me.user.coinBalance}`);

  // 余额不足
  r = await req('/posts', { method: 'POST', token: TA, body: { assessmentResultId: rid, bounty: 999999 } });
  assert(r.status === 402, '余额不足返回 402');

  // 追加悬赏 20
  const balanceAfterFreeze = me.user.coinBalance;
  r = await req(`/posts/${pid}/append`, { method: 'POST', token: TA, body: { amount: 20 } });
  assert(r.status === 200 && r.data.coinBalance === balanceAfterFreeze - 20,
    `追加 20 冻结成功，余额→${r.data.coinBalance}`);

  // bob 帮解读
  r = await req(`/posts/${pid}/responses`, { method: 'POST', token: TB, body: { content: '你是典型的分析型人格，思考深入，注意别过度内耗，多给自己放松的空间。' } });
  assert(r.status === 201, `bob 解读成功 id=${r.data.response?.id}`);
  const respId = r.data.response.id;

  // 不能给自己解读
  r = await req(`/posts/${pid}/responses`, { method: 'POST', token: TA, body: { content: '自己给自己写一条解读看看行不行呢' } });
  assert(r.status === 400, '给自己解读返回 400');
  // 重复解读
  r = await req(`/posts/${pid}/responses`, { method: 'POST', token: TB, body: { content: '同一条帖子我还想再解读一次内容' } });
  assert(r.status === 409, '重复解读返回 409');
  // 非发帖人不能采纳
  r = await req(`/posts/${pid}/responses/${respId}/accept`, { method: 'POST', token: TB });
  assert(r.status === 403, '非发帖人采纳返回 403');

  // bob 采纳前余额
  const bobBefore = (await req('/me', { token: TB })).data.user.coinBalance;

  // alice 采纳，结算 50
  r = await req(`/posts/${pid}/responses/${respId}/accept`, { method: 'POST', token: TA });
  assert(r.status === 200, `采纳结算成功: ${r.data.message}`);

  // 重复采纳
  r = await req(`/posts/${pid}/responses/${respId}/accept`, { method: 'POST', token: TA });
  assert(r.status === 400, '重复采纳返回 400');
  // 已结算帖不能再解读
  r = await req(`/posts/${pid}/responses`, { method: 'POST', token: TB, body: { content: '结束后再补一条解读试试看内容' } });
  assert(r.status === 400, '已结算帖不能再解读');

  // 余额校验
  const aliceAfter = (await req('/me', { token: TA })).data.user.coinBalance;
  const bobAfter = (await req('/me', { token: TB })).data.user.coinBalance;
  assert(aliceAfter === balanceAfterFreeze - 20, `alice 余额保持 ${aliceAfter}（冻结款不退回，已结算给解读人）`);
  assert(bobAfter === bobBefore + 50, `bob 余额 +50: ${bobBefore}→${bobAfter}`);

  // 帖子详情双面卡片数据
  const detail = (await req(`/posts/${pid}`)).data.post;
  assert(detail.status === 'settled' && detail.acceptedResponseId === respId, '帖子状态 settled 且记录采纳 ID');
  assert(detail.assessment.profile.code === detail.assessment.typeCode, '正面：测评结果+类型画像完整');
  assert(detail.responses.length === 1 && detail.responses[0].resolver.nickname === '鲍勃', '背面：他人解读+解读人信息完整');

  // 列表
  const list = (await req('/posts?status=settled')).data.posts;
  assert(list.length >= 1 && list.some(p => p.id === pid), '首页列表包含已结算帖');

  // 流水
  const aliceTx = (await req('/me/transactions', { token: TA })).data.transactions;
  console.log('   alice 流水:', aliceTx.map(t => `${t.type}(${t.change})`).join(' '));
  assert(aliceTx.some(t => t.type === 'register_gift' && t.change > 0), '流水含注册赠送');
  assert(aliceTx.some(t => t.type === 'post_freeze' && t.change === -30), '流水含发帖冻结 -30');
  assert(aliceTx.some(t => t.type === 'bounty_append' && t.change === -20), '流水含追加冻结 -20');
  // 流水余额连续性
  let continuous = true;
  for (const t of aliceTx) {
    // balanceAfter 必须非负
    if (t.balanceAfter < 0) continuous = false;
  }
  assert(continuous, '所有流水变动后余额非负');

  const bobTx = (await req('/me/transactions', { token: TB })).data.transactions;
  assert(bobTx.some(t => t.type === 'settle_income' && t.change === 50), 'bob 流水含采纳收入 +50');

  // 个人中心
  const overview = (await req('/me/overview', { token: TA })).data;
  assert(overview.myPosts.some(p => p.id === pid), '个人中心含我发布的帖');
  const bobOv = (await req('/me/overview', { token: TB })).data;
  assert(bobOv.myResponses.some(x => x.postId === pid && x.earned === true), '个人中心含我解读且已中标记录');

  console.log(process.exitCode ? '\n❌ 存在失败用例' : '\n🎉 全部用例通过');
})().catch(e => { console.error(e); process.exit(1); });
