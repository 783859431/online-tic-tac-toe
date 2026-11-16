package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type State int

const (
	WaitStart State = iota
	GameStart
	GameOver
)

type Game struct {
	conn   net.Conn
	cells  [3][3]byte
	row    int
	col    int
	flag   byte
	state  State
	myturn bool
}

func game_init(game *Game) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			game.cells[i][j] = '.'
		}

	}
	game.row = 0
	game.col = 0
	game.myturn = false
	game.state = GameStart
}

func disp(game *Game) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if i == game.row && j == game.col {
				fmt.Printf("[%c]", game.cells[i][j])
			} else {
				fmt.Printf(" %c ", game.cells[i][j])
			}

		}
		fmt.Print("\n")
	}
}

func move(game *Game, c byte) int {
	switch c {
	case 'w':
		game.row--
		if game.row < 0 {
			game.row = 0
		}
		return 0
	case 's':
		game.row++
		if game.row >= 3 {
			game.row = 2
		}
		return 0
	case 'a':
		game.col--
		if game.col < 0 {
			game.col = 0
		}
		return 0
	case 'd':
		game.col++
		if game.col >= 3 {
			game.col = 2
		}
		return 0
	case 'e':

		if game.cells[game.row][game.col] != '.' {
			return 0
		}
		game.cells[game.row][game.col] = game.flag
		s := "set %d %d"
		s = fmt.Sprintf(s, game.row, game.col)
		game.conn.Write([]byte(s))
		game.myturn = false
		return 0
	}
	return 0
}

func main() {

	conn, err := net.Dial("tcp", "165.154.227.243:8080")
	game = new(Game)
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}
	defer conn.Close() // 确保关闭
	// 2. 发送数据
	fmt.Println("连接成功")
	game.conn = conn
	go handle_cmsg(conn)
	handle_input()
	//game_start()
	//foo()
}

func handle_cmsg(conn net.Conn) {

	buf := make([]byte, 1024)
	for true {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("接收失败:", err)
			return
		}
		s := string(buf[:n])
		handle_cmd_str(s)

	}

}

var game *Game

func handle_cmd_str(cmd string) {
	i := strings.Index(cmd, " ")
	fmt.Print(cmd)
	var head = ""
	var body = ""
	if i == -1 {
		head = cmd
		fmt.Println("没找到分隔符")
	} else {
		head = cmd[:i]   // "key"
		body = cmd[i+1:] // "value:extra"

	}
	switch head {
	case "msg":
		fmt.Print(body)
	case "start":
		//fmt.Print(body)
		ss := strings.Split(body, " ")
		game.flag = ss[0][0]
		game_init(game)
		game.myturn, _ = strconv.ParseBool(ss[1])
		if game.myturn {
			fmt.Println("you first")
		}
		disp(game)
	case "set":
		ss := strings.Split(body, " ")
		x, _ := strconv.Atoi(ss[0])
		y, _ := strconv.Atoi(ss[1])
		if game.flag == 'x' {
			game.cells[x][y] = 'o'
		} else {
			game.cells[x][y] = 'x'
		}
		game.myturn = true
		disp(game)
	case "over":
		fmt.Println(body)
		game.state = GameOver
	case "lost":
		fmt.Println(body)
		game.state = WaitStart

	default:
		//	fmt.Println("unknow server" + cmd)
		//	fmt.Print("head:" + head)
		//	fmt.Print(body)
	}
}
func handle_input() {
	scanner := bufio.NewScanner(os.Stdin)
	for true {
		scanner.Scan()
		line := scanner.Text()
		switch game.state {
		case WaitStart:
			game.conn.Write([]byte(line))
		case GameStart:
			if !game.myturn {
				continue
			}
			flags := strings.Fields(line)
			for idx, f := range flags {
				if f[0] == 'e' && idx != (len(flags)-1) {
					flags = flags[:idx+1]
					break
				}
			}
			for _, f := range flags {
				move(game, f[0])
			}
			disp(game)
		case GameOver:
			if line == "r" {
				game.conn.Write([]byte("restart"))
			} else {
				game.conn.Write([]byte("exit"))
				game.state = WaitStart
			}
		default:
			game.conn.Write([]byte(line))
		}
	}

}
