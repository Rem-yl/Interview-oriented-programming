"""
Redis Session 测试套件 - 使用 pytest

==============================================================================
如何运行服务
==============================================================================

前置条件: 启动 Redis
    docker run -d --name redis -p 6379:6379 redis:alpine

方式 1: 运行单个服务器（用于基础功能测试）
    cd ../redis-session
    PORT=8091 SERVERID=server-1 go run main.go

方式 2: 运行多个服务器（用于完整测试套件）
    # 终端 1
    cd ../redis-session
    PORT=8091 SERVERID=server-1 go run main.go

    # 终端 2
    PORT=8092 SERVERID=server-2 go run main.go

    # 终端 3
    PORT=8093 SERVERID=server-3 go run main.go

可选: 启动 Nginx (Round Robin)
    cd ..
    docker run -d --name nginx-redis -p 8090:80 \
      -v $(pwd)/docker/nginx-redis.conf:/etc/nginx/conf.d/default.conf:ro \
      nginx:alpine

==============================================================================
如何运行测试
==============================================================================

前置条件:
    pip install -r requirements.txt

基础用法:
    pytest test_redis_session.py -v                 # 详细输出
    pytest test_redis_session.py -v -s              # 显示 print 输出

运行特定测试:
    pytest test_redis_session.py::TestBasicFunctionality -v
    pytest test_redis_session.py::TestBasicFunctionality::test_login_success -v

使用标记:
    pytest test_redis_session.py -m multi_server -v    # 只运行多服务器测试

生成报告:
    pytest test_redis_session.py --html=report.html --self-contained-html

查看 Redis 数据:
    redis-cli
    KEYS session:*              # 查看所有 Session
    GET session:<uuid>          # 查看具体 Session
    TTL session:<uuid>          # 查看剩余过期时间

==============================================================================
测试说明
==============================================================================

测试覆盖范围:
- 10+ 个测试用例
- 3 个测试类：基础功能、跨服务器共享、Redis 特性

注意事项:
- 必须先启动 Redis
- 单服务器测试需要至少 1 个服务器运行（端口 8091）
- 跨服务器测试需要至少 2 个服务器运行（端口 8091, 8092, 8093）

与 Sticky Session 的区别:
- 使用 Round Robin 负载均衡（不需要 ip_hash）
- 所有服务器都能获取到 Session
- 服务器宕机不影响 Session
"""

import pytest
import requests
import redis
import json


# ==============================================================================
# Fixtures - 测试准备和清理
# ==============================================================================

@pytest.fixture(scope="session")
def base_url():
    """单服务器基础URL"""
    return "http://localhost:8091"


@pytest.fixture(scope="session")
def multi_server_urls():
    """多服务器URL列表"""
    return [
        "http://localhost:8091",
        "http://localhost:8092",
        "http://localhost:8093"
    ]


@pytest.fixture(scope="session")
def redis_client():
    """Redis 客户端"""
    client = redis.Redis(host='localhost', port=6379, db=0, decode_responses=True)
    yield client
    # 测试结束后清理（可选）
    # client.flushdb()


@pytest.fixture(scope="function")
def new_session():
    """为每个测试创建新的 requests.Session 对象"""
    session = requests.Session()
    yield session
    session.close()


@pytest.fixture(scope="session")
def check_redis_running(redis_client):
    """检查 Redis 是否运行"""
    try:
        redis_client.ping()
        return True
    except redis.exceptions.ConnectionError:
        pytest.skip("Redis 未运行，请先启动: docker run -d --name redis -p 6379:6379 redis:alpine")


@pytest.fixture(scope="session")
def check_server_running(base_url, check_redis_running):
    """检查服务器是否运行"""
    try:
        resp = requests.get(f"{base_url}/profile", timeout=2)
        return True
    except requests.exceptions.ConnectionError:
        pytest.skip("服务器未运行，请先启动: PORT=8091 SERVERID=server-1 go run main.go")
    except Exception as e:
        pytest.skip(f"服务器检查失败: {e}")


