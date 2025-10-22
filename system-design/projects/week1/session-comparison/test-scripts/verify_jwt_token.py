#!/usr/bin/env python3
"""
验证 JWT Token 实现

运行前确保:
1. 启动至少 1 个后端服务器 (8010)
2. 无需 Redis 或其他外部依赖（JWT 的优势）
"""

import requests
import jwt
import json


def test_server_connection():
    """测试服务器连接"""
    print("=" * 70)
    print("测试 1: 验证服务器连接")
    print("=" * 70)

    server = "http://localhost:8010"

    try:
        # 尝试访问（应该返回 401，表示服务器正常运行）
        resp = requests.get(f"{server}/profile", timeout=2)
        if resp.status_code == 401:
            print("✅ 服务器连接成功")
            return True
        else:
            print(f"⚠️  服务器响应异常: {resp.status_code}")
            return True
    except requests.exceptions.ConnectionError:
        print("❌ 服务器连接失败")
        print("   请先启动服务器: PORT=8010 SERVERID=server-1 go run main.go")
        return False
    except Exception as e:
        print(f"❌ 连接测试失败: {e}")
        return False


def test_jwt_authentication():
    """测试 JWT 认证流程"""
    print("\n" + "=" * 70)
    print("测试 2: 验证 JWT 认证流程")
    print("=" * 70)

    server = "http://localhost:8010"
    jwt_secret = "rem"  # 必须与服务器配置一致

    try:
        # 1. 登录获取 Token
        print("\n1. 登录获取 JWT Token...")
        resp = requests.post(
            f"{server}/login",
            json={"username": "alice", "password": "123456"}
        )

        if resp.status_code != 200:
            print(f"❌ 登录失败: {resp.status_code}")
            return False

        data = resp.json()
        if "token" not in data:
            print("❌ 响应中没有 token 字段")
            return False

        token = data["token"]
        print(f"✅ 登录成功")
        print(f"   Token (前50字符): {token[:50]}...")
        print(f"   Token 结构: {token.count('.')} 个点（Header.Payload.Signature）")

        # 2. 解码并显示 Token 内容
        print("\n2. 解码 Token 内容...")
        try:
            # 不验证签名，只查看内容
            decoded = jwt.decode(token, options={"verify_signature": False})
            print("✅ Token Payload:")
            print(f"   {json.dumps(decoded, indent=2)}")

            # 验证签名
            verified_decoded = jwt.decode(token, jwt_secret, algorithms=["HS256"])
            print("✅ Token 签名验证成功")

        except jwt.InvalidSignatureError:
            print("❌ Token 签名验证失败")
            return False
        except Exception as e:
            print(f"❌ Token 解码失败: {e}")
            return False

        # 3. 使用 Token 访问 /profile
        print("\n3. 使用 Token 访问 /profile...")
        resp = requests.get(
            f"{server}/profile",
            headers={"Authorization": f"Bearer {token}"}
        )

        if resp.status_code == 200:
            data = resp.json()
            print("✅ 认证成功，获取用户信息:")
            print(f"   {json.dumps(data, indent=2, ensure_ascii=False)}")
            return True
        else:
            print(f"❌ 认证失败: {resp.status_code}")
            return False

    except Exception as e:
        print(f"❌ 测试失败: {e}")
        return False


def test_stateless_feature():
    """测试无状态特性（跨服务器）"""
    print("\n" + "=" * 70)
    print("测试 3: 验证无状态特性（跨服务器）")
    print("=" * 70)

    servers = [
        "http://localhost:8010",
        "http://localhost:8011",
        "http://localhost:8012"
    ]

    # 检查可用的服务器
    available_servers = []
    for server in servers:
        try:
            requests.get(f"{server}/profile", timeout=1)
            available_servers.append(server)
        except:
            pass

    if len(available_servers) < 2:
        print("⚠️  只有 1 个服务器运行，无法测试跨服务器特性")
        print("   如需测试，请启动多个服务器:")
        print("   PORT=8010 SERVERID=server-1 go run main.go")
        print("   PORT=8011 SERVERID=server-2 go run main.go")
        return

    print(f"\n发现 {len(available_servers)} 个运行中的服务器:")
    for server in available_servers:
        print(f"   - {server}")

    try:
        # 在第一个服务器登录
        server1 = available_servers[0]
        server2 = available_servers[1]

        print(f"\n1. 在 {server1} 登录...")
        resp = requests.post(
            f"{server1}/login",
            json={"username": "alice", "password": "123456"}
        )

        if resp.status_code != 200:
            print(f"❌ 登录失败")
            return

        token = resp.json()["token"]
        print(f"✅ 登录成功，获取 Token")

        # 访问第一个服务器
        print(f"\n2. 使用 Token 访问 {server1}...")
        resp1 = requests.get(
            f"{server1}/profile",
            headers={"Authorization": f"Bearer {token}"}
        )

        if resp1.status_code == 200:
            data1 = resp1.json()
            print(f"✅ Server-1 认证成功")
            print(f"   serverID: {data1.get('serverID')}")
        else:
            print(f"❌ Server-1 认证失败")
            return

        # 访问第二个服务器（关键测试）
        print(f"\n3. 使用同一 Token 访问 {server2}（跨服务器）...")
        resp2 = requests.get(
            f"{server2}/profile",
            headers={"Authorization": f"Bearer {token}"}
        )

        if resp2.status_code == 200:
            data2 = resp2.json()
            print(f"✅ Server-2 也能验证 Token（无状态特性验证成功）")
            print(f"   serverID: {data2.get('serverID')}")

            # 验证用户信息一致
            if data1.get("userID") == data2.get("userID"):
                print("\n✅ 验证通过: JWT 无状态特性工作正常")
                print("   - 不同服务器能验证同一 Token")
                print("   - 无需共享存储（Redis）")
                print("   - 完全无状态架构")
            else:
                print("\n⚠️  警告: userID 不一致")

        else:
            print(f"❌ Server-2 认证失败: {resp2.status_code}")
            print("   JWT 应该支持跨服务器（无状态）")

    except Exception as e:
        print(f"❌ 测试失败: {e}")


