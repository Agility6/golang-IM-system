package main

func main() {
	// 项目的入口
	server := NewServer("127.0.0.1", 8899)
	server.Start()
}
