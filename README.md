# 镜像人格 · MBTI 人格测评应用

通过标准化 MBTI 测评定位性格类型，并借助他人的解读视角获得更立体的自我认知。

- **前端**：Vue 3 + Vite + Vue Router（Composition API，`<script setup>`）
- **后端**：Go 1.22 + Gin + GORM
- **数据库**：MySQL 8
- **鉴权**：JWT（bcrypt 密码哈希）

## 功能与页面

### 首页 `/`（测评流 + 求解读广场）
- Hero 入口：开始测评 / 发布求解读
- 求解读信息流，支持「全部 / 征集中 / 已采纳」筛选
- **双面翻转卡片**：
  - 🔮 正面「测评结果 + 类型画像」：MBTI 类型、E/I、S/N、T/F、J/P 四维倾向占比条、性格特质、优势、成长盲区
  - 💬 背面「他人解读 + 洞察视角」：解读列表、帮 TA 解读、发帖人采纳并结算、追加悬赏
- 登录 / 注册弹窗（注册即赠 **100 测评币**）

### 测评页 `/quiz`
- 28 道二选一题目（每维度 7 题），进度条 + 题目跳转点
- 提交后服务端计算四维占比（两极百分比四舍五入后互补为 100%）与 16 型人格画像

### 个人中心 `/me`
- 人格类型档案：类型代码 + 名称 + 四维倾向占比
- 测评币余额、测评次数、发布数、已结算数、解读累计收入
- 三个记录页签：我发布的 / 我解读的 / 测评币流水

## 测评币循环生态

| 动作 | 资金流 | 事务保障 |
|---|---|---|
| 注册 | +100 赠送 | 建用户 + 赠送流水同一事务 |
| 发布求解读 | 冻结扣除悬赏（余额不足返回 402） | `SELECT … FOR UPDATE` 行锁扣款 + 冻结流水 |
| 追加悬赏 | 再次冻结追加金额 | 同上，校验帖子归属与 open 状态 |
| 帮 TA 解读 | 不涉及资金 | 禁止自评、每人限一条、仅 open 帖 |
| 发起人采纳 | 冻结悬赏一次性结算给解读人 | 行锁入账 + 收入流水 + 帖子置 settled，全程同一事务 |

所有账目动作在数据库事务内完成，并写入带 `balanceAfter` 的**测评币流水表**，保证余额可追溯、不会透支或重复结算。

## 数据模型

- `users` 用户（coin_balance 余额）
- `assessment_results` 测评结果（type_code + 四维原始分 JSON）
- `interpretation_posts` 求解读帖（bounty 悬赏、status: open/settled、accepted_response_id）
- `interpretation_responses` 解读回应
- `coin_transactions` 测评币流水（register_gift / post_freeze / bounty_append / settle_income）

## 本地运行

### 1. 数据库
MySQL 本地启动后（默认 root/123456）：
```sql
CREATE DATABASE mbti_app DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```
表结构由 GORM `AutoMigrate` 自动创建。连接串在 `backend/config/config.go` 修改。

### 2. 后端（:8088）
```bash
cd backend
# Go 位于 C:\mine\code\tool\Google\go\go1.22.1\sdk\bin
go run .
```

### 3. 前端（:5310）
```bash
cd frontend
npm install
npm run dev
```
Vite 将 `/api` 代理到 `http://127.0.0.1:8088`。打开 http://127.0.0.1:5310 即可。

## RESTful API 一览

| 方法 | 路径 | 鉴权 | 说明 |
|---|---|---|---|
| POST | `/api/auth/register` | | 注册（赠 100 币） |
| POST | `/api/auth/login` | | 登录 |
| GET | `/api/quiz/questions` | | 题库 |
| POST | `/api/quiz/submit` | ✔ | 提交答案 |
| GET | `/api/quiz/latest` | ✔ | 最近一次测评 |
| GET | `/api/posts` | 可选 | 求解读列表（?status=&mine=） |
| GET | `/api/posts/:id` | 可选 | 帖子详情（双面卡片数据） |
| POST | `/api/posts` | ✔ | 发布（冻结悬赏） |
| POST | `/api/posts/:id/append` | ✔ | 追加悬赏 |
| POST | `/api/posts/:id/responses` | ✔ | 帮 TA 解读 |
| POST | `/api/posts/:id/responses/:rid/accept` | ✔ | 采纳并结算 |
| GET | `/api/me` | ✔ | 个人中心概览 |
| GET | `/api/me/overview` | ✔ | 我的发布/解读 |
| GET | `/api/me/transactions` | ✔ | 测评币流水 |

## 测试

- `backend/smoke-test.mjs`：端到端接口测试（Node 运行），覆盖注册赠送 → 冻结 → 追加 → 解读 → 采纳结算的完整资金链路，共 30+ 断言。
- `frontend/shoot.mjs`：基于 Edge DevTools Protocol 的无头浏览器截图脚本。

> 后端使用 8088、前端使用 5310 是为避开本机已占用的 8080 / 5173 端口。
