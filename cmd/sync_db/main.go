package main

import (
	"log"
	"os"
	"os/exec"
)

const localRedisContainer = "vvgo-redis-1"
const redisHost = "root@redis-1.infra.vvgo.org"

func main() {
	for _, cmd := range []*exec.Cmd{
		exec.Command("ssh", redisHost, "/root/bin/backup-redis"),
		exec.Command("scp", redisHost+":/var/lib/redis/dump.rdb", "."),
		exec.Command("docker", "stop", localRedisContainer),
		exec.Command("docker", "cp", "dump.rdb", "vvgo-redis-1:/data/dump.rdb"),
		exec.Command("docker", "start", localRedisContainer),
		exec.Command("docker", "exec", localRedisContainer, "redis-cli", "SET", "__db_name", "localhost"),
	} {
		log.Println("Executing:", cmd.String())
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatal("command failed: ", err)
		}
	}
	log.Println("Executing cleanup tasks.")
	if err := os.Remove("dump.rdb"); err != nil {
		log.Println(err)
	}
	log.Println("Sync completed successfully!")
	os.Exit(0)
}
