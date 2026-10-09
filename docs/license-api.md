# WR Tool 激活服务接口约定

客户端（Wails 桌面程序）与你的线上发码服务之间的契约。**服务端由你实现**，客户端按本文改造。

## 1. 术语与标识

| 名称 | 含义 | 格式 |
| --- | --- | --- |
| `machineCode` | 客户机指纹，界面展示并可复制 | 16 位 base32，4 位一组：`HSYM-A476-ZP4Z-5MII` |
| `normalizedMachineCode` | 参与签名的规范化形式 | 去掉 `-` 与空格、转大写 |
| `payload` | 被签名的授权内容字符串 | 见 §3 |
| `signature` | 服务端私钥对 `payload` 的签名 | Ed25519，64 字节，base64 std |
| `licenseKey` | 交付给客户的"激活码" | 见 §4，链路 C 才需要 |

## 2. 为什么必须换掉现在的 HMAC

当前 `internal/license` 用 `HMAC-SHA256(Seed, machineCode)`，**校验和发码共用同一个 Seed**，而 Seed 编译进了客户端二进制。把发码搬到服务器并不解决这个问题：客户端仍要持有 Seed，从二进制里抠出来就能给任意机器码造码。

正确做法是**非对称**：服务端持 Ed25519 **私钥**发码，客户端只内置**公钥**验签。客户端被逆向也只能验签，不能造码。

私钥生成（在服务器上做一次，别在开发机）：

```
openssl genpkey -algorithm ed25519 -out license_private.pem
openssl pkey -in license_private.pem -pubout -out license_public.pem
```

- 私钥权限 600，不入 git，不进备份盘明文。
- 公钥以常量写进客户端（32 字节 raw 或 PEM 均可）。
- `payload` 首段带密钥版本（`WRTOOL1K1`），客户端可同时内置 2~3 个公钥以便轮换。

## 3. 签名内容（payload 规范）

```
WRTOOL1|<keyVersion>|<normalizedMachineCode>|<issuedAt>|<expiresAt>|<plan>
```

| 字段 | 说明 | 例 |
| --- | --- | --- |
| `keyVersion` | 密钥版本 | `K1` |
| `normalizedMachineCode` | §1 | `HSYMA476ZP4Z5MII` |
| `issuedAt` | 签发时间 UTC | `20261008T102300Z` |
| `expiresAt` | 授权到期 UTC；**永久授权填 `PERMANENT`** | `20271008T000000Z` |
| `plan` | 套餐标记，便于以后区分买断/年费 | `buy` / `year` / `trial` |

`signature = Ed25519.Sign(priv, []byte(payload))`，对字节直接签，不做二次哈希。

## 4. 三条可选交付链路

选一条为主、其余按需保留。**推荐 A**。

### A. 在线激活（客户点一下就好）

```
POST /v1/activate
Content-Type: application/json

{
  "appId":        "wr_tool",
  "machineCode":  "HSYM-A476-ZP4Z-5MII",
  "clientVersion":"1.1.0",
  "osVersion":    "Windows 10 1809",
  "trialStartAt": "2026-10-08T18:23:53+08:00",
  "nonce":        "c9f1...（客户端随机 16 字节 hex）"
}
```

```
200 OK
{
  "payload":   "WRTOOL1|K1|HSYMA476ZP4Z5MII|20261008T102300Z|PERMANENT|buy",
  "signature": "base64(64 bytes)",
  "customer":  "张三 / 某某律所",
  "serverTime":"2026-10-08T18:23:55+08:00"
}
```

客户端用内置公钥验签，通过后写 `%APPDATA%\WR Tool\license.json`（见 §5），此后**离线可继续用**，因为校验只依赖公钥和本地文件。

### B. 离线授权文件（客户机不能上网时用）

你在管理端输入机器码 → 服务端返回同样的 `payload`+`signature` → 生成 `license.json` 文件 → 微信发给客户 → 客户放到 `%APPDATA%\WR Tool\license.json`。客户端加一个"从授权文件导入"入口。内容与 §5 完全一致，因此 A 激活过的文件可以直接给 B 复用。

### C. 短激活码（保留现在"手输 25 位码"的体验）

Ed25519 签名是 64 字节 = base32 **104 个字符**，**不能截断后还能验签**，所以"短码"必须由服务端在线判定，客户端不能自证：

