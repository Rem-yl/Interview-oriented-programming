"""
Sticky Session 测试套件 - 使用 pytest

==============================================================================
如何运行服务
==============================================================================

方式 1: 运行单个服务器（用于基础功能测试）
    cd ../sticky-session
    PORT=8081 SERVER_ID=server-1 go run main.go

方式 2: 运行多个服务器（用于完整测试套件）
    # 终端 1
    cd ../sticky-session
    PORT=8081 SERVER_ID=server-1 go run main.go

    # 终端 2
    cd ../sticky-session
    PORT=8082 SERVER_ID=server-2 go run main.go

    # 终端 3
    cd ../sticky-session
    PORT=8083 SERVER_ID=server-3 go run main.go

==============================================================================
如何运行测试
==============================================================================

前置条件:
    pip install -r requirements.txt

基础用法:
    pytest test_sticky_session.py -v                 # 详细输出
    pytest test_sticky_session.py -v -s              # 显示 print 输出

运行特定测试:
    pytest test_sticky_session.py::TestBasicFunctionality -v
    pytest test_sticky_session.py::TestBasicFunctionality::test_login_success -v

使用标记:
    pytest test_sticky_session.py -m multi_server -v    # 只运行多服务器测试
    pytest test_sticky_session.py -m "not multi_server" # 跳过多服务器测试

生成报告:
    pytest test_sticky_session.py --html=report.html --self-contained-html

调试:
    pytest test_sticky_session.py --pdb              # 失败时进入调试器
    pytest test_sticky_session.py -l --tb=short      # 显示局部变量

==============================================================================
测试说明
==============================================================================

测试覆盖范围:
- 13 个测试用例
- 4 个测试类：基础功能、Cookie 机制、多服务器隔离、数据一致性

注意事项:
- 单服务器测试需要至少 1 个服务器运行（端口 8081）
- 多服务器测试需要至少 2 个服务器运行（端口 8081, 8082, 8083）
- 如果服务器未运行，相关测试会被自动跳过
"""

import pytest
import requests


# ==============================================================================
# Fixtures - 测试准备和清理
# ==============================================================================

@pytest.fixture(scope="session")
def base_url():
    """单服务器基础URL"""
    return "http://localhost:8081"


@pytest.fixture(scope="session")
def multi_server_urls():
    """多服务器URL列表"""
    return [
        "http://localhost:8081",
        "http://localhost:8082",
        "http://localhost:8083"
    ]


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
        pytest.skip("服务器未运行，请先启动: go run main.go")
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
            "  PORT=8081 SERVER_ID=server-1 go run main.go\n"
            "  PORT=8082 SERVER_ID=server-2 go run main.go\n"
            "  PORT=8083 SERVER_ID=server-3 go run main.go"
        )

    return available


# ==============================================================================
# 测试类 1: 基础功能测试
# ==============================================================================

class TestBasicFunctionality:
    """测试基础功能：登录、Session、Cookie"""

    def test_login_success(self, base_url, check_server_running):
        """测试 1.1: 登录成功并返回 Cookie"""
        resp = requests.post(
            f"{base_url}/login",
            json={"username": "alice", "password": "123456"}
        )

        assert resp.status_code == 200, "登录应该返回 200"
        assert "session_id" in resp.cookies, "应该返回 session_id Cookie"

        session_id = resp.cookies["session_id"]
        assert len(session_id) > 0, "Session ID 不能为空"

    def test_profile_with_valid_session(self, base_url, check_server_running, new_session):
        """测试 1.2: 使用有效 Session 访问 /profile"""
        # 登录
        resp = new_session.post(
            f"{base_url}/login",
            json={"username": "bob", "password": "123456"}
        )
        assert resp.status_code == 200

        # 访问 profile
        resp = new_session.get(f"{base_url}/profile")

        assert resp.status_code == 200, "应该成功获取用户信息"

        data = resp.json()
        assert "username" in data, "响应应包含 username"
        assert "server_id" in data, "响应应包含 server_id"
        assert "login_time" in data, "响应应包含 login_time"
        assert data["username"] == "bob", "用户名应该是 bob"

    def test_profile_without_cookie(self, base_url, check_server_running):
        """测试 1.3: 无 Cookie 访问 /profile 应该失败"""
        resp = requests.get(f"{base_url}/profile")

        assert resp.status_code == 401, "无 Cookie 应该返回 401"

        data = resp.json()
        assert "error" in data, "响应应包含错误信息"
        assert data["error"] == "Not authenticated", "错误信息应该是 Not authenticated"

    def test_profile_with_invalid_session_id(self, base_url, check_server_running):
        """测试 1.4: 使用无效的 Session ID 应该失败"""
        fake_session_id = "invalid-session-id-12345"
        resp = requests.get(
            f"{base_url}/profile",
            cookies={"session_id": fake_session_id}
        )

        assert resp.status_code == 401, "无效 Session ID 应该返回 401"

        data = resp.json()
        assert "error" in data, "响应应包含错误信息"
        assert data["error"] == "Session not found", "错误信息应该是 Session not found"

    def test_login_response_contains_server_info(self, base_url, check_server_running):
        """测试 1.5: 登录响应包含服务器信息（可选）"""
        resp = requests.post(
            f"{base_url}/login",
            json={"username": "charlie", "password": "123456"}
        )

        assert resp.status_code == 200
        data = resp.json()
        # 这个测试是可选的，取决于你的实现
        # assert "server_id" in data, "响应可以包含 server_id"


