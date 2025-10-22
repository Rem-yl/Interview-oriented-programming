"""
JWT Token 测试套件 - 使用 pytest

==============================================================================
如何运行服务
==============================================================================

无需前置条件 - JWT 是完全无状态的，不依赖任何外部服务

方式 1: 运行单个服务器（用于基础功能测试）
    cd ../jwt-token
    PORT=8010 SERVERID=server-1 go run main.go

方式 2: 运行多个服务器（用于验证无状态特性）
    # 终端 1
    cd ../jwt-token
    PORT=8010 SERVERID=server-1 go run main.go

    # 终端 2
    PORT=8011 SERVERID=server-2 go run main.go

    # 终端 3
    PORT=8012 SERVERID=server-3 go run main.go

==============================================================================
如何运行测试
==============================================================================

前置条件:
    pip install -r requirements.txt

基础用法:
    pytest test_jwt_token.py -v                 # 详细输出
    pytest test_jwt_token.py -v -s              # 显示 print 输出

运行特定测试:
    pytest test_jwt_token.py::TestBasicFunctionality -v
    pytest test_jwt_token.py::TestBasicFunctionality::test_login_success -v

使用标记:
    pytest test_jwt_token.py -m multi_server -v    # 只运行多服务器测试
    pytest test_jwt_token.py -m stateless -v       # 只运行无状态特性测试

生成报告:
    pytest test_jwt_token.py --html=report.html --self-contained-html

在线解码 JWT Token:
    访问 https://jwt.io/
    粘贴测试中返回的 Token，查看 Payload 内容

==============================================================================
测试说明
==============================================================================

测试覆盖范围:
- 12+ 个测试用例
- 4 个测试类：基础功能、无状态特性、Token 验证、安全性测试

注意事项:
- JWT 完全无状态，不需要 Redis 或任何外部依赖
- 单服务器测试需要至少 1 个服务器运行（端口 8010）
- 无状态测试需要至少 2 个服务器运行（端口 8010, 8011, 8012）

与 Session 方案的区别:
- 完全无状态，Token 自包含所有信息
- 不需要任何外部存储（Redis、数据库）
- Token 无法主动失效，只能等待过期
- 跨服务器天然支持（无需 ip_hash 或共享存储）
"""

import pytest
import requests
import jwt
import time


# ==============================================================================
# Fixtures - 测试准备和清理
# ==============================================================================

@pytest.fixture(scope="session")
def base_url():
    """单服务器基础URL"""
    return "http://localhost:8010"


@pytest.fixture(scope="session")
def multi_server_urls():
    """多服务器URL列表"""
    return [
        "http://localhost:8010",
        "http://localhost:8011",
        "http://localhost:8012"
    ]


@pytest.fixture(scope="session")
def jwt_secret():
    """JWT 密钥（必须与服务器配置一致）"""
    return "rem"


@pytest.fixture(scope="function")
def new_session():
    """为每个测试创建新的 requests.Session 对象"""
    session = requests.Session()
    yield session
    session.close()


@pytest.fixture(scope="session")
def check_server_running(base_url):
    """检查服务器是否运行"""
    try:
        resp = requests.get(f"{base_url}/profile", timeout=2)
        return True
    except requests.exceptions.ConnectionError:
        pytest.skip("服务器未运行，请先启动: PORT=8010 SERVERID=server-1 go run main.go")
    except Exception as e:
        pytest.skip(f"服务器检查失败: {e}")


@pytest.fixture(scope="session")
def check_multi_servers(multi_server_urls):
    """检查多服务器是否运行"""
    available = []
    for url in multi_server_urls:
        try:
            requests.get(f"{url}/profile", timeout=1)
            available.append(url)
        except:
            pass

    if len(available) < 2:
        pytest.skip(
            "需要至少2个服务器运行，请启动:\n"
            "  PORT=8010 SERVERID=server-1 go run main.go\n"
            "  PORT=8011 SERVERID=server-2 go run main.go\n"
            "  PORT=8012 SERVERID=server-3 go run main.go"
        )

    return available


# ==============================================================================
# 测试类 1: 基础功能测试
# ==============================================================================

