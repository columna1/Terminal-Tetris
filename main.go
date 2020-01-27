package main

import (
	"strconv"
	"math"
	"math/rand"
	"fmt"
	"os"
	"time"
	"github.com/gdamore/tcell"
)

var(
	//consts
	colors = [7]int32{
		0x00FFFF,//Cyan I
		0x0000FF,//Blue J
		0xFF9600,//Orange L
		0xFFFF00,//Yellow O
		0x00FF00,//Green S
		0xFF00FF,//Purple T
		0xFF0000,//Red Z
	}
	colors256 = [7]tcell.Color{
		tcell.ColorDarkCyan,//Cyan I
		tcell.ColorBlue,//Blue J
		tcell.ColorOrange,//Orange L
		tcell.ColorYellow,//Yellow O
		tcell.ColorGreen,//Green S
		tcell.ColorPurple,//Purple T
		tcell.ColorRed,//Red Z
	}

	pieces = [][][]uint8{
		{	
			{0,1,0,0},//I piece
			{0,1,0,0},
			{0,1,0,0},
			{0,1,0,0},
		},
		{
			{2,2,0},//J piece
			{0,2,0},
			{0,2,0},
		},
		{
			{0,3,0},//L piece
			{0,3,0},
			{3,3,0},
		},
		{
			{4,4},//O piece
			{4,4},
		},
		{
			{0,5,0},//S piece
			{5,5,0},
			{5,0,0},
		},
		{
			{0,6,0},//T piece
			{6,6,0},
			{0,6,0},
		},
		{
			{7,0,0},//Z piece
			{7,7,0},
			{0,7,0},
		},
	}

	//game state
	gameField [10][40]uint8

	pieceQueue []uint8
	currentPiece [][]uint8
	currentPos [2]uint8//x,y
	currentRot,currentPieceType,holdPiece uint8

	pieceDropTimer,pieceLockTimer,pieceDropInterval time.Duration

	lastClearWasTetris,countLockTimer bool

	//scoring
	level int = 1
	points int

)

const(
	
)

//Piece manupulation//

func printPiece(piece [][]uint8){
	for y := 0; y < len(piece); y++{
		str := ""
		for x :=0;x<len(piece[y]);x++{
			s := strconv.Itoa(int(piece[x][y]))
			str += s
		}
		fmt.Println(str)
	}
}

func rotateCW(piece [][]uint8) [][]uint8 {
	//assumes that the 2d slice is square
	n := len(piece)

	x := int(math.Floor(float64(n)/2))
	y := n - 1
	for i := 0;i < x;i++ {
		for j := i;j<y-i;j++ {
			temp := piece[i][j]
			piece[i][j] = piece[j][(y-i)]
			piece[j][(y-i)] = piece[(y-i)][(y-j)]
			piece[(y-i)][(y-j)] = piece[(y-j)][i]
			piece[(y-j)][i] = temp
		}
	}
	return piece
}

func rotateCCW(piece [][]uint8) [][]uint8 {
	//assumes that the 2d slice is square
	n := len(piece)

	x := int(math.Floor(float64(n)/2))
	y := n - 1
	for i := 0;i < x;i++ {
		for j := i;j<y-i;j++ {
			temp := piece[i][j]
			piece[i][j] = piece[(y-j)][i]
			piece[(y-j)][i] = piece[(y-i)][(y-j)]
			piece[(y-i)][(y-j)] = piece[j][(y-i)]
			piece[j][(y-i)] = temp
		}
	}
	return piece
}



//gamefield/gameplay manipulation functions//

func genNewPieces(){
	//7bag algorithm
	t := [7]uint8{1,2,3,4,5,6,7}
	rand.Shuffle(len(t),func(i,j int){
		t[i],t[j] = t[j],t[i]
	})

	for i:= range t {
		pieceQueue = append(pieceQueue,t[i])
	}
}

func getNextPiece() uint8{
	if len(pieceQueue) < 7 {
		genNewPieces()
	}
	x := uint8(0)
	x,pieceQueue = pieceQueue[0],pieceQueue[1:]
	return x
}

func checkColision(piece [][]uint8,field [10][40]uint8,x,y uint8) bool{
	for xx := range piece {
		for yy := range piece[xx] {
			ox,oy := uint8(xx)+x,uint8(yy)+y
			if piece[xx][yy] != 0 {
				if ox > 9 || ox < 0 || oy > 39 || oy < 0 {
					return false
				}else if field[ox][oy] != 0 {
					return false
				}
			}
		}
	}
	return true
}

func spawnPiece(n uint8){
	currentPiece = make([][]uint8,len(pieces[n-1]))
	for i := range currentPiece {
		currentPiece[i] = append([]uint8(nil),pieces[n-1][i]...)
	}
	currentPos = [2]uint8{4-uint8(math.Floor(float64(len(pieces[n-1])/2))),19}
	currentRot = 0
	currentPieceType = n
}