# ==============================================================================
# 测试类 2: Cookie 机制测试
# ==============================================================================

class TestCookieMechanism:
    """测试 Cookie 机制和属性"""

    def test_cookie_attributes(self, base_url, check_server_running):
        """测试 2.1: 验证 Cookie 的属性设置"""
        resp = requests.post(
            f"{base_url}/login",
            json={"username": "alice", "password": "123456"}
        )

        set_cookie_header = resp.headers.get("Set-Cookie", "")

        # 验证 Cookie 属性
        assert "session_id=" in set_cookie_header, "Cookie 名称应该是 session_id"
        assert "Path=/" in set_cookie_header, "Path 应该设置为 /"
        assert "Max-Age=3600" in set_cookie_header, "Max-Age 应该是 3600"
        assert "HttpOnly" in set_cookie_header, "应该设置 HttpOnly 标志"

    def test_session_object_auto_cookie_management(self, base_url, check_server_running):
        """测试 2.2: requests.Session 自动管理 Cookie"""
        session = requests.Session()

        # 登录前 - 没有 Cookie
        assert len(session.cookies) == 0, "登录前不应该有 Cookie"

        # 登录
        session.post(
            f"{base_url}/login",
            json={"username": "dave", "password": "123456"}
        )

        # 登录后 - 应该有 session_id Cookie
        assert "session_id" in session.cookies, "登录后应该自动保存 Cookie"

        # 访问 profile - Session 对象应该自动携带 Cookie
        resp = session.get(f"{base_url}/profile")
        assert resp.status_code == 200, "Session 对象应该自动携带 Cookie"


# ==============================================================================
# 测试类 3: 多服务器 Session 隔离测试
# ==============================================================================

