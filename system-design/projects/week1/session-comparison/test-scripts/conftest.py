"""
Pytest 共享配置和 Fixtures
"""

import pytest


def pytest_addoption(parser):
    """添加自定义命令行选项"""
    parser.addoption(
        "--base-port",
        action="store",
        default="8081",
        help="单服务器端口（默认: 8081）"
    )
    parser.addoption(
        "--multi-ports",
        action="store",
        default="8081,8082,8083",
        help="多服务器端口列表，逗号分隔（默认: 8081,8082,8083）"
    )
    parser.addoption(
        "--skip-multi",
        action="store_true",
        default=False,
        help="跳过多服务器测试"
    )


@pytest.fixture(scope="session")
def config(request):
    """获取命令行配置"""
    return {
        "base_port": request.config.getoption("--base-port"),
        "multi_ports": request.config.getoption("--multi-ports").split(","),
        "skip_multi": request.config.getoption("--skip-multi"),
    }


def pytest_collection_modifyitems(config, items):
    """修改测试收集：根据配置跳过某些测试"""
    if config.getoption("--skip-multi"):
        skip_multi = pytest.mark.skip(reason="--skip-multi 选项已启用")
        for item in items:
            if "multi_server" in item.keywords:
                item.add_marker(skip_multi)


def pytest_report_header(config):
    """在测试报告头部添加信息"""
    return [
        f"Base Port: {config.getoption('--base-port')}",
        f"Multi Ports: {config.getoption('--multi-ports')}",
        f"Skip Multi-Server Tests: {config.getoption('--skip-multi')}",
    ]
