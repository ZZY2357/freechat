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

``` sh
# 注册（成功返回 token）
curl -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123"}'

# 登录
curl -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123"}'

# 用 token 发表评论（作者取自 token，请求体里的 name 会被忽略）
curl -X POST http://localhost:8080/comments \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <token>' \
  -d '{"content":"hello","os":"Linux"}'
```

## 接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| GET | `/` | 否 | 评论页（含登录/注册弹窗） |
| GET | `/comments` | 否 | 评论列表 |
| POST | `/comments` | 是 | 发表评论，返回最新列表 |
| POST | `/auth/register` | 否 | 注册，成功返回 `201` + token |
| POST | `/auth/login` | 否 | 登录，成功返回 `200` + token |
| GET | `/auth/me` | 是 | 校验 token，返回当前用户名 |

鉴权方式为请求头 `Authorization: Bearer <token>`，token 有效期 7 天。
用户名规则：3-20 个字符，忽略大小写判重；密码：6-72 字节。

## 配置

| 环境变量 | 必填 | 说明 |
| --- | --- | --- |
| `JWT_SECRET` | 生产必填 | JWT 签名密钥。未设置时使用内置默认密钥并打印告警，**请勿用于生产环境** |

``` sh
JWT_SECRET="$(head -c 32 /dev/urandom | base64)" ./freechat
```