class TestBasicFunctionality:
    """测试基础功能：登录、Token 生成、认证"""

    def test_login_success(self, base_url, check_server_running):
        """测试 1.1: 登录成功并返回 JWT Token"""
        resp = requests.post(
            f"{base_url}/login",
            json={"username": "alice", "password": "123456"}
        )

        assert resp.status_code == 200, "登录应该返回 200"
        data = resp.json()
        assert "token" in data, "响应应包含 token 字段"

        token = data["token"]
        assert len(token) > 0, "Token 不能为空"
        assert token.count('.') == 2, "JWT Token 应该有 3 部分（Header.Payload.Signature）"

    def test_token_structure(self, base_url, check_server_running, jwt_secret):
        """测试 1.2: 验证 JWT Token 结构"""
        # 登录获取 Token
        resp = requests.post(
            f"{base_url}/login",
            json={"username": "bob", "password": "123456"}
        )
        token = resp.json()["token"]

        # 解码 Token（不验证签名）
        decoded = jwt.decode(token, options={"verify_signature": False})

        # 验证必要字段
        assert "userid" in decoded, "Token 应包含 userid"
        assert "username" in decoded, "Token 应包含 username"
        assert decoded["username"] == "bob", "Username 应该匹配"
        assert "exp" in decoded, "Token 应包含过期时间"
        assert "iat" in decoded, "Token 应包含签发时间"
        assert "iss" in decoded, "Token 应包含签发者"
        assert decoded["iss"] == "jwt-session", "签发者应该是 jwt-session"

    def test_profile_with_valid_token(self, base_url, check_server_running):
        """测试 1.3: 使用有效 Token 访问 /profile"""
        # 登录获取 Token
        resp = requests.post(
            f"{base_url}/login",
            json={"username": "charlie", "password": "123456"}
        )
        token = resp.json()["token"]

        # 使用 Token 访问 profile
        resp = requests.get(
            f"{base_url}/profile",
            headers={"Authorization": f"Bearer {token}"}
        )

        assert resp.status_code == 200, "应该成功获取用户信息"

        data = resp.json()
        assert "userID" in data, "响应应包含 userID"
        assert "username" in data, "响应应包含 username"
        assert data["username"] == "charlie", "用户名应该匹配"

    def test_profile_without_token(self, base_url, check_server_running):
        """测试 1.4: 无 Token 访问 /profile 应该失败"""
        resp = requests.get(f"{base_url}/profile")

        assert resp.status_code == 401, "无 Token 应该返回 401"

        data = resp.json()
        assert "error" in data, "响应应包含错误信息"

    def test_profile_with_invalid_token(self, base_url, check_server_running):
        """测试 1.5: 使用无效 Token 应该失败"""
        fake_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaWQiOiIxMjMiLCJ1c2VybmFtZSI6ImZha2UifQ.invalid"

        resp = requests.get(
            f"{base_url}/profile",
            headers={"Authorization": f"Bearer {fake_token}"}
        )

        assert resp.status_code == 401, "无效 Token 应该返回 401"

        data = resp.json()
        assert "error" in data, "响应应包含错误信息"


# ==============================================================================
# 测试类 2: 无状态特性测试
# ==============================================================================

class TestStatelessFeature:
    """测试 JWT 的无状态特性 - 这是 JWT 的核心优势"""

    @pytest.mark.multi_server
    @pytest.mark.stateless
    def test_cross_server_authentication(self, check_multi_servers):
        """测试 2.1: 跨服务器认证（无状态验证）"""
        servers = check_multi_servers
        server1 = servers[0]
        server2 = servers[1]

        # 在 Server-1 登录
        resp = requests.post(
            f"{server1}/login",
            json={"username": "alice", "password": "123456"}
        )
        assert resp.status_code == 200
        token = resp.json()["token"]

        # 访问 Server-1（应该成功）
        resp1 = requests.get(
            f"{server1}/profile",
            headers={"Authorization": f"Bearer {token}"}
        )
        assert resp1.status_code == 200, "Server-1 应该验证 Token 成功"

        # 访问 Server-2（应该也成功 - 这是 JWT 无状态的关键）
        resp2 = requests.get(
            f"{server2}/profile",
            headers={"Authorization": f"Bearer {token}"}
        )
        assert resp2.status_code == 200, "Server-2 也应该验证 Token 成功（无状态）"

        # 验证返回的 username 相同
        data1 = resp1.json()
        data2 = resp2.json()
        assert data1["username"] == data2["username"], "用户名应该相同"
        assert data1["userID"] == data2["userID"], "用户 ID 应该相同"

    @pytest.mark.multi_server
    @pytest.mark.stateless
    def test_round_robin_all_servers(self, check_multi_servers):
        """测试 2.2: 轮询访问所有服务器（验证无状态）"""
        servers = check_multi_servers

        # 登录获取 Token
        resp = requests.post(
            f"{servers[0]}/login",
            json={"username": "test_user", "password": "123456"}
        )
        token = resp.json()["token"]

        # 轮询访问所有服务器
        success_count = 0
        server_ids = set()

        for i in range(len(servers) * 2):  # 每个服务器访问 2 次
            server = servers[i % len(servers)]
            resp = requests.get(
                f"{server}/profile",
                headers={"Authorization": f"Bearer {token}"}
            )

            if resp.status_code == 200:
                success_count += 1
                data = resp.json()
                server_ids.add(data.get("serverID"))

        # 所有请求都应该成功
        assert success_count == len(servers) * 2, "所有请求都应该成功（JWT 无状态）"
        # 应该访问到不同的服务器
        assert len(server_ids) >= 2, f"应该访问到至少 2 个不同的服务器，实际: {server_ids}"

    @pytest.mark.stateless
    def test_no_server_state(self, base_url, check_server_running):
        """测试 2.3: 验证服务器不存储任何状态"""
        # 登录两次，应该返回不同的 Token（因为 userID 每次生成）
        resp1 = requests.post(
            f"{base_url}/login",
            json={"username": "alice", "password": "123456"}
        )
        token1 = resp1.json()["token"]

        resp2 = requests.post(
            f"{base_url}/login",
            json={"username": "alice", "password": "123456"}
        )
        token2 = resp2.json()["token"]

        # Token 应该不同（因为 userID 是 UUID，每次生成）
        assert token1 != token2, "每次登录应该生成不同的 Token"

        # 两个 Token 都应该有效
        resp = requests.get(
            f"{base_url}/profile",
            headers={"Authorization": f"Bearer {token1}"}
        )
        assert resp.status_code == 200, "第一个 Token 应该有效"

        resp = requests.get(
            f"{base_url}/profile",
            headers={"Authorization": f"Bearer {token2}"}
        )
        assert resp.status_code == 200, "第二个 Token 应该有效"


