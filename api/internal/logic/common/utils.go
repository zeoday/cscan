package common

import (
	"cscan/pkg/utils"
)

// IsIPAddress 判断是否为IP地址
func IsIPAddress(s string) bool {
	return utils.IsIPAddress(s)
}

// GetRootDomain 获取根域名
func GetRootDomain(domain string) string {
	return utils.GetRootDomain(domain)
}

// IsValidDomain 检查是否是有效的域名
func IsValidDomain(domain string) bool {
	return utils.IsValidDomain(domain)
}

// UniqueStrings 去重字符串切片
func UniqueStrings(slice []string) []string {
	return utils.UniqueStrings(slice)
}
