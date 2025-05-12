import sys
import asyncio
from jsinspector import JSInspector, inspect_html_file

# 异步使用示例
async def async_example(file_path):
    print(f"检查文件: {file_path}")

    # 使用异步上下文管理器
    async with JSInspector(headless=True) as inspector:
        errors = await inspector.inspect_file(file_path)

        if not errors:
            print("没有发现JavaScript错误")
        else:
            print(f"发现 {len(errors)} 个JavaScript错误:")
            for i, error in enumerate(errors, 1):
                print(f"\n错误 #{i}:")
                print(f"消息: {error.message}")
                if error.line_number > 0:
                    print(f"位置: 行 {error.line_number}, 列 {error.column_number}")
                if error.stack_trace:
                    print(f"堆栈跟踪: {error.stack_trace}")

# 同步使用示例
def sync_example(file_path):
    print(f"检查文件: {file_path}")

    try:
        # 使用同步辅助函数
        errors = inspect_html_file(file_path)

        if not errors:
            print("没有发现JavaScript错误")
        else:
            print(f"发现 {len(errors)} 个JavaScript错误:")
            for i, error in enumerate(errors, 1):
                print(f"\n错误 #{i}:")
                print(f"消息: {error.message}")
                if error.line_number > 0:
                    print(f"位置: 行 {error.line_number}, 列 {error.column_number}")
                if error.stack_trace:
                    print(f"堆栈跟踪: {error.stack_trace}")
    except Exception as e:
        print(f"检查时出错: {e}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("请提供HTML文件的绝对路径")
        # /mnt/h/code/test_project/tetris2.html
        print("用法: python run_jsinspector.py /path/to/your/file.html")
        sys.exit(1)

    html_file_path = sys.argv[1]

    # 使用同步版本
    sync_example(html_file_path)

    # 或者使用异步版本
    # asyncio.run(async_example(html_file_path))