# ==============================================================================
# 测试类 3: Token 验证测试
# ==============================================================================

class TestTokenValidation:
    """测试 Token 验证：签名、过期、格式"""

    def test_token_signature_verification(self, base_url, check_server_running, jwt_secret):
        """测试 3.1: Token 签名验证"""
        # 登录获取 Token
        resp = requests.post(
            f"{base_url}/login",
            json={"username": "alice", "password": "123456"}
        )
        token = resp.json()["token"]

        # 验证签名
        try:
            decoded = jwt.decode(token, jwt_secret, algorithms=["HS256"])
            assert "userid" in decoded
            assert "username" in decoded
        except jwt.InvalidSignatureError:
            pytest.fail("Token 签名验证失败")

    def test_tampered_token(self, base_url, check_server_running, jwt_secret):
        """测试 3.2: 篡改的 Token 应该被拒绝"""
        # 登录获取 Token
        resp = requests.post(
            f"{base_url}/login",
            json={"username": "alice", "password": "123456"}
        )
        token = resp.json()["token"]

        # 解码 Token
        decoded = jwt.decode(token, options={"verify_signature": False})

        # 修改 Payload（篡改用户名）
        decoded["username"] = "hacker"

        # 使用错误的密钥重新签名（模拟篡改）
        tampered_token = jwt.encode(decoded, "wrong-secret", algorithm="HS256")

        # 使用篡改的 Token 访问
        resp = requests.get(
            f"{base_url}/profile",
            headers={"Authorization": f"Bearer {tampered_token}"}
        )

        # 应该被拒绝
        assert resp.status_code == 401, "篡改的 Token 应该被拒绝"

    def test_wrong_bearer_format(self, base_url, check_server_running):
        """测试 3.3: 错误的 Bearer 格式"""
        resp = requests.post(
            f"{base_url}/login",
            json={"username": "alice", "password": "123456"}
        )
        token = resp.json()["token"]

        # 使用错误的前缀（Basic 而不是 Bearer）
        resp = requests.get(
            f"{base_url}/profile",
            headers={"Authorization": f"Basic {token}"}
        )

        # 应该失败（TrimPrefix 会留下 "Basic " + token，导致验证失败）
        assert resp.status_code == 401, "错误的 Authorization 格式应该被拒绝"


# ==============================================================================
# 测试类 4: 安全性测试
# ==============================================================================

class TestSecurity:
    """测试安全特性"""

    def test_token_contains_no_sensitive_data(self, base_url, check_server_running):
        """测试 4.1: Token 中不应包含敏感信息（密码等）"""
        resp = requests.post(
            f"{base_url}/login",
            json={"username": "alice", "password": "secret123"}
        )
        token = resp.json()["token"]

        # 解码 Token
        decoded = jwt.decode(token, options={"verify_signature": False})

        # 验证不包含密码
        assert "password" not in decoded, "Token 不应包含密码"
        assert "secret" not in str(decoded).lower(), "Token 不应包含任何密码相关信息"

    def test_token_expiration_time(self, base_url, check_server_running):
        """测试 4.2: Token 应该有过期时间"""
        resp = requests.post(
            f"{base_url}/login",
            json={"username": "alice", "password": "123456"}
        )
        token = resp.json()["token"]

        # 解码 Token
        decoded = jwt.decode(token, options={"verify_signature": False})

        # 验证过期时间
        assert "exp" in decoded, "Token 应该包含过期时间"

        exp_time = decoded["exp"]
        iat_time = decoded["iat"]

        # 过期时间应该是 2 小时后（根据服务器配置）
        expected_duration = 2 * 60 * 60  # 2 小时
        actual_duration = exp_time - iat_time

        assert abs(actual_duration - expected_duration) < 10, \
            f"Token 过期时间应该是 2 小时，实际: {actual_duration / 3600:.2f} 小时"


# ==============================================================================
# Pytest 配置
# ==============================================================================

def pytest_configure(config):
    """Pytest 配置"""
    config.addinivalue_line(
        "markers", "multi_server: 需要多服务器的测试"
    )
    config.addinivalue_line(
        "markers", "stateless: 测试无状态特性"
    )