@pytest.fixture(scope="session")
def check_multi_servers(multi_server_urls, check_redis_running):
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
            "  PORT=8091 SERVERID=server-1 go run main.go\n"
            "  PORT=8092 SERVERID=server-2 go run main.go\n"
            "  PORT=8093 SERVERID=server-3 go run main.go"
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
        assert "sessionID" in resp.cookies, "应该返回 sessionID Cookie"

        session_id = resp.cookies["sessionID"]
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
        assert "sessionID" in data, "响应应包含 sessionID"
        assert "serverID" in data, "响应应包含 serverID"
        assert "loginTime" in data, "响应应包含 loginTime"

    def test_profile_without_cookie(self, base_url, check_server_running):
        """测试 1.3: 无 Cookie 访问 /profile 应该失败"""
        resp = requests.get(f"{base_url}/profile")

        assert resp.status_code == 401, "无 Cookie 应该返回 401"

        data = resp.json()
        assert "error" in data, "响应应包含错误信息"

    def test_profile_with_invalid_session_id(self, base_url, check_server_running):
        """测试 1.4: 使用无效的 Session ID 应该失败"""
        fake_session_id = "invalid-session-id-12345"
        resp = requests.get(
            f"{base_url}/profile",
            cookies={"sessionID": fake_session_id}
        )

        assert resp.status_code == 401, "无效 Session ID 应该返回 401"

        data = resp.json()
        assert "error" in data, "响应应包含错误信息"


# ==============================================================================
# 测试类 2: 跨服务器 Session 共享测试
# ==============================================================================

class TestCrossServerSharing:
    """测试跨服务器 Session 共享 - Redis Session 的核心特性"""

    @pytest.mark.multi_server
    def test_session_sharing_between_servers(self, check_multi_servers):
        """测试 2.1: Session 在不同服务器间应该共享"""
        servers = check_multi_servers
        server1 = servers[0]
        server2 = servers[1]

        # 在 Server-1 登录
        resp = requests.post(
            f"{server1}/login",
            json={"username": "alice", "password": "123456"}
        )
        assert resp.status_code == 200

        session_id = resp.cookies.get("sessionID")
        assert session_id is not None, "应该获取到 Session ID"

        # 访问 Server-1（应该成功）
        resp1 = requests.get(
            f"{server1}/profile",
            cookies={"sessionID": session_id}
        )
        assert resp1.status_code == 200, "Server-1 应该找到 Session"

        # 访问 Server-2（应该也成功 - 这是与 Sticky Session 的关键区别）
        resp2 = requests.get(
            f"{server2}/profile",
            cookies={"sessionID": session_id}
        )
        assert resp2.status_code == 200, "Server-2 也应该找到 Session（Session 共享）"

        # 验证两个响应的 sessionID 相同
        data1 = resp1.json()
        data2 = resp2.json()
        assert data1["sessionID"] == data2["sessionID"], "两个服务器返回的 sessionID 应该相同"

    @pytest.mark.multi_server
    def test_round_robin_with_redis_session(self, check_multi_servers):
        """测试 2.2: Round Robin + Redis Session（所有请求都应该成功）"""
        servers = check_multi_servers

        # 登录到 Server-1
        session = requests.Session()
        resp = session.post(
            f"{servers[0]}/login",
            json={"username": "test_user", "password": "123456"}
        )
        assert resp.status_code == 200

        # 手动轮询访问不同服务器（模拟 Round Robin）
        success_count = 0
        for i in range(len(servers) * 2):  # 每个服务器访问 2 次
            server = servers[i % len(servers)]
            resp = session.get(f"{server}/profile")

            if resp.status_code == 200:
                success_count += 1

        # 所有请求都应该成功（与 Sticky Session 不同）
        assert success_count == len(servers) * 2, "所有请求都应该成功（Round Robin + Redis）"

    @pytest.mark.multi_server
    def test_created_server_vs_handled_server(self, check_multi_servers):
        """测试 2.3: 创建 Session 的服务器 vs 处理请求的服务器"""
        servers = check_multi_servers
        server1 = servers[0]
        server2 = servers[1]

        # 在 Server-1 登录
        resp = requests.post(
            f"{server1}/login",
            json={"username": "alice", "password": "123456"}
        )
        session_id = resp.cookies.get("sessionID")

        # 在 Server-2 访问
        resp = requests.get(
            f"{server2}/profile",
            cookies={"sessionID": session_id}
        )

        data = resp.json()

        # serverID: 创建 Session 的服务器
        # handledBy: 处理此请求的服务器
        # 在 Redis Session 中，这两个可以不同
        assert "serverID" in data, "应该包含创建 Session 的服务器 ID"
        assert "handledBy" in data, "应该包含处理请求的服务器 ID"


