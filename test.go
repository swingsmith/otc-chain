package main
import (
	"testing"
)

// 假设networkFunc是一个网络调用
func networkFunc(a, b int) int {
	return a + b
}

// 本地单测一般不会进行网络调用，所以要mock住networkFunc
func Test_MockNetworkFunc(t *testing.T) {

}
