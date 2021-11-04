package main

import (
	"log"
	"os"
	"os/exec"
)

const localRedisContainer = "vvgo_redis_1"
const remoteRedisContainer = "redis-prod"

func main() {
	for _, cmd := range []*exec.Cmd{
		exec.Command("docker", "stop", localRedisContainer),
		exec.Command("ssh", "vvgo.org", "docker", "exec", remoteRedisContainer, "redis-cli", "save"),
		exec.Command("ssh", "vvgo.org", "docker", "cp", remoteRedisContainer+":/data/dump.rdb", "dump.rdb"),
		exec.Command("scp", "vvgo.org:dump.rdb", "."),
		exec.Command("ssh", "vvgo.org", "rm", "dump.rdb"),
		exec.Command("docker", "cp", "dump.rdb", "vvgo_redis_1:/data/dump.rdb"),
		exec.Command("docker", "start", localRedisContainer),
		exec.Command("docker", "exec", localRedisContainer, "redis-cli", "SET", "__db_name", "localhost"),
	} {
		log.Println("executing:", cmd.String())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatal("command failed: ", err)
		}
	}

}
