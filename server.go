package main

import (
	"fmt"
	"io"
	"net"
	"slices"
	"strconv"
	"strings"
)

func handleClient(pl *Player) {
	defer pl.conn.Close() // 函数退出时执行

	fmt.Println("客户端连接：", pl.conn.RemoteAddr())

	buf := make([]byte, 4096)

	for {
		n, err := pl.conn.Read(buf)

		if err != nil {
			player_lost(pl)
			if err == io.EOF {
				//
				fmt.Println("客户端断开：", pl.conn.RemoteAddr())
			} else {
				fmt.Println("读取错误：", err)
			}
			return
		}

		data := buf[:n]
		cmd := fmt.Sprintf("%s", data)
		handle_cmd(cmd, pl)
	}
}

func main2() {

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("TCP服务器启动，监听 :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept错误：", err)
			continue
		}
		p := create_player(conn)
		// 每个客户端使用一个 goroutine
		go handleClient(p)
	}
}

type Player struct {
	id       int
	name     string
	ref_room *Room
	conn     net.Conn
	flg      byte
}

var pls []*Player
var rooms []*Room
var max_players int = 100
var max_rooms int = 100
var room_ids = 1
var player_ids = 1

func join_room(id int, pl *Player) {
	r := find_room(id)
	join_room_(r, pl)
}

func player_lost(pl *Player) {
	r := pl.ref_room
	if r != nil {
		r.players = slices.DeleteFunc(r.players,
			func(v *Player) bool {
				return v == pl
			})
		pl.ref_room = nil
		if len(r.players) == 0 {
			destroy_room(r)
		}
		for _, v := range r.players {
			v.conn.Write([]byte("lost player leave server"))
		}

	}
	pls = slices.DeleteFunc(pls, func(p1 *Player) bool { return pl == p1 })
}

func join_room_(r *Room, pl *Player) {
	if pl.ref_room != nil {
		if pl.ref_room == r {
			s := "msg already in room\n"
			pl.conn.Write([]byte(s))

		} else {
			s := "msg exit current room first\n"
			pl.conn.Write([]byte(s))

		}
		return
	}

	if r != nil {
		if r.state == 1 {
			//对局进行中
			s := "msg game is playing\n"
			pl.conn.Write([]byte(s))
			return
		}
		if len(r.players) == 2 {
			//房间已满
			s := "msg room is full\n"
			pl.conn.Write([]byte(s))
			return
		}
		r.players = append(r.players, pl)
		pl.ref_room = r
		//s := "msg join room success\n"
		//pl.conn.Write([]byte(s))
		if len(r.players) == 2 {
			start_game(r)
		}
	} else {
		s := "msg room dose not exist\n"
		pl.conn.Write([]byte(s))
	}

}

func exit_room(pl *Player) {
	r := pl.ref_room
	if r != nil {
		r.players = slices.DeleteFunc(r.players,
			func(v *Player) bool {
				return v == pl
			})
		pl.ref_room = nil
		if len(r.players) == 0 {
			destroy_room(r)
		}
		s := "msg exit room{id=%d}\n"
		s = fmt.Sprintf(s, r.id)
		pl.conn.Write([]byte(s))
	} else {
		s := "msg you have not join any room\n"
		pl.conn.Write([]byte(s))
	}

}

func find_room(id int) *Room {
	for i := 0; i < len(rooms); i++ {
		if rooms[i].id == id {
			return rooms[i]
		}
	}
	return nil
}

func list_rooms_str() string {
	s := ""
	for i := 0; i < len(rooms); i++ {
		s += fmt.Sprintf("room{id:%d,state:%d,pls:%d}\n",
			rooms[i].id,
			rooms[i].state,
			len(rooms[i].players))
	}
	return s
}

func list_rooms(pl *Player) {
	s := list_rooms_str()
	s = "msg " + s
	pl.conn.Write([]byte(s))
}

