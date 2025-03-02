package cdn_ecs

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestAA(t *testing.T) {
	domain := "www.52st.com" // 替换成你要查询的域名

	// 查询 NS 记录
	records, err := net.LookupCNAME(domain)
	if err != nil {
		fmt.Printf("查询 NS 记录失败: %v\n", err)
		return
	}
	fmt.Println(records)

	fmt.Printf("域名 %s 的 NS 服务器地址:\n", domain)
	for _, r := range records {
		fmt.Println(r)
	}
}

func TestDoCDNCheck(t *testing.T) {
	fmt.Println(DoEcsQuery("jss.bastatic.com"))
}

func TestName(t *testing.T) {
	fmt.Println(EDNSQuery("cdn-lbyhhmhj.sched.sma.tdnsstic1.cn.", "22.33.44.0", ToServerAddr("ns1.stc2.tdnsstic1.cn")))
}
func TestLookupNSServer(t *testing.T) {
	type args struct {
		domain string
	}
	tests := []struct {
		name      string
		args      args
		wantAddrs string
		wantErr   bool
	}{
		{"1", args{"www.a.shifen.com"}, "a.shifen.com", false},
		{"2", args{"cdn-lbyhhmhj.sched.sma.tdnsstic1.cn"}, "stc2.tdnsstic1.cn", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAddrs, err := LookupNSServer(tt.args.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupNSServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(gotAddrs[0], tt.wantAddrs) {
				t.Errorf("LookupNSServer() gotAddrs = %v, want %v", gotAddrs, tt.wantAddrs)
			}
		})
	}
}

func TestLookupCNAME(t *testing.T) {
	type args struct {
		domain string
	}
	tests := []struct {
		name      string
		args      args
		wantCname string
		wantErr   bool
	}{
		{"1", args{"www.baidu.com"}, "www.a.shifen.com.", false},
		{"2", args{"www.52st.com"}, "cdn-lbyhhmhj.sched.sma.tdnsstic1.cn.", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCname, err := LookupCNAME(tt.args.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupCNAME() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCname != tt.wantCname {
				t.Errorf("LookupCNAME() gotCname = %v, want %v", gotCname, tt.wantCname)
			}
		})
	}
}

func TestDoCDNCheck1(t *testing.T) {
	type args struct {
		domain string
	}
	tests := []struct {
		name         string
		args         args
		wantIsNotCDN bool
		wantErr      bool
	}{
		{"1", args{"www.baidu.com"}, false, false},
		{"2", args{"pay.52st.com"}, false, false},
		{"3", args{"www.52st.com"}, false, false},
		{"4", args{"mt.52st.com"}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsNotCDN, err := DoCDNCheck(tt.args.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("DoCDNCheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotIsNotCDN != tt.wantIsNotCDN {
				t.Errorf("DoCDNCheck() gotIsNotCDN = %v, want %v", gotIsNotCDN, tt.wantIsNotCDN)
			}
		})
	}
}
