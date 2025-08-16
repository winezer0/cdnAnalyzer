import ipaddress
import json
from collections import Counter
from collections import defaultdict
from typing import List
import tldextract

from file_utils import read_csv_to_dict


def ip_list_to_cidrs_optimal(ip_list: List[str]) -> List[str]:
    """
    将 IP 列表合并为最优的 CIDR 表示（全部输出为 CIDR 格式，如 /32, /31, /30）。

    特点：
    - 自动去重、排序、验证 IPv4
    - 使用 collapse_addresses 智能聚合（支持跨缺失 IP 合并，如 /31）
    - 输出统一为 CIDR 格式，不再返回 'a-b' 范围
    - 高效、标准、适合防火墙/安全策略使用

    Args:
        ip_list (List[str]): IP 地址列表，例如 ['1.1.1.1', '1.1.1.2', '1.1.1.3']

    Returns:
        List[str]: CIDR 网段列表，例如 ['1.1.1.1/32', '1.1.1.2/31']
    """
    if not ip_list:
        return []

    # 解析并去重 IPv4 地址
    valid_ips = set()
    for ip_str in ip_list:
        try:
            ip = ipaddress.IPv4Address(ip_str.strip())
            valid_ips.add(ip)
        except Exception:
            print(f"跳过无效 IP: {ip_str}")
            continue

    if not valid_ips:
        return []

    # 转为 /32 网络列表
    networks = [ipaddress.ip_network(f"{ip}/32") for ip in valid_ips]

    # 使用 collapse_addresses 自动合并为最小 CIDR（会生成 /31, /30 等）
    collapsed = list(ipaddress.collapse_addresses(networks))

    # 转为字符串并排序（按网络地址）
    return sorted([str(net) for net in collapsed], key=lambda x: ipaddress.ip_network(x).network_address)
def ip_list_split_by_c(ip_list: List[str], max_missing: int = 1) -> List[str]:
    """按 C 段分组后，分别进行范围或 CIDR 合并"""
    if not ip_list:
        return []

    # 按 /24 分组
    groups = defaultdict(list)
    for ip_str in ip_list:
        try:
            ip = ipaddress.ip_address(ip_str.strip())
            if isinstance(ip, ipaddress.IPv4Address):
                c_net = ipaddress.ip_network(f"{ip}/24", strict=False)
                groups[str(c_net.network_address)].append(ip_str)
        except:
            continue
    return groups

def extract_main_domain(url_or_domain):
    # 从给定的 URL 或域名中提取主域名 (注册域名)。
    try:
        # tldextract 会处理 URL 和域名，并自动处理 PSL
        extracted = tldextract.extract(url_or_domain)
        # 主域名是 subdomain 为空时的 domain + suffix
        main_domain = f"{extracted.domain}.{extracted.suffix}"
        # 如果 domain 或 suffix 为空，可能输入无效
        if extracted.domain and extracted.suffix:
            return main_domain
        else:
            return None
    except Exception as e:
        print(f"提取域名时出错: {e}")
        return None


def format_string_list(string):
    # 将格式如 "[a b c]" 或 "a b c" 的字符串转换为去重后的列表。
    if not isinstance(string, str):
        return []

    # 去除中括号并清理空白
    cleaned = string.strip().strip("[]")
    if not cleaned:
        return []

    # 按空白字符分割，过滤空字符串，去重并保持相对顺序
    items = cleaned.split()
    return list(set(items))


def format_main_domains(domains):
    if len(domains) == 0:
        return domains

    seen = set()
    for doamin in domains:
        main_domain = extract_main_domain(doamin)
        seen.add(main_domain)
    return list(seen)


def merge_company_ips(clean_dicts):
    company_ips_dict = {}
    for clean_dict in clean_dicts:
        company = clean_dict["CdnCompany"]
        ip_list = clean_dict["A"]
        if company not in company_ips_dict:
            company_ips_dict[company] = []
        company_ips_dict[company].extend(ip_list)
    return company_ips_dict

def merge_cname_ips(clean_dicts):
    cname_ips_dict = {}
    for clean_dict in clean_dicts:
        company = clean_dict["CNAME"]
        ip_list = clean_dict["A"]
        if company not in cname_ips_dict:
            cname_ips_dict[company] = []
        cname_ips_dict[company].extend(ip_list)
    return cname_ips_dict

def format_dicts_ips_cname_to_list(raw_dicts):
    format_dicts = []
    for need_dict in raw_dicts:
        ip_list = format_string_list(need_dict["A"])
        cname_list = format_main_domains(format_string_list(need_dict["CNAME"]))
        need_dict["A"] = ip_list
        need_dict["CNAME"] = cname_list
        format_dicts.append(need_dict)
    return format_dicts

def split_one_more_cname_dicts(format_dicts):
    one_cname_dicts = []
    more_cname_dicts = []
    for format_dict in format_dicts:
        cname_list = format_dict["CNAME"]
        if len(cname_list) > 1:
            print(f"存在多 cname_list:{cname_list} -> {format_dict}")
            more_cname_dicts.append(format_dict)
        elif len(cname_list) == 1:
            format_dict["CNAME"] = cname_list[0]
            one_cname_dicts.append(format_dict)
    return one_cname_dicts,more_cname_dicts