# ==============================================================================
# 测试类 3: Redis 特性测试
# ==============================================================================

class TestRedisFeatures:
    """测试 Redis 特性：TTL、续期、数据格式"""

    def test_session_stored_in_redis(self, base_url, check_server_running, redis_client, new_session):
        """测试 3.1: Session 应该存储在 Redis 中"""
        # 登录
        resp = new_session.post(
            f"{base_url}/login",
            json={"username": "redis_test", "password": "123456"}
        )
        session_id = resp.cookies.get("sessionID")

        # 检查 Redis 中是否存在
        redis_key = f"session:{session_id}"
        exists = redis_client.exists(redis_key)
        assert exists == 1, f"Session 应该存储在 Redis 中: {redis_key}"

        # 获取 Redis 数据
        data = redis_client.get(redis_key)
        assert data is not None, "应该能从 Redis 获取 Session 数据"

        # 验证数据格式
        session_data = json.loads(data)
        assert "session_id" in session_data
        assert "login_time" in session_data
        assert "server_id" in session_data

    def test_session_ttl(self, base_url, check_server_running, redis_client, new_session):
        """测试 3.2: Session 应该有 TTL（30 分钟）"""
        # 登录
        resp = new_session.post(
            f"{base_url}/login",
            json={"username": "ttl_test", "password": "123456"}
        )
        session_id = resp.cookies.get("sessionID")

        # 检查 TTL
        redis_key = f"session:{session_id}"
        ttl = redis_client.ttl(redis_key)

        # TTL 应该接近 1800 秒（30 分钟）
        assert ttl > 1700 and ttl <= 1800, f"TTL 应该接近 1800 秒，实际: {ttl}"

    def test_session_renewal(self, base_url, check_server_running, redis_client, new_session):
        """测试 3.3: 访问 Session 应该续期"""
        import time

        # 登录
        resp = new_session.post(
            f"{base_url}/login",
            json={"username": "renewal_test", "password": "123456"}
        )
        session_id = resp.cookies.get("sessionID")
        redis_key = f"session:{session_id}"

        # 获取初始 TTL
        ttl_before = redis_client.ttl(redis_key)

        # 等待 3 秒（让 TTL 减少）
        time.sleep(3)

        # 在访问前检查 TTL 已经减少
        ttl_before_access = redis_client.ttl(redis_key)
        assert ttl_before_access < ttl_before, f"等待后 TTL 应该减少: {ttl_before} -> {ttl_before_access}"

        # 访问 profile（应该触发续期）
        new_session.get(f"{base_url}/profile")

        # 获取续期后的 TTL
        ttl_after = redis_client.ttl(redis_key)

        # TTL 应该被重置为接近 1800（比访问前的 TTL 更大）
        assert ttl_after > ttl_before_access, f"访问后 TTL 应该被续期: {ttl_before_access} -> {ttl_after}"
        assert ttl_after > 1700, f"续期后的 TTL 应该接近 1800，实际: {ttl_after}"


# ==============================================================================
# Pytest 配置
# ==============================================================================

def pytest_configure(config):
    """Pytest 配置"""
    config.addinivalue_line(
        "markers", "multi_server: 需要多服务器的测试"
    )
