D=data
BINDATA=$(BD)/wingo.png.syso $(BD)/wingo.png.c \
				$(BD)/close.png.syso $(BD)/close.png.c \
				$(BD)/maximize.png.syso $(BD)/maximize.png.c \
				$(BD)/minimize.png.syso $(BD)/minimize.png.c \
				$(BD)/DejaVuSans.ttf.syso $(BD)/DejaVuSans.ttf.c \
				$(BD)/wingo.wav.syso $(BD)/wingo.wav.c

install: supported
	go install -p 6 . ./cursors ./focus \
		./frame ./heads ./hook ./layout ./logger ./misc ./prompt ./render \
		./stack ./text ./wingo-cmd ./wini ./wm ./workspace ./xclient

gofmt:
	gofmt -w *.go cursors/*.go focus/*.go frame/*.go \
		heads/*.go hook/*.go layout/*.go logger/*.go misc/*.go prompt/*.go \
		render/*.go stack/*.go text/*.go wingo-cmd/*.go wini/*.go wm/*.go \
		workspace/*.go xclient/*.go
	colcheck *.go */*.go

cmd:
	go install github.com/BurntSushi/wingo/wingo-cmd

supported:
	scripts/generate-supported | gofmt > ewmh_supported.go

loc:
	find ./ -name '*.go' \
		-and -not -wholename './tests*' -print \
		| sort | xargs wc -l

tags:
	find ./ \( -name '*.go' \
					   -and -not -wholename './tests/*' \) \
			 -print0 \
	| xargs -0 gotags > TAGS

push:
	git push origin master
	git push github master