```
POST /v1/verify   { "machineCode": "...", "licenseKey": "ABCDE-..." }
→ 200 { "payload": "...", "signature": "..." }      // 服务端认这个码，回一张票据
→ 403 { "code": "INVALID_KEY" }
```

服务端保存 `licenseKey`（或它本身的 HMAC），校验通过后把 §3 的 payload 签好回给客户端。客户端拿到票据后离线可用，但**票据要带 `expiresAt`**（比如 30 天），到期需再联网续一次——否则一张票据截图转发就等于永久授权。

> 结论：想要"客户完全离线 + 短码"是不可能的三选二。要么在线（A/C），要么离线但用长码/文件（B）。

## 5. 客户端本地文件

`%APPDATA%\WR Tool\license.json`（安装版路径，见 `internal/license.StorePath`）：

```json
{
  "machineCode": "HSYMA476ZP4Z5MII",
  "payload":     "WRTOOL1|K1|HSYMA476ZP4Z5MII|20261008T102300Z|PERMANENT|buy",
  "signature":   "base64...",
  "activatedAt": "2026-10-08T18:24:00+08:00",
  "lastSeen":    "2026-10-08T09:11:02+08:00"
}
```

- `machineCode` 不匹配当前机器 → 视为未激活（防止拷别人的文件）。
- `lastSeen` 每次启动更新；若当前时间 < `lastSeen - 容忍 24h` → 判定改过系统时钟，按未激活处理。
- 试用期起点单独存在 `%APPDATA%\WR Tool\trial.json`，与授权文件互不影响。

## 6. 服务端数据表（MySQL 8）

```sql
CREATE TABLE wr_licenses (
  id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  machine_code  CHAR(16) NOT NULL,
  customer      VARCHAR(128) NOT NULL DEFAULT '',
  plan          ENUM('buy','year','trial') NOT NULL DEFAULT 'buy',
  issued_at     DATETIME NOT NULL,
  expires_at    DATETIME NULL,
  revoked       TINYINT NOT NULL DEFAULT 0,
  note          VARCHAR(255) NOT NULL DEFAULT '',
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_machine (machine_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 7. 错误码

| HTTP | `code` | 含义 | 客户端提示 |
| --- | --- | --- | --- |
| 400 | `BAD_MACHINE_CODE` | 机器码格式不对 | 请重新复制机器码 |
| 401 | `UNAUTHORIZED` | 管理接口缺/错 key | （仅管理端） |
| 403 | `REVOKED` | 该机器码已被吊销 | 联系提供方 |
| 403 | `INVALID_KEY` | 激活码不正确 | 激活码无效，请核对机器码后重试 |
| 409 | `BOUND_ELSEWHERE` | 该激活码已绑别的机器 | 联系提供方换机 |
| 429 | `RATE_LIMITED` | 请求过快 | 稍后再试 |
| 500 | `INTERNAL` | 服务端异常 | 激活失败，请稍后重试或联系提供方 |

## 8. 安全要求（服务端）

1. 只上 HTTPS，客户端固定校验证书（或至少校验域名 + 不做明文降级）。
2. 发码接口要有鉴权：管理端 `key` 走环境变量，别写死在代码里。
3. `nonce` 服务端记录 5 分钟，重复请求直接拒，防重放。
4. 记录访问日志：机器码、IP、结果——用来判断有没有人批量试码。
5. 私钥泄露时的应急：把 `keyVersion` 升到 `K2`，客户端发新版内置新公钥并拒绝 `K1`。

## 9. 客户端待改造清单（接口定了我这边动手）

1. `internal/license`：删除 `Seed` 与 HMAC 分支，改为 `VerifyPayload(pubKeys, payload, signature)`。
2. `license.json` 结构换成 §5；`Status()` 增加 `expiresAt` 与 `lastSeen` 判断。
3. `app.go`：`VerifyLicense(key)` → 链路 C 的在线校验；新增 `ActivateOnline()`（链路 A）与 `ImportLicenseFile()`（链路 B）。
4. `LicenseModal.vue`：三个入口（在线激活 / 粘贴激活码 / 导入授权文件），保留"加微信 wangran38 + 复制机器码"的引导文案。
5. 域名与 `appId` 需要你提供，客户端要写进配置常量。
