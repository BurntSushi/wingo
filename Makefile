BD=bindata
D=data
BINDATA=$(BD)/wingo.png.syso \
				$(BD)/close.png.syso $(BD)/maximize.png.syso $(BD)/minimize.png.syso \
				$(BD)/DejaVuSans.ttf.syso \
				$(BD)/wingo.wav.syso
# BINDATA=$(BD)/wingo.png.go \ 
				# $(BD)/close.png.go $(BD)/maximize.png.go $(BD)/minimize.png.go \ 
				# $(BD)/DejaVuSans.ttf.go \ 
				# $(BD)/wingo.wav.go 

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

$(BD)/%.syso: $(BD)/%.S
	as -o $(BD)/$*.syso $(BD)/$*.S

$(BD)/%.png.S: $(D)/%.png
	scripts/mkSData `python2 -c 'print "$*".title()'`Png $(D)/$*.png \
		> $(BD)/$*.png.S

$(BD)/%.ttf.S: $(D)/%.ttf
	scripts/mkSData `python2 -c 'print "$*".title()'`Ttf $(D)/$*.ttf \
	 	> $(BD)/$*.ttf.S

$(BD)/%.wav.S: $(D)/%.wav
	scripts/mkSData `python2 -c 'print "$*".title()'`Wav $(D)/$*.wav \
	 	> $(BD)/$*.wav.S

loc:
	find ./ -name '*.go' \
		-and -not -wholename './tests*' \
		-and -not -wholename './bindata*' -print \
		| sort | xargs wc -l

tags:
	find ./ \( -name '*.go' \
					   -and -not -wholename './tests/*' \
						 -and -not -wholename './bindata/*' \) \
			 -print0 \
	| xargs -0 gotags > TAGS

push:
	git push origin master
	git push github master

