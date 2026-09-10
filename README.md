# freechat

[![GitHub license](https://img.shields.io/github/license/ZZY2357/freechat?style=for-the-badge)](https://github.com/ZZY2357/freechat/blob/master/LICENSE)

可搭建的聊天服务

# 快速上手

``` sh
git clone https://github.com/ZZY2357/freechat.git
cd freechat
go build
nohup ./freechat &
```

# 注册与登录

服务使用 JWT 做身份认证，密码以 bcrypt 摘要存储，不落明文。

JWT 由服务端通过 **HttpOnly Cookie** 下发，**不**出现在响应体里：

- `HttpOnly` —— 页面 JS 读不到凭证，XSS 无法把它带走
- `SameSite=Lax` —— 跨站 POST 不携带 Cookie，抦住写接口的 CSRF
- `Secure` —— 由 `COOKIE_SECURE` 控制，上 HTTPS 后应设为 `true`

``` sh
# 注册（凭证写入 Cookie，响应体只回用户名）
curl -c jar.txt -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123"}'

# 登录
curl -c jar.txt -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123"}'

# 带 Cookie 发表评论（作者取自凭证，请求体里的 name 会被忽略）
curl -b jar.txt -X POST http://localhost:8080/comments \
  -H 'Content-Type: application/json' \
  -d '{"content":"hello","os":"Linux"}'

# 退出登录（HttpOnly Cookie 前端删不掉，必须请求服务端）
curl -b jar.txt -X POST http://localhost:8080/auth/logout
```

## 接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| GET | `/` | 否 | 评论页（含登录/注册弹窗） |
| GET | `/comments` | 否 | 评论列表 |
| POST | `/comments` | 是 | 发表评论，返回最新列表 |
| POST | `/auth/register` | 否 | 注册成功返回 `201`，并 `Set-Cookie` |
| POST | `/auth/login` | 否 | 登录成功返回 `200`，并 `Set-Cookie` |
| POST | `/auth/logout` | 否 | 清除认证 Cookie |
| GET | `/auth/me` | 是 | 校验凭证，返回当前用户名（前端用它判断登录态） |

鉴权优先读 `Authorization: Bearer <token>`，没有则回退到认证 Cookie。
因此浏览器走 Cookie，脚本和 API 客户端仍可用 Bearer 头。token 有效期 7 天。
用户名规则：3-20 个字符，忽略大小写判重；密码：6-72 字节。

## 配置

| 环境变量 | 必填 | 说明 |
| --- | --- | --- |
| `JWT_SECRET` | 生产必填 | JWT 签名密钥。未设置时使用内置默认密钥并打印告警，**请勿用于生产环境** |
| `COOKIE_SECURE` | 上 HTTPS 后必填 | 设为 `true` 给认证 Cookie 加 `Secure`。**HTTP 下必须保持关闭**，否则浏览器不会回传 Cookie |
| `ALLOWED_ORIGINS` | 按需 | 允许携带凭证跨域调用的来源白名单，逗号分隔。未配置时不下发任何 CORS 头，即仅允许同源 |

> 认证改为 Cookie 后，响应里不能再回 `Access-Control-Allow-Origin: *`——
> 浏览器不允许 `*` 与凭证共存。需要被别的域名跨域调用时用 `ALLOWED_ORIGINS` 显式列出来源。

``` sh
JWT_SECRET="$(head -c 32 /dev/urandom | base64)" \
COOKIE_SECURE=true \
ALLOWED_ORIGINS="https://chat.example.com" \
./freechat
```

