import difflib
import os
import argparse
import base64
import sys

class CodeDiffMerge:
    def __init__(self):
        self.differ = difflib.Differ()
    
    def compare_text(self, text1, text2):
        """比较两段文本的差异"""
        lines1 = text1.splitlines()
        lines2 = text2.splitlines()
        diff = list(self.differ.compare(lines1, lines2))
        return '\n'.join(diff)
    
    def compare_files(self, file1, file2):
        """比较两个文件的差异"""
        with open(file1, 'r', encoding='utf-8') as f1, open(file2, 'r', encoding='utf-8') as f2:
            text1 = f1.read()
            text2 = f2.read()
        return self.compare_text(text1, text2)
    
    def merge_text(self, base, text1, text2):
        """合并两段文本"""
        merger = difflib.Differ()
        base_lines = base.splitlines()
        text1_lines = text1.splitlines()
        text2_lines = text2.splitlines()
        
        diff1 = list(merger.compare(base_lines, text1_lines))
        diff2 = list(merger.compare(base_lines, text2_lines))
        
        merged = []
        for line in diff1:
            if line.startswith('  ') or line.startswith('+ '):
                merged.append(line[2:])
        for line in diff2:
            if line.startswith('+ '):
                merged.append(line[2:])
        
        return '\n'.join(merged)
    
    def merge_files(self, base_file, file1, file2, output_file):
        """合并两个文件并输出到新文件"""
        with open(base_file, 'r', encoding='utf-8') as f_base, \
             open(file1, 'r', encoding='utf-8') as f1, \
             open(file2, 'r', encoding='utf-8') as f2:
            base = f_base.read()
            text1 = f1.read()
            text2 = f2.read()
        
        merged = self.merge_text(base, text1, text2)
        
        with open(output_file, 'w', encoding='utf-8') as f_out:
            f_out.write(merged)
    
    def auto_detect(self, input1, input2):
        """自动检测输入是文件路径还是文本内容"""
        if os.path.isfile(input1) and os.path.isfile(input2):
            return self.compare_files(input1, input2)
        else:
            return self.compare_text(input1, input2)


def main():
    parser = argparse.ArgumentParser(description='代码差异比较和合并工具')
    parser.add_argument('--task', choices=['compare', 'merge', 'auto'], default='auto', 
                        help='执行的任务类型: compare(比较), merge(合并), auto(自动检测)')
    parser.add_argument('--input1', required=True, help='第一个输入(文件路径或Base64编码的文本)')
    parser.add_argument('--input2', required=True, help='第二个输入(文件路径或Base64编码的文本)')
    parser.add_argument('--base', help='合并时的基准文件或Base64编码的基准文本')
    parser.add_argument('--output', help='输出文件路径')
    parser.add_argument('--base64', action='store_true', help='输入是否为Base64编码')
    
    args = parser.parse_args()
    
    diff_merge = CodeDiffMerge()
    
    # 处理Base64编码的输入
    input1 = args.input1
    input2 = args.input2
    base = args.base
    
    if args.base64:
        try:
            input1 = base64.b64decode(input1).decode('utf-8')
            input2 = base64.b64decode(input2).decode('utf-8')
            if base:
                base = base64.b64decode(base).decode('utf-8')
        except Exception as e:
            sys.stderr.write(f"Base64解码错误: {e}\n")
            sys.exit(1)
    
    # 执行任务
    if args.task == 'compare' or (args.task == 'auto' and not args.base):
        if args.base64 or not (os.path.isfile(input1) and os.path.isfile(input2)):
            result = diff_merge.compare_text(input1, input2)
        else:
            result = diff_merge.compare_files(input1, input2)
        
        # 直接输出结果，不添加额外内容
        print(result)
    
    elif args.task == 'merge' or (args.task == 'auto' and args.base):
        if not args.output:
            sys.stderr.write("合并操作需要指定输出文件路径\n")
            sys.exit(1)
            
        if args.base64 or not (os.path.isfile(input1) and os.path.isfile(input2) and os.path.isfile(base)):
            result = diff_merge.merge_text(base, input1, input2)
            with open(args.output, 'w', encoding='utf-8') as f_out:
                f_out.write(result)
        else:
            diff_merge.merge_files(base, input1, input2, args.output)
        
        print(f"合并结果已保存到: {args.output}")


if __name__ == "__main__":
    '''
    # 比较两个文本文件
python script.py --task compare --input1 file1.txt --input2 file2.txt

# 比较两段Base64编码的文本
python script.py --task compare --input1 SGVsbG8= --input2 SGVsbG8gV29ybGQ= --base64

# 合并文件
python script.py --task merge --base base.txt --input1 file1.txt --input2 file2.txt --output merged.txt

# 自动检测模式
python script.py --task auto --input1 file1.txt --input2 file2.txt
    '''
    main()