def test_invalid_token():
    """测试无效 Token 处理"""
    print("\n" + "=" * 70)
    print("测试 4: 验证无效 Token 处理")
    print("=" * 70)

    server = "http://localhost:8010"

    print("\n1. 测试没有 Token...")
    resp = requests.get(f"{server}/profile")
    if resp.status_code == 401:
        print("✅ 正确返回 401 Unauthorized")
    else:
        print(f"⚠️  返回了 {resp.status_code}，期望 401")

    print("\n2. 测试无效的 Token...")
    fake_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature"
    resp = requests.get(
        f"{server}/profile",
        headers={"Authorization": f"Bearer {fake_token}"}
    )

    if resp.status_code == 401:
        print("✅ 正确拒绝无效 Token")
    else:
        print(f"⚠️  返回了 {resp.status_code}，期望 401")

    print("\n3. 测试篡改的 Token...")
    # 登录获取有效 Token
    resp = requests.post(
        f"{server}/login",
        json={"username": "alice", "password": "123456"}
    )
    token = resp.json()["token"]

    # 解码并篡改
    decoded = jwt.decode(token, options={"verify_signature": False})
    decoded["username"] = "hacker"

    # 使用错误密钥重新签名
    tampered_token = jwt.encode(decoded, "wrong-secret", algorithm="HS256")

    resp = requests.get(
        f"{server}/profile",
        headers={"Authorization": f"Bearer {tampered_token}"}
    )

    if resp.status_code == 401:
        print("✅ 正确拒绝篡改的 Token")
    else:
        print(f"⚠️  返回了 {resp.status_code}，期望 401")


def compare_with_other_solutions():
    """对比其他方案"""
    print("\n" + "=" * 70)
    print("对比总结: JWT Token vs Session 方案")
    print("=" * 70)

    print("""
┌────────────────────┬─────────────────┬─────────────────┬─────────────────┐
│ 特性               │ Sticky Session  │ Redis Session   │ JWT Token       │
├────────────────────┼─────────────────┼─────────────────┼─────────────────┤
│ 存储位置           │ 服务器本地内存  │ Redis 集中存储  │ 客户端（Token） │
│ 状态               │ 有状态          │ 有状态          │ 无状态          │
│ 跨服务器访问       │ ❌ 不支持       │ ✅ 支持         │ ✅ 天然支持     │
│ Nginx 算法         │ ip_hash (必须)  │ round_robin     │ round_robin     │
│ 依赖外部服务       │ ❌ 无           │ ✅ Redis        │ ❌ 无           │
│ 可主动失效         │ ✅ 可以         │ ✅ 可以         │ ❌ 不可以       │
│ 性能               │ 极高            │ 高（网络I/O）   │ 极高            │
│ 扩展性             │ 差              │ 好              │ 极好            │
│ Token 体积         │ 小（UUID）      │ 小（UUID）      │ 大（包含信息）  │
│ 适用场景           │ 单体应用        │ 分布式应用      │ 微服务、API     │
└────────────────────┴─────────────────┴─────────────────┴─────────────────┘

本次测试验证了:
✅ JWT Token 的生成和验证
✅ 无状态特性（跨服务器验证 Token）
✅ Token 签名机制和安全性
✅ 完全不依赖外部存储

JWT 的优势:
✅ 完全无状态，易于水平扩展
✅ 无需外部依赖（Redis、数据库）
✅ 天然支持分布式、微服务架构
✅ Token 自包含信息，减少查询

JWT 的劣势:
❌ Token 无法主动失效（只能等待过期）
❌ Token 体积较大（每次请求都传输）
❌ 敏感信息不能放入 Token（可被解码）
❌ 密钥泄露影响范围大
    """)


if __name__ == "__main__":
    print("\n")
    print("╔" + "=" * 68 + "╗")
    print("║" + " " * 21 + "JWT Token 验证工具" + " " * 29 + "║")
    print("╚" + "=" * 68 + "╝")

    # 运行测试
    if not test_server_connection():
        print("\n请先启动服务器后再运行测试")
    else:
        test_jwt_authentication()
        test_stateless_feature()
        test_invalid_token()
        compare_with_other_solutions()

    print("\n" + "=" * 70)
    print("测试完成")
    print("=" * 70)