func create_room(pl *Player) *Room {
	if pl.ref_room != nil {
		s := "msg already in room\n"
		pl.conn.Write([]byte(s))
		return nil
	}
	r := new(Room)
	r.id = room_ids
	room_ids++
	rooms = append(rooms, r)
	join_room_(r, pl)
	s := "msg room_id=%d\n"
	s = fmt.Sprintf(s, r.id)
	pl.conn.Write([]byte(s))
	return r
}

func create_player(conn net.Conn) *Player {
	p := new(Player)
	p.conn = conn
	p.id = player_ids
	player_ids++
	return p
}

func destroy_player(p *Player) {
	if p.ref_room != nil {
		exit_room(p)
	}
	//p.conn.Close()
	pls = slices.DeleteFunc(pls, func(p1 *Player) bool { return p == p1 })

}

func destroy_room(r *Room) {
	rooms = slices.DeleteFunc(rooms, func(r1 *Room) bool { return r1 == r })

}

func check_win2(g *Room, flg byte, row int, col int) bool {
	return check_col2(g, flg, col) ||
		check_row2(g, flg, row) ||
		check_main2(g, flg) ||
		check_sub2(g, flg)
}

func check_row2(g *Room, flg byte, row int) bool {
	for i := 0; i < 3; i++ {
		if g.cells[row][i] != flg {
			return false
		}
	}
	return true
}

func check_col2(g *Room, flg byte, col int) bool {
	for i := 0; i < 3; i++ {
		if g.cells[i][col] != flg {
			return false
		}

	}
	return true
}

func check_main2(g *Room, flg byte) bool {
	for i := 0; i < 3; i++ {
		if g.cells[i][i] != flg {
			return false
		}
	}
	return true
}

func check_sub2(g *Room, flg byte) bool {
	for i := 0; i < 3; i++ {
		if g.cells[2-i][i] != flg {
			return false
		}
	}
	return true
}

func start_game(room *Room) {
	room.players[0].flg = 'x'
	room.players[1].flg = 'o'
	room.step = 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			room.cells[i][j] = '.'
		}
	}
	for i := 0; i < len(room.players); i++ {
		s := "start %c %t"
		s = fmt.Sprintf(s, room.players[i].flg, room.players[i].flg == 'x')
		room.players[i].conn.Write([]byte(s))
	}
}

func handle_cmd(cmd string, p *Player) {

	i := strings.Index(cmd, " ")
	var head = ""
	var body = ""
	if i == -1 {
		head = cmd
	} else {
		head = cmd[:i]   // "key"
		body = cmd[i+1:] // "value:extra"

	}
	switch head {
	case "join":
		id, err := strconv.Atoi(body)
		if err == nil {
			join_room(id, p)
		}
	case "exit":
		exit_room(p)
	case "create":
		create_room(p)
	case "list":
		list_rooms(p)
	case "restart":
		start_game(p.ref_room)
		//	case "msg":
	case "set":
		p.ref_room.step++
		ss := strings.Split(body, " ")
		x, _ := strconv.Atoi(ss[0])
		y, _ := strconv.Atoi(ss[1])
		p.ref_room.cells[x][y] = p.flg
		for _, v := range p.ref_room.players {
			if v != p {
				v.conn.Write([]byte(cmd))
			}
		}
		if check_win2(p.ref_room, p.flg, x, y) {
			p.conn.Write([]byte("over you win\n"))
			for _, v := range p.ref_room.players {
				if v != p {
					v.conn.Write([]byte("over you lose\n"))
				}
			}
		}
		if p.ref_room.step == 9 {
			for _, v := range p.ref_room.players {

				v.conn.Write([]byte("over end in a draw\n"))

			}
		}
	default:
		p.conn.Write([]byte("msg unknow command\n"))

	}

}

type Room struct {
	id      int
	players []*Player
	state   int
	cells   [3][3]byte
	step    int
}

func main() {

	main2()
	//game_start()
	//foo()
}

// a：左移 w :上移 s:下移 d:右移 e:落子
// list 查看所有房间
// create 创建并加入新房间
// join id id为房间号，加入某个房间
// exit 退出房间
// r 重新开始{对局结束后才能使用}
