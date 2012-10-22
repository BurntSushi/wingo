package bindata

var (
	DejavusansTtf []byte
	WingoWav      []byte
	WingoPng      []byte
	ClosePng      []byte
	MinimizePng   []byte
	MaximizePng   []byte
)

func getDejavusansTtf() []byte // defined in DejaVuSans.ttf.c
func getWingoWav() []byte      // defined in wingo.wav.c
func getWingoPng() []byte      // defined in wingo.png.c
func getClosePng() []byte      // defined in close.png.c
func getMinimizePng() []byte   // defined in minimize.png.c
func getMaximizePng() []byte   // defined in maximize.png.c

func init() {
	DejavusansTtf = getDejavusansTtf()
	WingoWav = getWingoWav()
	WingoPng = getWingoPng()
	ClosePng = getClosePng()
	MinimizePng = getMinimizePng()
	MaximizePng = getMaximizePng()
}