class TestMultiServerIsolation:
    """测试多服务器环境下的 Session 隔离"""

    def test_session_isolation_between_servers(self, check_multi_servers):
        """测试 3.1: Session 在不同服务器间应该隔离"""
        servers = check_multi_servers
        server1 = servers[0]
        server2 = servers[1]

        # 在 Server-1 登录
        resp = requests.post(
            f"{server1}/login",
            json={"username": "alice", "password": "123456"}
        )
        assert resp.status_code == 200

        session_id = resp.cookies.get("session_id")
        assert session_id is not None, "应该获取到 Session ID"

        # 访问 Server-1（应该成功）
        resp1 = requests.get(
            f"{server1}/profile",
            cookies={"session_id": session_id}
        )
        assert resp1.status_code == 200, "Server-1 应该找到 Session"

        # 访问 Server-2（应该失败 - Session 隔离）
        resp2 = requests.get(
            f"{server2}/profile",
            cookies={"session_id": session_id}
        )
        assert resp2.status_code == 401, "Server-2 不应该找到 Session（Session 隔离）"

    def test_debug_sessions_endpoint(self, check_multi_servers):
        """测试 3.2: /debug/sessions 接口返回正确的 Session 信息"""
        servers = check_multi_servers

        # 在每个服务器上创建 Session
        for i, server in enumerate(servers):
            requests.post(
                f"{server}/login",
                json={"username": f"user{i}", "password": "123456"}
            )

        # 验证每个服务器的 Session
        for server in servers:
            resp = requests.get(f"{server}/debug/sessions")
            assert resp.status_code == 200

            data = resp.json()
            assert "server_id" in data, "应该包含 server_id"
            assert "session_count" in data, "应该包含 session_count"
            assert "sessions" in data, "应该包含 sessions 列表"
            assert data["session_count"] >= 1, "应该至少有 1 个 Session"

    def test_session_distribution_across_servers(self, check_multi_servers):
        """测试 3.3: 验证 Session 在服务器间的分布"""
        servers = check_multi_servers

        # 在多个服务器上创建 Session
        created_sessions = {}
        for i in range(6):
            server = servers[i % len(servers)]
            resp = requests.post(
                f"{server}/login",
                json={"username": f"user{i}", "password": "123456"}
            )
            session_id = resp.cookies.get("session_id")
            created_sessions[session_id] = server

        # 验证 Session 只存在于创建它的服务器上
        for session_id, creator_server in created_sessions.items():
            for server in servers:
                resp = requests.get(
                    f"{server}/profile",
                    cookies={"session_id": session_id}
                )

                if server == creator_server:
                    assert resp.status_code == 200, f"创建服务器应该找到 Session"
                else:
                    assert resp.status_code == 401, f"其他服务器不应该找到 Session"


# ==============================================================================
# 测试类 4: Session 数据一致性
# ==============================================================================

class TestSessionDataConsistency:
    """测试 Session 数据的一致性"""

    def test_session_data_persistence(self, base_url, check_server_running, new_session):
        """测试 4.1: Session 数据在多次请求间保持一致"""
        # 登录
        resp = new_session.post(
            f"{base_url}/login",
            json={"username": "frank", "password": "123456"}
        )
        assert resp.status_code == 200

        # 多次访问 profile，验证数据一致
        results = []
        for _ in range(5):
            resp = new_session.get(f"{base_url}/profile")
            assert resp.status_code == 200
            results.append(resp.json())

        # 验证所有响应的用户名都一致
        usernames = [r["username"] for r in results]
        assert len(set(usernames)) == 1, "所有请求返回的用户名应该一致"
        assert usernames[0] == "frank", "用户名应该是 frank"

    def test_different_users_different_sessions(self, base_url, check_server_running):
        """测试 4.2: 不同用户应该有不同的 Session"""
        # 用户1登录
        session1 = requests.Session()
        session1.post(
            f"{base_url}/login",
            json={"username": "user1", "password": "123456"}
        )

        # 用户2登录
        session2 = requests.Session()
        session2.post(
            f"{base_url}/login",
            json={"username": "user2", "password": "123456"}
        )

        # 获取 Session ID
        sid1 = session1.cookies.get("session_id")
        sid2 = session2.cookies.get("session_id")

        assert sid1 != sid2, "不同用户应该有不同的 Session ID"

        # 验证各自能访问自己的数据
        resp1 = session1.get(f"{base_url}/profile")
        resp2 = session2.get(f"{base_url}/profile")

        assert resp1.json()["username"] == "user1"
        assert resp2.json()["username"] == "user2"


# ==============================================================================
# Pytest 配置和钩子
# ==============================================================================

def pytest_configure(config):
    """Pytest 配置"""
    config.addinivalue_line(
        "markers", "slow: 标记慢速测试"
    )
    config.addinivalue_line(
        "markers", "multi_server: 需要多服务器的测试"
    )


# 为多服务器测试添加标记
pytest.mark.multi_server(TestMultiServerIsolation)


# ==============================================================================
# 测试报告钩子（可选）
# ==============================================================================

@pytest.hookimpl(tryfirst=True, hookwrapper=True)
def pytest_runtest_makereport(item, call):
    """在测试报告中添加额外信息"""
    outcome = yield
    rep = outcome.get_result()

    # 在失败的测试中添加服务器信息
    if rep.when == "call" and rep.failed:
        if hasattr(item, 'funcargs'):
            if 'base_url' in item.funcargs:
                rep.sections.append(("Server", item.funcargs['base_url']))
