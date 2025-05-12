import os
import asyncio
import tempfile
from dataclasses import dataclass
from typing import List, Optional, Dict, Any
from playwright.async_api import async_playwright, Page, Error

'''
安装库：
pip install playwright
playwright install chromium

'''

@dataclass
class JSError:
    """JavaScript错误类"""
    message: str
    line_number: int = 0
    column_number: int = 0
    url: str = ""
    stack_trace: str = ""


class JSInspector:
    """JavaScript代码检查器，使用无头浏览器检查HTML文件中的JavaScript错误"""

    def __init__(self, timeout: int = 15000, headless: bool = True):
        """
        初始化JavaScript检查器

        Args:
            timeout: 超时时间（毫秒）
            headless: 是否使用无头模式
        """
        self.timeout = timeout
        self.headless = headless
        self.errors = []
        self._playwright = None
        self._browser = None

    async def __aenter__(self):
        """异步上下文管理器入口"""
        await self.start()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """异步上下文管理器退出"""
        await self.close()

    async def start(self):
        """启动playwright和浏览器"""
        self._playwright = await async_playwright().start()
        self._browser = await self._playwright.chromium.launch(headless=self.headless)

    async def close(self):
        """关闭浏览器和playwright"""
        if self._browser:
            await self._browser.close()
            self._browser = None

        if self._playwright:
            await self._playwright.stop()
            self._playwright = None

    async def inspect_file(self, file_path: str) -> List[JSError]:
        """
        检查指定HTML文件中的JavaScript错误

        Args:
            file_path: HTML文件的绝对路径

        Returns:
            检测到的JavaScript错误列表
        """
        if not os.path.exists(file_path):
            raise FileNotFoundError(f"文件不存在: {file_path}")

        if not self._browser:
            await self.start()

        # 重置错误列表
        self.errors = []

        # 创建新的浏览器上下文和页面
        context = await self._browser.new_context()
        page = await context.new_page()

        # 配置错误捕获
        await self._setup_error_listeners(page, file_path)

        try:
            # 导航到文件
            file_url = f"file://{file_path}"
            await page.goto(file_url, timeout=self.timeout)

            # 等待一段时间以确保所有脚本都有机会执行
            await asyncio.sleep(2)

        except Error as e:
            # 捕获页面加载错误
            self.errors.append(JSError(
                message=f"页面加载错误: {str(e)}",
                url=file_path
            ))
        finally:
            # 关闭上下文和页面
            await context.close()

        return self.errors

    async def _setup_error_listeners(self, page: Page, file_path: str):
        """配置页面的错误监听器"""

        # 监听控制台错误
        async def handle_console_message(msg):
            if msg.type == "error":
                self.errors.append(JSError(
                    message=msg.text,
                    url=file_path
                ))

        page.on("console", handle_console_message)

        # 监听页面错误
        async def handle_page_error(error):
            stack = str(error.stack) if hasattr(error, "stack") else ""

            # 尝试从堆栈跟踪中提取行号和列号
            line_number = 0
            column_number = 0

            if stack:
                import re
                # 尝试匹配类似于 "at file:///path/to/file.html:10:15" 的模式
                matches = re.search(r":(\d+):(\d+)", stack)
                if matches:
                    line_number = int(matches.group(1))
                    column_number = int(matches.group(2))

            self.errors.append(JSError(
                message=str(error),
                line_number=line_number,
                column_number=column_number,
                url=file_path,
                stack_trace=stack
            ))

        page.on("pageerror", handle_page_error)

        # 注入JavaScript来捕获未处理的拒绝承诺
        await page.evaluate("""
            window.addEventListener('unhandledrejection', event => {
                console.error('Unhandled promise rejection:', event.reason);
            });
        """)


# 辅助函数：同步版本的检查函数
def inspect_html_file(file_path: str, timeout: int = 15000, headless: bool = True) -> List[JSError]:
    """
    同步版本的HTML文件JavaScript错误检查函数

    Args:
        file_path: HTML文件的绝对路径
        timeout: 超时时间（毫秒）
        headless: 是否使用无头模式

    Returns:
        检测到的JavaScript错误列表
    """
    async def _run():
        async with JSInspector(timeout=timeout, headless=headless) as inspector:
            return await inspector.inspect_file(file_path)

    return asyncio.run(_run())