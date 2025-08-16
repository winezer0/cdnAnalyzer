#!/usr/bin/env python
# encoding: utf-8
import csv
import os
import os.path


def string_encoding(data: bytes):
    # 简单的判断文件编码类型
    # 说明：UTF兼容ISO8859-1和ASCII，GB18030兼容GBK，GBK兼容GB2312，GB2312兼容ASCII
    CODES = ['UTF-8', 'GB18030', 'BIG5']
    # UTF-8 BOM前缀字节
    UTF_8_BOM = b'\xef\xbb\xbf'

    # 遍历编码类型
    for code in CODES:
        try:
            data.decode(encoding=code)
            if 'UTF-8' == code and data.startswith(UTF_8_BOM):
                return 'UTF-8-SIG'
            return code
        except UnicodeDecodeError:
            continue
    return 'unknown'


def file_encoding(file_path: str):
    # 获取文件编码类型
    if not os.path.exists(file_path):
        return "utf-8"
    with open(file_path, 'rb') as f:
        return string_encoding(f.read())



def path_is_exist(file_path):
    # 判断文件是否存在
    return os.path.exists(file_path) if file_path else False


def file_is_empty(file_path):
    # 判断一个文件是否为空
    return not path_is_exist(file_path) or not os.path.getsize(file_path)


def read_csv_to_dict_auto(csv_file, mode="r", encoding=None):
    """
    读个CSV文件到字典格式(会自动拼接表头)
    :param csv_file:
    :param mode:
    :param encoding:
    :return:
    """
    if file_is_empty(csv_file):
        return None

    # 自动获取文件编码
    encoding = encoding if encoding else file_encoding(csv_file)

    with open(csv_file, mode=mode, encoding=encoding, newline='') as csvfile:
        # 方案1、使用 reader
        # reader = csv.reader(csvfile)
        # title = next(reader)  # 读取第一行
        # row_list = [dict(zip(title, row)) for row in reader]

        # 自动分析分隔符
        dialect = csv.Sniffer().sniff(csvfile.read(1024))
        csvfile.seek(0)

        # 方案2、使用 DictReader
        reader = csv.DictReader(csvfile, dialect=dialect)
        row_list = [row for row in reader]
        return row_list


def read_csv_to_dict_man(csv_file, mode="r", encoding=None, delimiter=","):
    """
    读个CSV文件到字典格式(会自动拼接表头)
    :param csv_file: 文件名
    :param mode: 读取模式
    :param encoding: 文件读取编码
    :param delimiter: CSV分隔符
    :return:
    """
    if file_is_empty(csv_file):
        return None

    # 自动获取文件编码
    encoding = encoding if encoding else file_encoding(csv_file)

    with open(csv_file, mode=mode, encoding=encoding, newline='') as csvfile:
        # 方案1、使用 reader
        reader = csv.reader(csvfile, delimiter=delimiter)
        title = next(reader)  # 读取第一行
        row_list = [dict(zip(title, row)) for row in reader]
        return row_list


def read_csv_to_dict(csv_file, mode="r", encoding=None, delimiter=","):
    try:
        data = read_csv_to_dict_auto(csv_file, mode=mode, encoding=encoding)
    except Exception:
        data = read_csv_to_dict_man(csv_file, mode=mode, encoding=encoding, delimiter=delimiter)
    return data

def auto_make_dir(path, is_file=False):
    # 自动创建目录  如果输入的是文件路径,就创建上一级目录
    directory = os.path.dirname(os.path.abspath(path)) if is_file else path
    # print(f"auto_make_dir:{directory}")
    if not os.path.exists(directory):
        os.makedirs(directory)
        return True
    return False


def write_line(file_path, data_list, encoding="utf-8", new_line=True, mode="w+"):
    auto_make_dir(file_path, is_file=True)
    # 判断输入的是字符串还是列表
    data_list = [data_list] if isinstance(data_list, str) else data_list
    # 文本文件写入数据 默认追加
    with open(file_path, mode=mode, encoding=encoding) as f_open:
        data_list = [f"{data.strip()}\n" for data in data_list] if new_line else data_list
        f_open.writelines(data_list)
        f_open.close()
