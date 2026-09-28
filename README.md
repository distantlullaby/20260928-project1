# 镜我 · MBTI 人格测评与解读社区

帮助用户通过标准化测评定位性格类型，并借助他人的解读视角获得更立体的自我认知。

- **前端**：Vue 3 + Vite + Pinia + Vue Router + Axios，调用 RESTful API
- **后端**：Go 1.22 + Gin + GORM（MySQL 8）
- **核心机制**：测评币循环生态，发帖冻结与采纳结算均由 **数据库事务 + 行级锁（SELECT ... FOR UPDATE）** 保障账目安全

---

## 功能一览

### 两个 Web 页面

**首页（测评流 + 求解读大厅）**
- 32 题标准化测评入口（E/I、S/N、T/F、J/P 四维各 8 题，逐题作答、进度可视）
- 求解读列表，卡片可翻面：
  - **正面 = 测评结果 + 类型画像**（类型徽章、四维占比条、画像关键词/优势/方向）
  - **背面 = 他人解读 + 洞察视角**（解读列表、帮 TA 解读、发起人确认采纳）
- 发布「求解读」（选择测评结果、填写困惑、设置悬赏）
- 悬赏追加、关闭退款

**个人中心**
- 人格类型档案：当前类型 + E/I、S/N、T/F、J/P 四维倾向占比
- 测评币钱包：可用余额 + 冻结中金额
- 我的求解读 / 我的解读 / 测评历史 / 测评币流水 四个 Tab

### 测评币循环

| 动作 | 币的变化 |
| --- | --- |
| 注册 | 赠送 200 初始测评币 |
| 发布求解读 | 悬赏币从「可用」冻结，余额不足返回 402 |
| 追加悬赏 | 再次冻结并累加到请求 |
| 他人「帮 TA 解读」 | 提交文字解读（不涉及币） |
| 发起人「确认采纳」 | 解冻并核销悬赏，等额结算给解读人 |
| 关闭未采纳请求 | 冻结币全额退回 |

事务内顺序：锁定用户行 → 校验余额/状态 → 改写余额/冻结 → 写流水 → 提交，并发下不会超支或重复结算。

---

## 目录结构

```
20260928-project1-B/
├─ backend/
│  ├─ main.go              # 入口、路由
│  ├─ config/              # 配置（端口/DSN/JWT/初始币）
│  ├─ database/            # DB 初始化、自动迁移、题库/画像/演示数据种子
│  ├─ models/              # 用户/题目/画像/测评结果/请求/解读/币流水
│  ├─ handlers/            # HTTP 处理器
│  ├─ services/            # 测评币事务逻辑（核心）
│  ├─ middleware/          # JWT 鉴权
│  └─ utils/               # JWT、MBTI 计分
├─ frontend/
│  └─ src/
│     ├─ views/            # 首页/测评/结果/详情/登录/个人中心
│     ├─ components/       # 双面卡片、四维占比条、类型徽章、Toast
│     ├─ stores/           # Pinia 用户态
│     ├─ api/ router/ composables/ styles/
├─ run-backend.ps1
└─ run-frontend.ps1
```

## 数据表

- `users` 用户（含可用币 `coin_balance`、冻结币 `frozen_balance`、当前 `mbti_type`）
- `questions` 32 道测评题
- `type_profiles` 16 型人格画像
- `assessment_results` 测评结果（含四维占比 JSON）
- `interpretation_requests` 求解读请求（悬赏额、状态、采纳的解读）
- `interpretations` 解读回应
- `coin_transactions` 测评币流水（变动额、变动后余额快照、类型、关联 ID）

---

## 本地运行

### 环境

- Go：`C:\mine\code\tool\Google\go\go1.22.1\sdk\bin`（脚本已内置）
- MySQL 8：本地已启动，`root / 123456`
- Node：通过 nvm 管理（v22 已验证）

### 1. 建库（仅首次）

```sql
CREATE DATABASE mbti_assess DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

> 表结构由 GORM AutoMigrate 自动创建；首次启动自动灌入 32 题、16 型画像与演示数据。

### 2. 启动后端

```powershell
./run-backend.ps1
# 或：cd backend; go run .
# 监听 http://127.0.0.1:8090 （8080 被同机其他项目占用）
```

### 3. 启动前端

```powershell
./run-frontend.ps1
# 或：cd frontend; npm install; npm run dev
# 默认 http://127.0.0.1:5173 ，被占用时 Vite 自动顺延，请看终端实际输出
```

前端通过 Vite proxy 把 `/api` 转发到 `http://127.0.0.1:8090`。

### 配置覆盖（可选环境变量）

`DB_HOST` `DB_PORT` `DB_USER` `DB_PASS` `DB_NAME` `APP_ADDR` `JWT_SECRET` `INITIAL_COINS`

---

## 演示账号（密码均为 demo123）

| 用户名 | 昵称 | 类型 | 说明 |
| --- | --- | --- | --- |
| `linxia` | 林小夏 | INFP | 发布 1 条悬赏中请求 |
| `zhouye` | 周野 | ENTJ | 1 条请求已采纳结算 |
| `chenan` | 陈安 | ISFJ | 活跃解读人，也发布 1 条悬赏 |

---

## RESTful API 摘要

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| POST | `/api/auth/register` | - | 注册并赠送 200 币 |
| POST | `/api/auth/login` | - | 登录获取 JWT |
| GET | `/api/me` | ✓ | 当前用户与余额 |
| GET | `/api/questions` | - | 测评题 |
| GET | `/api/profiles` / `?code=INFP` | - | 类型画像 |
| POST | `/api/assessments` | ✓ | 提交答卷、计分存结果 |
| GET | `/api/results/:id` | - | 单条测评结果+画像 |
| GET | `/api/my/results` | ✓ | 我的测评历史 |
| GET | `/api/requests` | - | 求解读列表（status/mine/page） |
| GET | `/api/requests/:id` | - | 详情含全部解读 |
| POST | `/api/requests` | ✓ | 发布并冻结悬赏 |
| POST | `/api/requests/:id/append` | ✓ | 追加悬赏冻结 |
| POST | `/api/requests/:id/close` | ✓ | 关闭并退款 |
| POST | `/api/requests/:id/interpretations` | ✓ | 帮 TA 解读 |
| POST | `/api/requests/:id/interpretations/:interpId/accept` | ✓ | 采纳并结算 |
| GET | `/api/my/interpretations` | ✓ | 我解读过的 |
| GET | `/api/my/transactions` | ✓ | 测评币流水 |

请求头携带：`Authorization: Bearer <token>`。