def filter_dicts_when_value_eqs_any_keys(csv_dicts, except_keys = ["IsCdn"], except_values = ["true"]):
    need_dicts = []
    for csv_dict in csv_dicts:
        if any(str(csv_dict.get(key)).lower() == str(except_value).lower() for key in except_keys for except_value in except_values):
                need_dicts.append(csv_dict)
    return need_dicts

def filter_dicts_when_value_no_empty(csv_dicts, except_keys = ["CNAME"]):
    need_dicts = []
    for csv_dict in csv_dicts:
        if all(csv_dict.get(key, None) for key in except_keys):
                need_dicts.append(csv_dict)
    return need_dicts


def filter_dicts_when_value_has_any_keys(csv_dicts, except_keys = ["CNAME"], except_values = ["cdn"]):
    need_dicts = []
    for csv_dict in csv_dicts:
        keys1 = [str(csv_dict.get(key)) for key in except_keys]
        if any(str(value).lower() in str(keys1).lower() for value in except_values):
                need_dicts.append(csv_dict)
    return need_dicts

def filter_dicts_by_keys(csv_dicts, store_keys = ["CNAME"]):
    new_dicts = []
    for csv_dict in csv_dicts:
        new_dict = {}
        for key in store_keys:
            new_dict[key] = csv_dict.get(key, '')
        new_dicts.append(new_dict)
    return new_dicts

def filter_by_ip_frequency_limit(company_ips_dict, limit = 3):
    new_company_ips_dict = {}
    for cname, ip_list in company_ips_dict.items():
        ip_frequency = Counter(ip_list)
        new_ip_list = []
        for ip, count in ip_frequency.items():
            if ip != '127.0.0.1' and count >= limit:
                new_ip_list.append(ip)
        if len(new_ip_list) > 0:
            new_company_ips_dict[cname] = new_ip_list
    return new_company_ips_dict


def dict_ip_list_to_ip_cidr_list(company_ips_dict):
    company_cidr_list_dict = {}
    for company, ip_list in company_ips_dict.items():
        cip_list = ip_list_split_by_c(ip_list)
        # print(f"cname:{cname} ip_list:{ip_list} cip_list:{cip_list}")
        all_results = []
        for cid, ips in cip_list.items():
            cidr_results = ip_list_to_cidrs_optimal(ips)
            all_results.extend(cidr_results)
            # print(f"cid:{cid} cidr list:{all_results}")
        all_results = list(set(all_results))
        company_cidr_list_dict[company] = sorted(all_results)
    return company_cidr_list_dict


def get_dicts_values(cdn_dicts, key="CNAME"):
    all_results = []
    for cdn_dict in cdn_dicts:
        if key in cdn_dict.keys():
            results = cdn_dict.get(key)
            if isinstance(results, list):
                all_results.extend(results)
            else:
                all_results.append(results)
    return list(set(all_results))

def init_cdn_cname_info(cdn_dicts):
    # 创建CDN CNAMEs的全部集合
    all_cname_list = []
    # 创建CDN CNAMEs 和 CDN company直接的映射字典
    cname_company_dict = {}
    for is_cdn_dict in cdn_dicts:
        cname_list = format_main_domains(format_string_list(is_cdn_dict["CNAME"]))
        all_cname_list.extend(cname_list)
        # 填充CNAME和Company关系
        cdn_company = is_cdn_dict["CdnCompany"]
        for cname in cname_list:
            if cname not in cname_company_dict:
                cname_company_dict[cname] = []
            cname_company_dict[cname].append(cdn_company)
    return cname_company_dict, all_cname_list

