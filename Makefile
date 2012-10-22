BD=bindata
D=data
BINDATA=$(BD)/wingo.png.go \
				$(BD)/close.png.go $(BD)/maximize.png.go $(BD)/minimize.png.go \
				$(BD)/DejaVuSans.ttf.go \
				$(BD)/wingo.wav.go

install: bindata supported
	go install -p 6 . ./bindata ./cursors ./focus \
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

bindata: $(BINDATA)

$(BD)/%.png.go: $(D)/%.png
	go-bindata -f `python2 -c 'print "$*".title()'`Png \
		-i $(D)/$*.png -o $(BD)/$*.png.go -p bindata
	gofmt -w $(BD)/$*.png.go

$(BD)/%.ttf.go: $(D)/%.ttf
	go-bindata -f `python2 -c 'print "$*".title()'`Ttf \
		-i $(D)/$*.ttf -o $(BD)/$*.ttf.go -p bindata
	gofmt -w $(BD)/$*.ttf.go

$(BD)/%.wav.go: $(D)/%.wav
	go-bindata -f `python2 -c 'print "$*".title()'`Wav \
		-i $(D)/$*.wav -o $(BD)/$*.wav.go -p bindata
	gofmt -w $(BD)/$*.wav.go

loc:
	find ./ -name '*.go' -and -not -wholename './tests*' -and -not -wholename './bindata*' -print | sort | xargs wc -l

tags:
	find ./ \( -name '*.go' \
					   -and -not -wholename './tests/*' \
						 -and -not -wholename './bindata/*' \) \
			 -print0 \
	| xargs -0 gotags > TAGS

push:
	git push origin master
	git push github master