func checkForCompletedLines(field [10][40]uint8) [10][40]uint8{
	for y := 39; y > 0 ;y--{
		com := true
		for x := 0; x < 10 ;x++{
			if field[x][y] == 0{
				com = false
				break
			}
		}
		if com {
			//line has been completed, move everything down by a line
			for l := y;l < 39;l++{
				if l < 39 {
					for x := 0; x < 10 ;x++{
						field[x][y] = field[x][y+1]
					}
				}
			}
			//score
		}
	}
	return field
}

func placePiece(piece [][]uint8,field [10][40]uint8,x,y uint8) [10][40]uint8{
	for xx := range piece {
		for yy := range piece[xx] {
			ox,oy := uint8(xx)+x,uint8(yy)+y
			if piece[xx][yy] != 0 {
				//fmt.Println(ox,oy,xx,yy,piece[xx][yy])
				field[ox][oy] = piece[xx][yy]
			}
		}
	}
	//check if any lines have been completed
	field = checkForCompletedLines(field)
	//spawn new piece
	spawnPiece(getNextPiece())
	//os.Exit(0)
	return field
}

//Gameplay logic//

func update(dt time.Duration){
	pieceDropInterval = time.Duration(math.Pow((0.8-((float64(level)-1)*0.007)),(float64(level)-1)))
	pieceDropInterval *= time.Second
	pieceDropInterval /= 4
	pieceDropTimer += dt
	if countLockTimer{
		pieceLockTimer += dt
		if pieceLockTimer >= time.Second/2{
			//place piece into gamefield
			gameField = placePiece(currentPiece,gameField,currentPos[0],currentPos[1])
			countLockTimer = false
			pieceLockTimer = 0
		}
	}
	if pieceDropTimer >= pieceDropInterval {
		//drop the piece down by one square
		if checkColision(currentPiece,gameField,currentPos[0],currentPos[1]+1) {
			//success, we can move our piece down
			currentPos[1]++
		}else{
			//we have hit something
			countLockTimer = true
		}
		pieceDropTimer -= pieceDropInterval
	}
}


//drawing methods//

func drawGameField(s tcell.Screen,field [10][40]uint8,xoff,yoff int){
	st := tcell.StyleDefault
	for x := 0; x < 10 ; x++{
		for y := 0; y < 20 ; y++{
			if field[x][y+20] !=0{
				if s.Colors() > 256 {
					st = st.Background(tcell.NewHexColor(colors[field[x][y+20]-1]))
				}else if s.Colors() > 1 {
					st = st.Background(tcell.Color(colors256[field[x][y+20]-1]))
				}
				s.SetCell((xoff+x)*2,yoff+y,st,' ')
				s.SetCell((xoff+x)*2+1,yoff+y,st,' ')
			}
		}
	}
	//s.Show()
}

func drawPiece(s tcell.Screen,piece [][]uint8,xoff,yoff int){
	st := tcell.StyleDefault

	for x := range piece {
		for y := range piece[x] {
			if piece[x][y] != 0{
				if s.Colors() > 256 {
					st = st.Background(tcell.NewHexColor(colors[piece[x][y]-1]))
				}else if s.Colors() > 1 {
					st = st.Background(tcell.Color(colors256[piece[x][y]-1]))
				}
				s.SetCell((xoff+x)*2,yoff+y,st,' ')
				s.SetCell((xoff+x)*2+1,yoff+y,st,' ')
			}
		}
	}

	st = st.Background(tcell.Color(colors[0]))
}

func drawRect(s tcell.Screen,x,y,w,h int) {
	for i := 0; i < w; i++{
		for j := 0; j < h; j++{
			s.SetCell(x+i,y+j,tcell.StyleDefault,rune(65))
		}
	}
}


func main() {
	spawnPiece(getNextPiece())

	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s,e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e = s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	var dx,dy = 2,2

	drawRect(s,dx,dy,10,5)
	quit := make(chan struct{})
	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(quit)
					return
				case tcell.KeyRune:
					switch ev.Rune() {
						case 'q':
							close(quit)
							return
					}
				case tcell.KeyLeft:
					dx--
				case tcell.KeyRight:
					dx++
				case tcell.KeyUp:
					rotateCW(currentPiece)
				case tcell.KeyCtrlL:
					s.Sync()
				}
			case *tcell.EventResize:
				s.Sync()
			}
		}
	}()
	lt := time.Now()
loop:
	for {
		select {
		case <-quit:
			break loop
		case <-time.After(time.Millisecond * 32):
			s.Clear()
			//drawRect(s,dx,dy,10,5)
			dt := time.Now().Sub(lt)
			lt = time.Now()
			update(dt)
			drawPiece(s,currentPiece,int(currentPos[0]),int(currentPos[1]-20))
			drawGameField(s,gameField,0,0)
			s.Show()
		}
		
	}

	s.Fini()

	//fmt.Println(gameField)
	for i := 0;i<10;i++{
		s := ""
		for j := 0;j<40;j++{
			s += strconv.Itoa(int(gameField[i][j]))
		}
		fmt.Println(s)
	}
}