if __name__ == '__main__':
    print(f'Hi')  # 按 Ctrl+F8 切换断点。
    csv_file = r"C:\Users\WINDOWS\Desktop\2.csv"
    csv_dicts = read_csv_to_dict(csv_file, mode="r", encoding=None, delimiter=",")

    CDN_CNAME_KEYS = ["cdn", "cloud", "waf", "dns", "yun", "dos", "dun"]
    if True:
        # 处理company已确认的CDN结果
        is_cdn_dicts = filter_dicts_when_value_eqs_any_keys(csv_dicts, except_keys = ["IsCdn"], except_values = ["true"])
        # 只处理包含CNAME的项, 没有CNAME还有资产，说明就是通过IP查询出来的
        is_cdn_dicts = filter_dicts_when_value_no_empty(is_cdn_dicts, except_keys = ["CNAME"])
        # 初始化 提前获取所有CName和企业对应关系
        CNAME_COMPANY_DICT, all_cname_list = init_cdn_cname_info(is_cdn_dicts)
        CNAME_FREQUENCY = dict(Counter(all_cname_list))

        # 整理数据，只保留仅一个 cname 的数据， 对于多个cname的数据进行输出提示
        format_is_cdn_dicts = format_dicts_ips_cname_to_list(is_cdn_dicts)
        one_cname_cdn_dicts, _ = split_one_more_cname_dicts(format_is_cdn_dicts)

        # 按 company 合并IP列表
        company_ips_dict = merge_company_ips(one_cname_cdn_dicts)
        # 统计IP出现的频率，当IP出现频率大于等于3时才作为CDN的IP处理
        company_ips_dict = filter_by_ip_frequency_limit(company_ips_dict, limit=3)
        # 整理每个CNAME的IP列表为CIDR格式
        company_cidr_list_dict =  dict_ip_list_to_ip_cidr_list(company_ips_dict)

        # 整理出非常疑似 CDN Cname 的项
        like_cdn_cname_list = []
        # 当cname中包含关键字或者频率较高时作为CDN的 cnames 处理
        for cname, count in CNAME_FREQUENCY.items():
            if str(cname).lower() in str(CDN_CNAME_KEYS).lower():
                like_cdn_cname_list.append(cname)
            if count > 2:
                like_cdn_cname_list.append(cname)

        # 将所有疑似Cdn的cname添加到数据中
        should_cdn_cname_dict = {}
        for like_cdn_cname in like_cdn_cname_list:
            cdn_company_list = CNAME_COMPANY_DICT.get(cname)
            cdn_company_list = list(set(cdn_company_list))
            if len(cdn_company_list) == 1:
                cdn_company = cdn_company_list[0]
                if cdn_company not in should_cdn_cname_dict:
                    should_cdn_cname_dict[cdn_company] = []
                should_cdn_cname_dict[cdn_company].append(like_cdn_cname)

        should_added_data = {
            "cdn": { "ip": company_cidr_list_dict, "cname":should_cdn_cname_dict},
            "waf": {},
            "cloud": {}
        }

        # 输出为 Json 格式字符串
        with open("should_added.json", "w", encoding="utf-8") as f:
            json.dump(should_added_data, f, indent=2, allow_nan=True, ensure_ascii=False)


    if True:
        # 提取没有 CdnCompany 但是存在CNAME 的项，
        none_company_dicts = filter_dicts_when_value_eqs_any_keys(csv_dicts, except_keys= ["CdnCompany"], except_values = [""])
        none_company_dicts = filter_dicts_when_value_no_empty(none_company_dicts, except_keys = ["CNAME"])

        # 提取 其中 IpSizeIsCdn 为 true 的项目，比较可能是CDN
        ip_size_like_cdn_dicts = filter_dicts_when_value_eqs_any_keys(none_company_dicts, except_keys= ["IpSizeIsCdn"], except_values = ["true"])
        ip_size_like_cdn_dicts = filter_dicts_by_keys(ip_size_like_cdn_dicts, store_keys = ["A", "CNAME"])
        # print(f"ip_size_like_cdn_dicts: {len(ip_size_like_cdn_dicts)}")

        # 提取 其中 cname 为 中包含 cdn|waf|dns关键字的项目，比较可能是CDN
        cname_like_cdn_dicts = filter_dicts_when_value_has_any_keys(none_company_dicts, except_keys= ["CNAME"],  except_values=CDN_CNAME_KEYS)
        cname_like_cdn_dicts = filter_dicts_by_keys(cname_like_cdn_dicts, store_keys = ["A", "CNAME"])
        # print(f"cname_like_cdn_dicts: {len(cname_like_cdn_dicts)}")
        possible_cdn_dicts = ip_size_like_cdn_dicts + cname_like_cdn_dicts
        # print(f"possible_cdn_dicts:{possible_cdn_dicts}")
        possible_cdn_dicts = format_dicts_ips_cname_to_list(possible_cdn_dicts)
        # print(f"possible_cdn_dicts:{possible_cdn_dicts}")

        # 将所有疑似Cdn的cname添加到数据中
        possible_cname_dict = {}
        for possible_cdn_dict in possible_cdn_dicts:
            for cname in possible_cdn_dict["CNAME"]:
                possible_cname_dict[cname] = [cname]

        one_cname_possible_cdn_dicts, more_cname_possible_cdn_dicts = split_one_more_cname_dicts(possible_cdn_dicts)

        # 生成CNAMEs IP关系 和 Cnames Cnames 关系
        one_cname_ips_possible_cdn_dict = merge_cname_ips(one_cname_possible_cdn_dicts)
        # print(one_cname_ips_possible_cdn_dict)
        # 整理每个CNAME的IP列表为CIDR格式
        one_cname_cidr_possible_cdn_dict =   dict_ip_list_to_ip_cidr_list(one_cname_ips_possible_cdn_dict)
        # print(one_cname_cidr_possible_cdn_dict)

        possible_data = {
            "cdn": {"cname": possible_cname_dict, "ip": one_cname_cidr_possible_cdn_dict},
            "waf": {},
            "cloud": {}
        }
        print(possible_data)

        # 输出为 Json 格式字符串
        with open("possible_cdn.json", "w", encoding="utf-8") as f:
            json.dump(possible_data, f, indent=2, allow_nan=True, ensure_ascii=False)
