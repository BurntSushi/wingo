package misc

var (
	DejavusansTtf []byte
	WingoWav      []byte
	WingoPng      []byte
	ClosePng      []byte
	MinimizePng   []byte
	MaximizePng   []byte
)

func ReadData() {
	DejavusansTtf = DataFile("DejaVuSans.ttf")
	WingoWav = DataFile("wingo.wav")
	WingoPng = DataFile("wingo.png")
	ClosePng = DataFile("close.png")
	MinimizePng = DataFile("minimize.png")
	MaximizePng = DataFile("maximize.png")
}
